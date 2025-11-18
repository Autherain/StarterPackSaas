# HTTPS Certificate Setup

This setup supports **both development and production** using the same `docker-compose.yml`.

## 🔧 Development Setup (Using .dev Domain)

Perfect for local development with HTTPS.

### 1. Generate Development Certificates

```bash
./scripts/generate_dev_cert.sh
```

This will:

- Install `mkcert` (if not already installed)
- Generate self-signed certificates for `*.starterpack.dev`
- Install the CA in your system's trust store (so browsers trust it)

### 2. Update /etc/hosts

Add the dev domain to point to localhost:

```bash
sudo sh -c 'echo "127.0.0.1 api.starterpack.dev" >> /etc/hosts'
```

### 3. Run Docker Compose

No environment variables needed for dev:

```bash
docker compose up
```

### 4. Access Your API

Open in your browser:

```
https://api.starterpack.dev
```

Your browser will trust the certificate! ✅

### How it Works (Dev Mode)

- `DOMAIN` defaults to `starterpack.dev`
- `ACME_EMAIL` is not set
- Traefik uses static certificates from `./services/traefik/certs/`
- No Let's Encrypt involved

---

## 🚀 Production Setup (Using Real Domain)

For deployment on a VPS with automatic Let's Encrypt certificates.

### 1. Set Environment Variables

Create a `.env` file or export variables:

```bash
export DOMAIN="yourdomain.com"
export ACME_EMAIL="your-email@example.com"
```

Or create `.env` file:

```env
DOMAIN=yourdomain.com
ACME_EMAIL=your-email@example.com
```

### 2. Configure DNS

Point your domain to your server's IP:

```
api.yourdomain.com  →  your.server.ip.address
```

### 3. Ensure Ports are Open

Make sure these ports are accessible:

- **Port 80**: Required for Let's Encrypt HTTP challenge
- **Port 443**: HTTPS traffic

### 4. Run Docker Compose

```bash
docker compose up -d
```

### 5. Access Your API

```
https://api.yourdomain.com
```

Let's Encrypt will automatically:

- Generate a valid SSL certificate
- Store it in `./var/data/traefik/acme.json`
- Renew it automatically before expiration

### How it Works (Production Mode)

- `DOMAIN` is set to your real domain
- `ACME_EMAIL` is set → Traefik uses Let's Encrypt
- The special syntax `${ACME_EMAIL:+letsencrypt}` means: "use letsencrypt resolver only if ACME_EMAIL is set"
- Traefik automatically manages certificates

---

## 📋 Summary: Dev vs Prod

| Aspect | Development | Production |
|--------|-------------|------------|
| **Domain** | `api.starterpack.dev` | `api.yourdomain.com` |
| **Env Vars** | None needed | `DOMAIN`, `ACME_EMAIL` |
| **Certificates** | Self-signed (mkcert) | Let's Encrypt |
| **DNS** | `/etc/hosts` → `127.0.0.1` | DNS A record → VPS IP |
| **Browser Trust** | ✅ Yes (after mkcert install) | ✅ Yes (valid cert) |

---

## 🔧 Troubleshooting

### Development Issues

**Certificate not trusted by browser:**

- Run `mkcert -install` to install the CA
- Restart your browser

**"Connection refused":**

- Check `/etc/hosts` has the entry
- Make sure docker compose is running

**Wrong certificate shown:**

- Clear browser cache
- Check `services/traefik/certs/` has the `.pem` and `.key` files

### Production Issues

**Let's Encrypt not generating certificate:**

Check the logs:

```bash
docker compose logs traefik
```

Common issues:

- **DNS not configured**: Make sure domain points to your server
- **Port 80 blocked**: Let's Encrypt needs it for HTTP challenge
- **ACME_EMAIL not set**: Required for Let's Encrypt
- **Rate limiting**: Let's Encrypt allows 5 certs/week per domain

**Testing with Let's Encrypt Staging:**

To avoid rate limits while testing, use staging server:

```bash
# Add to traefik command in docker-compose.yml:
- "--certificatesresolvers.letsencrypt.acme.caserver=https://acme-staging-v02.api.letsencrypt.org/directory"
```

---

## 🎯 Quick Commands Reference

**Development:**

```bash
# One-time setup
./scripts/generate_dev_cert.sh
sudo sh -c 'echo "127.0.0.1 api.starterpack.dev" >> /etc/hosts'

# Every time
docker compose up
# Access: https://api.starterpack.dev
```

**Production:**

```bash
# One-time setup
export DOMAIN="yourdomain.com"
export ACME_EMAIL="your-email@example.com"

# Every time
docker compose up -d
# Access: https://api.yourdomain.com
```

---

## 📁 Important Files

- `docker-compose.yml` - Auto-detects dev vs prod based on env vars
- `services/traefik/traefik.yaml` - Traefik configuration
- `services/traefik/certs/` - Dev certificates (gitignored)
- `var/data/traefik/acme.json` - Production certificates (gitignored, auto-generated)
- `scripts/generate_dev_cert.sh` - Certificate generator for development

---

## 🎉 That's It

Same codebase, different environment:

- **Dev**: Use `.dev` domain with self-signed certs
- **Prod**: Use real domain with Let's Encrypt

Just change the env vars and voilà! 🚀
