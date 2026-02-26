#!/usr/bin/env bash
set -euo pipefail

REPO_URL="git@github.com:apeace/purl.git"
REPO_DIR="/home/ubuntu/purl"
SSH_KEY_PATH="/home/ubuntu/.ssh/deploy_key"

# Prompt for the deploy key before doing anything else.
echo "Paste your GitHub deploy key (private key), then press Ctrl+D on a new line:"
echo ""
DEPLOY_KEY=$(cat)

if [ -z "$DEPLOY_KEY" ]; then
  echo "ERROR: No key provided."
  exit 1
fi

echo ""
echo "==> Installing deploy key..."
mkdir -p /home/ubuntu/.ssh
chmod 700 /home/ubuntu/.ssh
echo "$DEPLOY_KEY" > "$SSH_KEY_PATH"
chmod 600 "$SSH_KEY_PATH"

# Tell SSH to use this key for GitHub.
cat >> /home/ubuntu/.ssh/config <<EOF

Host github.com
  IdentityFile $SSH_KEY_PATH
  StrictHostKeyChecking accept-new
EOF
chmod 600 /home/ubuntu/.ssh/config

echo "==> Installing Docker..."
sudo apt-get update -q
sudo apt-get install -y -q ca-certificates curl gnupg

sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" \
  | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null

sudo apt-get update -q
sudo apt-get install -y -q docker-ce docker-ce-cli containerd.io docker-compose-plugin

echo "==> Enabling Docker service..."
sudo systemctl enable --now docker

echo "==> Adding ubuntu to docker group (re-login required for this to take effect)..."
sudo usermod -aG docker ubuntu

# Configure ufw if it is active.
if sudo ufw status | grep -q "Status: active"; then
  echo "==> Configuring ufw firewall..."
  sudo ufw allow 22/tcp
  sudo ufw allow 80/tcp
  sudo ufw allow 443/tcp
  sudo ufw reload
fi

echo "==> Setting up bash aliases for ubuntu user..."
cat >> /home/ubuntu/.bashrc <<'EOF'

export EDITOR="vi"

# Git aliases
alias \
  c='git commit' \
  s='git status' \
  ch='git checkout' \
  b='git branch' \
  d='git diff' \
  pull='git pull' \
  push='git push'

# Tmux
# See: https://tmuxcheatsheet.com/
# See also: https://hamvocke.com/blog/a-quick-and-easy-guide-to-tmux/

# Create or attach to session with given name, e.g. "sesh work"
alias o='tmux new-session -A -s'

# Kill the given tmux session, e.g. "tmuxrm work"
alias tmuxkill='tmux kill-session -t'
EOF

echo "==> Cloning repository to $REPO_DIR..."
if [ -d "$REPO_DIR" ]; then
  echo "   Directory already exists, skipping clone."
else
  git clone "$REPO_URL" "$REPO_DIR"
fi

echo "==> Creating api/.env placeholder..."
cat > "$REPO_DIR/api/.env" <<'EOF'
DATABASE_URL=postgres://purl:CHANGE_ME@postgres:5432/purl?sslmode=disable
REDIS_URL=redis://redis:6379
PORT=9090
EOF

echo "==> Creating deploy/.env placeholder..."
cat > "$REPO_DIR/deploy/.env" <<'EOF'
POSTGRES_PASSWORD=CHANGE_ME
EOF

echo ""
echo "========================================"
echo " Setup complete!"
echo "========================================"
echo ""
echo "Next steps:"
echo ""
echo "1. Edit API environment variables:"
echo "   nano $REPO_DIR/api/.env"
echo ""
echo "2. Set the Postgres password (must match DATABASE_URL above):"
echo "   nano $REPO_DIR/deploy/.env"
echo ""
echo "3. Set your email in init-letsencrypt.sh:"
echo "   nano $REPO_DIR/deploy/init-letsencrypt.sh"
echo ""
echo "4. Log out and back in so the docker group takes effect, then:"
echo "   bash $REPO_DIR/deploy/init-letsencrypt.sh"
echo ""
echo "5. Bring up the full stack:"
echo "   docker compose -f $REPO_DIR/deploy/docker-compose.prod.yml up -d"
echo ""
echo "6. Add cert-reload crontab entry (crontab -e):"
echo "   0 6 * * * docker compose -f $REPO_DIR/deploy/docker-compose.prod.yml exec nginx nginx -s reload"
echo ""
