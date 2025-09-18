package com.example.blockchain.controller;

import com.example.blockchain.model.*;
import com.example.blockchain.service.TransferService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

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
    public ResponseEntity<WorldState> approveTransfer(@RequestBody ApproveRequest request) {
        return ResponseEntity.ok(transferService.approveTransfer(request));
    }

    @GetMapping("/status/{transactionId}")
    public ResponseEntity<WorldState> getTransferStatus(@PathVariable String transactionId) {
        return ResponseEntity.ok(transferService.getTransferStatus(transactionId));
    }

    @GetMapping("/list")
    public ResponseEntity<List<WorldState>> getTransfers(
            @RequestParam(required = false) String approverId,
            @RequestParam(required = false) String status
    ) {
        return ResponseEntity.ok(transferService.getTransfers(approverId, status));
    }

    @GetMapping("/transactions")
    public ResponseEntity<List<Transaction>> getAllTransactions() {
        return ResponseEntity.ok(transferService.getAllTransactions());
    }

    @GetMapping("/blocks")
    public ResponseEntity<List<Block>> getAllBlocks() {
        return ResponseEntity.ok(transferService.getAllBlocks());
    }

    @GetMapping("/pending-approvals")
    public ResponseEntity<List<WorldState>> getPendingApprovals(@RequestParam String userId) {
        return ResponseEntity.ok(transferService.getPendingApprovals(userId));
    }
}
