package com.example.blockchain.controller;

import com.example.blockchain.model.ApproveRequest;
import com.example.blockchain.model.Contract;
import com.example.blockchain.service.ContractService;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/v1/contracts")
@CrossOrigin(origins = "*")
public class ContractV1Controller {
    private static final Logger logger = LoggerFactory.getLogger(ContractV1Controller.class);
    private final ContractService contractService;
    private final ObjectMapper objectMapper;

    public ContractV1Controller(ContractService contractService, ObjectMapper objectMapper) {
        this.contractService = contractService;
        this.objectMapper = objectMapper;
    }

    @PostMapping(consumes = MediaType.MULTIPART_FORM_DATA_VALUE)
    public ResponseEntity<?> createContractFormData(
        @RequestParam(value = "file", required = false) MultipartFile file,
        @RequestParam("contract") String contractJson
    ) {
        try {
            logger.info("Creating contract with form-data: {}", contractJson);
            Contract contract = objectMapper.readValue(contractJson, Contract.class);

            // TEMP: Skip authentication for testing
            // User currentUser = userService.getCurrentUser();
            // contract.setBuyer(currentUser.getUsername());

            Contract created = contractService.createContract(contract, file);
            logger.info("Contract created successfully with ID: {}", created.getContractId());
            return ResponseEntity.ok(created);
        } catch (IllegalArgumentException e) {
            logger.warn("Invalid request data: {}", e.getMessage());
            return ResponseEntity.badRequest().body(e.getMessage());
        } catch (Exception e) {
            logger.error("Error creating contract", e);
            return ResponseEntity.internalServerError().body("Error creating contract: " + e.getMessage());
        }
    }

    @PostMapping(consumes = MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<?> createContractJson(@RequestBody Contract contract) {
        try {
            logger.info("Creating contract with JSON: {}", contract);

            Contract created = contractService.createContract(contract, null);
            logger.info("Contract created successfully with ID: {}", created.getContractId());
            return ResponseEntity.ok(created);
        } catch (IllegalArgumentException e) {
            logger.warn("Invalid request data: {}", e.getMessage());
            return ResponseEntity.badRequest().body(e.getMessage());
        } catch (Exception e) {
            logger.error("Error creating contract", e);
            return ResponseEntity.internalServerError().body("Error creating contract: " + e.getMessage());
        }
    }

    @GetMapping
    public ResponseEntity<?> getContracts() {
        try {
            // TEMP: Skip authentication for testing
            List<Contract> contracts = contractService.getContracts();
            return ResponseEntity.ok(contracts);
        } catch (Exception e) {
            logger.error("Error getting contracts", e);
            return ResponseEntity.internalServerError().body("Error getting contracts: " + e.getMessage());
        }
    }

    @GetMapping("/{id}")
    public ResponseEntity<?> getContract(@PathVariable("id") String contractId) {
        logger.info("ContractV1Controller.getContract called with ID: {}", contractId);
        try {
            Contract contract = contractService.getContract(contractId);
            if (contract == null) {
                logger.info("Contract not found: {}", contractId);
                return ResponseEntity.notFound().build();
            }
            return ResponseEntity.ok(contract);
        } catch (IllegalArgumentException e) {
            logger.warn("Invalid contract ID: {}", e.getMessage());
            return ResponseEntity.badRequest().body(e.getMessage());
        } catch (RuntimeException e) {
            logger.info("Contract {} not found: {}", contractId, e.getMessage());
            return ResponseEntity.notFound().build();
        } catch (Exception e) {
            logger.error("Error getting contract {}: {}", contractId, e.getMessage());
            return ResponseEntity.internalServerError().body("Error getting contract: " + e.getMessage());
        }
    }

    @PostMapping("/{id}/approve-bank")
    public ResponseEntity<?> approveContractByBank(
        @PathVariable("id") String contractId,
        @RequestBody Map<String, String> requestBody
    ) {
        try {
            String bankId = requestBody.get("bankId");
            if (bankId == null || bankId.isEmpty()) {
                return ResponseEntity.badRequest().body("Bank ID is required");
            }

            logger.info("Bank {} approving contract {}", bankId, contractId);
            Contract approved = contractService.approveContractByBank(contractId, bankId);
            logger.info("Contract approved by bank successfully with ID: {}", approved.getContractId());
            return ResponseEntity.ok(approved);
        } catch (IllegalArgumentException e) {
            logger.warn("Invalid request data: {}", e.getMessage());
            return ResponseEntity.badRequest().body(e.getMessage());
        } catch (Exception e) {
            logger.error("Error approving contract by bank", e);
            return ResponseEntity.internalServerError().body("Error approving contract by bank: " + e.getMessage());
        }
    }

    @PostMapping("/{id}/approve")
    public ResponseEntity<?> approveContract(
        @PathVariable("id") String contractId,
        @RequestBody Map<String, String> requestBody
    ) {
        try {
            String supplierId = requestBody.get("supplierId");
            if (supplierId == null || supplierId.isEmpty()) {
                return ResponseEntity.badRequest().body("Supplier ID is required");
            }

            logger.info("Supplier {} approving contract {}", supplierId, contractId);
            Contract approved = contractService.approveContract(contractId, supplierId);
            logger.info("Contract approved by supplier successfully with ID: {}", approved.getContractId());
            return ResponseEntity.ok(approved);
        } catch (IllegalArgumentException e) {
            logger.warn("Invalid request data: {}", e.getMessage());
            return ResponseEntity.badRequest().body(e.getMessage());
        } catch (RuntimeException e) {
            logger.warn("Business logic error: {}", e.getMessage());
            return ResponseEntity.badRequest().body(e.getMessage());
        } catch (Exception e) {
            logger.error("Error approving contract by supplier", e);
            return ResponseEntity.internalServerError().body("Error approving contract by supplier: " + e.getMessage());
        }
    }

    @GetMapping("/tokens/all")
    public ResponseEntity<?> getAllTokens() {
        try {
            List<Map<String, Object>> tokens = contractService.getAllTokens();
            return ResponseEntity.ok(tokens);
        } catch (Exception e) {
            logger.error("Error getting all tokens", e);
            return ResponseEntity.internalServerError().body("Error getting all tokens: " + e.getMessage());
        }
    }

    @GetMapping("/balances/all")
    public ResponseEntity<?> getAllBalances() {
        try {
            List<Map<String, Object>> balances = contractService.getAllBalances();
            return ResponseEntity.ok(balances);
        } catch (Exception e) {
            logger.error("Error getting all balances", e);
            return ResponseEntity.internalServerError().body("Error getting all balances: " + e.getMessage());
        }
    }

    @GetMapping("/balances/token/{tokenId}")
    public ResponseEntity<?> getBalancesByToken(@PathVariable("tokenId") String tokenId) {
        try {
            List<Map<String, Object>> balances = contractService.getBalancesByToken(tokenId);
            return ResponseEntity.ok(balances);
        } catch (Exception e) {
            logger.error("Error getting balances by token", e);
            return ResponseEntity.internalServerError().body("Error getting balances by token: " + e.getMessage());
        }
    }
}
