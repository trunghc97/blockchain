package com.example.blockchain.model;

import lombok.Data;

@Data
public class Approver {
    private String id;
    private String type; // "BANK" or "SUPPLIER"
    private String status;
    private String timestamp;
}