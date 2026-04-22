# Capacity Headroom Observability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a lightweight capacity headroom view to the ops dashboard so operators can see upstream client cache, usage-record queue, concurrency, and DB/Redis headroom without increasing dashboard query cost.

**Architecture:** Keep the new signal path runtime-backed and compact. Expose a narrow set of runtime stats from the upstream client cache and usage-record worker pool, aggregate those with existing system/concurrency metrics inside the ops service, extend the dashboard overview contract, and render a focused `Capacity Headroom` card in the existing ops dashboard.

**Tech Stack:** Go services and Wire DI, Vue 3 + TypeScript + Vitest, existing ops dashboard APIs, PostgreSQL/Redis-backed ops data with runtime in-memory stats

---

### Task 1: Add Backend Runtime Capacity Stats Providers

**Files:**
- Create: `/mnt/c/web/code/openapi/backend/internal/service/ops_capacity_headroom.go`
- Create: `/mnt/c/web/code/openapi/backend/internal/service/ops_capacity_headroom_test.go`
- Modify: `/mnt/c/web/code/openapi/backend/internal/service/ops_service.go`
- Modify: `/mnt/c/web/code/openapi/backend/internal/service/wire.go`
- Modify: `/mnt/c/web/code/openapi/backend/cmd/server/wire.go`
- Modify: `/mnt/c/web/code/openapi/backend/cmd/server/wire_gen.go`
- Modify: `/mnt/c/web/code/openapi/backend/internal/repository/http_upstream.go`
- Modify: `/mnt/c/web/code/openapi/backend/internal/service/usage_record_worker_pool.go`
- Test: `/mnt/c/web/code/openapi/backend/internal/repository/http_upstream_test.go`
- Test: `/mnt/c/web/code/openapi/backend/internal/service/usage_record_worker_pool_test.go`

- [ ] **Step 1: Write failing tests for the new runtime stats shapes**

Add backend tests that describe the new runtime stats contracts before implementation:

```go
func TestBuildCapacityHeadroom_UsesRuntimeStats(t *testing.T) {
    overview := &OpsDashboardOverview{SystemMetrics: &OpsSystemMetricsSnapshot{}}
    upstream := opsUpstreamRuntimeStats{CurrentClients: 120, MaxClients: 5000}
    usage := UsageRecordWorkerPoolStats{MaxConcurrency: 128, WaitingTasks: 3}

    got := buildCapacityHeadroom(overview, upstream, usage, nil)

    require.NotNil(t, got)
    require.Equal(t, 120, got.UpstreamClients.CurrentClients)
    require.Equal(t, 3, got.UsageRecordPool.WaitingTasks)
}
```

- [ ] **Step 2: Run the targeted backend tests and confirm they fail**

Run:

```bash
cd /mnt/c/web/code/openapi
go test ./backend/internal/service ./backend/internal/repository -run 'TestBuildCapacityHeadroom|TestHTTPUpstream|TestUsageRecordWorkerPool' -count=1
```

Expected:

- New tests fail because the runtime capacity stats types and aggregation functions do not exist yet.

- [ ] **Step 3: Add narrow runtime stats interfaces instead of changing the broad `HTTPUpstream` interface**

Implement a small runtime-only stats surface so existing mocks are not broken:

```go
type OpsUpstreamRuntimeStatsProvider interface {
    OpsRuntimeStats() OpsUpstreamRuntimeStats
}

type OpsUsageRecordRuntimeStatsProvider interface {
    Stats() UsageRecordWorkerPoolStats
}
```

Implementation notes:

- Put the capacity-headroom models and aggregation helpers in `ops_capacity_headroom.go`.
- Add a concrete `OpsRuntimeStats()` method on `httpUpstreamService`.
- Reuse the existing `Stats()` method on `UsageRecordWorkerPool`.
- Inject these providers into `OpsService` with setter methods if needed, so we do not have to widen unrelated constructors early.

- [ ] **Step 4: Wire the runtime providers into `OpsService`**

Update DI so the ops service can read the runtime stats:

```go
opsService := service.NewOpsService(...)
opsService.SetUpstreamRuntimeStatsProvider(httpUpstream)
opsService.SetUsageRecordRuntimeStatsProvider(usageRecordWorkerPool)
```

Run:

```bash
cd /mnt/c/web/code/openapi/backend/cmd/server
go generate .
```

Expected:

- `wire_gen.go` updates cleanly and compiles with the new provider wiring.

- [ ] **Step 5: Run the focused backend tests again**

Run:

```bash
cd /mnt/c/web/code/openapi
go test ./backend/internal/service ./backend/internal/repository -run 'TestBuildCapacityHeadroom|TestHTTPUpstream|TestUsageRecordWorkerPool' -count=1
```

Expected:

- The new runtime stats tests pass.
- Existing upstream and worker-pool tests still pass.

- [ ] **Step 6: Commit the runtime stats backend slice**

Run:

```bash
cd /mnt/c/web/code/openapi
git add \
  backend/internal/service/ops_capacity_headroom.go \
  backend/internal/service/ops_capacity_headroom_test.go \
  backend/internal/service/ops_service.go \
  backend/internal/service/wire.go \
  backend/cmd/server/wire.go \
  backend/cmd/server/wire_gen.go \
  backend/internal/repository/http_upstream.go \
  backend/internal/repository/http_upstream_test.go \
  backend/internal/service/usage_record_worker_pool.go \
  backend/internal/service/usage_record_worker_pool_test.go
git commit -m "feat: expose runtime capacity stats for ops"
```

### Task 2: Extend Ops Overview With Capacity Headroom

**Files:**
- Modify: `/mnt/c/web/code/openapi/backend/internal/service/ops_dashboard_models.go`
- Modify: `/mnt/c/web/code/openapi/backend/internal/service/ops_dashboard.go`
- Modify: `/mnt/c/web/code/openapi/backend/internal/service/ops_port.go`
- Test: `/mnt/c/web/code/openapi/backend/internal/service/ops_repo_mock_test.go`
- Test: `/mnt/c/web/code/openapi/backend/internal/service/ops_health_score_test.go`
- Test: `/mnt/c/web/code/openapi/backend/internal/service/ops_capacity_headroom_test.go`

- [ ] **Step 1: Write failing tests for overview contract extension**

Add tests that prove the dashboard overview now includes `capacity_headroom`:

```go
func TestGetDashboardOverview_AttachesCapacityHeadroom(t *testing.T) {
    svc := newOpsServiceForTest(...)
    overview, err := svc.GetDashboardOverview(ctx, filter)
    require.NoError(t, err)
    require.NotNil(t, overview.CapacityHeadroom)
    require.NotNil(t, overview.CapacityHeadroom.UpstreamClients)
}
```

- [ ] **Step 2: Run the new overview tests and confirm failure**

Run:

```bash
cd /mnt/c/web/code/openapi
go test ./backend/internal/service -run 'TestGetDashboardOverview_AttachesCapacityHeadroom|TestBuildCapacityHeadroom' -count=1
```

Expected:

- Tests fail because the overview payload does not yet include the new field.

- [ ] **Step 3: Add the new DTOs and attach them in `GetDashboardOverview`**

Implement the contract with compact UI-friendly fields:

```go
type OpsCapacityHeadroom struct {
    UpstreamClients *OpsCapacityHeadroomUpstreamClients `json:"upstream_clients,omitempty"`
    UsageRecordPool *OpsCapacityHeadroomUsageRecordPool `json:"usage_record_pool,omitempty"`
    Concurrency     *OpsCapacityHeadroomConcurrency     `json:"concurrency,omitempty"`
    DB              *OpsCapacityHeadroomDB             `json:"db,omitempty"`
    Redis           *OpsCapacityHeadroomRedis          `json:"redis,omitempty"`
}
```

Implementation notes:

- Keep this in service-layer models used by the dashboard overview.
- Build from existing system metrics plus the new runtime providers.
- Fail open: if runtime stats are unavailable, overview still renders.

- [ ] **Step 4: Re-run focused backend tests**

Run:

```bash
cd /mnt/c/web/code/openapi
go test ./backend/internal/service -run 'TestGetDashboardOverview_AttachesCapacityHeadroom|TestBuildCapacityHeadroom|TestComputeDashboardHealthScore' -count=1
```

Expected:

- Overview tests pass.
- Existing health-score tests remain green.

- [ ] **Step 5: Commit the overview contract slice**

Run:

```bash
cd /mnt/c/web/code/openapi
git add \
  backend/internal/service/ops_dashboard_models.go \
  backend/internal/service/ops_dashboard.go \
  backend/internal/service/ops_port.go \
  backend/internal/service/ops_repo_mock_test.go \
  backend/internal/service/ops_health_score_test.go \
  backend/internal/service/ops_capacity_headroom_test.go
git commit -m "feat: add capacity headroom to ops overview"
```

### Task 3: Render the Capacity Headroom Card in the Ops Dashboard

**Files:**
- Create: `/mnt/c/web/code/openapi/frontend/src/views/admin/ops/components/OpsCapacityHeadroomCard.vue`
- Create: `/mnt/c/web/code/openapi/frontend/src/views/admin/ops/components/__tests__/OpsCapacityHeadroomCard.spec.ts`
- Modify: `/mnt/c/web/code/openapi/frontend/src/api/admin/ops.ts`
- Modify: `/mnt/c/web/code/openapi/frontend/src/views/admin/ops/OpsDashboard.vue`
- Modify: `/mnt/c/web/code/openapi/frontend/src/views/admin/ops/components/OpsDashboardHeader.vue`
- Modify: `/mnt/c/web/code/openapi/frontend/src/i18n/locales/en.ts`

- [ ] **Step 1: Write the failing frontend component test**

Create a focused spec for the new card:

```ts
it('renders headroom rows with usage and remaining capacity', async () => {
  const wrapper = mount(OpsCapacityHeadroomCard, {
    props: {
      capacityHeadroom: {
        upstream_clients: { current_clients: 120, max_clients: 5000, usage_percent: 2.4, remaining_clients: 4880 },
        usage_record_pool: { running_workers: 8, max_concurrency: 128, waiting_tasks: 0, queue_size: 16384, queue_usage_percent: 0 },
      },
    },
  })

  expect(wrapper.text()).toContain('120 / 5000')
  expect(wrapper.text()).toContain('4880')
})
```

- [ ] **Step 2: Run the new frontend test and confirm failure**

Run:

```bash
cd /mnt/c/web/code/openapi
pnpm --dir frontend exec vitest run src/views/admin/ops/components/__tests__/OpsCapacityHeadroomCard.spec.ts
```

Expected:

- Test fails because the component and contract types do not exist yet.

- [ ] **Step 3: Extend the frontend API types and build the card**

Implementation notes:

- Add `capacity_headroom` types to `frontend/src/api/admin/ops.ts`.
- Create `OpsCapacityHeadroomCard.vue` as a compact status card, not a chart.
- Use one row per domain: upstream clients, usage queue, concurrency, DB, Redis.
- Show `current / max`, usage percent, and remaining headroom where meaningful.
- Use the same green/yellow/red status language already used in `OpsDashboardHeader.vue`.

- [ ] **Step 4: Mount the card near the top of the existing dashboard**

Recommended placement:

- Keep `OpsDashboardHeader.vue` focused on summary chips.
- Insert the new card in `OpsDashboard.vue` just below the header and before the main grid rows.

Example target shape:

```vue
<OpsCapacityHeadroomCard
  v-if="opsEnabled && !(loading && !hasLoadedOnce)"
  :capacity-headroom="overview?.capacity_headroom ?? null"
/>
```

- [ ] **Step 5: Run frontend tests and typecheck**

Run:

```bash
cd /mnt/c/web/code/openapi
pnpm --dir frontend exec vitest run src/views/admin/ops/components/__tests__/OpsCapacityHeadroomCard.spec.ts
pnpm --dir frontend run typecheck
```

Expected:

- New card spec passes.
- Frontend typecheck passes.

- [ ] **Step 6: Commit the frontend slice**

Run:

```bash
cd /mnt/c/web/code/openapi
git add \
  frontend/src/views/admin/ops/components/OpsCapacityHeadroomCard.vue \
  frontend/src/views/admin/ops/components/__tests__/OpsCapacityHeadroomCard.spec.ts \
  frontend/src/api/admin/ops.ts \
  frontend/src/views/admin/ops/OpsDashboard.vue \
  frontend/src/views/admin/ops/components/OpsDashboardHeader.vue \
  frontend/src/i18n/locales/en.ts
git commit -m "feat: add ops capacity headroom card"
```

### Task 4: Run Integrated Regression Checks

**Files:**
- Inspect: `/mnt/c/web/code/openapi/backend/internal/service/ops_dashboard.go`
- Inspect: `/mnt/c/web/code/openapi/frontend/src/views/admin/ops/OpsDashboard.vue`

- [ ] **Step 1: Run the targeted backend regression suite**

Run:

```bash
cd /mnt/c/web/code/openapi
go test ./backend/internal/service ./backend/internal/repository -run 'TestGetDashboardOverview|TestBuildCapacityHeadroom|TestHTTPUpstream|TestUsageRecordWorkerPool|TestComputeDashboardHealthScore' -count=1
```

Expected:

- All targeted backend tests pass.

- [ ] **Step 2: Run the targeted frontend regression suite**

Run:

```bash
cd /mnt/c/web/code/openapi
pnpm --dir frontend exec vitest run \
  src/views/admin/ops/components/__tests__/OpsCapacityHeadroomCard.spec.ts \
  src/views/admin/ops/components/__tests__/OpsOpenAITokenStatsCard.spec.ts
```

Expected:

- The new capacity-headroom card spec passes.
- Existing ops component spec still passes.

- [ ] **Step 3: Build confidence with a manual local smoke check**

Run:

```bash
cd /mnt/c/web/code/openapi
pnpm --dir frontend run build
go test ./backend/internal/service -run TestGetDashboardOverview_AttachesCapacityHeadroom -count=1
```

Expected:

- Frontend builds successfully.
- Backend overview test passes with the new contract.

- [ ] **Step 4: Final commit**

Run:

```bash
cd /mnt/c/web/code/openapi
git status --short
git add backend frontend
git commit -m "feat: surface ops capacity headroom"
```

Expected:

- Only intended implementation files are committed.
