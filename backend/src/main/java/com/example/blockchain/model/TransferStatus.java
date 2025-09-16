package com.example.blockchain.model;

import lombok.Data;

import java.time.LocalDateTime;
import java.util.List;

@Data
public class TransferStatus {
    private String reqId;
    private String fromUser;
    private String toAccount;
    private double amount;
    private String description;
    private String status; // PENDING, PARTIALLY_APPROVED, EXECUTED
    private List<String> approvers;
    private List<String> approvedBy;
    private LocalDateTime createdAt;
}
