package com.example.blockchain.model;

import lombok.Data;

import java.util.List;

@Data
public class TransferRequest {
    private String tokenId;
    private String from;
    private String to;
    private double amount;
}
