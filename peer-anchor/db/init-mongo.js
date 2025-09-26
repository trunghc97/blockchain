// MongoDB initialization script for Anchor Peer
// This script sets up the initial database state for the Anchor peer node

db = db.getSiblingDB('blockchain_anchor');

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

// Insert initial users for Anchor
db.users.insertMany([
  {
    id: "ANCHOR001",
    username: "anchor_admin",
    password: "$2a$10$hashed_password_here",
    role: "ANCHOR",
    name: "Anchor Company Administrator",
    createdAt: new Date()
  },
  {
    id: "ANCHOR002",
    username: "anchor_user",
    password: "$2a$10$hashed_password_here",
    role: "ANCHOR",
    name: "Anchor Company User",
    createdAt: new Date()
  }
]);

// Create genesis block for Anchor peer
db.blocks.insertOne({
  blockNumber: 1,
  timestamp: new Date(),
  events: [],
  previousHash: "genesis",
  hash: "genesis_hash_anchor",
  merkleRoot: "",
  peerId: "anchor-peer-1"
});

print("Anchor Peer database initialized successfully");
