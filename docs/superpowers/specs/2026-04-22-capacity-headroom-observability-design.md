# Capacity Headroom Observability Design

Date: 2026-04-22
Scope: `bridgemind.pro` API/backend observability enhancement for capacity headroom
Status: Approved for implementation after user review

## Background

Recent production work improved API capacity headroom by making gateway pool settings explicit and switching upstream connection-pool isolation to `proxy`. That change improved the system's capacity posture, but the current ops dashboard still makes it easier to see that the system is busy than to understand how much safe headroom remains.

Today, the dashboard already exposes useful signals such as:

- DB connection usage and max open connections
- Redis connection usage and configured pool size
- Goroutine count
- Total concurrency queue depth
- Platform/group/account concurrency and availability

However, the dashboard still lacks a focused view that answers:

- How close the system is to exhausting upstream client cache capacity
- Whether the async usage-record pipeline is building backlog
- Whether concurrency pressure is account-level, queue-level, or system-level
- Which layer is most likely to become the next scaling bottleneck

## Goals

- Add a compact, high-signal "capacity headroom" view to the ops dashboard.
- Expose the most actionable runtime capacity signals without adding heavy database queries.
- Reuse existing ops infrastructure and patterns wherever possible.
- Keep page latency stable and avoid introducing monitoring overhead that harms production traffic.

## Non-Goals

- No rebuild of the entire ops dashboard.
- No new external monitoring stack in this phase.
- No advanced predictive forecasting or autoscaling automation in this phase.
- No wide alerting redesign in this phase.

## Current Technical Findings

### Existing dashboard strengths

The current dashboard and services already provide:

- Overview snapshots and cached dashboard data
- A concurrency card with platform/group/account availability and load
- DB and Redis connection pool usage on the header
- System health and background job heartbeat visibility

Relevant files:

- [OpsDashboard.vue](/mnt/c/web/code/openapi/frontend/src/views/admin/ops/OpsDashboard.vue)
- [OpsDashboardHeader.vue](/mnt/c/web/code/openapi/frontend/src/views/admin/ops/components/OpsDashboardHeader.vue)
- [ops_dashboard.go](/mnt/c/web/code/openapi/backend/internal/service/ops_dashboard.go)
- [ops_metrics_collector.go](/mnt/c/web/code/openapi/backend/internal/service/ops_metrics_collector.go)

### Existing gap: capacity signals are incomplete

The current system metrics snapshot already includes:

- `db_conn_active`, `db_conn_idle`, `db_conn_waiting`
- `db_max_open_conns`
- `redis_conn_total`, `redis_conn_idle`
- `redis_pool_size`
- `goroutine_count`
- `concurrency_queue_depth`

But it does not yet expose two of the most important runtime headroom signals:

1. Upstream client cache usage versus `gateway.max_upstream_clients`
2. Usage record worker pool runtime state, such as:
   - running workers
   - max concurrency
   - waiting tasks
   - queue size
   - sync fallback count
   - dropped queue-full count

These two signals are especially important because they indicate whether the system is quietly approaching operational limits even when request success rate still looks healthy.

### Existing gap: "busy" is visible, "remaining capacity" is not

The current UI exposes raw current values, but it does not consistently compute or present:

- percentage used
- percentage remaining
- warning/critical states aligned with capacity semantics
- a single place that compares the most capacity-sensitive subsystems side by side

## Options Considered

### Option 1: Frontend-only summary using existing fields

Pros:

- Lowest engineering risk
- No backend contract changes

Cons:

- Cannot show upstream client cache usage
- Cannot show usage record worker-pool backlog
- Misses the most important newly relevant capacity signals

### Option 2: Add a lightweight runtime capacity snapshot and render one new dashboard card

Pros:

- Best ratio of insight to implementation cost
- Uses mostly in-memory/runtime stats instead of heavy SQL
- Gives operators one place to judge "how much headroom is left"

Cons:

- Requires coordinated backend and frontend changes
- Needs careful API design to stay compact and stable

### Option 3: Skip UI improvements and only add capacity alert rules

Pros:

- Helps operators react sooner

Cons:

- Thresholds would be harder to tune without first seeing the signals clearly
- Gives less day-to-day capacity visibility

## Chosen Design

Use Option 2.

We will add a lightweight runtime capacity snapshot to the ops data model and render a compact `Capacity Headroom` card in the existing ops dashboard.

## Architecture

### Backend design

Add a new runtime-oriented capacity summary object to the ops service layer. This object should be attached to the dashboard overview response so the frontend can fetch it together with existing overview data.

The new summary should aggregate four capacity domains:

1. **Upstream Clients**
   - current cached upstream clients
   - configured `max_upstream_clients`
   - utilization percent
   - remaining client slots

2. **Usage Record Pool**
   - running workers
   - max concurrency
   - waiting tasks
   - configured queue size
   - waiting percent
   - dropped queue-full count
   - sync fallback count

3. **Concurrency Pressure**
   - aggregate used concurrency
   - aggregate max concurrency
   - overall utilization percent
   - total waiting queue depth
   - account availability ratio if available from existing concurrency/availability data

4. **DB / Redis Pool Headroom**
   - DB open connections versus max open
   - DB waiting count
   - Redis total connections versus pool size
   - derived utilization percent

This summary should be computed from runtime state and existing already-fetched overview/system metrics whenever possible. It should avoid new expensive queries on hot dashboard paths.

### Backend data sources

Expected primary data sources:

- Existing system metrics snapshot and config-derived limits in [ops_dashboard.go](/mnt/c/web/code/openapi/backend/internal/service/ops_dashboard.go)
- Runtime DB and Redis pool stats already persisted via [ops_metrics_collector.go](/mnt/c/web/code/openapi/backend/internal/service/ops_metrics_collector.go)
- Usage record pool runtime stats from [usage_record_worker_pool.go](/mnt/c/web/code/openapi/backend/internal/service/usage_record_worker_pool.go)
- Upstream client cache stats from [http_upstream.go](/mnt/c/web/code/openapi/backend/internal/repository/http_upstream.go)
- Existing concurrency and availability data already used by the dashboard

### Proposed backend contract

Add a new object on the dashboard overview payload:

```json
{
  "capacity_headroom": {
    "upstream_clients": {
      "current_clients": 120,
      "max_clients": 5000,
      "usage_percent": 2.4,
      "remaining_clients": 4880
    },
    "usage_record_pool": {
      "running_workers": 8,
      "max_concurrency": 128,
      "waiting_tasks": 0,
      "queue_size": 16384,
      "queue_usage_percent": 0,
      "dropped_queue_full": 0,
      "sync_fallback_tasks": 0
    },
    "concurrency": {
      "used_capacity": 37,
      "max_capacity": 145,
      "usage_percent": 25.5,
      "waiting_in_queue": 0,
      "availability_percent": 94
    },
    "db": {
      "open_connections": 11,
      "max_open_connections": 256,
      "usage_percent": 4.3,
      "waiting": 0
    },
    "redis": {
      "total_connections": 14,
      "pool_size": 4096,
      "usage_percent": 0.3
    }
  }
}
```

Field names may differ slightly during implementation, but the contract should stay compact and directly UI-friendly.

### Frontend design

Add a new `Capacity Headroom` card near the top of the ops dashboard, adjacent to the existing overview area rather than as a new page.

The card should:

- show one row per capacity domain
- display `current / max`
- show percentage used
- show remaining headroom when meaningful
- use the existing status color vocabulary:
  - green for safe
  - yellow for warning
  - red for critical
- degrade gracefully when a signal is unavailable

The card should prioritize operator readability over chart complexity. This is a status card, not a trend chart.

### Status thresholds

The first implementation should use simple local thresholds:

- **Safe:** under 70%
- **Warning:** 70% to under 90%
- **Critical:** 90% and above

Special handling:

- DB waiting count greater than 0 should elevate severity
- Usage record dropped tasks greater than 0 should elevate severity
- Concurrency waiting queue above 0 should at least surface as warning when sustained

Thresholds can later become configurable if needed, but they do not need to be admin-configurable in this phase.

## File/Module Impact

Likely backend files to modify:

- [ops_port.go](/mnt/c/web/code/openapi/backend/internal/service/ops_port.go)
- [ops_dashboard.go](/mnt/c/web/code/openapi/backend/internal/service/ops_dashboard.go)
- [ops_metrics_collector.go](/mnt/c/web/code/openapi/backend/internal/service/ops_metrics_collector.go)
- [usage_record_worker_pool.go](/mnt/c/web/code/openapi/backend/internal/service/usage_record_worker_pool.go)
- [http_upstream.go](/mnt/c/web/code/openapi/backend/internal/repository/http_upstream.go)

Likely frontend files to modify:

- [ops.ts](/mnt/c/web/code/openapi/frontend/src/api/admin/ops.ts)
- [OpsDashboard.vue](/mnt/c/web/code/openapi/frontend/src/views/admin/ops/OpsDashboard.vue)
- [OpsDashboardHeader.vue](/mnt/c/web/code/openapi/frontend/src/views/admin/ops/components/OpsDashboardHeader.vue)

If a new focused component is created, prefer a dedicated file such as:

- `frontend/src/views/admin/ops/components/OpsCapacityHeadroomCard.vue`

## Performance Constraints

- No heavy new SQL should be added to the dashboard hot path.
- Runtime stats should be read from in-memory services or already-collected snapshots whenever possible.
- The dashboard should not require additional slow requests for initial render if the new object can be included in the existing overview response.
- If a live runtime stats read is needed, it should be best-effort and fail-open.

## Error Handling

- Missing runtime stats must not break the dashboard.
- When a signal is unavailable, show `-` or `No data` rather than treating it as failure.
- If the runtime-only portion fails, the rest of the overview must still render.

## Testing Plan

### Backend

- Unit tests for capacity summary calculation:
  - normal values
  - missing values
  - zero-capacity guards
  - waiting/dropped-task severity escalation
- Tests for upstream cache stats exposure
- Tests for usage record pool stats exposure

### Frontend

- Component tests for:
  - normal rendering
  - warning/critical state rendering
  - missing data rendering
- Contract compatibility tests if existing ops overview types are extended

### Manual verification

- Open admin ops dashboard and confirm the new card renders
- Verify the card values are consistent with existing DB/Redis/concurrency cards
- Verify dashboard fetch latency does not materially regress
- Verify no production API traffic path is affected

## Risks

- If runtime services do not currently expose stats cleanly, the first implementation may require small adapter methods.
- If the overview payload grows too much, the dashboard response could become noisier than necessary.
- If thresholds are too aggressive, the new card could create visual fatigue. The initial version should stay conservative.

## Rollout

1. Implement backend runtime capacity summary
2. Extend overview contract and frontend types
3. Add the capacity headroom card
4. Run tests
5. Deploy and verify dashboard render plus latency

## Decision

Proceed with a lightweight runtime-backed `Capacity Headroom` card that makes system remaining capacity visible without adding heavy monitoring overhead.
