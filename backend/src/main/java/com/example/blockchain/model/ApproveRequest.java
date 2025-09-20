package com.example.blockchain.model;

import lombok.Data;

@Data
public class ApproveRequest {
    private String contractId;
    private String approverId;
}