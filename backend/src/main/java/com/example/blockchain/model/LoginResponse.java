package com.example.blockchain.model;

import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class LoginResponse {
    private String userId;
    private String username;
    private String role;
    private String token;
}