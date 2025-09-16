// MongoDB Initialization Script
// Run this script to initialize database with sample data

// Switch to blockchain database
db = db.getSiblingDB('blockchain');

// Create collections
db.createCollection('users');
db.createCollection('transactions');
db.createCollection('blocks');
db.createCollection('worldstate');

// Insert sample users
db.users.insertMany([
  {
    "_id": "user1",
    "name": "Nguyễn Văn A",
    "accountNo": "1234567890"
  },
  {
    "_id": "user2",
    "name": "Trần Thị B",
    "accountNo": "1234567891"
  },
  {
    "_id": "user3",
    "name": "Lê Văn C",
    "accountNo": "1234567892"
  },
  {
    "_id": "user4",
    "name": "Phạm Thị D",
    "accountNo": "1234567893"
  },
  {
    "_id": "user5",
    "name": "Hoàng Văn E",
    "accountNo": "1234567894"
  }
]);


// Create indexes for better performance
db.users.createIndex({ "accountNo": 1 }, { unique: true });
db.transactions.createIndex({ "transaction_id": 1 });
db.transactions.createIndex({ "status": 1 });
db.blocks.createIndex({ "block_number": 1 }, { unique: true });
db.worldstate.createIndex({ "transaction_id": 1 }, { unique: true });
db.worldstate.createIndex({ "status": 1 });

print("Database initialization completed!");
print("Collections created:");
print("- users: " + db.users.countDocuments());
print("- transactions: " + db.transactions.countDocuments());
print("- blocks: " + db.blocks.countDocuments());
print("- worldstate: " + db.worldstate.countDocuments());
