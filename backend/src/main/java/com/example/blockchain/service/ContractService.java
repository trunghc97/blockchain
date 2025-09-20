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
import java.util.Objects;
import java.util.UUID;
import java.util.Map;
import java.util.stream.Collectors;

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

        // Calculate total amount from suppliers
        if (contract.getSuppliers() != null && !contract.getSuppliers().isEmpty()) {
            double total = contract.getSuppliers().stream()
                .filter(Objects::nonNull)
                .mapToDouble(supplier -> supplier.getAmount() != null ? supplier.getAmount() : 0.0)
                .sum();
            contract.setTotalAmount(total);
        }

        // Get supplier names before saving
        enrichContractWithSupplierNames(contract);

        // Set initial values
        contract.setContractId(UUID.randomUUID().toString());
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

        return mongoTemplate.save(contract);
    }

    public List<Contract> getContracts() {
        List<Contract> contracts = mongoTemplate.findAll(Contract.class);
        return contracts.stream()
            .map(this::enrichContractWithSupplierNames)
            .collect(Collectors.toList());
    }

    public List<Contract> getContractsByUser(String userId) {
        if (!StringUtils.hasText(userId)) {
            throw new IllegalArgumentException("User ID cannot be empty");
        }

        // Find contracts where user is either buyer (anchor) or a supplier
        Criteria buyerCriteria = Criteria.where("buyer").is(userId);
        Criteria supplierCriteria = Criteria.where("suppliers.supplierId").is(userId);

        Query query = new Query(new Criteria().orOperator(buyerCriteria, supplierCriteria));

        List<Contract> contracts = mongoTemplate.find(query, Contract.class);
        return contracts.stream()
            .map(this::enrichContractWithSupplierNames)
            .collect(Collectors.toList());
    }

    public Contract getContract(String contractId) {
        if (!StringUtils.hasText(contractId)) {
            throw new IllegalArgumentException("Contract ID cannot be empty");
        }

        Contract contract = mongoTemplate.findOne(
            Query.query(Criteria.where("contractId").is(contractId)),
            Contract.class
        );

        if (contract != null) {
            return enrichContractWithSupplierNames(contract);
        }

        return contract;
    }

    private Contract enrichContractWithSupplierNames(Contract contract) {
        if (contract == null || contract.getSuppliers() == null) {
            return contract;
        }

        // Get all supplier IDs from the contract
        List<String> supplierIds = contract.getSuppliers().stream()
            .filter(Objects::nonNull)
            .map(supplier -> supplier.getSupplierId())
            .filter(StringUtils::hasText)
            .distinct()
            .collect(Collectors.toList());

        if (supplierIds.isEmpty()) {
            return contract;
        }

        // Fetch all suppliers in one query
        List<User> suppliers = mongoTemplate.find(
            Query.query(Criteria.where("id").in(supplierIds)),
            User.class
        );

        // Create a map for quick lookup
        Map<String, String> supplierIdToNameMap = suppliers.stream()
            .filter(Objects::nonNull)
            .collect(Collectors.toMap(User::getId, User::getUsername));

        // Update supplier names in contract
        contract.getSuppliers().forEach(supplier -> {
            if (supplier != null && supplier.getSupplierId() != null) {
                String supplierName = supplierIdToNameMap.get(supplier.getSupplierId());
                if (supplierName != null) {
                    supplier.setName(supplierName);
                }
            }
        });

        return contract;
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