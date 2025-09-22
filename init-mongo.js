// Switch to admin database to create user
db = db.getSiblingDB('admin');

// Authenticate as root user
db.auth('root', 'example');

// Switch to blockchain database
db = db.getSiblingDB('blockchain');

// Create collections for Contract-Token Management System
print('Creating collections for Contract-Token Management System...');

// Core collections
db.createCollection('users');        // User authentication
db.createCollection('contracts');    // Contract data (world state)
db.createCollection('tokens');       // Token data (world state)
db.createCollection('balances');     // Token balances (world state)

// Legacy collections (keeping for compatibility)
db.createCollection('events');       // Blockchain events
db.createCollection('blocks');       // Block data

print('Collections created successfully');

// Only insert users if collection is empty
if (db.users.countDocuments() === 0) {
    print('Creating initial users for Contract-Token system...');

    // Insert bank user (token issuer)
    db.users.insertOne({
        id: 'BANK001',
        username: 'bank',
        password: '$2a$10$mUtYfT1CpUHZhDL7hgQ4W.Z400PK78v1vsOxtWV6fCqmAIgqYMfJK', // 123456
        role: 'BANK'
    });

    // Insert anchor user (contract creator)
    db.users.insertOne({
        id: 'ANCHOR001',
        username: 'anchor',
        password: '$2a$10$mUtYfT1CpUHZhDL7hgQ4W.Z400PK78v1vsOxtWV6fCqmAIgqYMfJK', // 123456
        role: 'ANCHOR'
    });

    // Insert supplier users (token receivers)
    const suppliers = [
        { id: 'SUPPLIER001', username: 'supplier1' },
        { id: 'SUPPLIER002', username: 'supplier2' },
        { id: 'SUPPLIER003', username: 'supplier3' },
        { id: 'SUPPLIER004', username: 'supplier4' },
        { id: 'SUPPLIER005', username: 'supplier5' }
    ];

    suppliers.forEach(supplier => {
        db.users.insertOne({
            id: supplier.id,
            username: supplier.username,
            password: '$2a$10$mUtYfT1CpUHZhDL7hgQ4W.Z400PK78v1vsOxtWV6fCqmAIgqYMfJK', // 123456
            role: 'SUPPLIER'
        });
    });

    print('Initial users created successfully');
    print('Available users:');
    print('- Bank: bank / 123456 (BANK001)');
    print('- Anchor: anchor / 123456 (ANCHOR001)');
    print('- Suppliers: supplier1-supplier5 / 123456 (SUPPLIER001-SUPPLIER005)');
} else {
    print('Users collection already has data, skipping user initialization');
}

// Create indexes for optimal performance
print('Creating database indexes...');

// Users collection indexes
const userIndexes = db.users.getIndexes();
if (!userIndexes.some(index => index.name === 'username_1')) {
    db.users.createIndex({ "username": 1 }, { unique: true });
    print('Created unique index on users.username');
}

if (!userIndexes.some(index => index.name === 'id_1')) {
    db.users.createIndex({ "id": 1 }, { unique: true });
    print('Created unique index on users.id');
}

if (!userIndexes.some(index => index.name === 'role_1')) {
    db.users.createIndex({ "role": 1 });
    print('Created index on users.role');
}

// Contracts collection indexes
const contractIndexes = db.contracts.getIndexes();
if (!contractIndexes.some(index => index.name === '_id_')) {
    db.contracts.createIndex({ "_id": 1 }, { unique: true });
    print('Created unique index on contracts._id');
}

if (!contractIndexes.some(index => index.name === 'anchorId_1')) {
    db.contracts.createIndex({ "anchorId": 1 });
    print('Created index on contracts.anchorId');
}

if (!contractIndexes.some(index => index.name === 'supplierId_1')) {
    db.contracts.createIndex({ "supplierId": 1 });
    print('Created index on contracts.supplierId');
}

if (!contractIndexes.some(index => index.name === 'bankId_1')) {
    db.contracts.createIndex({ "bankId": 1 });
    print('Created index on contracts.bankId');
}

if (!contractIndexes.some(index => index.name === 'approved_1')) {
    db.contracts.createIndex({ "approved": 1 });
    print('Created index on contracts.approved');
}

// Tokens collection indexes
const tokenIndexes = db.tokens.getIndexes();
if (!tokenIndexes.some(index => index.name === '_id_')) {
    db.tokens.createIndex({ "_id": 1 }, { unique: true });
    print('Created unique index on tokens._id');
}

if (!tokenIndexes.some(index => index.name === 'contractId_1')) {
    db.tokens.createIndex({ "contractId": 1 });
    print('Created index on tokens.contractId');
}

if (!tokenIndexes.some(index => index.name === 'issuer_1')) {
    db.tokens.createIndex({ "issuer": 1 });
    print('Created index on tokens.issuer');
}

if (!tokenIndexes.some(index => index.name === 'owner_1')) {
    db.tokens.createIndex({ "owner": 1 });
    print('Created index on tokens.owner');
}

// Balances collection indexes (critical for performance)
const balanceIndexes = db.balances.getIndexes();
if (!balanceIndexes.some(index => index.name === 'tokenId_1_account_1')) {
    db.balances.createIndex({ "tokenId": 1, "account": 1 }, { unique: true });
    print('Created compound unique index on balances(tokenId, account)');
}

if (!balanceIndexes.some(index => index.name === 'balance_1')) {
    db.balances.createIndex({ "balance": 1 });
    print('Created index on balances.balance');
}

// Legacy collections indexes (keeping for compatibility)
const eventIndexes = db.events.getIndexes();
if (!eventIndexes.some(index => index.name === 'included_1')) {
    db.events.createIndex({ "included": 1 });
    print('Created index on events.included');
}

if (!eventIndexes.some(index => index.name === 'contractId_1')) {
    db.events.createIndex({ "contractId": 1 });
    print('Created index on events.contractId');
}

const blockIndexes = db.blocks.getIndexes();
if (!blockIndexes.some(index => index.name === 'blockNumber_1')) {
    db.blocks.createIndex({ "blockNumber": 1 }, { unique: true });
    print('Created unique index on blocks.blockNumber');
}

print('Database indexes created successfully');

// Final summary
print('\n=== BLOCKCHAIN CONTRACT-TOKEN SYSTEM INITIALIZATION COMPLETE ===');
print('Collections created:');
print('- users: User authentication data');
print('- contracts: Contract world state');
print('- tokens: Token world state');
print('- balances: Token balance tracking');
print('- events: Legacy blockchain events');
print('- blocks: Legacy block data');

print('\nAvailable test users:');
print('BANK:     bank / 123456 (ID: BANK001)');
print('ANCHOR:   anchor / 123456 (ID: ANCHOR001)');
print('SUPPLIERS: supplier1-supplier5 / 123456 (IDs: SUPPLIER001-SUPPLIER005)');

print('\nSystem is ready for testing!');
print('Use docker-compose up --build to start all services.');
print('Access frontend at: http://localhost:4200');
print('===============================================================\n');