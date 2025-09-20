package com.example.blockchain.service;

import com.example.blockchain.model.Contract;
import org.springframework.data.mongodb.core.MongoTemplate;
import org.springframework.data.mongodb.core.query.Criteria;
import org.springframework.data.mongodb.core.query.Query;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;
import org.springframework.util.StringUtils;

import java.util.Date;
import java.util.List;
import java.util.Objects;
import java.util.UUID;

@Service
public class ContractService {
    private final MongoTemplate mongoTemplate;

    public ContractService(MongoTemplate mongoTemplate) {
        this.mongoTemplate = mongoTemplate;
    }

    public Contract createContract(Contract contract, MultipartFile file) {
        if (contract == null) {
            throw new IllegalArgumentException("Contract cannot be null");
        }

        // Set initial values
        contract.setContractId(UUID.randomUUID().toString());
        contract.setStatus("PENDING");
        contract.setCreatedAt(new Date());
        contract.setUpdatedAt(new Date());

        // Set initial supplier status
        if (contract.getSuppliers() != null) {
            contract.getSuppliers().forEach(supplier -> {
                if (supplier != null) {
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

        return mongoTemplate.save(contract);
    }

    public List<Contract> getContracts() {
        return mongoTemplate.findAll(Contract.class);
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

        if (contract.getSuppliers() == null || contract.getSuppliers().isEmpty()) {
            throw new RuntimeException("Contract has no suppliers");
        }

        // Update supplier status
        boolean supplierFound = false;
        for (var supplier : contract.getSuppliers()) {
            if (supplier != null && supplierId.equals(supplier.getSupplierId())) {
                supplier.setStatus("APPROVED");
                supplierFound = true;
                break;
            }
        }

        if (!supplierFound) {
            throw new RuntimeException("Supplier not found in contract: " + supplierId);
        }

        // Check if all suppliers approved
        boolean allApproved = contract.getSuppliers().stream()
            .filter(Objects::nonNull)
            .allMatch(s -> "APPROVED".equals(s.getStatus()));

        if (allApproved) {
            contract.setStatus("APPROVED");
        }

        contract.setUpdatedAt(new Date());
        return mongoTemplate.save(contract);
    }
}