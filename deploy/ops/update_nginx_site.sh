#!/usr/bin/env bash
set -euo pipefail

SITE_SRC="/home/ubuntu/bridgemind/deploy/ops/nginx/bridgemind.pro.conf"
SITE_DST="/etc/nginx/sites-available/bridgemind.pro"
BACKUP_DIR="/home/ubuntu/bridgemind/deploy/ops/nginx-backups"
STAMP="$(date +%Y%m%d-%H%M%S)"

mkdir -p "$BACKUP_DIR"
sudo cp "$SITE_DST" "$BACKUP_DIR/bridgemind.pro.$STAMP.conf"
sudo cp "$SITE_SRC" "$SITE_DST"
sudo nginx -t
sudo systemctl reload nginx

echo "Nginx config updated and reloaded successfully."
echo "Backup saved to: $BACKUP_DIR/bridgemind.pro.$STAMP.conf"
