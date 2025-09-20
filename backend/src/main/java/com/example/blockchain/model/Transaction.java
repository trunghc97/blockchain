package com.example.blockchain.model;

import lombok.Data;
import java.util.List;

@Data
public class Transaction {
    private String id;
    private String contractId;
    private String type;
    private String buyer;
    private String bank;
    private List<Supplier> suppliers;
    private double totalAmount;
    private String description;
    private String approverID;
    private String status;
    private String timestamp;
    private boolean included;
}