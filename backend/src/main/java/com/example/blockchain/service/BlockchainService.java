package com.example.blockchain.service;

import com.example.blockchain.model.*;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpMethod;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class BlockchainService {
    private final RestTemplate restTemplate;
    private final String blockchainUrl;

    public BlockchainService(
            RestTemplate restTemplate,
            @Value("${blockchain.url:http://localhost:8081}") String blockchainUrl
    ) {
        this.restTemplate = restTemplate;
        this.blockchainUrl = blockchainUrl;
    }

    public Map<String, Object> createContract(Map<String, Object> contractData) {
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);

        HttpEntity<Map<String, Object>> request = new HttpEntity<>(contractData, headers);

        ResponseEntity<Map> response = restTemplate.exchange(
                blockchainUrl + "/contract/create",
                HttpMethod.POST,
                request,
                Map.class
        );
        return response.getBody();
    }

    public Map<String, Object> approveContract(String contractId, String supplierId) {
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);

        Map<String, Object> approvalData = new HashMap<>();
        approvalData.put("contractId", contractId);
        approvalData.put("supplierId", supplierId);

        HttpEntity<Map<String, Object>> request = new HttpEntity<>(approvalData, headers);

        ResponseEntity<Map> response = restTemplate.exchange(
                blockchainUrl + "/contract/approve",
                HttpMethod.POST,
                request,
                Map.class
        );
        return response.getBody();
    }

    public List<Map<String, Object>> listContracts() {
        ResponseEntity<List> response = restTemplate.exchange(
                blockchainUrl + "/contract/list",
                HttpMethod.GET,
                null,
                List.class
        );
        return response.getBody();
    }

    public com.example.blockchain.model.LedgerResponse queryLedger(String contractId) {
        ResponseEntity<Map> response = restTemplate.exchange(
                blockchainUrl + "/contract/" + contractId + "/ledger",
                HttpMethod.GET,
                null,
                Map.class
        );
        com.example.blockchain.model.LedgerResponse ledgerResponse = new com.example.blockchain.model.LedgerResponse();
        ledgerResponse.setData(response.getBody());
        return ledgerResponse;
    }

    public List<Map<String, Object>> getUsers() {
        ResponseEntity<List> response = restTemplate.exchange(
                blockchainUrl + "/users",
                HttpMethod.GET,
                null,
                List.class
        );
        return response.getBody();
    }
}
