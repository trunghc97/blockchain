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
    private final UserService userService;

    public ContractService(MongoTemplate mongoTemplate, BlockchainService blockchainService, UserService userService) {
        this.mongoTemplate = mongoTemplate;
        this.blockchainService = blockchainService;
        this.userService = userService;
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
        contract.setStatus("PENDING_BANK_APPROVAL");
        contract.setBankApproved(false);
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
            contractData.put("bankId", contract.getBankId() != null ? contract.getBankId() : "BANK001"); // use provided bankId or default
            contractData.put("bankApproved", contract.getBankApproved());
            contractData.put("amount", contract.getTotalAmount());
            contractData.put("suppliers", contract.getSuppliers());
            contractData.put("approvers", contract.getSuppliers() != null
                ? contract.getSuppliers().stream().map(s -> s.getSupplierId()).toList()
                : java.util.Collections.emptyList());

            System.out.println("DEBUG: Calling blockchain service with data: " + contractData);
            // Call blockchain service to create contract (without token creation)
            Map<String, Object> blockchainResponse = blockchainService.createContract(contractData);
            System.out.println("DEBUG: Blockchain response: " + blockchainResponse);

            if (blockchainResponse != null && "success".equals(blockchainResponse.get("status"))) {
                // Save contract to local MongoDB for immediate availability
                mongoTemplate.save(contract, "contracts");

                // Get the created contract from blockchain service to return complete data
                String createdContractId = (String) blockchainResponse.get("contractId");
                if (createdContractId != null) {
                    Map<String, Object> blockchainContract = blockchainService.getContract(createdContractId);
                    if (blockchainContract != null) {
                        Contract localContract = convertBlockchainContractToLocal(blockchainContract);
                        // Update local MongoDB with blockchain data
                        mongoTemplate.save(localContract, "contracts");
                        return localContract;
                    }
                }
                // Return the local contract if blockchain data is not available
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
        // Query contracts from MongoDB local first
        List<Contract> localContracts = mongoTemplate.findAll(Contract.class, "contracts");

        // Group contracts by contractId to remove duplicates, keeping the most complete one
        Map<String, Contract> contractMap = new java.util.HashMap<>();

        for (Contract contract : localContracts) {
            String contractId = contract.getContractId();
            Contract existing = contractMap.get(contractId);

            if (existing == null) {
                contractMap.put(contractId, contract);
            } else {
                // Keep the contract with more complete information
                Contract betterContract = chooseBetterContract(existing, contract);
                contractMap.put(contractId, betterContract);
            }
        }

        return new java.util.ArrayList<>(contractMap.values());
    }

    private Contract chooseBetterContract(Contract contract1, Contract contract2) {
        // Prefer contract with more complete information
        int score1 = calculateCompletenessScore(contract1);
        int score2 = calculateCompletenessScore(contract2);

        if (score1 > score2) {
            return contract1;
        } else if (score2 > score1) {
            return contract2;
        } else {
            // Same score, prefer local contract (UUID) over blockchain contract (ObjectID)
            boolean isContract1Local = isLocalContract(contract1);
            boolean isContract2Local = isLocalContract(contract2);

            if (isContract1Local && !isContract2Local) {
                return contract1;
            } else if (isContract2Local && !isContract1Local) {
                return contract2;
            } else {
                // Both same type or both blockchain, prefer Contract2 (later processed)
                return contract2;
            }
        }
    }

    private boolean isLocalContract(Contract contract) {
        // Local contracts have UUID-style IDs, blockchain contracts have MongoDB ObjectIDs
        String id = contract.getId();
        if (id == null) return false;

        // UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (36 chars)
        // MongoDB ObjectID: 24 hex chars
        return id.length() == 36 && id.contains("-");
    }

    private int calculateCompletenessScore(Contract contract) {
        int score = 0;
        if (contract.getDescription() != null && !contract.getDescription().isEmpty()) score += 2;
        if (contract.getBuyer() != null && !contract.getBuyer().isEmpty()) score += 1;
        if (contract.getSuppliers() != null && !contract.getSuppliers().isEmpty()) score += 3;
        if (contract.getTotalAmount() != null && contract.getTotalAmount() > 0) score += 2;
        if (contract.getBankId() != null && !contract.getBankId().isEmpty()) score += 1;
        return score;
    }

    private void saveContractToLocal(Contract contract) {
        // Save contract to local MongoDB if not exists
        Query query = new Query(Criteria.where("contractId").is(contract.getContractId()));
        Contract existing = mongoTemplate.findOne(query, Contract.class, "contracts");
        if (existing == null) {
            mongoTemplate.save(contract, "contracts");
        }
    }

    private Contract convertBlockchainContractToLocal(Map<String, Object> blockchainContract) {
        Contract contract = new Contract();
        contract.setId((String) blockchainContract.get("_id"));
        contract.setContractId((String) blockchainContract.get("_id"));
        contract.setDescription((String) blockchainContract.get("description"));
        contract.setBuyer((String) blockchainContract.get("anchorId"));
        contract.setBankId((String) blockchainContract.get("bankId"));

        // Set bank approval status
        Boolean bankApproved = (Boolean) blockchainContract.get("bankApproved");
        contract.setBankApproved(bankApproved != null ? bankApproved : false);

        // Set status based on approval fields
        Boolean approved = (Boolean) blockchainContract.get("approved");
        if (approved != null && approved) {
            contract.setStatus("EXECUTED");
        } else if (bankApproved != null && bankApproved) {
            contract.setStatus("BANK_APPROVED");
        } else {
            contract.setStatus("PENDING_BANK_APPROVAL");
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
        if (allBlockchainContracts == null) {
            return java.util.Collections.emptyList();
        }
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

        // Bank users can see all contracts
        if (username != null && username.startsWith("BANK")) {
            return true;
        }

        // Suppliers can only see contracts that have been approved by bank
        if (contract.getSuppliers() != null) {
            boolean isSupplier = contract.getSuppliers().stream()
                .anyMatch(supplier -> username.equals(supplier.getSupplierId()));

            if (isSupplier) {
                // Suppliers can only see contracts that have been bank-approved
                return contract.getBankApproved() != null && contract.getBankApproved();
            }
        }

        return false;
    }

    public Contract getContract(String contractId) {
        if (!StringUtils.hasText(contractId)) {
            throw new IllegalArgumentException("Contract ID cannot be empty");
        }

        // Query contract from local database first
        Contract localContract = mongoTemplate.findById(contractId, Contract.class);
        if (localContract != null) {
            return localContract;
        }

        // Fallback to blockchain service if not found locally
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

        // Verify contract has been approved by bank before allowing supplier approval
        if (contract.getBankApproved() == null || !contract.getBankApproved()) {
            throw new RuntimeException("Contract must be approved by bank before supplier approval: " + contractId);
        }

        // Verify supplier is part of this contract
        boolean isValidSupplier = contract.getSuppliers() != null &&
            contract.getSuppliers().stream()
                .anyMatch(supplier -> {
                    // Check exact supplierId match
                    if (supplier.getSupplierId() != null && supplierId.equals(supplier.getSupplierId())) {
                        return true;
                    }

                    // Check if supplierId is null and supplier name matches user username
                    if (supplier.getSupplierId() == null && supplier.getName() != null) {
                        try {
                            // Get user by supplierId to get their username
                            User user = mongoTemplate.findOne(
                                Query.query(Criteria.where("id").is(supplierId)),
                                User.class,
                                "users"
                            );
                            return user != null && supplier.getName().equals(user.getUsername());
                        } catch (Exception e) {
                            // If user lookup fails, don't allow
                            return false;
                        }
                    }

                    return false;
                });

        if (!isValidSupplier) {
            throw new RuntimeException("Supplier " + supplierId + " is not authorized to approve contract " + contractId);
        }

        try {
            // Call blockchain service to approve contract
            Map<String, Object> blockchainResponse = blockchainService.approveContract(contractId, supplierId);

            if (blockchainResponse != null && "success".equals(blockchainResponse.get("status"))) {
                // Update local contract status for quick access
                if (contract.getSuppliers() != null) {
                    for (var supplier : contract.getSuppliers()) {
                        if (supplier != null) {
                            // Check exact supplierId match
                            if (supplierId.equals(supplier.getSupplierId())) {
                                supplier.setStatus("APPROVED");
                                break;
                            }
                            // Check if supplierId is null and supplier name matches user username
                            else if (supplier.getSupplierId() == null && supplier.getName() != null) {
                                try {
                                    User user = mongoTemplate.findOne(
                                        Query.query(Criteria.where("id").is(supplierId)),
                                        User.class,
                                        "users"
                                    );
                                    if (user != null && supplier.getName().equals(user.getUsername())) {
                                        supplier.setStatus("APPROVED");
                                        break;
                                    }
                                } catch (Exception e) {
                                    // Continue to next supplier if lookup fails
                                }
                            }
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

    public Contract approveContractByBank(String contractId, String bankId) {
        if (!StringUtils.hasText(contractId)) {
            throw new IllegalArgumentException("Contract ID cannot be empty");
        }
        if (!StringUtils.hasText(bankId)) {
            throw new IllegalArgumentException("Bank ID cannot be empty");
        }

        Contract contract = getContract(contractId);
        if (contract == null) {
            throw new RuntimeException("Contract not found: " + contractId);
        }

        // Verify bank has permission to approve this contract
        if (!bankId.equals(contract.getBankId())) {
            throw new RuntimeException("Bank " + bankId + " does not have permission to approve contract " + contractId);
        }

        try {
            // Call blockchain service to approve contract by bank
            Map<String, Object> blockchainResponse = blockchainService.approveContractByBank(contractId, bankId);

            if (blockchainResponse != null && "success".equals(blockchainResponse.get("status"))) {
                // Update local contract status
                contract.setBankApproved(true);
                contract.setStatus("BANK_APPROVED");

                String now = LocalDateTime.now().format(DateTimeFormatter.ISO_LOCAL_DATE_TIME) + "Z";
                contract.setUpdatedAt(now);
                return mongoTemplate.save(contract);
            } else {
                throw new RuntimeException("Failed to approve contract by bank on blockchain");
            }
        } catch (Exception e) {
            System.err.println("Error calling blockchain service: " + e.getMessage());
            throw new RuntimeException("Blockchain service unavailable", e);
        }
    }

    public List<Map<String, Object>> getAllTokens() {
        try {
            return blockchainService.getAllTokens();
        } catch (Exception e) {
            System.err.println("Error getting all tokens: " + e.getMessage());
            throw new RuntimeException("Blockchain service unavailable", e);
        }
    }

    public List<Map<String, Object>> getAllBalances() {
        try {
            // Get all balances by combining balances from all accounts
            List<Map<String, Object>> allBalances = new java.util.ArrayList<>();

            // Get all users to iterate through their balances
            List<User> users = userService.getUsers();
            for (User user : users) {
                String userId = user.getId();
                if (userId != null) {
                    List<Map<String, Object>> userBalances = blockchainService.getBalancesByAccount(userId);
                    if (userBalances != null) {
                        allBalances.addAll(userBalances);
                    }
                }
            }

            return allBalances;
        } catch (Exception e) {
            System.err.println("Error getting all balances: " + e.getMessage());
            throw new RuntimeException("Error retrieving balances", e);
        }
    }

    public List<Map<String, Object>> getBalancesByToken(String tokenId) {
        try {
            return blockchainService.getBalancesByToken(tokenId);
        } catch (Exception e) {
            System.err.println("Error getting balances by token: " + e.getMessage());
            throw new RuntimeException("Blockchain service unavailable", e);
        }
    }
}
