const crypto = require('crypto');
const fs = require('fs');
const path = require('path');

// Generate ECDSA key pairs for PBFT orderers
function generateOrdererKeys() {
    const orderers = ['ord1', 'ord2', 'ord3'];

    orderers.forEach(ordererId => {
        console.log(`Generating keys for ${ordererId}...`);

        // Create directory
        const keyDir = path.join(__dirname, '..', 'secrets', ordererId);
        if (!fs.existsSync(keyDir)) {
            fs.mkdirSync(keyDir, { recursive: true });
        }

        // Generate ECDSA key pair (prime256v1 curve)
        const { privateKey, publicKey } = crypto.generateKeyPairSync('ec', {
            namedCurve: 'prime256v1',
            publicKeyEncoding: {
                type: 'spki',
                format: 'pem'
            },
            privateKeyEncoding: {
                type: 'pkcs8',
                format: 'pem'
            }
        });

        // Write private key
        const privateKeyPath = path.join(keyDir, 'private.pem');
        fs.writeFileSync(privateKeyPath, privateKey);
        console.log(`  Private key saved to ${privateKeyPath}`);

        // Write public key
        const publicKeyPath = path.join(keyDir, 'public.pem');
        fs.writeFileSync(publicKeyPath, publicKey);
        console.log(`  Public key saved to ${publicKeyPath}`);
    });

    console.log('All orderer keys generated successfully!');
}

// Run the generator
generateOrdererKeys();
