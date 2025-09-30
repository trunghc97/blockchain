package com.example.blockchain.config;

import com.mongodb.client.MongoClients;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.mongo.MongoProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;
import org.springframework.data.mongodb.MongoDatabaseFactory;
import org.springframework.data.mongodb.core.MongoTemplate;
import org.springframework.data.mongodb.core.SimpleMongoClientDatabaseFactory;

@Configuration
public class MongoConfig {

    @Value("${mongodb.public.host}")
    private String publicHost;

    @Value("${mongodb.public.port}")
    private int publicPort;

    @Value("${mongodb.public.database}")
    private String publicDatabase;

    @Value("${mongodb.public.username}")
    private String publicUsername;

    @Value("${mongodb.public.password}")
    private String publicPassword;

    @Bean
    @Primary
    public MongoTemplate mongoTemplate() {
        String connectionString = String.format("mongodb://%s:%s@%s:%d/%s?authSource=admin",
            publicUsername, publicPassword, publicHost, publicPort, "blockchain_private");
        MongoDatabaseFactory factory = new SimpleMongoClientDatabaseFactory(MongoClients.create(connectionString), "blockchain_private");
        return new MongoTemplate(factory);
    }

    @Bean(name = "publicMongoTemplate")
    public MongoTemplate publicMongoTemplate() {
        String connectionString = String.format("mongodb://%s:%s@%s:%d/%s?authSource=admin",
            publicUsername, publicPassword, publicHost, publicPort, publicDatabase);
        MongoDatabaseFactory factory = new SimpleMongoClientDatabaseFactory(MongoClients.create(connectionString), publicDatabase);
        return new MongoTemplate(factory);
    }
}
