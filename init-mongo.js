// Switch to admin database to create user
db = db.getSiblingDB('admin');

// Authenticate as root user
db.auth('root', 'example');

// Switch to blockchain database
db = db.getSiblingDB('blockchain');

// Create collections if not exists
db.createCollection('users');
db.createCollection('contracts');  // For Java backend to store contracts
db.createCollection('events');     // For Go service to store blockchain events
db.createCollection('blocks');     // For Go service to store blocks

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
const contractIndexes = db.contracts.getIndexes();
const eventIndexes = db.events.getIndexes();
const blockIndexes = db.blocks.getIndexes();

if (!userIndexes.some(index => index.name === 'username_1')) {
    db.users.createIndex({ "username": 1 }, { unique: true });
    print('Created index on users.username');
}

if (!contractIndexes.some(index => index.name === 'contractId_1')) {
    db.contracts.createIndex({ "contractId": 1 }, { unique: true });
    print('Created index on contracts.contractId');
}

if (!contractIndexes.some(index => index.name === 'buyer_1')) {
    db.contracts.createIndex({ "buyer": 1 });
    print('Created index on contracts.buyer');
}

if (!eventIndexes.some(index => index.name === 'included_1')) {
    db.events.createIndex({ "included": 1 });
    print('Created index on events.included');
}

if (!eventIndexes.some(index => index.name === 'contract_id_1')) {
    db.events.createIndex({ "contract_id": 1 });
    print('Created index on events.contract_id');
}

if (!blockIndexes.some(index => index.name === 'blockNumber_1')) {
    db.blocks.createIndex({ "blockNumber": 1 }, { unique: true });
    print('Created index on blocks.blockNumber');
}