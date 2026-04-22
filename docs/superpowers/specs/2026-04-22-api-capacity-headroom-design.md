# API Capacity Headroom Design

Date: 2026-04-22
Scope: `bridgemind.pro` production API/backend capacity optimization
Status: Approved for implementation after user review

## Background

The current production system is healthy and responsive. Recent checks showed:

- `sub2api` CPU usage is low and memory usage is modest.
- PostgreSQL and Redis connection pools are already sized generously through `deploy/.env`.
- The production `config.yaml` is intentionally minimal, so many gateway capacity controls currently rely on code defaults instead of explicit production settings.

The goal is to improve future concurrency headroom without lowering current response speed.

## Goals

- Increase single-node capacity headroom for future traffic growth.
- Avoid regressions in response latency and perceived responsiveness.
- Make gateway capacity behavior explicit in production config instead of relying on implicit defaults.
- Keep rollout and rollback simple.

## Non-Goals

- No database schema or query changes in this phase.
- No horizontal scaling in this phase.
- No broad refactor of routing, account scheduling, or business logic.

## Current Technical Findings

### Production resource headroom

- Application container CPU and memory usage are currently low.
- Database pool is already configured to `DATABASE_MAX_OPEN_CONNS=256` and `DATABASE_MAX_IDLE_CONNS=128`.
- Redis pool is already configured to `REDIS_POOL_SIZE=4096` and `REDIS_MIN_IDLE_CONNS=256`.

These numbers suggest the current bottleneck risk is not host resource exhaustion.

### Upstream connection pool fragmentation

The gateway default isolation mode is `account_proxy`.

- Default isolation is defined in [backend/internal/config/config.go](/mnt/c/web/code/openapi/backend/internal/config/config.go#L1418).
- In `account` and `account_proxy` modes, the upstream transport pool sizes are clamped to the account concurrency value in [backend/internal/repository/http_upstream.go](/mnt/c/web/code/openapi/backend/internal/repository/http_upstream.go#L607).

This means many accounts can end up with many small isolated upstream pools. That is good for strict isolation, but it can reduce connection reuse, increase TLS handshakes, and waste capacity headroom.

### Existing async capacity protections

Usage record processing already uses a bounded worker pool with autoscaling defaults in [backend/internal/service/usage_record_worker_pool.go](/mnt/c/web/code/openapi/backend/internal/service/usage_record_worker_pool.go#L356).

This is already a good protection against unbounded goroutine growth and does not appear to be the first lever to pull.

## Options Considered

### Option 1: Keep `account_proxy`, only make defaults explicit

Pros:

- Lowest risk.
- No change in upstream sharing behavior.

Cons:

- Capacity improvement likely small.
- Does not address connection-pool fragmentation.

### Option 2: Switch to `proxy` isolation and make gateway limits explicit

Pros:

- Best expected gain in capacity headroom on the current single node.
- Improves connection reuse for accounts that share the same proxy path.
- Reduces transport/client fragmentation and handshake churn.

Cons:

- Slightly broader behavior change than pure config-explicitness.
- Requires careful post-release verification to ensure latency does not worsen.

### Option 3: Horizontal scale to multiple app instances

Pros:

- Highest absolute headroom.

Cons:

- Higher operational complexity.
- Not necessary yet given current resource usage.

## Chosen Design

Use Option 2 with a guarded rollout.

### Planned production changes

1. Set `gateway.connection_pool_isolation` to `proxy`.
2. Make gateway capacity-related settings explicit in production `config.yaml` instead of relying on code defaults.
3. Keep current PostgreSQL and Redis pool settings unchanged for this phase.
4. Keep current business routing, account selection, and quotas unchanged.

### Gateway settings to make explicit

The production config should explicitly set:

- `gateway.connection_pool_isolation`
- `gateway.max_idle_conns`
- `gateway.max_idle_conns_per_host`
- `gateway.max_conns_per_host`
- `gateway.idle_conn_timeout_seconds`
- `gateway.max_upstream_clients`
- `gateway.client_idle_ttl_seconds`
- `gateway.concurrency_slot_ttl_minutes`
- `gateway.usage_record.*` autoscaling settings

The exact values should start from current code defaults unless a production-specific reason appears during implementation.

## Why this should not reduce response speed

This design is intended to preserve or improve latency, not trade latency away for throughput.

Reasons:

- Shared upstream pools should improve connection reuse.
- Better reuse reduces avoidable TCP/TLS setup work.
- We are not reducing DB or Redis pool sizes.
- We are not increasing request path complexity.
- We are keeping rollback simple if real-world latency is worse.

The only acceptable outcome is latency staying the same or improving. If verification suggests slower responses, the change should be rolled back immediately.

## Rollout Plan

1. Back up the current production `config.yaml`.
2. Update production `config.yaml` with explicit gateway capacity settings.
3. Restart the application service only.
4. Verify application health and public site availability.
5. Run a real API request check.
6. Inspect logs for transport/client cache issues, dropped tasks, or unexpected upstream failures.

## Verification Plan

The change is considered successful only if all of the following are true:

- `GET /health` returns healthy.
- `https://bridgemind.pro/` loads correctly.
- `https://bridgemind.pro/api/v1/settings/public` returns successfully.
- A real API request succeeds.
- No obvious increase in first-byte or end-to-end latency is observed.
- No new warnings suggest upstream client cache churn, queue drops, or transport instability.

## Rollback Plan

If latency, stability, or correctness regresses:

1. Restore the previous `config.yaml`.
2. Restart the application container.
3. Re-run the same health and API checks.

Rollback must be possible within a single config restore and service restart.

## Risks

- If some traffic patterns depend on stricter per-account transport isolation, shared proxy-level pooling could reveal edge cases.
- If production traffic is dominated by direct connections with very uneven proxy/account distribution, gains may be smaller than expected.
- If undocumented production overrides exist outside `config.yaml`, explicit settings must be checked carefully before rollout.

## Decision

Proceed with a guarded production config optimization that prioritizes preserving response speed while improving concurrency headroom.
