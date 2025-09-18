package com.example.blockchain.model;

import lombok.Data;

import java.util.List;

@Data
public class TransferRequest {
    private String transactionId;  // Thêm trường này
    private String fromAccount;    // Đổi tên từ fromUser
    private String toAccount;
    private double amount;
    private String description;
    private List<String> approvers;
}
