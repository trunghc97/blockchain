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

    public Map<String, Object> getContract(String contractId) {
        ResponseEntity<Map> response = restTemplate.exchange(
                blockchainUrl + "/contract/" + contractId,
                HttpMethod.GET,
                null,
                Map.class
        );
        return response.getBody();
    }

    public Map<String, Object> getToken(String tokenId) {
        ResponseEntity<Map> response = restTemplate.exchange(
                blockchainUrl + "/token/" + tokenId,
                HttpMethod.GET,
                null,
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
                blockchainUrl + "/contract/" + contractId + "/approve",
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

    public LedgerResponse queryLedger(String contractId) {
        ResponseEntity<Map> response = restTemplate.exchange(
                blockchainUrl + "/contract/" + contractId + "/ledger",
                HttpMethod.GET,
                null,
                Map.class
        );
        LedgerResponse ledgerResponse = new LedgerResponse();
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

    public Map<String, Object> transferToken(Map<String, Object> transferData) {
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);

        HttpEntity<Map<String, Object>> request = new HttpEntity<>(transferData, headers);

        ResponseEntity<Map> response = restTemplate.exchange(
                blockchainUrl + "/token/transfer",
                HttpMethod.POST,
                request,
                Map.class
        );
        return response.getBody();
    }

    public List<Map<String, Object>> getTokensIssuedByBank(String bankId) {
        System.out.println("DEBUG: Calling blockchain service for tokens issued by: " + bankId);
        try {
            ResponseEntity<List> response = restTemplate.exchange(
                    blockchainUrl + "/token/issued/" + bankId,
                    HttpMethod.GET,
                    null,
                    List.class
            );
            System.out.println("DEBUG: Response: " + response.getBody());
            return response.getBody();
        } catch (Exception e) {
            System.out.println("DEBUG: Error calling blockchain: " + e.getMessage());
            throw e;
        }
    }

    public List<Map<String, Object>> getAllSuppliers() {
        ResponseEntity<List> response = restTemplate.exchange(
                blockchainUrl + "/suppliers",
                HttpMethod.GET,
                null,
                List.class
        );
        return response.getBody();
    }
}
