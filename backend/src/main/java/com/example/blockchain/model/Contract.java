package com.example.blockchain.model;

import lombok.Data;
import org.springframework.data.mongodb.core.mapping.Document;

import java.util.Date;
import java.util.List;

@Data
@Document(collection = "contracts")
public class Contract {
    private String id;
    private String contractId;
    private String description;
    private List<SupplierAmount> suppliers;
    private String status;
    private String fileUrl;
    private Date createdAt;
    private Date updatedAt;
}