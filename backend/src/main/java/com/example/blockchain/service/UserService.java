package com.example.blockchain.service;

import com.example.blockchain.model.User;
import lombok.RequiredArgsConstructor;
import org.springframework.data.mongodb.core.MongoTemplate;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
@RequiredArgsConstructor
public class UserService {

    private final MongoTemplate mongoTemplate;

    public List<User> getAllUsers() {
        return mongoTemplate.findAll(User.class);
    }
}
