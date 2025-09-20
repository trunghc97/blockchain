package com.example.blockchain.utils;

import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;

public class PasswordGenerator {
    public static void main(String[] args) {
        BCryptPasswordEncoder encoder = new BCryptPasswordEncoder(10);
        String password = "123456";
        String hash = encoder.encode(password);
        System.out.println("Password hash for '" + password + "': " + hash);
        
        // Verify
        boolean matches = encoder.matches(password, hash);
        System.out.println("Verification: " + matches);
    }
}
