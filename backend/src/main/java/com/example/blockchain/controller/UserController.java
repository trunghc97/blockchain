package com.example.blockchain.controller;

import com.example.blockchain.model.LoginRequest;
import com.example.blockchain.model.LoginResponse;
import com.example.blockchain.model.User;
import com.example.blockchain.service.UserService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api")
@CrossOrigin(origins = "*")
public class UserController {
    private final UserService userService;

    public UserController(UserService userService) {
        this.userService = userService;
    }

    @PostMapping("/auth/login")
    public ResponseEntity<LoginResponse> login(@RequestBody LoginRequest request) {
        return ResponseEntity.ok(userService.login(request));
    }

    @GetMapping("/users")
    public ResponseEntity<List<User>> getUsers() {
        return ResponseEntity.ok(userService.getUsers());
    }

    @GetMapping("/users/current")
    public ResponseEntity<User> getCurrentUser(@RequestHeader("Authorization") String token) {
        // Giả sử token format là "Bearer <token>"
        String jwtToken = token.replace("Bearer ", "");
        User user = userService.getCurrentUser(jwtToken);
        return ResponseEntity.ok(user);
    }

    @GetMapping("/users/suppliers")
    public ResponseEntity<List<User>> getSuppliers() {
        return ResponseEntity.ok(userService.getUsers().stream()
                .filter(user -> "SUPPLIER".equals(user.getRole()))
                .toList());
    }
}