const bcrypt = require('bcryptjs');

const password = '123456';
const salt = bcrypt.genSaltSync(10);
const hash = bcrypt.hashSync(password, salt);

console.log('Hashed password:', hash);
