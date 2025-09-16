package com.example.blockchain.controller;

import com.example.blockchain.model.TransferRequest;
import com.example.blockchain.model.WorldState;
import com.example.blockchain.service.TransferService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/transfer")
@RequiredArgsConstructor
public class TransferController {

    private final TransferService transferService;

    @PostMapping("/create")
    public ResponseEntity<WorldState> createTransfer(@RequestBody TransferRequest request) {
        return ResponseEntity.ok(transferService.createTransfer(request));
    }

    @PostMapping("/approve")
    public ResponseEntity<WorldState> approveTransfer(@RequestBody TransferRequest request) {
        return ResponseEntity.ok(transferService.approveTransfer(request));
    }

    @GetMapping("/status/{transactionId}")
    public ResponseEntity<WorldState> getTransferStatus(@PathVariable String transactionId) {
        return ResponseEntity.ok(transferService.getTransferStatus(transactionId));
    }
}
