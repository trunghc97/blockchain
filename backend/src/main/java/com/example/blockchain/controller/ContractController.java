package com.example.blockchain.controller;

import com.example.blockchain.model.ApproveRequest;
import com.example.blockchain.model.Contract;
import com.example.blockchain.model.LedgerResponse;
import com.example.blockchain.model.TransferRequest;
import com.example.blockchain.service.ContractService;
import com.example.blockchain.service.PeerRoutingService;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.List;
import java.util.Map;
import java.util.HashMap;

@RestController
@RequestMapping("/api/contracts")
@CrossOrigin(origins = "*")
public class ContractController {
    private static final Logger logger = LoggerFactory.getLogger(ContractController.class);
    private final ContractService contractService;
    private final PeerRoutingService peerRoutingService;
    private final ObjectMapper objectMapper;

    public ContractController(
        ContractService contractService,
        PeerRoutingService peerRoutingService,
        ObjectMapper objectMapper
    ) {
        this.contractService = contractService;
        this.peerRoutingService = peerRoutingService;
        this.objectMapper = objectMapper;
    }

    @PostMapping
    public ResponseEntity<?> createContract(
        @RequestParam(value = "file", required = false) MultipartFile file,
        @RequestParam("contract") String contractJson
    ) {
        try {
            Contract contract = objectMapper.readValue(contractJson, Contract.class);

            // TEMP: Skip authentication for testing
            // User currentUser = userService.getCurrentUser();
            // contract.setBuyer(currentUser.getUsername());

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

    @GetMapping("/public")
    public ResponseEntity<?> getContractsPublic() {
        try {
            // Return all contracts for testing (bypass authentication)
            List<Contract> contracts = contractService.getContracts();
            return ResponseEntity.ok(contracts);
        } catch (Exception e) {
            logger.error("Error getting contracts", e);
            return ResponseEntity.internalServerError().body("Error getting contracts: " + e.getMessage());
        }
    }

    @GetMapping
    public ResponseEntity<?> getContracts() {
        try {
            // TEMP: Skip authentication for testing
            // User currentUser = userService.getCurrentUser();
            // if (currentUser == null) {
            //     System.out.println("ContractController: currentUser is null");
            //     return ResponseEntity.badRequest().body("User not authenticated");
            // }

            // TEMP: Test with all contracts
            List<Contract> contracts = contractService.getContracts(); // contractService.getContractsByUser(currentUser.getUsername());
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
    public ResponseEntity<?> approveContract(@PathVariable("id") String contractId, @RequestBody ApproveRequest request) {
        try {
            // TEMP: Skip authentication for testing
            // User currentUser = userService.getCurrentUser();
            // if (currentUser == null) {
            //     return ResponseEntity.badRequest().body("User not authenticated");
            // }

            // Use supplierId from request body, fallback to hardcoded for testing
            String supplierId = request != null && request.getApproverId() != null
                ? request.getApproverId()
                : "SUPPLIER001";

            Contract approved = contractService.approveContract(contractId, supplierId);
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

    @GetMapping("/{id}/ledger")
    public ResponseEntity<?> getContractLedger(@PathVariable("id") String contractId) {
        try {
            // Call peer routing service to get ledger data
            LedgerResponse ledgerResponse = peerRoutingService.queryLedger(contractId);

            if (ledgerResponse == null || ledgerResponse.getData() == null) {
                return ResponseEntity.notFound().build();
            }

            return ResponseEntity.ok(ledgerResponse.getData());
        } catch (IllegalArgumentException e) {
            logger.warn("Invalid contract ID: {}", e.getMessage());
            return ResponseEntity.badRequest().body(e.getMessage());
        } catch (Exception e) {
            logger.error("Error getting contract ledger", e);
            return ResponseEntity.internalServerError().body("Error getting contract ledger: " + e.getMessage());
        }
    }

    @PostMapping("/tokens/transfer")
    public ResponseEntity<?> transferToken(@RequestBody TransferRequest transferRequest) {
        try {
            // TEMP: Skip authentication for testing
            // User currentUser = userService.getCurrentUser();
            // if (currentUser == null) {
            //     return ResponseEntity.badRequest().body("User not authenticated");
            // }

            // Call blockchain service to transfer token
            Map<String, Object> transferData = new HashMap<>();
            transferData.put("tokenId", transferRequest.getTokenId());
            transferData.put("from", transferRequest.getFrom());
            transferData.put("to", transferRequest.getTo());
            transferData.put("amount", transferRequest.getAmount());

            Map<String, Object> blockchainResponse = peerRoutingService.transferToken(transferData);

            if (blockchainResponse != null && "success".equals(blockchainResponse.get("status"))) {
                return ResponseEntity.ok(blockchainResponse);
            } else {
                return ResponseEntity.badRequest().body("Transfer failed");
            }
        } catch (IllegalArgumentException e) {
            logger.warn("Invalid transfer data: {}", e.getMessage());
            return ResponseEntity.badRequest().body(e.getMessage());
        } catch (Exception e) {
            logger.error("Error transferring token", e);
            return ResponseEntity.internalServerError().body("Error transferring token: " + e.getMessage());
        }
    }

    @GetMapping("/issued/{bankId}")
    public ResponseEntity<?> getTokensIssuedByBank(@PathVariable("bankId") String bankId) {
        System.out.println("DEBUG: getTokensIssuedByBank called with bankId: " + bankId);
        try {
            List<Map<String, Object>> tokens = peerRoutingService.getTokensIssuedByBank(bankId);
            return ResponseEntity.ok(tokens);
        } catch (Exception e) {
            logger.error("Error getting tokens issued by bank", e);
            return ResponseEntity.internalServerError().body("Error getting tokens: " + e.getMessage());
        }
    }

    @GetMapping("/test")
    public ResponseEntity<?> testEndpoint() {
        return ResponseEntity.ok("Test endpoint works");
    }
}
