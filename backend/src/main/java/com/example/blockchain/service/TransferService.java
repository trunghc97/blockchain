package com.example.blockchain.service;

import com.example.blockchain.model.TransferRequest;
import com.example.blockchain.model.WorldState;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

@Service
@RequiredArgsConstructor
public class TransferService {

    private final RestTemplate restTemplate;

    @Value("${blockchain.service.url}")
    private String blockchainServiceUrl;

    public WorldState createTransfer(TransferRequest request) {
        return restTemplate.postForObject(
            blockchainServiceUrl + "/tx/create",
            request,
            WorldState.class
        );
    }

    public WorldState approveTransfer(TransferRequest request) {
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
}
