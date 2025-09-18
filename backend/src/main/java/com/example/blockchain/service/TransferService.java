package com.example.blockchain.service;

import com.example.blockchain.model.*;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.data.mongodb.core.MongoTemplate;
import org.springframework.data.mongodb.core.query.Query;
import org.springframework.data.mongodb.core.query.Criteria;
import org.springframework.http.HttpMethod;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.*;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class TransferService {
    private static final Logger log = LoggerFactory.getLogger(TransferService.class);

    private final RestTemplate restTemplate;
    private final MongoTemplate mongoTemplate;

    @Value("${blockchain.service.url}")
    private String blockchainServiceUrl;

    public WorldState createTransfer(TransferRequest request) {
        // Tạo transaction ID nếu chưa có
        if (request.getTransactionId() == null) {
            request.setTransactionId(UUID.randomUUID().toString());
        }

        // Map request sang format của blockchain service
        Map<String, Object> blockchainRequest = new HashMap<>();
        blockchainRequest.put("transaction_id", request.getTransactionId());
        blockchainRequest.put("from_account", request.getFromAccount());
        blockchainRequest.put("to_account", request.getToAccount());
        blockchainRequest.put("amount", request.getAmount());
        blockchainRequest.put("approvers", request.getApprovers());

        try {
            log.debug("Sending request to blockchain service: {}", blockchainRequest);

            // Gọi blockchain service
            WorldState response = restTemplate.postForObject(
                blockchainServiceUrl + "/tx/create",
                blockchainRequest,
                WorldState.class
            );

            log.debug("Received response from blockchain service: {}", response);

            // Kiểm tra và log nếu có lỗi
            if (response == null || response.getTransactionId() == null) {
                log.error("Invalid response from blockchain service: {}", response);
                throw new RuntimeException("Failed to create transfer: Invalid response from blockchain service");
            }

            return response;
        } catch (Exception e) {
            log.error("Error creating transfer", e);
            throw new RuntimeException("Failed to create transfer: " + e.getMessage());
        }
    }

    public WorldState approveTransfer(ApproveRequest request) {
        // Map request sang format của blockchain service
        Map<String, Object> blockchainRequest = new HashMap<>();
        blockchainRequest.put("transaction_id", request.getTransactionId());
        blockchainRequest.put("approver_id", request.getApproverUserId());
        blockchainRequest.put("from_account", request.getFromAccount());
        blockchainRequest.put("to_account", request.getToAccount());
        blockchainRequest.put("amount", request.getAmount());

        return restTemplate.postForObject(
            blockchainServiceUrl + "/tx/approve",
            blockchainRequest,
            WorldState.class
        );
    }

    public WorldState getTransferStatus(String transactionId) {
        return restTemplate.getForObject(
            blockchainServiceUrl + "/tx/status/" + transactionId,
            WorldState.class
        );
    }

    public List<WorldState> getTransfers(String approverId, String status) {
        if (approverId == null) {
            return Collections.emptyList();
        }

        // Gọi API blockchain service để lấy danh sách giao dịch
        String url = String.format("%s/tx/pending-approvals?user_id=%s", blockchainServiceUrl, approverId);

        ResponseEntity<List<WorldState>> response = restTemplate.exchange(
            url,
            HttpMethod.GET,
            null,
            new ParameterizedTypeReference<List<WorldState>>() {}
        );

        List<WorldState> transfers = response.getBody();
        if (transfers == null) {
            return Collections.emptyList();
        }

        // Lọc theo status nếu có
        if (status != null) {
            String[] statuses = status.split(",");
            Set<String> statusSet = new HashSet<>(Arrays.asList(statuses));
            transfers = transfers.stream()
                .filter(tx -> statusSet.contains(tx.getStatus()))
                .collect(Collectors.toList());
        }

        return transfers;
    }

    public List<Transaction> getAllTransactions() {
        return mongoTemplate.findAll(Transaction.class);
    }

    public List<Block> getAllBlocks() {
        return mongoTemplate.findAll(Block.class);
    }

    public List<WorldState> getPendingApprovals(String userId) {
        ResponseEntity<List<WorldState>> response = restTemplate.exchange(
            blockchainServiceUrl + "/tx/pending-approvals?user_id=" + userId,
            HttpMethod.GET,
            null,
            new ParameterizedTypeReference<List<WorldState>>() {}
        );
        return response.getBody();
    }
}
