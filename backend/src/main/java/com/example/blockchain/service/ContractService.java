package com.example.blockchain.service;

import com.example.blockchain.model.Contract;
import com.example.blockchain.model.SupplierAmount;
import com.example.blockchain.model.User;
import org.springframework.data.mongodb.core.MongoTemplate;
import org.springframework.data.mongodb.core.query.Criteria;
import org.springframework.data.mongodb.core.query.Query;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;
import org.springframework.util.StringUtils;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.Date;
import java.util.List;
import java.util.Map;
import java.util.HashMap;
import java.util.UUID;
import java.util.Objects;

@Service
public class ContractService {
    private final MongoTemplate mongoTemplate;
    private final BlockchainService blockchainService;

    public ContractService(MongoTemplate mongoTemplate, BlockchainService blockchainService) {
        this.mongoTemplate = mongoTemplate;
        this.blockchainService = blockchainService;
    }

    public Contract createContract(Contract contract, MultipartFile file) {
        if (contract == null) {
            throw new IllegalArgumentException("Contract cannot be null");
        }

        // Calculate total amount from suppliers
        if (contract.getSuppliers() != null && !contract.getSuppliers().isEmpty()) {
            double total = contract.getSuppliers().stream()
                .filter(Objects::nonNull)
                .mapToDouble(supplier -> supplier.getAmount() != null ? supplier.getAmount() : 0.0)
                .sum();
            contract.setTotalAmount(total);
        }

        // Set initial values
        if (contract.getContractId() == null || contract.getContractId().isEmpty()) {
            contract.setContractId(UUID.randomUUID().toString());
        }
        contract.setStatus("PENDING");
        contract.setWordState("CREATED"); // Set initial word state

        // Set timestamps as ISO string format
        String now = LocalDateTime.now().format(DateTimeFormatter.ISO_LOCAL_DATE_TIME) + "Z";
        contract.setCreatedAt(now);
        contract.setUpdatedAt(now);

        // Set initial supplier status
        if (contract.getSuppliers() != null) {
            contract.getSuppliers().forEach(supplier -> {
                if (supplier != null && supplier.getStatus() == null) {
                    supplier.setStatus("PENDING");
                }
            });
        }

        // Handle file upload if needed
        if (file != null && !file.isEmpty()) {
            String fileName = StringUtils.cleanPath(Objects.requireNonNull(file.getOriginalFilename()));
            // TODO: Implement file storage
            contract.setFileUrl("/uploads/" + fileName);
        }

        try {
            // Prepare contract data for blockchain service
            Map<String, Object> contractData = new HashMap<>();
            contractData.put("id", contract.getContractId());
            contractData.put("description", contract.getDescription());
            contractData.put("anchorId", contract.getBuyer()); // buyer is anchor
            contractData.put("supplierId", contract.getSuppliers() != null && !contract.getSuppliers().isEmpty()
                ? contract.getSuppliers().get(0).getSupplierId() : ""); // primary supplier
            contractData.put("bankId", "BANK001"); // hardcoded for now
            contractData.put("amount", contract.getTotalAmount());
            contractData.put("suppliers", contract.getSuppliers());
            contractData.put("approvers", contract.getSuppliers() != null
                ? contract.getSuppliers().stream().map(s -> s.getSupplierId()).toList()
                : java.util.Collections.emptyList());

            System.out.println("DEBUG: Calling blockchain service with data: " + contractData);
            // Call blockchain service to create contract
            Map<String, Object> blockchainResponse = blockchainService.createContract(contractData);
            System.out.println("DEBUG: Blockchain response: " + blockchainResponse);

            if (blockchainResponse != null && "success".equals(blockchainResponse.get("status"))) {
                // Contract is now managed by blockchain service only
                // Get the created contract from blockchain service to return complete data
                String createdContractId = (String) blockchainResponse.get("contractId");
                if (createdContractId != null) {
                    Map<String, Object> blockchainContract = blockchainService.getContract(createdContractId);
                    if (blockchainContract != null) {
                        return convertBlockchainContractToLocal(blockchainContract);
                    }
                }
                // Fallback to original contract if blockchain query fails
                return contract;
            } else {
                throw new RuntimeException("Failed to create contract on blockchain");
            }
        } catch (Exception e) {
            System.err.println("Error calling blockchain service: " + e.getMessage());
            e.printStackTrace();
            throw new RuntimeException("Blockchain service unavailable", e);
        }
    }


    public List<Contract> getContracts() {
        // Query contracts from blockchain service since that's now the authoritative source
        List<Map<String, Object>> blockchainContracts = blockchainService.listContracts();
        return blockchainContracts.stream()
            .map(this::convertBlockchainContractToLocal)
            .collect(java.util.stream.Collectors.toList());
    }

    private Contract convertBlockchainContractToLocal(Map<String, Object> blockchainContract) {
        Contract contract = new Contract();
        contract.setId((String) blockchainContract.get("_id"));
        contract.setContractId((String) blockchainContract.get("_id"));
        contract.setDescription((String) blockchainContract.get("description"));
        contract.setBuyer((String) blockchainContract.get("anchorId"));

        // Set status based on approved field
        Boolean approved = (Boolean) blockchainContract.get("approved");
        if (approved != null && approved) {
            contract.setStatus("EXECUTED");
        } else {
            contract.setStatus("PENDING");
        }

        // Safely handle amount field - could be null from blockchain
        Object amountObj = blockchainContract.get("amount");
        double amount = 0.0;
        if (amountObj instanceof Number) {
            amount = ((Number) amountObj).doubleValue();
        } else if (amountObj != null) {
            try {
                amount = Double.parseDouble(amountObj.toString());
            } catch (NumberFormatException e) {
                // Log warning but continue with 0.0
                System.err.println("Warning: Invalid amount format in contract " + contract.getContractId() + ": " + amountObj);
            }
        }
        contract.setTotalAmount(amount);
        contract.setFileUrl((String) blockchainContract.get("fileUrl"));
        contract.setCreatedAt((String) blockchainContract.get("createdAt"));
        contract.setUpdatedAt((String) blockchainContract.get("createdAt")); // Use createdAt as updatedAt
        contract.setWordState("CREATED");

        // Convert suppliers from blockchain format to local format
        @SuppressWarnings("unchecked")
        List<Map<String, Object>> blockchainSuppliers = (List<Map<String, Object>>) blockchainContract.get("suppliers");
        if (blockchainSuppliers != null) {
            List<SupplierAmount> suppliers = blockchainSuppliers.stream()
                .map(this::convertBlockchainSupplierToLocal)
                .collect(java.util.stream.Collectors.toList());
            contract.setSuppliers(suppliers);
        }

        return contract;
    }

    private SupplierAmount convertBlockchainSupplierToLocal(Map<String, Object> blockchainSupplier) {
        SupplierAmount supplier = new SupplierAmount();
        // Try different possible keys for supplierId
        String supplierId = (String) blockchainSupplier.get("supplierId");
        if (supplierId == null) {
            supplierId = (String) blockchainSupplier.get("supplierid");
        }
        supplier.setSupplierId(supplierId);
        supplier.setName((String) blockchainSupplier.get("name"));
        // Safely handle supplier amount field - could be null from blockchain
        Object supplierAmountObj = blockchainSupplier.get("amount");
        double supplierAmount = 0.0;
        if (supplierAmountObj instanceof Number) {
            supplierAmount = ((Number) supplierAmountObj).doubleValue();
        } else if (supplierAmountObj != null) {
            try {
                supplierAmount = Double.parseDouble(supplierAmountObj.toString());
            } catch (NumberFormatException e) {
                // Log warning but continue with 0.0
                System.err.println("Warning: Invalid supplier amount format for supplier " + supplier.getSupplierId() + ": " + supplierAmountObj);
            }
        }
        supplier.setAmount(supplierAmount);
        supplier.setStatus((String) blockchainSupplier.get("status"));
        return supplier;
    }

    public List<Contract> getContractsByUser(String username) {
        // Get all contracts from blockchain service and filter by user
        List<Map<String, Object>> allBlockchainContracts = blockchainService.listContracts();
        return allBlockchainContracts.stream()
            .map(this::convertBlockchainContractToLocal)
            .filter(contract -> isUserInContract(contract, username))
            .collect(java.util.stream.Collectors.toList());
    }

    private boolean isUserInContract(Contract contract, String username) {
        // Check if user is the buyer (anchor)
        if (username.equals(contract.getBuyer())) {
            return true;
        }

        // Check if user is a supplier
        if (contract.getSuppliers() != null) {
            return contract.getSuppliers().stream()
                .anyMatch(supplier -> username.equals(supplier.getSupplierId()));
        }

        return false;
    }

    public Contract getContract(String contractId) {
        if (!StringUtils.hasText(contractId)) {
            throw new IllegalArgumentException("Contract ID cannot be empty");
        }

        // Query contract from blockchain service
        Map<String, Object> blockchainContract = blockchainService.getContract(contractId);
        if (blockchainContract == null) {
            return null;
        }

        return convertBlockchainContractToLocal(blockchainContract);
    }


    public Contract approveContract(String contractId, String supplierId) {
        if (!StringUtils.hasText(contractId)) {
            throw new IllegalArgumentException("Contract ID cannot be empty");
        }
        if (!StringUtils.hasText(supplierId)) {
            throw new IllegalArgumentException("Supplier ID cannot be empty");
        }

        Contract contract = getContract(contractId);
        if (contract == null) {
            throw new RuntimeException("Contract not found: " + contractId);
        }

        try {
            // Call blockchain service to approve contract
            Map<String, Object> blockchainResponse = blockchainService.approveContract(contractId, supplierId);

            if (blockchainResponse != null && "success".equals(blockchainResponse.get("status"))) {
                // Update local contract status for quick access
                if (contract.getSuppliers() != null) {
                    for (var supplier : contract.getSuppliers()) {
                        if (supplier != null && supplierId.equals(supplier.getSupplierId())) {
                            supplier.setStatus("APPROVED");
                            break;
                        }
                    }

                    // Check if all suppliers approved
                    boolean allApproved = contract.getSuppliers().stream()
                        .filter(Objects::nonNull)
                        .allMatch(s -> "APPROVED".equals(s.getStatus()));

                    if (allApproved) {
                        contract.setStatus("APPROVED");
                    }
                }

                String now = LocalDateTime.now().format(DateTimeFormatter.ISO_LOCAL_DATE_TIME) + "Z";
                contract.setUpdatedAt(now);
                return mongoTemplate.save(contract);
            } else {
                throw new RuntimeException("Failed to approve contract on blockchain");
            }
        } catch (Exception e) {
            System.err.println("Error calling blockchain service: " + e.getMessage());
            throw new RuntimeException("Blockchain service unavailable", e);
        }
    }
}
