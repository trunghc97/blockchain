package com.example.blockchain.model;

import lombok.Data;
import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.mapping.Document;

import java.time.LocalDateTime;
import java.util.List;

@Data
@Document(collection = "transactions")
public class Transaction {
    @Id
    private String id;
    private String transactionId;
    private String fromAccount;
    private String toAccount;
    private double amount;
    private String status;
    private String type; // CREATE, APPROVE, EXECUTE
    private String approverId;
    private LocalDateTime timestamp;
}
