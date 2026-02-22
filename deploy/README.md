# Deployment

Single EC2 instance running Docker Compose. Nginx terminates SSL and routes:
- `app.justpurl.com` → Vue SPA (built into the nginx image)
- `api.justpurl.com` → Go API container

All services restart automatically on crash and on reboot.

## Initial Server Setup

### 1. Provision the EC2 instance

Launch an Ubuntu 22.04+ instance, open ports 22, 80, and 443 in the security group, and point DNS for both `app.justpurl.com` and `api.justpurl.com` at the instance's public IP. DNS must resolve before running the SSL bootstrap.

### 2. Run the setup script

Add the deploy key to the GitHub repo (Settings → Deploy keys), then SSH into the instance and run:

```sh
curl -fsSL https://raw.githubusercontent.com/apeace/purl/main/deploy/setup.sh | bash
```

The script will immediately prompt you to paste the private deploy key. After you paste it and press Ctrl+D, it installs the key at `~/.ssh/deploy_key`, configures SSH to use it for GitHub, installs Docker, enables it on boot, adds `ubuntu` to the `docker` group, clones the repo to `/home/ubuntu/purl`, and creates placeholder `.env` files.

Log out and back in so the `docker` group takes effect.

### 3. Configure environment variables

**`/home/ubuntu/purl/api/.env`** — application secrets:

```sh
nano /home/ubuntu/purl/api/.env
```

```
DATABASE_URL=postgres://purl:CHANGE_ME@postgres:5432/purl?sslmode=disable
REDIS_URL=redis://redis:6379
PORT=9090
```

**`/home/ubuntu/purl/deploy/.env`** — Postgres password (must match `DATABASE_URL`):

```sh
nano /home/ubuntu/purl/deploy/.env
```

```
POSTGRES_PASSWORD=CHANGE_ME
```

### 4. Bootstrap SSL certificates

Edit `deploy/init-letsencrypt.sh` and set your email address at the top of the file, then run it:

```sh
nano /home/ubuntu/purl/deploy/init-letsencrypt.sh   # set EMAIL=
bash /home/ubuntu/purl/deploy/init-letsencrypt.sh
```

This script handles the chicken-and-egg problem: nginx needs certs to start, but certbot needs nginx for the ACME challenge. It does so by generating temporary self-signed certs, starting nginx, then replacing them with real Let's Encrypt certs.

### 5. Start the full stack

```sh
docker compose -f /home/ubuntu/purl/deploy/docker-compose.prod.yml up -d
```

### 6. Add the cert-reload cron job

Certbot renews certificates automatically every 12 hours, but nginx must be reloaded to pick them up. Add this to the crontab (`crontab -e`):

```
0 6 * * * docker compose -f /home/ubuntu/purl/deploy/docker-compose.prod.yml exec nginx nginx -s reload
```

---

## Deploying Updates

Rebuild and restart only the changed service without downtime to others:

```sh
cd /home/ubuntu/purl
git pull

# Redeploy the API:
docker compose -f deploy/docker-compose.prod.yml up -d --build api

# Redeploy the frontend (rebuilds the nginx image):
docker compose -f deploy/docker-compose.prod.yml up -d --build nginx
```

To rebuild and restart everything:

```sh
docker compose -f deploy/docker-compose.prod.yml up -d --build
```

---

## Common Maintenance

All commands run from `/home/ubuntu/purl` on the server.

**Check service status:**
```sh
docker compose -f deploy/docker-compose.prod.yml ps
```

**View logs:**
```sh
# All services
docker compose -f deploy/docker-compose.prod.yml logs -f

# One service
docker compose -f deploy/docker-compose.prod.yml logs -f api
```

**Restart a service:**
```sh
docker compose -f deploy/docker-compose.prod.yml restart api
```

**Stop everything:**
```sh
docker compose -f deploy/docker-compose.prod.yml down
```

**Start everything:**
```sh
docker compose -f deploy/docker-compose.prod.yml up -d
```

**Open a psql shell:**
```sh
docker compose -f deploy/docker-compose.prod.yml exec postgres psql -U purl purl
```

**Force cert renewal (outside the normal schedule):**
```sh
docker compose -f deploy/docker-compose.prod.yml exec certbot certbot renew --force-renewal
docker compose -f deploy/docker-compose.prod.yml exec nginx nginx -s reload
```
