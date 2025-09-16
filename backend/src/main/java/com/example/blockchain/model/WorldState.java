package com.example.blockchain.model;

import lombok.Data;
import java.time.LocalDateTime;

@Data
public class WorldState {
    private String id;
    private String transactionId;
    private String fromAccount;
    private String toAccount;
    private double amount;
    private String status;
    private int approvalCount;
    private LocalDateTime lastUpdated;
}
