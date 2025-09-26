// MongoDB initialization script for Main Bank Peer
// This script sets up the initial database state for the Main Bank peer node

db = db.getSiblingDB('blockchain_main_bank');

// Create collections
db.createCollection('contracts');
db.createCollection('tokens');
db.createCollection('balances');
db.createCollection('events');
db.createCollection('blocks');
db.createCollection('users');

// Create indexes for performance
db.contracts.createIndex({ "_id": 1 });
db.contracts.createIndex({ "bankId": 1 });
db.contracts.createIndex({ "approved": 1 });
db.contracts.createIndex({ "bankApproved": 1 });

db.tokens.createIndex({ "_id": 1 });
db.tokens.createIndex({ "contractId": 1 });
db.tokens.createIndex({ "issuer": 1 });
db.tokens.createIndex({ "owner": 1 });

db.balances.createIndex({ "tokenId": 1, "account": 1 }, { unique: true });
db.balances.createIndex({ "account": 1 });

db.events.createIndex({ "contractId": 1 });
db.events.createIndex({ "tokenId": 1 });
db.events.createIndex({ "eventType": 1 });
db.events.createIndex({ "timestamp": 1 });

db.blocks.createIndex({ "blockNumber": 1 }, { unique: true });
db.blocks.createIndex({ "timestamp": 1 });

db.users.createIndex({ "id": 1 }, { unique: true });
db.users.createIndex({ "username": 1 }, { unique: true });
db.users.createIndex({ "role": 1 });

// Insert initial users for Main Bank
db.users.insertMany([
  {
    id: "BANK001",
    username: "mainbank_admin",
    password: "$2a$10$hashed_password_here", // In production, use proper hashing
    role: "BANK",
    name: "Main Bank Administrator",
    createdAt: new Date()
  },
  {
    id: "BANK002",
    username: "mainbank_user",
    password: "$2a$10$hashed_password_here",
    role: "BANK",
    name: "Main Bank User",
    createdAt: new Date()
  }
]);

// Create genesis block for Main Bank peer
db.blocks.insertOne({
  blockNumber: 1,
  timestamp: new Date(),
  events: [],
  previousHash: "genesis",
  hash: "genesis_hash_main_bank",
  merkleRoot: "",
  peerId: "main-bank-peer-1"
});

print("Main Bank Peer database initialized successfully");
