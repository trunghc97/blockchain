package com.example.blockchain.model;

import lombok.Data;
import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.mapping.Document;
import org.springframework.data.mongodb.core.mapping.Field;

@Data
@Document(collection = "users")  // Chỉ định rõ collection name
public class User {
    @Id
    private String mongoId; // MongoDB _id

    @Field("id")
    private String id; // Custom id field

    private String username;
    private String password;
    private String role; // "ANCHOR", "SUPPLIER", "BANK"
}
