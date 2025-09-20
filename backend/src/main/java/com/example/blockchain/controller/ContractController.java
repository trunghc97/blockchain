package com.example.blockchain.controller;

import com.example.blockchain.model.Contract;
import com.example.blockchain.service.ContractService;
import com.example.blockchain.service.UserService;
import com.example.blockchain.model.User;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.List;

@RestController
@RequestMapping("/api/contracts")
@CrossOrigin(origins = "*")
public class ContractController {
    private static final Logger logger = LoggerFactory.getLogger(ContractController.class);
    private final ContractService contractService;
    private final UserService userService;
    private final ObjectMapper objectMapper;

    public ContractController(
        ContractService contractService,
        UserService userService,
        ObjectMapper objectMapper
    ) {
        this.contractService = contractService;
        this.userService = userService;
        this.objectMapper = objectMapper;
    }

    @PostMapping
    public ResponseEntity<?> createContract(
        @RequestParam(value = "file", required = false) MultipartFile file,
        @RequestParam("contract") String contractJson
    ) {
        try {
            Contract contract = objectMapper.readValue(contractJson, Contract.class);
            Contract created = contractService.createContract(contract, file);
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
            List<Contract> contracts = contractService.getContracts();
            return ResponseEntity.ok(contracts);
        } catch (Exception e) {
            logger.error("Error getting contracts", e);
            return ResponseEntity.internalServerError().body("Error getting contracts: " + e.getMessage());
        }
    }

    @GetMapping("/{id}")
    public ResponseEntity<?> getContract(@PathVariable("id") String contractId) {
        try {
            Contract contract = contractService.getContract(contractId);
            if (contract == null) {
                return ResponseEntity.notFound().build();
            }
            return ResponseEntity.ok(contract);
        } catch (IllegalArgumentException e) {
            logger.warn("Invalid contract ID: {}", e.getMessage());
            return ResponseEntity.badRequest().body(e.getMessage());
        } catch (Exception e) {
            logger.error("Error getting contract", e);
            return ResponseEntity.internalServerError().body("Error getting contract: " + e.getMessage());
        }
    }

    @PostMapping("/{id}/approve")
    public ResponseEntity<?> approveContract(@PathVariable("id") String contractId) {
        try {
            User currentUser = userService.getCurrentUser();
            if (currentUser == null) {
                return ResponseEntity.badRequest().body("User not authenticated");
            }

            Contract approved = contractService.approveContract(contractId, currentUser.getId());
            return ResponseEntity.ok(approved);
        } catch (IllegalArgumentException e) {
            logger.warn("Invalid request data: {}", e.getMessage());
            return ResponseEntity.badRequest().body(e.getMessage());
        } catch (RuntimeException e) {
            logger.warn("Business logic error: {}", e.getMessage());
            return ResponseEntity.badRequest().body(e.getMessage());
        } catch (Exception e) {
            logger.error("Error approving contract", e);
            return ResponseEntity.internalServerError().body("Error approving contract: " + e.getMessage());
        }
    }
}