// Switch to admin database to create user
db = db.getSiblingDB('admin');

// Authenticate as root user
db.auth('root', 'example');

// Switch to blockchain database
db = db.getSiblingDB('blockchain');

// Create collections if not exists
db.createCollection('users');
db.createCollection('transactions');
db.createCollection('blocks');
db.createCollection('world_state');

// Only insert users if collection is empty
if (db.users.countDocuments() === 0) {
    // Insert anchor user
    db.users.insertOne({
        username: 'anchor',
        password: '$2a$10$mUtYfT1CpUHZhDL7hgQ4W.Z400PK78v1vsOxtWV6fCqmAIgqYMfJK', // 123456
        role: 'ANCHOR'
    });

    // Insert supplier users
    const suppliers = Array.from({length: 10}, (_, i) => ({
        username: `supplier${i + 1}`,
        password: '$2a$10$mUtYfT1CpUHZhDL7hgQ4W.Z400PK78v1vsOxtWV6fCqmAIgqYMfJK', // 123456
        role: 'SUPPLIER'
    }));

    db.users.insertMany(suppliers);

    print('Initial users created successfully');
} else {
    print('Users collection already has data, skipping initialization');
}

// Create indexes if not exists
const userIndexes = db.users.getIndexes();
const transactionIndexes = db.transactions.getIndexes();
const blockIndexes = db.blocks.getIndexes();
const worldStateIndexes = db.world_state.getIndexes();

if (!userIndexes.some(index => index.name === 'username_1')) {
    db.users.createIndex({ "username": 1 }, { unique: true });
    print('Created index on users.username');
}

if (!transactionIndexes.some(index => index.name === 'transactionId_1')) {
    db.transactions.createIndex({ "transactionId": 1 }, { unique: true });
    print('Created index on transactions.transactionId');
}

if (!blockIndexes.some(index => index.name === 'blockNumber_1')) {
    db.blocks.createIndex({ "blockNumber": 1 }, { unique: true });
    print('Created index on blocks.blockNumber');
}

if (!worldStateIndexes.some(index => index.name === 'transactionId_1')) {
    db.world_state.createIndex({ "transactionId": 1 }, { unique: true });
    print('Created index on world_state.transactionId');
}