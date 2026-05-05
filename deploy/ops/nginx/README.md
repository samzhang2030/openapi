# Nginx Site Maintenance

Canonical site config lives in `ops/nginx/bridgemind.pro.conf`.

Apply changes safely:

```bash
cd /home/ubuntu/bridgemind/deploy
./ops/update_nginx_site.sh
```

What the script does:
- Backs up the active `/etc/nginx/sites-available/bridgemind.pro`
- Copies the canonical config into place
- Runs `nginx -t`
- Reloads Nginx only if the config test passes

Latest backup files are stored in `ops/nginx-backups/`.
