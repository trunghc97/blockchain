package com.example.blockchain.model;

import lombok.Data;

@Data
public class TransferRequest {
    private String transactionId;
    private String fromAccount;
    private String toAccount;
    private double amount;
    private String approverId;
}
