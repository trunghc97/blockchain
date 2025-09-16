package com.example.blockchain.model;

import lombok.Data;
import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.mapping.Document;

import java.time.LocalDateTime;
import java.util.List;

@Data
@Document(collection = "blocks")
public class Block {
    @Id
    private String id;
    private long blockNumber;
    private LocalDateTime timestamp;
    private String previousHash;
    private String hash;
    private List<Transaction> transactions;
}
