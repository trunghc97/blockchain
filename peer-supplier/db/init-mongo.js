// MongoDB initialization script for Supplier Peer
// This script sets up the initial database state for the Supplier peer node

db = db.getSiblingDB('blockchain_supplier');

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

// Insert initial users for Suppliers
db.users.insertMany([
  {
    id: "SUP001",
    username: "supplier1",
    password: "$2a$10$hashed_password_here",
    role: "SUPPLIER",
    name: "ABC Corporation",
    createdAt: new Date()
  },
  {
    id: "SUP002",
    username: "supplier2",
    password: "$2a$10$hashed_password_here",
    role: "SUPPLIER",
    name: "XYZ Ltd",
    createdAt: new Date()
  },
  {
    id: "SUP003",
    username: "supplier3",
    password: "$2a$10$hashed_password_here",
    role: "SUPPLIER",
    name: "Tech Solutions Inc",
    createdAt: new Date()
  }
]);

// Create genesis block for Supplier peer
db.blocks.insertOne({
  blockNumber: 1,
  timestamp: new Date(),
  events: [],
  previousHash: "genesis",
  hash: "genesis_hash_supplier",
  merkleRoot: "",
  peerId: "supplier-peer-1"
});

print("Supplier Peer database initialized successfully");
