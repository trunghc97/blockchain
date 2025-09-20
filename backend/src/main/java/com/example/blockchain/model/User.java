package com.example.blockchain.model;

import lombok.Data;
import org.springframework.data.mongodb.core.mapping.Document;

@Data
@Document(collection = "users")  // Chỉ định rõ collection name
public class User {
    private String id;
    private String username;
    private String password;
    private String role; // "ANCHOR", "SUPPLIER"
}