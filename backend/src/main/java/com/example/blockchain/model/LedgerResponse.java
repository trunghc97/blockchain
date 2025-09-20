package com.example.blockchain.model;

import lombok.Data;
import java.util.Map;

@Data
public class LedgerResponse {
    private Map<String, Object> data;
}
