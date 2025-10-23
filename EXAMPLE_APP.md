# Example Node.js Application Structure

This is an example of how your Node.js application should be structured to work with this deployment setup.

## Minimal Application

### package.json
```json
{
  "name": "myapp",
  "version": "1.0.0",
  "description": "My Node.js Application",
  "main": "index.js",
  "scripts": {
    "start": "node index.js",
    "dev": "nodemon index.js",
    "build": "echo 'No build step required'"
  },
  "dependencies": {
    "express": "^4.18.2",
    "pg": "^8.11.3",
    "dotenv": "^16.3.1"
  },
  "devDependencies": {
    "nodemon": "^3.0.1"
  }
}
```

### index.js
```javascript
require('dotenv').config();
const express = require('express');
const { Pool } = require('pg');

const app = express();
const port = process.env.PORT || 3000;

// Database connection
const pool = new Pool({
  host: process.env.DB_HOST,
  port: process.env.DB_PORT || 5432,
  database: process.env.DB_NAME,
  user: process.env.DB_USER,
  password: process.env.DB_PASSWORD,
});

// Middleware
app.use(express.json());

// Health check endpoint (required)
app.get('/health', (req, res) => {
  res.status(200).json({ 
    status: 'ok',
    environment: process.env.NODE_ENV,
    uptime: process.uptime()
  });
});

// Database health check
app.get('/health/db', async (req, res) => {
  try {
    await pool.query('SELECT 1');
    res.status(200).json({ status: 'ok', database: 'connected' });
  } catch (error) {
    res.status(503).json({ status: 'error', database: 'disconnected' });
  }
});

// Example route
app.get('/', (req, res) => {
  res.json({ 
    message: 'Welcome to My App',
    version: '1.0.0'
  });
});

// Example database query
app.get('/users', async (req, res) => {
  try {
    const result = await pool.query('SELECT * FROM users LIMIT 10');
    res.json(result.rows);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// Start server
app.listen(port, () => {
  console.log(`Server running on port ${port}`);
  console.log(`Environment: ${process.env.NODE_ENV}`);
});

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('SIGTERM signal received: closing HTTP server');
  pool.end();
  process.exit(0);
});
```

### .env.example
```bash
NODE_ENV=development
PORT=3000

DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp_dev
DB_USER=myapp_user
DB_PASSWORD=changeme

# Add your custom environment variables here
JWT_SECRET=your-secret-key
API_KEY=your-api-key
```

### .gitignore
```
node_modules/
.env
*.log
.DS_Store
dist/
build/
```

### README.md (for your app)
```markdown
# My Node.js Application

## Development

\`\`\`bash
npm install
cp .env.example .env
# Edit .env with your settings
npm run dev
\`\`\`

## Production

Application is deployed and managed by PM2 via Ansible.

Environment variables are managed in the deployment configuration.
```

## Application Requirements

Your application MUST:

1. ✅ **Listen on PORT environment variable**
   ```javascript
   const port = process.env.PORT || 3000;
   app.listen(port);
   ```

2. ✅ **Have a /health endpoint**
   ```javascript
   app.get('/health', (req, res) => {
     res.status(200).json({ status: 'ok' });
   });
   ```

3. ✅ **Use environment variables for configuration**
   - Database credentials
   - API keys
   - Application settings

4. ✅ **Handle SIGTERM for graceful shutdown**
   ```javascript
   process.on('SIGTERM', () => {
     // Cleanup and exit
   });
   ```

5. ✅ **Have a valid package.json with dependencies**

## Optional but Recommended

- `/health/db` - Database health check
- `/health/ready` - Readiness probe
- Error handling middleware
- Request logging
- Rate limiting
- CORS configuration

## Database Schema

Create initial schema using migrations:

### migrations/001_initial.sql
```sql
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  name VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

Run migrations after deployment:
```bash
ssh deploy@YOUR_SERVER_IP
cd /var/www/myapp/current
psql -h DB_HOST -U myapp_user -d myapp_production < migrations/001_initial.sql
```

## Testing Locally

```bash
# Start PostgreSQL locally (Docker)
docker run -d \
  --name postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=myapp_dev \
  -e POSTGRES_USER=myapp_user \
  -p 5432:5432 \
  postgres:15

# Run your app
npm install
npm run dev
```

## Deployment Flow

1. Push code to GitHub
2. Run: `./deploy.sh production deploy`
3. Ansible will:
   - Clone your repo
   - Install dependencies
   - Build (if needed)
   - Start/reload with PM2
   - Keep last 5 releases
