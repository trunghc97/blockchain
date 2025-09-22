package com.example.blockchain.service;

import com.example.blockchain.model.Contract;
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
                // Save contract locally for quick access
                return mongoTemplate.save(contract);
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
        return mongoTemplate.findAll(Contract.class, "contracts");
    }

    public List<Contract> getContractsByUser(String username) {
        // Find contracts where user is either buyer (anchor) or a supplier
        Criteria buyerCriteria = Criteria.where("buyer").is(username);
        Criteria supplierCriteria = Criteria.where("suppliers.supplierid").is(username);

        Query query = new Query(new Criteria().orOperator(buyerCriteria, supplierCriteria));
        return mongoTemplate.find(query, Contract.class);
    }

    public Contract getContract(String contractId) {
        if (!StringUtils.hasText(contractId)) {
            throw new IllegalArgumentException("Contract ID cannot be empty");
        }

        return mongoTemplate.findOne(
            Query.query(Criteria.where("contractId").is(contractId)),
            Contract.class
        );
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