# API Capacity Headroom Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Increase `bridgemind.pro` single-node API concurrency headroom by making gateway capacity settings explicit in production config and switching upstream connection-pool isolation from `account_proxy` to `proxy`, without reducing response speed.

**Architecture:** This is an ops-first change, not an application logic change. We will modify only the production gateway config at `/home/ubuntu/bridgemind/deploy/data/config.yaml`, keep existing DB and Redis pool sizing intact, restart the `sub2api` service, then verify health, public endpoints, and real API behavior with rollback ready.

**Tech Stack:** Go service config, Docker Compose, PostgreSQL, Redis, Cloudflare-fronted HTTPS, SSH-based production operations

---

### Task 1: Snapshot Current Production State

**Files:**
- Modify: `/home/ubuntu/bridgemind/deploy/data/config.yaml` (backup only in this task)
- Inspect: `/home/ubuntu/bridgemind/deploy/.env`
- Inspect: `/mnt/c/web/code/openapi/docs/superpowers/specs/2026-04-22-api-capacity-headroom-design.md`

- [ ] **Step 1: Confirm the current production config and env values**

Run:

```bash
sshpass -p 'ww731226.0' ssh -o StrictHostKeyChecking=no ubuntu@101.32.246.236 \
  "sed -n '1,220p' /home/ubuntu/bridgemind/deploy/data/config.yaml && \
   echo '---' && \
   grep -E '^(DATABASE_MAX_OPEN_CONNS|DATABASE_MAX_IDLE_CONNS|REDIS_POOL_SIZE|REDIS_MIN_IDLE_CONNS)=' /home/ubuntu/bridgemind/deploy/.env"
```

Expected:

- `config.yaml` is still minimal.
- `.env` still shows the current DB and Redis pool sizes.

- [ ] **Step 2: Create a timestamped backup of the production config**

Run:

```bash
sshpass -p 'ww731226.0' ssh -o StrictHostKeyChecking=no ubuntu@101.32.246.236 \
  "cp -a /home/ubuntu/bridgemind/deploy/data/config.yaml /home/ubuntu/bridgemind/deploy/data/config.yaml.bak-\$(date +%Y%m%d-%H%M%S)"
```

Expected:

- A new `config.yaml.bak-*` file exists on production.

- [ ] **Step 3: Record baseline service health before changes**

Run:

```bash
curl -fsS https://bridgemind.pro/api/v1/settings/public
curl -fsS https://bridgemind.pro/health
sshpass -p 'ww731226.0' ssh -o StrictHostKeyChecking=no ubuntu@101.32.246.236 \
  "docker stats --no-stream --format 'table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}'"
```

Expected:

- Public settings endpoint succeeds.
- Health endpoint succeeds.
- Container resource usage looks healthy.

- [ ] **Step 4: Commit plan/spec docs before touching production**

Run:

```bash
git -C /mnt/c/web/code/openapi add -f \
  docs/superpowers/specs/2026-04-22-api-capacity-headroom-design.md \
  docs/superpowers/plans/2026-04-22-api-capacity-headroom.md
git -C /mnt/c/web/code/openapi commit -m "docs: add api capacity headroom execution plan"
```

Expected:

- The plan commit succeeds or is a no-op if already committed.

### Task 2: Apply Explicit Gateway Capacity Settings

**Files:**
- Modify: `/home/ubuntu/bridgemind/deploy/data/config.yaml`
- Reference: `/mnt/c/web/code/openapi/backend/internal/config/config.go:1418`
- Reference: `/mnt/c/web/code/openapi/backend/internal/repository/http_upstream.go:607`

- [ ] **Step 1: Write the explicit gateway block into production config**

Target shape:

```yaml
gateway:
  connection_pool_isolation: proxy
  max_idle_conns: 2560
  max_idle_conns_per_host: 120
  max_conns_per_host: 1024
  idle_conn_timeout_seconds: 90
  max_upstream_clients: 5000
  client_idle_ttl_seconds: 900
  concurrency_slot_ttl_minutes: 30
  usage_record:
    worker_count: 128
    queue_size: 16384
    task_timeout_seconds: 5
    overflow_policy: sample
    overflow_sample_percent: 10
    auto_scale_enabled: true
    auto_scale_min_workers: 128
    auto_scale_max_workers: 512
    auto_scale_up_queue_percent: 70
    auto_scale_down_queue_percent: 15
    auto_scale_up_step: 32
    auto_scale_down_step: 16
    auto_scale_check_interval_seconds: 3
    auto_scale_cooldown_seconds: 10
```

- [ ] **Step 2: Apply the config change**

Run:

```bash
sshpass -p 'ww731226.0' ssh -o StrictHostKeyChecking=no ubuntu@101.32.246.236 \
  "python3 - <<'PY'
from pathlib import Path
path = Path('/home/ubuntu/bridgemind/deploy/data/config.yaml')
text = path.read_text()
gateway_block = '''
gateway:
    connection_pool_isolation: proxy
    max_idle_conns: 2560
    max_idle_conns_per_host: 120
    max_conns_per_host: 1024
    idle_conn_timeout_seconds: 90
    max_upstream_clients: 5000
    client_idle_ttl_seconds: 900
    concurrency_slot_ttl_minutes: 30
    usage_record:
        worker_count: 128
        queue_size: 16384
        task_timeout_seconds: 5
        overflow_policy: sample
        overflow_sample_percent: 10
        auto_scale_enabled: true
        auto_scale_min_workers: 128
        auto_scale_max_workers: 512
        auto_scale_up_queue_percent: 70
        auto_scale_down_queue_percent: 15
        auto_scale_up_step: 32
        auto_scale_down_step: 16
        auto_scale_check_interval_seconds: 3
        auto_scale_cooldown_seconds: 10
'''.strip()
lines = text.splitlines()
out = []
skip = False
for line in lines:
    if line.startswith('gateway:'):
        skip = True
        continue
    if skip and line and not line.startswith(' '):
        skip = False
    if not skip:
        out.append(line)
insert_at = len(out)
for i, line in enumerate(out):
    if line.startswith('turnstile:'):
        insert_at = i
        break
out = out[:insert_at] + [gateway_block] + out[insert_at:]
path.write_text('\\n'.join(out).rstrip() + '\\n')
PY
sed -n '1,220p' /home/ubuntu/bridgemind/deploy/data/config.yaml"
```

Expected:

- `config.yaml` contains the explicit `gateway` block with `connection_pool_isolation: proxy`.

- [ ] **Step 3: Sanity-check the edited config**

Run:

```bash
sshpass -p 'ww731226.0' ssh -o StrictHostKeyChecking=no ubuntu@101.32.246.236 \
  "python3 - <<'PY'
import yaml
from pathlib import Path
data = yaml.safe_load(Path('/home/ubuntu/bridgemind/deploy/data/config.yaml').read_text())
print(data['gateway']['connection_pool_isolation'])
print(data['gateway']['usage_record']['auto_scale_max_workers'])
PY"
```

Expected:

- Output includes `proxy`.
- Output includes `512`.

### Task 3: Restart the Service and Verify No Latency Regression

**Files:**
- Inspect: `/home/ubuntu/bridgemind/deploy/data/config.yaml`
- Inspect: Docker container `sub2api`

- [ ] **Step 1: Restart only the application container**

Run:

```bash
sshpass -p 'ww731226.0' ssh -o StrictHostKeyChecking=no ubuntu@101.32.246.236 \
  "cd /home/ubuntu/bridgemind/deploy && docker compose up -d sub2api"
```

Expected:

- `sub2api` is recreated or restarted cleanly.

- [ ] **Step 2: Wait for health to return**

Run:

```bash
sshpass -p 'ww731226.0' ssh -o StrictHostKeyChecking=no ubuntu@101.32.246.236 \
  "docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' sub2api"
curl -fsS https://bridgemind.pro/health
```

Expected:

- Container becomes `healthy` or `running`.
- Public health endpoint succeeds.

- [ ] **Step 3: Verify public site and settings still respond**

Run:

```bash
curl -fsS https://bridgemind.pro/api/v1/settings/public
curl -I https://bridgemind.pro/
```

Expected:

- Public settings return `200`.
- Homepage returns `200` with Cloudflare headers.

- [ ] **Step 4: Run a real API behavior check**

Run:

```bash
curl -fsS https://bridgemind.pro/v1/models \
  -H 'Authorization: Bearer REPLACE_WITH_TEMP_TEST_KEY'
```

Expected:

- A real authenticated API response succeeds.
- The request is not slower in an obvious way than before the change.

- [ ] **Step 5: Inspect logs for warnings that suggest degraded behavior**

Run:

```bash
sshpass -p 'ww731226.0' ssh -o StrictHostKeyChecking=no ubuntu@101.32.246.236 \
  "docker logs --since=10m sub2api 2>&1 | tail -n 200"
```

Expected:

- No obvious upstream client cache churn, transport failures, or usage-record drop bursts.

### Task 4: Roll Back Immediately If Response Speed or Stability Regresses

**Files:**
- Restore: `/home/ubuntu/bridgemind/deploy/data/config.yaml.bak-*`
- Restore target: `/home/ubuntu/bridgemind/deploy/data/config.yaml`

- [ ] **Step 1: Restore the most recent backup if needed**

Run:

```bash
sshpass -p 'ww731226.0' ssh -o StrictHostKeyChecking=no ubuntu@101.32.246.236 \
  "latest=\$(ls -1t /home/ubuntu/bridgemind/deploy/data/config.yaml.bak-* | head -n 1) && \
   cp -a \"\$latest\" /home/ubuntu/bridgemind/deploy/data/config.yaml && \
   echo \"restored \$latest\""
```

Expected:

- The last good backup is restored over the active config.

- [ ] **Step 2: Restart the service after rollback**

Run:

```bash
sshpass -p 'ww731226.0' ssh -o StrictHostKeyChecking=no ubuntu@101.32.246.236 \
  "cd /home/ubuntu/bridgemind/deploy && docker compose up -d sub2api"
```

Expected:

- Service comes back healthy on the previous config.

- [ ] **Step 3: Re-run health and public checks**

Run:

```bash
curl -fsS https://bridgemind.pro/health
curl -fsS https://bridgemind.pro/api/v1/settings/public
```

Expected:

- The site returns to the prior healthy state.
