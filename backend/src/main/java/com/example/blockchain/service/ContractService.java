package com.example.blockchain.service;

import com.example.blockchain.model.Contract;
import com.example.blockchain.model.User;
import org.springframework.data.mongodb.core.MongoTemplate;
import org.springframework.data.mongodb.core.query.Criteria;
import org.springframework.data.mongodb.core.query.Query;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;
import org.springframework.util.StringUtils;

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
        contract.setCreatedAt(new Date());
        contract.setUpdatedAt(new Date());

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
            contractData.put("contractId", contract.getContractId());
            contractData.put("description", contract.getDescription());
            contractData.put("buyer", contract.getBuyer());
            contractData.put("suppliers", contract.getSuppliers());
            contractData.put("totalAmount", contract.getTotalAmount());
            contractData.put("fileUrl", contract.getFileUrl());

            // Call blockchain service to create contract
            Map<String, Object> blockchainResponse = blockchainService.createContract(contractData);

            if (blockchainResponse != null && "success".equals(blockchainResponse.get("status"))) {
                // Save contract locally for quick access
                return mongoTemplate.save(contract);
            } else {
                throw new RuntimeException("Failed to create contract on blockchain");
            }
        } catch (Exception e) {
            System.err.println("Error calling blockchain service: " + e.getMessage());
            throw new RuntimeException("Blockchain service unavailable", e);
        }
    }


    public List<Contract> getContracts() {
        return mongoTemplate.findAll(Contract.class);
    }

    public List<Contract> getContractsByUser(String userId) {
        if (!StringUtils.hasText(userId)) {
            throw new IllegalArgumentException("User ID cannot be empty");
        }

        // Find contracts where user is either buyer (anchor) or a supplier
        Criteria buyerCriteria = Criteria.where("buyer").is(userId);
        Criteria supplierCriteria = Criteria.where("suppliers.supplierId").is(userId);

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

                contract.setUpdatedAt(new Date());
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