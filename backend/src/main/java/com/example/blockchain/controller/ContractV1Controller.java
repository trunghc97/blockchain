package com.example.blockchain.controller;

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
}
