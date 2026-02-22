#!/usr/bin/env bash
set -euo pipefail

# Your email for Let's Encrypt expiry notifications.
EMAIL="info@zeddunlimited.com"

DOMAINS=("app.justpurl.com" "api.justpurl.com")

# Compose project name is the directory name ("deploy"), so volumes get a "deploy_" prefix.
CERTS_VOLUME="deploy_certbot_certs"
WWW_VOLUME="deploy_certbot_www"

cd "$(dirname "$0")"

echo "==> Creating Docker volumes..."
docker volume create "$CERTS_VOLUME" >/dev/null
# certbot_www is used by the renewal certbot service in docker-compose.prod.yml
docker volume create "$WWW_VOLUME" >/dev/null

# Obtain real Let's Encrypt certificates using standalone mode.
# Certbot binds port 80 itself — no nginx required during initial cert acquisition.
for domain in "${DOMAINS[@]}"; do
  echo "==> Obtaining Let's Encrypt certificate for ${domain}..."
  docker run --rm \
    -v "$CERTS_VOLUME:/etc/letsencrypt" \
    -p 80:80 \
    certbot/certbot certonly \
      --standalone \
      --email "$EMAIL" \
      --agree-tos \
      --no-eff-email \
      --force-renewal \
      -d "$domain"
done

echo ""
echo "==> SSL bootstrap complete!"
echo ""
echo "Bring up the full stack with:"
echo "  docker compose -f deploy/docker-compose.prod.yml up -d"
echo ""
echo "Add this crontab entry (crontab -e) to reload nginx after cert renewal:"
echo "  0 6 * * * docker compose -f ~/purl/deploy/docker-compose.prod.yml exec nginx nginx -s reload"
