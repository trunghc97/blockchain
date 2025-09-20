package com.example.blockchain.model;

import lombok.Data;

@Data
public class SupplierAmount {
    private String supplierId;
    private Double amount;
    private String status;
}
