package com.example.blockchain.controller;

import com.example.blockchain.service.BlockchainService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/v1/suppliers")
@CrossOrigin(origins = "*")
public class SupplierController {
    private static final Logger logger = LoggerFactory.getLogger(SupplierController.class);
    private final BlockchainService blockchainService;

    public SupplierController(BlockchainService blockchainService) {
        this.blockchainService = blockchainService;
    }

    @GetMapping
    public ResponseEntity<?> getAllSuppliers() {
        try {
            List<Map<String, Object>> suppliers = blockchainService.getAllSuppliers();
            return ResponseEntity.ok(suppliers);
        } catch (Exception e) {
            logger.error("Error getting suppliers", e);
            return ResponseEntity.internalServerError().body("Error getting suppliers: " + e.getMessage());
        }
    }
}
