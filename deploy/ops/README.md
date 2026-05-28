# Production Ops Scripts

This directory stores production-oriented helper scripts for fork deployments.

## `update_production.sh`

Run this script on the production server from anywhere inside the cloned repo:

```bash
./deploy/ops/update_production.sh
```

What it does:

- fetches the latest `origin/main`
- preserves server-local tracked changes before the fast-forward update
- builds a clean image from the target git commit
- rewrites `deploy/docker-compose.override.yml` to use that image
- restarts only the `sub2api` service
- verifies container health and the local health endpoint

Optional environment variables:

- `GIT_REMOTE` default: `origin`
- `IMAGE_REPOSITORY` default: `openapi-prod`
- `SERVICE_NAME` default: `sub2api`
- `NODE_MAX_OLD_SPACE_SIZE` default: `4096`
- `FORCE_BUILD=1` to rebuild even if the image tag already exists locally

If you want the admin UI "更新" button to launch this script automatically, the running `sub2api` container must also have:

- `/var/run/docker.sock:/var/run/docker.sock`
- the repo bind-mounted into the container at the same absolute path as the host
- `UPDATE_EXTERNAL_UPDATER_COMMAND` pointing to this script
- `UPDATE_EXTERNAL_UPDATER_WORKING_DIRECTORY` pointing to the repo root
