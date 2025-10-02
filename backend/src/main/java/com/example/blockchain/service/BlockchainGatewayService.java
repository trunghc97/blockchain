package com.example.blockchain.service;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.*;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

@Service
public class BlockchainGatewayService {

    @Autowired
    private RestTemplate restTemplate;

    @Value("${blockchain.gateway.url:http://blockchain-gw:9090}")
    private String blockchainGatewayUrl;

    public String submitTransaction(String functionName, String... args) {
        String transactionId = UUID.randomUUID().toString();
        
        Map<String, Object> requestBody = new HashMap<>();
        requestBody.put("transaction_id", transactionId);
        requestBody.put("function_name", functionName);
        requestBody.put("args", args);
        requestBody.put("channel_id", "supply-chain-channel");
        requestBody.put("chaincode_name", "scf-chaincode");

        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);

        HttpEntity<Map<String, Object>> request = new HttpEntity<>(requestBody, headers);

        try {
            ResponseEntity<Map> response = restTemplate.postForEntity(
                blockchainGatewayUrl + "/submit-transaction", 
                request, 
                Map.class
            );

            if (response.getStatusCode() == HttpStatus.OK) {
                Map<String, Object> responseBody = response.getBody();
                Boolean success = (Boolean) responseBody.get("success");
                if (success != null && success) {
                    return (String) responseBody.get("block_number");
                } else {
                    throw new RuntimeException("Transaction failed: " + responseBody.get("message"));
                }
            } else {
                throw new RuntimeException("HTTP error: " + response.getStatusCode());
            }
        } catch (Exception e) {
            throw new RuntimeException("Failed to submit transaction to blockchain gateway", e);
        }
    }

    public String createContract(String anchorId, String supplierId, String amount) {
        return submitTransaction("CreateContract", anchorId, supplierId, amount);
    }

    public String approveContract(String contractId, String approverId) {
        return submitTransaction("ApproveContract", contractId, approverId);
    }

    public String finalizeContract(String contractId) {
        return submitTransaction("FinalizeContract", contractId);
    }

    public String issueToken(String contractId, String bankId, String amount) {
        return submitTransaction("IssueToken", contractId, bankId, amount);
    }

    public String transferToken(String tokenId, String fromAccount, String toAccount) {
        return submitTransaction("TransferToken", tokenId, fromAccount, toAccount);
    }

    public String settleToken(String tokenId) {
        return submitTransaction("SettleToken", tokenId);
    }

    public String getTransactionStatus(String transactionId) {
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);

        Map<String, String> requestBody = new HashMap<>();
        requestBody.put("transaction_id", transactionId);

        HttpEntity<Map<String, String>> request = new HttpEntity<>(requestBody, headers);

        try {
            ResponseEntity<Map> response = restTemplate.postForEntity(
                blockchainGatewayUrl + "/transaction-status", 
                request, 
                Map.class
            );

            if (response.getStatusCode() == HttpStatus.OK) {
                Map<String, Object> responseBody = response.getBody();
                return (String) responseBody.get("status");
            } else {
                throw new RuntimeException("HTTP error: " + response.getStatusCode());
            }
        } catch (Exception e) {
            throw new RuntimeException("Failed to get transaction status", e);
        }
    }
}

