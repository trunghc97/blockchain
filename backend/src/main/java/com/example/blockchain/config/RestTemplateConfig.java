package com.example.blockchain.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.client.SimpleClientHttpRequestFactory;
import org.springframework.web.client.RestTemplate;

@Configuration
public class RestTemplateConfig {

    @Bean
    public RestTemplate restTemplate() {
        RestTemplate restTemplate = new RestTemplate();
        // Disable default error handling for 4xx/5xx responses
        restTemplate.setErrorHandler(new org.springframework.web.client.DefaultResponseErrorHandler() {
            @Override
            public boolean hasError(org.springframework.http.client.ClientHttpResponse response) throws java.io.IOException {
                return false; // Never throw exceptions for HTTP errors
            }
        });
        return restTemplate;
    }

    // Custom error handler that doesn't throw exceptions
    private static class NoOpResponseErrorHandler implements org.springframework.web.client.ResponseErrorHandler {
        @Override
        public boolean hasError(org.springframework.http.client.ClientHttpResponse response) {
            return false; // Never consider it an error
        }

        @Override
        public void handleError(org.springframework.http.client.ClientHttpResponse response) {
            // Do nothing
        }
    }
}
