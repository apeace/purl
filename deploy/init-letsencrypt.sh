#!/usr/bin/env bash
set -euo pipefail

# Your email for Let's Encrypt expiry notifications.
EMAIL="you@example.com"

DOMAINS=("app.justpurl.com" "api.justpurl.com")

# Compose project name is the directory name ("deploy"), so volumes get a "deploy_" prefix.
CERTS_VOLUME="deploy_certbot_certs"
WWW_VOLUME="deploy_certbot_www"

cd "$(dirname "$0")"

echo "==> Creating Docker volumes..."
docker volume create "$CERTS_VOLUME" >/dev/null
docker volume create "$WWW_VOLUME" >/dev/null

# Generate temporary self-signed certs so nginx can start before real certs exist.
echo "==> Generating temporary self-signed certificates..."
for domain in "${DOMAINS[@]}"; do
  docker run --rm \
    -v "$CERTS_VOLUME:/etc/letsencrypt" \
    alpine:3.21 \
    sh -c "
      apk add --no-cache openssl >/dev/null 2>&1
      mkdir -p /etc/letsencrypt/live/${domain}
      openssl req -x509 -nodes -newkey rsa:2048 -days 1 \
        -keyout /etc/letsencrypt/live/${domain}/privkey.pem \
        -out    /etc/letsencrypt/live/${domain}/fullchain.pem \
        -subj '/CN=${domain}' >/dev/null 2>&1
    "
  echo "   Generated self-signed cert for ${domain}"
done

echo "==> Starting nginx with temporary certs..."
docker compose -f docker-compose.prod.yml up -d nginx

echo "==> Waiting for nginx to be ready..."
sleep 3

# Obtain real Let's Encrypt certificates for each domain.
for domain in "${DOMAINS[@]}"; do
  echo "==> Obtaining Let's Encrypt certificate for ${domain}..."
  docker run --rm \
    -v "$CERTS_VOLUME:/etc/letsencrypt" \
    -v "$WWW_VOLUME:/var/www/certbot" \
    certbot/certbot certonly \
      --webroot \
      --webroot-path=/var/www/certbot \
      --email "$EMAIL" \
      --agree-tos \
      --no-eff-email \
      --force-renewal \
      -d "$domain"
done

echo "==> Reloading nginx to pick up real certificates..."
docker compose -f docker-compose.prod.yml exec nginx nginx -s reload

echo ""
echo "==> SSL bootstrap complete!"
echo ""
echo "Bring up the full stack with:"
echo "  docker compose -f deploy/docker-compose.prod.yml up -d"
echo ""
echo "Add this crontab entry (crontab -e) to reload nginx after cert renewal:"
echo "  0 6 * * * docker compose -f ~/purl/deploy/docker-compose.prod.yml exec nginx nginx -s reload"
