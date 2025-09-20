package com.example.blockchain.model;

import lombok.Data;
import java.util.List;

@Data
public class Block {
    private String id;
    private long blockNumber;
    private String timestamp;
    private String previousHash;
    private String hash;
    private List<String> txIds;
}