package com.example.blockchain.service;

import com.example.blockchain.model.*;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpMethod;
import org.springframework.http.HttpStatus;
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
    private final String peerAnchorUrl;
    private final String peerSupplierUrl;
    private final String peerMainBankUrl;

    public BlockchainService(
            RestTemplate restTemplate,
            @Value("${peer.anchor.url:http://peer-anchor:8084}") String peerAnchorUrl,
            @Value("${peer.supplier.url:http://peer-supplier:8083}") String peerSupplierUrl,
            @Value("${peer.main-bank.url:http://peer-main-bank:8082}") String peerMainBankUrl
    ) {
        this.restTemplate = restTemplate;
        this.peerAnchorUrl = peerAnchorUrl;
        this.peerSupplierUrl = peerSupplierUrl;
        this.peerMainBankUrl = peerMainBankUrl;
    }

    // Contract operations - routed based on user role
    public Map<String, Object> createContract(Map<String, Object> contractData) {
        // Always route to peer-anchor for contract creation (only anchor can create contracts)
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);

        HttpEntity<Map<String, Object>> request = new HttpEntity<>(contractData, headers);

        ResponseEntity<Map> response = restTemplate.exchange(
                peerAnchorUrl + "/contract/create",
                HttpMethod.POST,
                request,
                Map.class
        );
        return response.getBody();
    }

    public Map<String, Object> getContract(String contractId) {
        // Route to peer-anchor for contract details
        ResponseEntity<Map> response = restTemplate.exchange(
                peerAnchorUrl + "/contract/" + contractId,
                HttpMethod.GET,
                null,
                Map.class
        );

        if (response.getStatusCode() == HttpStatus.NOT_FOUND) {
            // Contract not found
            return null;
        }

        return response.getBody();
    }

    public Map<String, Object> getToken(String tokenId) {
        try {
            // Route to peer-supplier for token details
            ResponseEntity<Map> response = restTemplate.exchange(
                    peerSupplierUrl + "/token/" + tokenId,
                    HttpMethod.GET,
                    null,
                    Map.class
            );
            return response.getBody();
        } catch (org.springframework.web.client.RestClientException e) {
            // Token not found, return null
            return null;
        }
    }

    public Map<String, Object> approveContract(String contractId, String supplierId) {
        // Route to peer-supplier for supplier approval
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);

        Map<String, Object> approvalData = new HashMap<>();
        approvalData.put("contractId", contractId);
        approvalData.put("supplierId", supplierId);

        HttpEntity<Map<String, Object>> request = new HttpEntity<>(approvalData, headers);

        ResponseEntity<Map> response = restTemplate.exchange(
                peerSupplierUrl + "/contract/" + contractId + "/approve",
                HttpMethod.POST,
                request,
                Map.class
        );
        return response.getBody();
    }

    public Map<String, Object> approveContractByBank(String contractId, String bankId) {
        // Route to peer-anchor for bank approval (contracts are created in peer-anchor)
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);

        Map<String, Object> approvalData = new HashMap<>();
        approvalData.put("contractId", contractId);
        approvalData.put("bankId", bankId);

        HttpEntity<Map<String, Object>> request = new HttpEntity<>(approvalData, headers);

        ResponseEntity<Map> response = restTemplate.exchange(
                peerAnchorUrl + "/contract/" + contractId + "/approve-bank",
                HttpMethod.POST,
                request,
                Map.class
        );
        return response.getBody();
    }

    public List<Map<String, Object>> listContracts() {
        // Route to peer-anchor for contract list
        ResponseEntity<List> response = restTemplate.exchange(
                peerAnchorUrl + "/contract/list",
                HttpMethod.GET,
                null,
                List.class
        );
        return response.getBody();
    }

    public LedgerResponse queryLedger(String contractId) {
        // Route to peer-anchor for ledger query
        ResponseEntity<Map> response = restTemplate.exchange(
                peerAnchorUrl + "/contract/" + contractId + "/ledger",
                HttpMethod.GET,
                null,
                Map.class
        );
        LedgerResponse ledgerResponse = new LedgerResponse();
        ledgerResponse.setData(response.getBody());
        return ledgerResponse;
    }

    public List<Map<String, Object>> getUsers() {
        // Route to peer-supplier for user data
        ResponseEntity<List> response = restTemplate.exchange(
                peerSupplierUrl + "/users",
                HttpMethod.GET,
                null,
                List.class
        );
        return response.getBody();
    }

    public Map<String, Object> transferToken(Map<String, Object> transferData) {
        // Route to peer-supplier for token transfer
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);

        HttpEntity<Map<String, Object>> request = new HttpEntity<>(transferData, headers);

        ResponseEntity<Map> response = restTemplate.exchange(
                peerSupplierUrl + "/token/transfer",
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
                    peerMainBankUrl + "/token/issued/" + bankId,
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
        // Route to peer-supplier for supplier data
        ResponseEntity<List> response = restTemplate.exchange(
                peerSupplierUrl + "/suppliers",
                HttpMethod.GET,
                null,
                List.class
        );
        return response.getBody();
    }

    public List<Map<String, Object>> getAllTokens() {
        // Route to peer-anchor for all tokens
        ResponseEntity<List> response = restTemplate.exchange(
                peerAnchorUrl + "/tokens",
                HttpMethod.GET,
                null,
                List.class
        );
        return response.getBody();
    }

    public List<Map<String, Object>> getBalancesByAccount(String accountId) {
        // Route to peer-supplier for account balances
        ResponseEntity<List> response = restTemplate.exchange(
                peerSupplierUrl + "/balances/account/" + accountId,
                HttpMethod.GET,
                null,
                List.class
        );
        return response.getBody();
    }

    public List<Map<String, Object>> getBalancesByToken(String tokenId) {
        // Route to peer-anchor for token balances
        ResponseEntity<List> response = restTemplate.exchange(
                peerAnchorUrl + "/balances/token/" + tokenId,
                HttpMethod.GET,
                null,
                List.class
        );
        return response.getBody();
    }

    public Map<String, Object> settleToken(Map<String, Object> settleData) {
        // Route to peer-supplier for token settlement
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);

        HttpEntity<Map<String, Object>> request = new HttpEntity<>(settleData, headers);

        ResponseEntity<Map> response = restTemplate.exchange(
                peerSupplierUrl + "/token/settle",
                HttpMethod.POST,
                request,
                Map.class
        );
        return response.getBody();
    }
}
