package com.example.blockchain.service;

import com.example.blockchain.model.*;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.*;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;
import org.springframework.web.client.RestClientException;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class PeerRoutingService {

    private final RestTemplate restTemplate;

    // Peer node URLs from application.yml
    @Value("${peer.main-bank.url:http://peer-main-bank:8082}")
    private String mainBankPeerUrl;

    @Value("${peer.supplier.url:http://peer-supplier:8083}")
    private String supplierPeerUrl;

    @Value("${peer.anchor.url:http://peer-anchor:8084}")
    private String anchorPeerUrl;

    public PeerRoutingService(RestTemplate restTemplate) {
        this.restTemplate = restTemplate;
    }

    /**
     * Route contract creation to appropriate peer based on business logic
     * Contract creation should go to Anchor peer
     */
    public Map<String, Object> createContract(Map<String, Object> contractData) {
        return callPeer(anchorPeerUrl, "/contract/create", HttpMethod.POST, contractData);
    }

    /**
     * Route contract approval to appropriate peer
     * Supplier approvals go to Supplier peer, Bank approvals go to Main Bank peer
     */
    public Map<String, Object> approveContract(String contractId, Map<String, Object> approvalData) {
        // Check if this is a bank approval or supplier approval
        String approverType = determineApproverType(approvalData);

        String peerUrl;
        if ("BANK".equals(approverType)) {
            peerUrl = mainBankPeerUrl;
        } else {
            peerUrl = supplierPeerUrl;
        }

        return callPeer(peerUrl, "/contract/" + contractId + "/approve", HttpMethod.POST, approvalData);
    }

    /**
     * Route bank approval to Main Bank peer
     */
    public Map<String, Object> approveContractByBank(String contractId, Map<String, Object> approvalData) {
        return callPeer(mainBankPeerUrl, "/contract/" + contractId + "/approve-bank", HttpMethod.POST, approvalData);
    }

    /**
     * Route token transfer to Supplier peer (since suppliers handle token circulation)
     */
    public Map<String, Object> transferToken(Map<String, Object> transferData) {
        return callPeer(supplierPeerUrl, "/token/transfer", HttpMethod.POST, transferData);
    }

    /**
     * Route token settlement to Main Bank peer
     */
    public Map<String, Object> settleToken(Map<String, Object> settleData) {
        return callPeer(mainBankPeerUrl, "/token/settle", HttpMethod.POST, settleData);
    }

    /**
     * Route token queries based on context
     */
    public List<Map<String, Object>> getTokensIssuedByBank(String bankId) {
        return callPeerForList(mainBankPeerUrl, "/token/issued/" + bankId, HttpMethod.GET, null);
    }

    /**
     * Route supplier queries to Supplier peer
     */
    public List<Map<String, Object>> getAllSuppliers() {
        return callPeerForList(supplierPeerUrl, "/suppliers", HttpMethod.GET, null);
    }

    /**
     * Route contract queries - can go to any peer since data should be synced
     * Default to Anchor peer for contract operations
     */
    public Map<String, Object> getContract(String contractId) {
        try {
            return callPeer(anchorPeerUrl, "/contract/" + contractId, HttpMethod.GET, null);
        } catch (RestClientException e) {
            // Contract not found or other client errors, return null instead of throwing exception
            return null;
        }
    }

    public List<Map<String, Object>> listContracts() {
        return callPeerForList(anchorPeerUrl, "/contract/list", HttpMethod.GET, null);
    }

    public LedgerResponse queryLedger(String contractId) {
        try {
            Map<String, Object> response = callPeer(anchorPeerUrl, "/contract/" + contractId + "/ledger", HttpMethod.GET, null);
            LedgerResponse ledgerResponse = new LedgerResponse();
            ledgerResponse.setData(response);
            return ledgerResponse;
        } catch (org.springframework.web.client.HttpClientErrorException.NotFound e) {
            // Contract not found, return empty ledger response
            LedgerResponse ledgerResponse = new LedgerResponse();
            ledgerResponse.setData(new HashMap<>());
            return ledgerResponse;
        }
    }

    /**
     * Route token queries - can go to any peer since data should be synced
     * Default to Supplier peer for token operations
     */
    public Map<String, Object> getToken(String tokenId) {
        try {
            return callPeer(supplierPeerUrl, "/token/" + tokenId, HttpMethod.GET, null);
        } catch (RestClientException e) {
            // Token not found or other client errors, return null instead of throwing exception
            return null;
        }
    }

    public List<Map<String, Object>> getAllTokens() {
        return callPeerForList(anchorPeerUrl, "/tokens", HttpMethod.GET, null);
    }

    /**
     * Route balance queries to appropriate peer based on account type
     */
    public List<Map<String, Object>> getBalancesByAccount(String accountId) {
        String peerUrl = determinePeerForAccount(accountId);
        return callPeerForList(peerUrl, "/balances/account/" + accountId, HttpMethod.GET, null);
    }

    public List<Map<String, Object>> getBalancesByToken(String tokenId) {
        return callPeerForList(supplierPeerUrl, "/balances/token/" + tokenId, HttpMethod.GET, null);
    }

    /**
     * Route users query - can go to any peer since user data should be consistent
     * Default to Main Bank peer
     */
    public List<Map<String, Object>> getUsers() {
        return callPeerForList(mainBankPeerUrl, "/users", HttpMethod.GET, null);
    }

    // Helper methods

    private String determineApproverType(Map<String, Object> approvalData) {
        // Check if bankId is present (bank approval) or supplierId (supplier approval)
        if (approvalData.containsKey("bankId")) {
            return "BANK";
        } else if (approvalData.containsKey("supplierId")) {
            return "SUPPLIER";
        }
        return "UNKNOWN";
    }

    private String determinePeerForAccount(String accountId) {
        // Route based on account ID prefix
        if (accountId.startsWith("BANK")) {
            return mainBankPeerUrl;
        } else if (accountId.startsWith("SUP")) {
            return supplierPeerUrl;
        } else if (accountId.startsWith("ANCHOR")) {
            return anchorPeerUrl;
        }
        // Default to supplier peer for token operations
        return supplierPeerUrl;
    }

    private Map<String, Object> callPeer(String peerUrl, String endpoint, HttpMethod method, Map<String, Object> requestData) {
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);

        HttpEntity<Map<String, Object>> request = new HttpEntity<>(requestData, headers);

        ResponseEntity<Map> response = restTemplate.exchange(
                peerUrl + endpoint,
                method,
                request,
                Map.class
        );
        return response.getBody();
    }

    @SuppressWarnings("unchecked")
    private List<Map<String, Object>> callPeerForList(String peerUrl, String endpoint, HttpMethod method, Map<String, Object> requestData) {
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);

        HttpEntity<Map<String, Object>> request = new HttpEntity<>(requestData, headers);

        ResponseEntity<List> response = restTemplate.exchange(
                peerUrl + endpoint,
                method,
                request,
                List.class
        );
        return response.getBody();
    }
}
