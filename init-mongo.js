// Initialize MongoDB databases and collections for Blockchain system
// This script runs only when database is first created or when manually executed

print('=== BLOCKCHAIN DATABASE INITIALIZATION SCRIPT ===');
print('This script should run automatically when MongoDB container starts with empty data directory');
print('If you see this message, the script is running manually');
print('');

// Switch to admin database
// In init script context, 'db' may not be available, so we need to handle both cases
if (typeof db !== 'undefined') {
    var adminDb = db.getSiblingDB('admin');

    // Authenticate as root user (only needed for manual execution)
    try {
        adminDb.auth('root', 'example');
        print('✓ Authenticated as root user');
    } catch (e) {
        print('Note: Authentication may not be required in init script context');
    }
} else {
    print('Running in init script context - authentication handled by MongoDB');
}

// Function to initialize database with specific collections
function initializeDatabase(dbName, collections) {
    print('\n=== Initializing database: ' + dbName + ' ===');

    // Switch to target database
    var targetDb;
    if (typeof db !== 'undefined') {
        targetDb = db.getSiblingDB(dbName);
    } else {
        // In init script context, use the global connection
        targetDb = connect(dbName);
    }

    // Get existing collections
    var existingCollections = targetDb.getCollectionNames();

    // Create specific collections for this database (only if they don't exist)
    print('Checking/creating collections for ' + dbName + '...');

    collections.forEach(function(collection) {
        if (existingCollections.includes(collection)) {
            print('Collection already exists: ' + collection);
        } else {
            targetDb.createCollection(collection);
            print('Created collection: ' + collection);
        }
    });

    print('Collections check/create completed for ' + dbName);
}

// Initialize databases with specific collections
initializeDatabase('blockchain_private', [
    'contracts',    // Contract data (anchor, mainbank, supplier)
    'tokens',       // Token data (supplier, anchor for monitoring)
    'balances',     // Token balances (all peers)
    'events',       // Private blockchain events (all peers)
    'blocks'        // Private blockchain blocks (all peers)
]); // Private blockchain database for all peers

initializeDatabase('blockchain_public', [
    'users',        // User authentication (public access)
    'events',       // Blockchain events (public)
    'blocks'        // Block data (public)
]); // Public blockchain database for users and public data

// Function to create users for a specific database
function createUsersForDatabase(dbName, userTypes) {
    print('\n--- Creating users for database: ' + dbName + ' ---');

    var targetDb;
    if (typeof db !== 'undefined') {
        targetDb = db.getSiblingDB(dbName);
    } else {
        targetDb = connect(dbName);
    }

    // Only create users if collection exists
    if (targetDb.getCollectionNames().includes('users')) {
        print('Checking/creating users for ' + dbName + '...');

        if (userTypes.includes('BANK')) {
            // Check if bank user exists
            var bankExists = targetDb.users.countDocuments({ id: 'BANK001' }) > 0;
            if (bankExists) {
                print('BANK user already exists');
            } else {
                // Insert bank user (token issuer)
                targetDb.users.insertOne({
                    id: 'BANK001',
                    username: 'bank',
                    password: '$2a$10$mUtYfT1CpUHZhDL7hgQ4W.Z400PK78v1vsOxtWV6fCqmAIgqYMfJK', // 123456
                    role: 'BANK'
                });
                print('Created BANK user');
            }
        }

        if (userTypes.includes('ANCHOR')) {
            // Check if anchor user exists
            var anchorExists = targetDb.users.countDocuments({ id: 'ANCHOR001' }) > 0;
            if (anchorExists) {
                print('ANCHOR user already exists');
            } else {
                // Insert anchor user (contract creator)
                targetDb.users.insertOne({
                    id: 'ANCHOR001',
                    username: 'anchor',
                    password: '$2a$10$mUtYfT1CpUHZhDL7hgQ4W.Z400PK78v1vsOxtWV6fCqmAIgqYMfJK', // 123456
                    role: 'ANCHOR'
                });
                print('Created ANCHOR user');
            }
        }

        if (userTypes.includes('SUPPLIERS')) {
            // Insert supplier users (check each one individually)
            const suppliers = [
                { id: 'SUPPLIER001', username: 'supplier1' },
                { id: 'SUPPLIER002', username: 'supplier2' },
                { id: 'SUPPLIER003', username: 'supplier3' },
                { id: 'SUPPLIER004', username: 'supplier4' },
                { id: 'SUPPLIER005', username: 'supplier5' }
            ];

            var createdCount = 0;
            suppliers.forEach(supplier => {
                var exists = targetDb.users.countDocuments({ id: supplier.id }) > 0;
                if (!exists) {
                    targetDb.users.insertOne({
                        id: supplier.id,
                        username: supplier.username,
                        password: '$2a$10$mUtYfT1CpUHZhDL7hgQ4W.Z400PK78v1vsOxtWV6fCqmAIgqYMfJK', // 123456
                        role: 'SUPPLIER'
                    });
                    createdCount++;
                }
            });

            if (createdCount > 0) {
                print('Created ' + createdCount + ' SUPPLIER users');
            } else {
                print('All SUPPLIER users already exist');
            }
        }

        print('Users check/create completed for ' + dbName);
    } else {
        print('Users collection not available in ' + dbName + ', skipping user initialization');
    }
}

// Create users for specific databases
createUsersForDatabase('blockchain_public', ['BANK', 'ANCHOR', 'SUPPLIERS']); // Public database - all users for public access
// Private database doesn't need users (internal peer operations)

print('\nAvailable users in all databases:');
print('- Bank: bank / 123456 (BANK001)');
print('- Anchor: anchor / 123456 (ANCHOR001)');
print('- Suppliers: supplier1-supplier5 / 123456 (SUPPLIER001-SUPPLIER005)');

// Function to create indexes for a database based on existing collections
function createIndexesForDatabase(dbName) {
    print('\n--- Creating indexes for database: ' + dbName + ' ---');

    var targetDb;
    if (typeof db !== 'undefined') {
        targetDb = db.getSiblingDB(dbName);
    } else {
        targetDb = connect(dbName);
    }
    var collections = targetDb.getCollectionNames();

    // Users collection indexes (only if collection exists)
    if (collections.includes('users')) {
        const userIndexes = targetDb.users.getIndexes();
        if (!userIndexes.some(index => index.name === 'username_1')) {
            targetDb.users.createIndex({ "username": 1 }, { unique: true });
            print('Created unique index on ' + dbName + '.users.username');
        } else {
            print('Index already exists: ' + dbName + '.users.username');
        }
        if (!userIndexes.some(index => index.name === 'id_1')) {
            targetDb.users.createIndex({ "id": 1 }, { unique: true });
            print('Created unique index on ' + dbName + '.users.id');
        } else {
            print('Index already exists: ' + dbName + '.users.id');
        }
        if (!userIndexes.some(index => index.name === 'role_1')) {
            targetDb.users.createIndex({ "role": 1 });
            print('Created index on ' + dbName + '.users.role');
        } else {
            print('Index already exists: ' + dbName + '.users.role');
        }
    }

    // Contracts collection indexes (only if collection exists)
    if (collections.includes('contracts')) {
        const contractIndexes = targetDb.contracts.getIndexes();
        if (!contractIndexes.some(index => index.name === '_id_')) {
            targetDb.contracts.createIndex({ "_id": 1 }, { unique: true });
            print('Created unique index on ' + dbName + '.contracts._id');
        } else {
            print('Index already exists: ' + dbName + '.contracts._id');
        }
        if (!contractIndexes.some(index => index.name === 'anchorId_1')) {
            targetDb.contracts.createIndex({ "anchorId": 1 });
            print('Created index on ' + dbName + '.contracts.anchorId');
        } else {
            print('Index already exists: ' + dbName + '.contracts.anchorId');
        }
        if (!contractIndexes.some(index => index.name === 'supplierId_1')) {
            targetDb.contracts.createIndex({ "supplierId": 1 });
            print('Created index on ' + dbName + '.contracts.supplierId');
        } else {
            print('Index already exists: ' + dbName + '.contracts.supplierId');
        }
        if (!contractIndexes.some(index => index.name === 'bankId_1')) {
            targetDb.contracts.createIndex({ "bankId": 1 });
            print('Created index on ' + dbName + '.contracts.bankId');
        } else {
            print('Index already exists: ' + dbName + '.contracts.bankId');
        }
        if (!contractIndexes.some(index => index.name === 'approved_1')) {
            targetDb.contracts.createIndex({ "approved": 1 });
            print('Created index on ' + dbName + '.contracts.approved');
        } else {
            print('Index already exists: ' + dbName + '.contracts.approved');
        }
        if (!contractIndexes.some(index => index.name === 'bankApproved_1')) {
            targetDb.contracts.createIndex({ "bankApproved": 1 });
            print('Created index on ' + dbName + '.contracts.bankApproved');
        } else {
            print('Index already exists: ' + dbName + '.contracts.bankApproved');
        }
        if (!contractIndexes.some(index => index.name === 'status_1')) {
            targetDb.contracts.createIndex({ "status": 1 });
            print('Created index on ' + dbName + '.contracts.status');
        } else {
            print('Index already exists: ' + dbName + '.contracts.status');
        }
    }

    // Tokens collection indexes (only if collection exists)
    if (collections.includes('tokens')) {
        const tokenIndexes = targetDb.tokens.getIndexes();
        if (!tokenIndexes.some(index => index.name === '_id_')) {
            targetDb.tokens.createIndex({ "_id": 1 }, { unique: true });
            print('Created unique index on ' + dbName + '.tokens._id');
        } else {
            print('Index already exists: ' + dbName + '.tokens._id');
        }
        if (!tokenIndexes.some(index => index.name === 'contractId_1')) {
            targetDb.tokens.createIndex({ "contractId": 1 });
            print('Created index on ' + dbName + '.tokens.contractId');
        } else {
            print('Index already exists: ' + dbName + '.tokens.contractId');
        }
        if (!tokenIndexes.some(index => index.name === 'issuer_1')) {
            targetDb.tokens.createIndex({ "issuer": 1 });
            print('Created index on ' + dbName + '.tokens.issuer');
        } else {
            print('Index already exists: ' + dbName + '.tokens.issuer');
        }
        if (!tokenIndexes.some(index => index.name === 'owner_1')) {
            targetDb.tokens.createIndex({ "owner": 1 });
            print('Created index on ' + dbName + '.tokens.owner');
        } else {
            print('Index already exists: ' + dbName + '.tokens.owner');
        }
    }

    // Balances collection indexes (only if collection exists)
    if (collections.includes('balances')) {
        const balanceIndexes = targetDb.balances.getIndexes();
        if (!balanceIndexes.some(index => index.name === 'tokenId_1_account_1')) {
            targetDb.balances.createIndex({ "tokenId": 1, "account": 1 }, { unique: true });
            print('Created compound unique index on ' + dbName + '.balances(tokenId, account)');
        } else {
            print('Index already exists: ' + dbName + '.balances(tokenId, account)');
        }
        if (!balanceIndexes.some(index => index.name === 'balance_1')) {
            targetDb.balances.createIndex({ "balance": 1 });
            print('Created index on ' + dbName + '.balances.balance');
        } else {
            print('Index already exists: ' + dbName + '.balances.balance');
        }
    }

    // Events collection indexes (only if collection exists)
    if (collections.includes('events')) {
        const eventIndexes = targetDb.events.getIndexes();
        if (!eventIndexes.some(index => index.name === 'included_1')) {
            targetDb.events.createIndex({ "included": 1 });
            print('Created index on ' + dbName + '.events.included');
        } else {
            print('Index already exists: ' + dbName + '.events.included');
        }
        if (!eventIndexes.some(index => index.name === 'contractId_1')) {
            targetDb.events.createIndex({ "contractId": 1 });
            print('Created index on ' + dbName + '.events.contractId');
        } else {
            print('Index already exists: ' + dbName + '.events.contractId');
        }
    }

    // Blocks collection indexes (only if collection exists)
    if (collections.includes('blocks')) {
        const blockIndexes = targetDb.blocks.getIndexes();
        if (!blockIndexes.some(index => index.name === 'height_1')) {
            targetDb.blocks.createIndex({ "height": 1 }, { unique: true });
            print('Created unique index on ' + dbName + '.blocks.height');
        } else {
            print('Index already exists: ' + dbName + '.blocks.height');
        }
    }


    print('Indexes created successfully for ' + dbName);
}

// Create indexes for all databases
createIndexesForDatabase('blockchain_private');
createIndexesForDatabase('blockchain_public');

print('\nDatabase indexes created successfully for all databases');

// Final summary
print('\n=== BLOCKCHAIN CONTRACT-TOKEN SYSTEM INITIALIZATION COMPLETE ===');
print('Dual-database architecture: Private + Public Blockchain');

print('\n🔒 blockchain_private (PRIVATE BLOCKCHAIN):');
print('   Collections: contracts, tokens, balances, events, blocks');
print('   Purpose: Internal peer operations (supplier, anchor, mainbank)');
print('   Access: Restricted to blockchain peers only');

print('\n🌐 blockchain_public (PUBLIC BLOCKCHAIN):');
print('   Collections: users, events, blocks');
print('   Purpose: Public user access and blockchain transparency');
print('   Access: Public API access');

print('\n👥 Available users (blockchain_public):');
print('BANK:       bank / 123456 (ID: BANK001)');
print('ANCHOR:     anchor / 123456 (ID: ANCHOR001)');
print('SUPPLIERS:  supplier1-supplier5 / 123456 (IDs: SUPPLIER001-SUPPLIER005)');

print('\n🔄 Peer Configuration:');
print('ANCHOR:     Connects to blockchain_private for contracts, events & blocks');
print('MAINBANK:   Connects to blockchain_private for contracts, tokens, balances');
print('SUPPLIER:   Connects to blockchain_private for contracts, tokens, balances');
print('ORDERER:    Connects to blockchain_public for events sync & blocks');
print('BACKEND:    Connects to both databases for full system integration');

print('\n✅ System is ready for testing!');
print('Use docker-compose up --build to start all services.');
print('Access frontend at: http://localhost:4200');
print('Login as "bank" to see the complete system.');
print('');
print('📝 NOTE: Init scripts only run when data directory is empty');
print('If you need to re-initialize, delete data volume and restart container');
print('=================================================================================\n');
