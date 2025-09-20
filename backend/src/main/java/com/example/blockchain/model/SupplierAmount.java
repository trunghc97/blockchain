package com.example.blockchain.model;

import lombok.Data;

@Data
public class SupplierAmount {
    private String supplierId;
    private String name;
    private Double amount;
    private String status;
}