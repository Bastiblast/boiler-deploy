# Examples

Real-world configuration examples for different Node.js frameworks and setups.

## Table of Contents

- [Next.js with pnpm (Tested)](#nextjs-with-pnpm)
- [Express with npm](#express-with-npm)
- [NestJS with yarn](#nestjs-with-yarn)
- [Nuxt.js with pnpm](#nuxtjs-with-pnpm)
- [Fastify API](#fastify-api)
- [Multiple Applications](#multiple-applications)
- [Environment Variables](#environment-variables)

## Next.js with pnpm

✅ **Tested on production VPS**

### Project Structure

```
myportfolio/
├── app/
├── public/
├── package.json
├── pnpm-lock.yaml
├── next.config.ts
└── middleware.ts
```

### package.json

```json
{
  "name": "myportfolio",
  "version": "0.1.0",
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start"
  },
  "dependencies": {
    "next": "15.0.0",
    "react": "19.0.0",
    "react-dom": "19.0.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "@types/react": "^19.0.0",
    "typescript": "^5.0.0"
  }
}
```

### Configuration

```yaml
# group_vars/all.yml
app_name: myportfolio
app_port: 3000
app_repo: "https://github.com/username/myportfolio.git"
app_branch: "main"
```

### Auto-Detection Result

- **Framework**: Next.js (detected from `next` dependency)
- **Package Manager**: pnpm (detected from `pnpm-lock.yaml`)
- **Build**: Required (detected from `build` script)
- **PM2 Mode**: Fork (1 instance)

### Deployed Configuration

**PM2 ecosystem.config.js (auto-generated):**
```javascript
{
  name: 'myportfolio',
  script: 'npm',
  args: 'start',
  instances: 1,
  exec_mode: 'fork',
  max_memory_restart: '512M',
  env: {
    NODE_ENV: 'production',
    PORT: 3000
  }
}
```

### Access

- Application: `http://your-vps-ip`
- With i18n: `http://your-vps-ip/en` or `/fr`

## Express with npm

### Project Structure

```
express-api/
├── src/
│   ├── routes/
│   ├── controllers/
│   └── index.js
├── package.json
├── package-lock.json
└── .env.example
```

### package.json

```json
{
  "name": "express-api",
  "version": "1.0.0",
  "main": "src/index.js",
  "scripts": {
    "start": "node src/index.js",
    "dev": "nodemon src/index.js"
  },
  "dependencies": {
    "express": "^4.18.0",
    "dotenv": "^16.0.0",
    "pg": "^8.11.0"
  }
}
```

### src/index.js

```javascript
const express = require('express');
const app = express();

app.use(express.json());

app.get('/api/health', (req, res) => {
  res.json({ status: 'ok' });
});

app.get('/api/users', (req, res) => {
  res.json({ users: [] });
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});
```

### Configuration

```yaml
# group_vars/all.yml
app_name: express-api
app_port: 3000
app_repo: "https://github.com/username/express-api.git"
app_branch: "main"

# group_vars/webservers.yml
ssl_enabled: true
ssl_certbot_email: "admin@yourdomain.com"
ssl_domains:
  - "api.yourdomain.com"
```

### Environment Variables

```bash
# On VPS: /var/www/express-api/shared/config/.env
DATABASE_URL=postgresql://user:pass@localhost:5432/apidb
JWT_SECRET=your-secret-key
NODE_ENV=production
PORT=3000
```

### Auto-Detection Result

- **Framework**: Express (detected from `express` dependency)
- **Package Manager**: npm (detected from `package-lock.json`)
- **Build**: Not required
- **Entry Point**: `src/index.js` (from `main` in package.json)
- **PM2 Mode**: Cluster (2 instances)

### Deployed Configuration

**PM2 ecosystem.config.js (auto-generated):**
```javascript
{
  name: 'express-api',
  script: './src/index.js',
  instances: 2,
  exec_mode: 'cluster',
  max_memory_restart: '512M',
  env: {
    NODE_ENV: 'production',
    PORT: 3000
  }
}
```

## NestJS with yarn

### Project Structure

```
nestjs-app/
├── src/
│   ├── main.ts
│   ├── app.module.ts
│   └── ...
├── dist/          # Build output
├── package.json
├── yarn.lock
└── nest-cli.json
```

### package.json

```json
{
  "name": "nestjs-app",
  "version": "1.0.0",
  "scripts": {
    "build": "nest build",
    "start": "node dist/main",
    "start:dev": "nest start --watch",
    "start:prod": "node dist/main"
  },
  "dependencies": {
    "@nestjs/common": "^10.0.0",
    "@nestjs/core": "^10.0.0",
    "@nestjs/platform-express": "^10.0.0",
    "reflect-metadata": "^0.1.13",
    "rxjs": "^7.8.1"
  },
  "devDependencies": {
    "@nestjs/cli": "^10.0.0",
    "@typescript-eslint/eslint-plugin": "^6.0.0",
    "typescript": "^5.0.0"
  }
}
```

### Configuration

```yaml
# group_vars/all.yml
app_name: nestjs-app
app_port: 3000
app_repo: "https://github.com/username/nestjs-app.git"
app_branch: "main"

# Override entry point after build
app_entry_file_override: "dist/main.js"
```

### Auto-Detection Result

- **Framework**: NestJS (detected from `@nestjs/core`)
- **Package Manager**: yarn (detected from `yarn.lock`)
- **Build**: Required (detected from `build` script)
- **Entry Point**: `dist/main.js`
- **PM2 Mode**: Cluster (2 instances)

### Deployment Flow

1. `yarn install --production`
2. `yarn build` → creates `dist/` directory
3. PM2 starts `dist/main.js`

## Nuxt.js with pnpm

### Project Structure

```
nuxt-app/
├── app/
├── pages/
├── components/
├── package.json
├── pnpm-lock.yaml
└── nuxt.config.ts
```

### package.json

```json
{
  "name": "nuxt-app",
  "version": "1.0.0",
  "scripts": {
    "build": "nuxt build",
    "start": "node .output/server/index.mjs",
    "dev": "nuxt dev"
  },
  "dependencies": {
    "nuxt": "^3.0.0",
    "vue": "^3.0.0"
  }
}
```

### Configuration

```yaml
# group_vars/all.yml
app_name: nuxt-app
app_port: 3000
app_repo: "https://github.com/username/nuxt-app.git"
app_branch: "main"
```

### Auto-Detection Result

- **Framework**: Nuxt.js (detected from `nuxt` dependency)
- **Package Manager**: pnpm (detected from `pnpm-lock.yaml`)
- **Build**: Required
- **PM2 Mode**: Fork (1 instance)

### Deployed Configuration

**PM2 ecosystem.config.js (auto-generated):**
```javascript
{
  name: 'nuxt-app',
  script: 'npm',
  args: 'start',
  instances: 1,
  exec_mode: 'fork',
  max_memory_restart: '512M',
  env: {
    NODE_ENV: 'production',
    PORT: 3000,
    HOST: '0.0.0.0'
  }
}
```

## Fastify API

### Project Structure

```
fastify-api/
├── src/
│   ├── server.js
│   ├── routes.js
│   └── plugins/
├── package.json
└── package-lock.json
```

### package.json

```json
{
  "name": "fastify-api",
  "version": "1.0.0",
  "main": "src/server.js",
  "scripts": {
    "start": "node src/server.js"
  },
  "dependencies": {
    "fastify": "^4.0.0",
    "fastify-cors": "^8.0.0",
    "pg": "^8.11.0"
  }
}
```

### src/server.js

```javascript
const fastify = require('fastify')({ logger: true });

fastify.register(require('fastify-cors'));

fastify.get('/health', async (request, reply) => {
  return { status: 'ok' };
});

const start = async () => {
  try {
    await fastify.listen({ 
      port: process.env.PORT || 3000,
      host: '0.0.0.0'
    });
  } catch (err) {
    fastify.log.error(err);
    process.exit(1);
  }
};

start();
```

### Auto-Detection Result

- **Framework**: Fastify (detected from `fastify` dependency)
- **Package Manager**: npm
- **Entry Point**: `src/server.js`
- **PM2 Mode**: Cluster (2 instances)

## Multiple Applications

Deploy multiple apps on the same VPS:

### Configuration

```yaml
# inventory/production/hosts.yml
all:
  children:
    webservers:
      hosts:
        vps-01:
          ansible_host: XX.XX.XX.XX
```

### Deploy First App

```bash
# Set first app config
vim group_vars/all.yml  # app_name: api, app_port: 3000

./deploy.sh deploy
```

### Deploy Second App

```bash
# Change config for second app
vim group_vars/all.yml  # app_name: frontend, app_port: 3001

./deploy.sh deploy
```

### Nginx Configuration

Both apps will be configured automatically:

- API: `http://your-vps-ip:3000`
- Frontend: `http://your-vps-ip:3001`

For custom domains:

```yaml
# group_vars/webservers.yml
ssl_domains:
  - "api.yourdomain.com"
  - "app.yourdomain.com"
```

## Environment Variables

### Simple .env File

```bash
# /var/www/myapp/shared/config/.env
DATABASE_URL=postgresql://user:pass@localhost/db
API_KEY=secret-key
NODE_ENV=production
PORT=3000
```

### Create via SSH

```bash
ssh deploy@your-vps-ip << 'EOF'
cat > /var/www/myapp/shared/config/.env << ENVEOF
DATABASE_URL=postgresql://myapp:password@localhost:5432/myapp_production
REDIS_URL=redis://localhost:6379
JWT_SECRET=$(openssl rand -base64 32)
NODE_ENV=production
ENVEOF
EOF
```

### Load in Application

**Next.js (automatic):**
```javascript
// .env is automatically loaded
console.log(process.env.DATABASE_URL);
```

**Express/Fastify:**
```javascript
require('dotenv').config();
console.log(process.env.DATABASE_URL);
```

### Sensitive Values (Ansible Vault)

For sensitive configs, use Ansible Vault:

```bash
# Encrypt a value
ansible-vault encrypt_string 'super-secret-password' --name 'db_password'
```

Add to `group_vars/dbservers.yml`:

```yaml
postgresql_users:
  - name: myapp
    password: !vault |
          $ANSIBLE_VAULT;1.1;AES256
          ...encrypted...
```

## Production Checklist

Before going to production:

- [ ] SSL certificates configured
- [ ] Custom domain pointed to VPS
- [ ] Database passwords changed
- [ ] Grafana admin password changed
- [ ] Environment variables set
- [ ] Backups configured
- [ ] Monitoring dashboards set up
- [ ] Firewall rules verified
- [ ] SSH keys only (no password auth)
- [ ] Application tested

## Performance Tuning

### High-Traffic API

```yaml
# group_vars/all.yml
pm2_instances: 4  # More instances
pm2_max_memory: "1G"

# group_vars/webservers.yml
nginx_worker_processes: 4
nginx_worker_connections: 2048
```

### Low-Memory VPS

```yaml
# group_vars/all.yml
pm2_instances: 1  # Single instance
pm2_max_memory: "256M"

# group_vars/dbservers.yml
postgresql_shared_buffers: "128MB"
postgresql_max_connections: 50
```

---

For more configuration options, see [Configuration Guide](CONFIGURATION.md).  
For troubleshooting, see [Troubleshooting Guide](TROUBLESHOOTING.md).
