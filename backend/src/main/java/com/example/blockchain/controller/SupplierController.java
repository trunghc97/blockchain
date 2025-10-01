package com.example.blockchain.controller;

import com.example.blockchain.model.User;
import com.example.blockchain.service.UserService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/api/v1/suppliers")
@CrossOrigin(origins = "*")
public class SupplierController {
    private static final Logger logger = LoggerFactory.getLogger(SupplierController.class);
    private final UserService userService;

    public SupplierController(UserService userService) {
        this.userService = userService;
    }

    @GetMapping
    public ResponseEntity<?> getAllSuppliers() {
        try {
            List<User> supplierUsers = userService.getUsers().stream()
                    .filter(user -> "SUPPLIER".equals(user.getRole()))
                    .collect(Collectors.toList());

            List<Map<String, Object>> suppliers = supplierUsers.stream()
                    .map(user -> {
                        Map<String, Object> supplier = new java.util.HashMap<>();
                        supplier.put("id", user.getId());
                        supplier.put("username", user.getUsername());
                        supplier.put("role", user.getRole());
                        return supplier;
                    })
                    .collect(Collectors.toList());

            return ResponseEntity.ok(suppliers);
        } catch (Exception e) {
            logger.error("Error getting suppliers", e);
            return ResponseEntity.internalServerError().body("Error getting suppliers: " + e.getMessage());
        }
    }
}
