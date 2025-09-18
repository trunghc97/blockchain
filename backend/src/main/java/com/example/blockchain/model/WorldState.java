package com.example.blockchain.model;

import lombok.Data;
import com.fasterxml.jackson.annotation.JsonFormat;
import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.Date;
import java.util.List;

@Data
public class WorldState {
    @JsonProperty("id")
    private String id;
    
    @JsonProperty("transaction_id")
    private String transactionId;
    
    @JsonProperty("from_account")
    private String fromAccount;
    
    @JsonProperty("to_account")
    private String toAccount;
    
    private double amount;
    private String status;
    private List<Approver> approvers;
    
    @JsonProperty("approval_count")
    private int approvalCount;
    
    @JsonProperty("supplier_ref")
    private String supplierRef;
    
    @JsonProperty("last_updated")
    @JsonFormat(pattern = "yyyy-MM-dd'T'HH:mm:ss.SSS'Z'")
    private Date lastUpdated;
}
