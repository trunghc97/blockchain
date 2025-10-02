package com.example.blockchain.controller;

import com.example.blockchain.service.PeerRoutingService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/v1/tokens")
@CrossOrigin(origins = "*")
public class TokenController {
    private static final Logger logger = LoggerFactory.getLogger(TokenController.class);
    private final PeerRoutingService peerRoutingService;

    public TokenController(PeerRoutingService peerRoutingService) {
        System.out.println("TokenController constructor called with peerRoutingService: " + peerRoutingService);
        this.peerRoutingService = peerRoutingService;
        System.out.println("TokenController initialized");
    }

    @GetMapping("/{tokenId}")
    public ResponseEntity<?> getToken(@PathVariable("tokenId") String tokenId) {
        try {
            logger.info("Getting token with ID: {}", tokenId);
            Map<String, Object> token = peerRoutingService.getToken(tokenId);
            if (token == null) {
                logger.warn("Token not found: {}", tokenId);
                return ResponseEntity.notFound().build();
            }
            return ResponseEntity.ok(token);
        } catch (IllegalArgumentException e) {
            logger.warn("Invalid token ID: {}", e.getMessage());
            return ResponseEntity.badRequest().body(e.getMessage());
        } catch (Exception e) {
            logger.error("Error getting token", e);
            return ResponseEntity.internalServerError().body("Error getting token: " + e.getMessage());
        }
    }

    @PostMapping("/transfer")
    public ResponseEntity<?> transferToken(@RequestBody Map<String, Object> transferData) {
        try {
            logger.info("Transferring token: {}", transferData);
            System.out.println("TokenController.transferToken called with: " + transferData);
            Map<String, Object> result = peerRoutingService.transferToken(transferData);
            System.out.println("TokenController.transferToken result: " + result);
            return ResponseEntity.ok(result);
        } catch (Exception e) {
            logger.error("Error transferring token", e);
            System.out.println("TokenController.transferToken exception: " + e.getMessage());
            e.printStackTrace();
            return ResponseEntity.internalServerError().body("Error transferring token: " + e.getMessage());
        }
    }

    @GetMapping("/issued/{bankId}")
    public ResponseEntity<?> getTokensIssuedByBank(@PathVariable("bankId") String bankId) {
        try {
            logger.info("Getting tokens issued by bank: {}", bankId);
            List<Map<String, Object>> tokens = peerRoutingService.getTokensIssuedByBank(bankId);
            return ResponseEntity.ok(tokens);
        } catch (Exception e) {
            logger.error("Error getting tokens issued by bank", e);
            return ResponseEntity.internalServerError().body("Error getting tokens issued by bank: " + e.getMessage());
        }
    }

    @GetMapping
    public ResponseEntity<?> getAllTokens() {
        try {
            logger.info("Getting all tokens");
            List<Map<String, Object>> tokens = peerRoutingService.getAllTokens();
            return ResponseEntity.ok(tokens);
        } catch (Exception e) {
            logger.error("Error getting all tokens", e);
            return ResponseEntity.internalServerError().body("Error getting all tokens: " + e.getMessage());
        }
    }

    @GetMapping("/balances/account/{accountId}")
    public ResponseEntity<?> getBalancesByAccount(@PathVariable("accountId") String accountId) {
        try {
            logger.info("Getting balances for account: {}", accountId);
            List<Map<String, Object>> balances = peerRoutingService.getBalancesByAccount(accountId);
            return ResponseEntity.ok(balances);
        } catch (Exception e) {
            logger.error("Error getting balances for account", e);
            return ResponseEntity.internalServerError().body("Error getting balances: " + e.getMessage());
        }
    }

    @GetMapping("/balances/token/{tokenId}")
    public ResponseEntity<?> getBalancesByToken(@PathVariable("tokenId") String tokenId) {
        try {
            logger.info("Getting balances for token: {}", tokenId);
            List<Map<String, Object>> balances = peerRoutingService.getBalancesByToken(tokenId);
            return ResponseEntity.ok(balances);
        } catch (Exception e) {
            logger.error("Error getting balances for token", e);
            return ResponseEntity.internalServerError().body("Error getting balances: " + e.getMessage());
        }
    }

    @PostMapping("/settle")
    public ResponseEntity<?> settleToken(@RequestBody Map<String, Object> settleData) {
        try {
            logger.info("Settling token: {}", settleData);
            Map<String, Object> result = peerRoutingService.settleToken(settleData);
            return ResponseEntity.ok(result);
        } catch (Exception e) {
            logger.error("Error settling token", e);
            return ResponseEntity.internalServerError().body("Error settling token: " + e.getMessage());
        }
    }
}
