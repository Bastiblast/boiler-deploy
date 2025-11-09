# Auto-Detection System

The deployment system automatically detects your application type and adapts configuration accordingly. No manual PM2 configuration needed!

## How It Works

During deployment, the system:

1. **Reads** your `package.json`
2. **Detects** framework from dependencies
3. **Identifies** package manager from lockfiles
4. **Determines** if build is needed
5. **Configures** PM2 optimally for your framework

## Supported Frameworks

### Next.js

**Detection:** Presence of `next` in `dependencies` or `devDependencies`

**PM2 Configuration:**
- Mode: `fork` (single instance)
- Script: `npm start`
- Reason: Next.js has built-in clustering

**Dependencies:** Full install (including devDependencies for build)

**Example package.json:**
```json
{
  "dependencies": {
    "next": "15.0.0",
    "react": "19.0.0"
  },
  "scripts": {
    "build": "next build",
    "start": "next start"
  }
}
```

**PM2 Config Generated:**
```javascript
{
  script: 'npm',
  args: 'start',
  instances: 1,
  exec_mode: 'fork'
}
```

### Nuxt.js

**Detection:** Presence of `nuxt` in dependencies

**PM2 Configuration:**
- Mode: `fork` (single instance)
- Script: `npm start`
- Host: `0.0.0.0`
- Reason: Nuxt.js has built-in clustering

**Dependencies:** Full install (including devDependencies for build)

**Example package.json:**
```json
{
  "dependencies": {
    "nuxt": "3.0.0"
  },
  "scripts": {
    "build": "nuxt build",
    "start": "nuxt start"
  }
}
```

### Express

**Detection:** Presence of `express` in dependencies

**PM2 Configuration:**
- Mode: `cluster` (multiple instances)
- Script: detected entry point
- Instances: configurable (default: 2)

**Dependencies:** Production only

**Example package.json:**
```json
{
  "dependencies": {
    "express": "4.18.0"
  },
  "main": "server.js"
}
```

**PM2 Config Generated:**
```javascript
{
  script: './server.js',
  instances: 2,
  exec_mode: 'cluster'
}
```

### Fastify

**Detection:** Presence of `fastify` in dependencies

**PM2 Configuration:**
- Mode: `cluster` (multiple instances)
- Script: detected entry point
- Instances: configurable (default: 2)

**Dependencies:** Production only

### NestJS

**Detection:** Presence of `@nestjs/core` in dependencies

**PM2 Configuration:**
- Mode: `cluster` (multiple instances)
- Script: detected entry point (usually `dist/main.js` after build)
- Instances: configurable (default: 2)

**Dependencies:** Production only

**Note:** NestJS requires build step

**Example package.json:**
```json
{
  "dependencies": {
    "@nestjs/core": "10.0.0"
  },
  "scripts": {
    "build": "nest build",
    "start:prod": "node dist/main"
  }
}
```

### Standard Node.js

**Detection:** No specific framework detected

**PM2 Configuration:**
- Mode: `cluster` (multiple instances)
- Script: detected entry point
- Instances: configurable (default: 2)

**Dependencies:** Production only

## Package Manager Detection

The system automatically detects your package manager:

### pnpm

**Detection:** Presence of `pnpm-lock.yaml`

**Commands Used:**
- Install: `pnpm install` (for Next.js/Nuxt) or `pnpm install --prod`
- Build: `pnpm run build`

**Example:**
```bash
# Your project has:
pnpm-lock.yaml

# System runs:
pnpm install
pnpm run build
```

### Yarn

**Detection:** Presence of `yarn.lock`

**Commands Used:**
- Install: `yarn install --production`
- Build: `yarn build`

### npm (default)

**Detection:** Presence of `package-lock.json` or default

**Commands Used:**
- Install: `npm install --production`
- Build: `npm run build`

## Build Detection

The system checks if a `build` script exists in `package.json`:

```json
{
  "scripts": {
    "build": "next build"  // ✅ Build detected
  }
}
```

**If build script exists:**
- Dependencies installed first
- Build command runs: `npm run build`, `pnpm run build`, or `yarn build`
- Then application starts

**If no build script:**
- Dependencies installed
- Application starts immediately

## Entry Point Detection

For standard Node.js applications, the system searches in this order:

1. `main` field in `package.json`
2. `index.js` in root
3. `server.js` in root
4. `app.js` in root
5. `src/index.js`
6. `src/server.js`
7. `src/app.js`

**Example:**
```json
{
  "main": "server.js"  // ✅ Entry point: server.js
}
```

Or auto-detected:
```
project/
├── server.js  // ✅ Auto-detected
└── package.json
```

## PM2 Configuration by Framework

### Next.js / Nuxt.js

```javascript
{
  name: 'myapp',
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

**Why fork mode?** Next.js and Nuxt.js handle clustering internally via worker threads.

### Express / Fastify / NestJS / Node.js

```javascript
{
  name: 'myapp',
  script: './index.js',
  instances: 2,
  exec_mode: 'cluster',
  max_memory_restart: '512M',
  env: {
    NODE_ENV: 'production',
    PORT: 3000
  }
}
```

**Why cluster mode?** Traditional Node.js apps benefit from PM2's clustering for multi-core utilization.

## Viewing Detection Results

During deployment, you'll see:

```
TASK [deploy-app : Display detected configuration]
ok: [vps-01] => 
  msg:
    - "Application Type: nextjs"
    - "Package Manager: pnpm"
    - "Needs Build: true"
    - "Entry File: N/A"
```

## Overriding Auto-Detection

If auto-detection doesn't work for your setup, you can override in `group_vars/all.yml`:

### Force Application Type

```yaml
# Options: nodejs, nextjs, nuxtjs, express, fastify, nestjs
app_type_override: "nextjs"
```

### Force Package Manager

```yaml
# Options: npm, pnpm, yarn
package_manager_override: "pnpm"
```

### Force Entry File

```yaml
# For custom entry points
app_entry_file_override: "dist/main.js"
```

## Detection Examples

### Example 1: Next.js with pnpm

**Your Project:**
```
project/
├── pnpm-lock.yaml
├── package.json
└── next.config.js
```

**package.json:**
```json
{
  "dependencies": {
    "next": "15.0.0"
  },
  "scripts": {
    "build": "next build",
    "start": "next start"
  }
}
```

**Detection Result:**
- Application Type: `nextjs`
- Package Manager: `pnpm`
- Needs Build: `true`
- PM2 Mode: `fork`

**Commands Run:**
```bash
pnpm install
pnpm run build
pm2 start ecosystem.config.js  # with fork mode
```

### Example 2: Express with npm

**Your Project:**
```
project/
├── package-lock.json
├── package.json
└── server.js
```

**package.json:**
```json
{
  "dependencies": {
    "express": "4.18.0"
  },
  "main": "server.js"
}
```

**Detection Result:**
- Application Type: `express`
- Package Manager: `npm`
- Needs Build: `false`
- Entry Point: `server.js`
- PM2 Mode: `cluster`

**Commands Run:**
```bash
npm install --production
pm2 start ecosystem.config.js  # with cluster mode
```

### Example 3: NestJS with yarn

**Your Project:**
```
project/
├── yarn.lock
├── package.json
└── nest-cli.json
```

**package.json:**
```json
{
  "dependencies": {
    "@nestjs/core": "10.0.0"
  },
  "scripts": {
    "build": "nest build",
    "start:prod": "node dist/main"
  }
}
```

**Detection Result:**
- Application Type: `nestjs`
- Package Manager: `yarn`
- Needs Build: `true`
- Entry Point: `dist/main.js` (after build)
- PM2 Mode: `cluster`

**Commands Run:**
```bash
yarn install --production
yarn build
pm2 start ecosystem.config.js  # with cluster mode
```

## Adding New Framework Support

To add support for a new framework, edit:

1. `roles/deploy-app/tasks/detect-app-type.yml` - Add detection logic
2. `roles/deploy-app/templates/ecosystem.config.YOURTYPE.js.j2` - Create PM2 template
3. `roles/deploy-app/tasks/main.yml` - Add PM2 config generation

Example addition:
```yaml
# detect-app-type.yml
- name: Detect application type
  set_fact:
    app_type: >-
      {%- if 'your-framework' in (package_json.dependencies | default({})) -%}
        your-framework
      {%- endif -%}
```

## Troubleshooting Auto-Detection

### Detection Not Working

**Check package.json exists:**
```bash
ssh deploy@your-vps 'cat /var/www/myapp/current/package.json'
```

**View detection output:**
```bash
./deploy.sh deploy -vv  # Verbose mode
```

### Wrong Framework Detected

**Check dependencies:**
```bash
ssh deploy@your-vps 'cat /var/www/myapp/current/package.json | jq .dependencies'
```

**Use override:**
```yaml
app_type_override: "express"  # Force specific type
```

### Wrong Package Manager

**Check lockfiles:**
```bash
ssh deploy@your-vps 'ls -la /var/www/myapp/current/*.lock* /var/www/myapp/current/*lock.yaml'
```

**Use override:**
```yaml
package_manager_override: "pnpm"
```

## Benefits

✅ **Zero Configuration** - No manual PM2 setup needed  
✅ **Optimal Performance** - Framework-specific optimizations  
✅ **Flexibility** - Works with 6+ frameworks  
✅ **Smart** - Detects package manager automatically  
✅ **Extensible** - Easy to add new frameworks  
✅ **Reliable** - Tested on production deployments  

---

For configuration options, see [Configuration Guide](CONFIGURATION.md).  
For troubleshooting, see [Troubleshooting Guide](TROUBLESHOOTING.md).
