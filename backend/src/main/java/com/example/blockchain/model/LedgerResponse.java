package com.example.blockchain.model;

import lombok.Data;
import java.util.List;

@Data
public class LedgerResponse {
    private List<Transaction> transactions;
    private List<Block> blocks;
}
