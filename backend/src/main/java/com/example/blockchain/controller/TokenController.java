package com.example.blockchain.controller;

import com.example.blockchain.service.BlockchainService;
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
    private final BlockchainService blockchainService;

    public TokenController(BlockchainService blockchainService) {
        this.blockchainService = blockchainService;
    }

    @GetMapping("/{tokenId}")
    public ResponseEntity<?> getToken(@PathVariable("tokenId") String tokenId) {
        try {
            logger.info("Getting token with ID: {}", tokenId);
            Map<String, Object> token = blockchainService.getToken(tokenId);
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
            Map<String, Object> result = blockchainService.transferToken(transferData);
            return ResponseEntity.ok(result);
        } catch (Exception e) {
            logger.error("Error transferring token", e);
            return ResponseEntity.internalServerError().body("Error transferring token: " + e.getMessage());
        }
    }

    @GetMapping("/issued/{bankId}")
    public ResponseEntity<?> getTokensIssuedByBank(@PathVariable("bankId") String bankId) {
        try {
            logger.info("Getting tokens issued by bank: {}", bankId);
            List<Map<String, Object>> tokens = blockchainService.getTokensIssuedByBank(bankId);
            return ResponseEntity.ok(tokens);
        } catch (Exception e) {
            logger.error("Error getting tokens issued by bank", e);
            return ResponseEntity.internalServerError().body("Error getting tokens issued by bank: " + e.getMessage());
        }
    }
}
