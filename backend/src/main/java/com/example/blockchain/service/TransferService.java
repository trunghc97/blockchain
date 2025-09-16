package com.example.blockchain.service;

import com.example.blockchain.model.*;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.data.mongodb.core.MongoTemplate;
import org.springframework.data.mongodb.core.query.Query;
import org.springframework.http.HttpMethod;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.util.List;

@Service
@RequiredArgsConstructor
public class TransferService {

    private final RestTemplate restTemplate;
    private final MongoTemplate mongoTemplate;

    @Value("${blockchain.service.url}")
    private String blockchainServiceUrl;

    public WorldState createTransfer(TransferRequest request) {
        return restTemplate.postForObject(
            blockchainServiceUrl + "/tx/create",
            request,
            WorldState.class
        );
    }

    public WorldState approveTransfer(ApproveRequest request) {
        return restTemplate.postForObject(
            blockchainServiceUrl + "/tx/approve",
            request,
            WorldState.class
        );
    }

    public WorldState getTransferStatus(String transactionId) {
        return restTemplate.getForObject(
            blockchainServiceUrl + "/tx/status/" + transactionId,
            WorldState.class
        );
    }

    public List<WorldState> getAllTransfers() {
        return mongoTemplate.findAll(WorldState.class);
    }

    public List<Transaction> getAllTransactions() {
        return mongoTemplate.findAll(Transaction.class);
    }

    public List<Block> getAllBlocks() {
        return mongoTemplate.findAll(Block.class);
    }
}
