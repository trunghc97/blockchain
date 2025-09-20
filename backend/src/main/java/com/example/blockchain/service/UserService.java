package com.example.blockchain.service;

import com.example.blockchain.model.LoginRequest;
import com.example.blockchain.model.LoginResponse;
import com.example.blockchain.model.User;
import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.SignatureAlgorithm;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.mongodb.core.MongoTemplate;
import org.springframework.data.mongodb.core.query.Criteria;
import org.springframework.data.mongodb.core.query.Query;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.stereotype.Service;

import java.util.Date;
import java.util.List;

@Service
public class UserService {
    private static final Logger logger = LoggerFactory.getLogger(UserService.class);
    private final MongoTemplate mongoTemplate;
    private final BCryptPasswordEncoder passwordEncoder;
    private static final String JWT_SECRET = "blockchain-secret-key";
    private static final long JWT_EXPIRATION = 86400000L; // 1 day

    public UserService(MongoTemplate mongoTemplate) {
        this.mongoTemplate = mongoTemplate;
        this.passwordEncoder = new BCryptPasswordEncoder();
    }

    public LoginResponse login(LoginRequest request) {
        User user = mongoTemplate.findOne(
            Query.query(Criteria.where("username").is(request.getUsername())),
            User.class,
            "users"
        );

        if (user == null || !passwordEncoder.matches(request.getPassword(), user.getPassword())) {
            throw new RuntimeException("Invalid username or password");
        }

        String token = generateToken(user);
        user.setPassword(null); // Don't send password back

        return LoginResponse.builder()
            .token(token)
            .userId(user.getId())
            .username(user.getUsername())
            .role(user.getRole())
            .build();
    }

    private String generateToken(User user) {
        return Jwts.builder()
            .setSubject(user.getUsername())
            .claim("userId", user.getId())
            .claim("role", user.getRole())
            .setIssuedAt(new Date())
            .setExpiration(new Date(System.currentTimeMillis() + JWT_EXPIRATION))
            .signWith(SignatureAlgorithm.HS512, JWT_SECRET)
            .compact();
    }

    public List<User> getUsers() {
        return mongoTemplate.findAll(User.class, "users");
    }

    public List<User> getSuppliers() {
        return mongoTemplate.find(Query.query(Criteria.where("role").is("SUPPLIER")), User.class, "users");
    }

    public User getCurrentUser(String token) {
        try {
            // Remove "Bearer " prefix if present
            if (token.startsWith("Bearer ")) {
                token = token.substring(7);
            }

            Claims claims = Jwts.parser()
                .setSigningKey(JWT_SECRET)
                .parseClaimsJws(token)
                .getBody();

            String username = claims.getSubject();
            return mongoTemplate.findOne(
                Query.query(Criteria.where("username").is(username)),
                User.class,
                "users"
            );
        } catch (Exception e) {
            logger.error("Error parsing token: ", e);
            throw new RuntimeException("Invalid token");
        }
    }

    public User getCurrentUser() {
        String username = SecurityContextHolder.getContext().getAuthentication().getName();
        return mongoTemplate.findOne(
            Query.query(Criteria.where("username").is(username)),
            User.class,
            "users"
        );
    }

    public static Claims parseToken(String token) {
        return Jwts.parser()
            .setSigningKey(JWT_SECRET)
            .parseClaimsJws(token)
            .getBody();
    }
}