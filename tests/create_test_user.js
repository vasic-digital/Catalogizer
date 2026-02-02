// Create a test user in the database
const sqlite3 = require('sqlite3').verbose();
const bcrypt = require('bcrypt');

// Database path - should match your app configuration
const DB_PATH = './catalog-api/data/catalogizer.db';

async function createTestUser() {
  return new Promise((resolve, reject) => {
    const db = new sqlite3.Database(DB_PATH, (err) => {
      if (err) {
        console.error('Error opening database:', err);
        reject(err);
        return;
      }
      console.log('Connected to SQLite database');
    });

    // First check if user already exists
    db.get('SELECT id FROM users WHERE username = ?', ['testuser'], (err, row) => {
      if (err) {
        console.error('Error checking user:', err);
        reject(err);
        return;
      }

      if (row) {
        console.log('Test user already exists with ID:', row.id);
        db.close();
        resolve(true);
        return;
      }

      // Create test user with hashed password
      const username = 'testuser';
      const email = 'test@example.com';
      const plainPassword = 'password123';
      
      // Hash password
      bcrypt.hash(plainPassword, 10, (err, hash) => {
        if (err) {
          console.error('Error hashing password:', err);
          reject(err);
          return;
        }

        // Insert user
        db.run(
          'INSERT INTO users (username, email, password_hash, salt, role_id, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, datetime("now"), datetime("now"))',
          [username, email, hash, 'salt123', 1, true], // role_id 1 = admin
          function(err) {
            if (err) {
              console.error('Error creating user:', err);
              reject(err);
              return;
            }
            
            console.log(`Test user created successfully: ${username} / ${plainPassword}`);
            console.log('User ID:', this.lastID);
            db.close();
            resolve(true);
          }
        );
      });
    });
  });
}

// Run the function
createTestUser()
  .then(() => {
    console.log('Test user creation completed');
    process.exit(0);
  })
  .catch((err) => {
    console.error('Failed to create test user:', err);
    process.exit(1);
  });