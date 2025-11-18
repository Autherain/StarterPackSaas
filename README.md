# StarterPack SaaS

A production-ready SaaS starter with automatic HTTPS via Traefik and Let's Encrypt.

## Quick Start

### Development (Local Machine)

```bash
# 1. Generate dev certificates
make dev-setup

# 2. Add to /etc/hosts (one-time)
sudo sh -c 'echo "127.0.0.1 api.starterpack.dev" >> /etc/hosts'

# 3. Start development server
make dev

# 4. Access your API
# Browser: https://api.starterpack.dev
# Dashboard: http://localhost:8888
```

### Production (VPS)

```bash
# 1. Set environment variables
export DOMAIN="yourdomain.com"
export ACME_EMAIL="your-email@example.com"

# 2. Make sure DNS is configured
# api.yourdomain.com → your VPS IP

# 3. Start production server
make prod

# 4. Access your API
# https://api.yourdomain.com
```

## The Magic ✨

The same `docker-compose.yml` works for both dev and production:

| Environment | Domain | Certificates | Setup |
|------------|--------|--------------|-------|
| **Dev** | `api.starterpack.dev` | Self-signed (mkcert) | No env vars needed |
| **Prod** | `api.yourdomain.com` | Let's Encrypt | Set `DOMAIN` + `ACME_EMAIL` |

### How It Works

**Line 74 is the key:**

```yaml
- "traefik.http.routers.api.tls.certresolver=${ACME_EMAIL:+letsencrypt}"
```

This bash parameter expansion means:

- `${ACME_EMAIL:+letsencrypt}` → If `ACME_EMAIL` is set, use "letsencrypt"
- `${ACME_EMAIL:+letsencrypt}` → If `ACME_EMAIL` is empty, use "" (empty = use static certs)

So:

- **Dev**: No `ACME_EMAIL` → certresolver is empty → Traefik uses static certs from `/certs/`
- **Prod**: Set `ACME_EMAIL` → certresolver is "letsencrypt" → Traefik uses Let's Encrypt

## Available Commands

```bash
make help        # Show all available commands
make dev-setup   # Generate dev certificates
make dev         # Start development server
make prod        # Start production server
make up          # Start in background
make down        # Stop containers
make logs        # Show logs
make clean       # Remove containers and volumes
```

## Documentation

- [CERTIFICATES.md](CERTIFICATES.md) - Detailed certificate setup guide
- Backend: See [backend/README.md](backend/README.md)

## Project Structure

```
.
├── backend/                  # Go API server
├── services/
│   └── traefik/
│       ├── traefik.yaml     # Traefik configuration
│       └── certs/           # Dev certificates (gitignored)
├── scripts/
│   └── generate_dev_cert.sh # Certificate generator
├── var/data/                # Runtime data (gitignored)
│   ├── api/                 # PostgreSQL data
│   └── traefik/acme.json    # Let's Encrypt certs
├── docker-compose.yml       # Works for dev AND prod
└── Makefile                 # Convenient commands
```

## Tech Stack

- **Backend**: Go
- **Database**: PostgreSQL
- **Reverse Proxy**: Traefik 3.1
- **HTTPS**: Let's Encrypt (prod) / mkcert (dev)
- **Container**: Docker Compose

## Support

Check the logs if something goes wrong:

```bash
make logs
```

For certificate issues, see [CERTIFICATES.md](CERTIFICATES.md).

---

Built with ❤️ for rapid SaaS development
