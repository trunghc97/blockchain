package com.example.blockchain.service;

import com.example.blockchain.model.*;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpMethod;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.util.List;

@Service
public class BlockchainService {
    private final RestTemplate restTemplate;
    private final String blockchainUrl;

    public BlockchainService(
            RestTemplate restTemplate,
            @Value("${blockchain.url}") String blockchainUrl
    ) {
        this.restTemplate = restTemplate;
        this.blockchainUrl = blockchainUrl;
    }

    public Contract createContract(Contract contract) {
        return restTemplate.postForObject(
                blockchainUrl + "/contract/create",
                contract,
                Contract.class
        );
    }

    public void approveContract(String contractId, String approverId) {
        restTemplate.postForObject(
                blockchainUrl + "/contract/approve",
                new ApprovalRequest(contractId, approverId),
                Void.class
        );
    }

    public List<Contract> listContracts() {
        ResponseEntity<List<Contract>> response = restTemplate.exchange(
                blockchainUrl + "/contract/list",
                HttpMethod.GET,
                null,
                new ParameterizedTypeReference<List<Contract>>() {}
        );
        return response.getBody();
    }

    public LedgerResponse queryLedger(String contractId) {
        return restTemplate.getForObject(
                blockchainUrl + "/ledger/query?contract_id=" + contractId,
                LedgerResponse.class
        );
    }

    public List<User> getUsers() {
        ResponseEntity<List<User>> response = restTemplate.exchange(
                blockchainUrl + "/users",
                HttpMethod.GET,
                null,
                new ParameterizedTypeReference<List<User>>() {}
        );
        return response.getBody();
    }

    private static class ApprovalRequest {
        private String contractId;
        private String approverId;

        public ApprovalRequest(String contractId, String approverId) {
            this.contractId = contractId;
            this.approverId = approverId;
        }

        public String getContractId() {
            return contractId;
        }

        public String getApproverId() {
            return approverId;
        }
    }
}
