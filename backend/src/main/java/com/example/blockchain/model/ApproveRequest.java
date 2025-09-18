package com.example.blockchain.model;

import lombok.Data;

@Data
public class ApproveRequest {
    private String transactionId;
    private String approverUserId;
    private String fromAccount;
    private String toAccount;
    private Double amount;
}
