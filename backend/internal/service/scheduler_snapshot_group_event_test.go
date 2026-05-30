//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type schedulerSnapshotGroupEventCache struct {
	buckets []SchedulerBucket
}

func (c *schedulerSnapshotGroupEventCache) GetSnapshot(context.Context, SchedulerBucket) ([]*Account, bool, error) {
	return nil, false, nil
}

func (c *schedulerSnapshotGroupEventCache) SetSnapshot(_ context.Context, bucket SchedulerBucket, _ []Account) error {
	c.buckets = append(c.buckets, bucket)
	return nil
}

func (c *schedulerSnapshotGroupEventCache) GetAccount(context.Context, int64) (*Account, error) {
	return nil, nil
}

func (c *schedulerSnapshotGroupEventCache) SetAccount(context.Context, *Account) error {
	return nil
}

func (c *schedulerSnapshotGroupEventCache) DeleteAccount(context.Context, int64) error {
	return nil
}

func (c *schedulerSnapshotGroupEventCache) UpdateLastUsed(context.Context, map[int64]time.Time) error {
	return nil
}

func (c *schedulerSnapshotGroupEventCache) TryLockBucket(context.Context, SchedulerBucket, time.Duration) (bool, error) {
	return true, nil
}

func (c *schedulerSnapshotGroupEventCache) UnlockBucket(context.Context, SchedulerBucket) error {
	return nil
}

func (c *schedulerSnapshotGroupEventCache) ListBuckets(context.Context) ([]SchedulerBucket, error) {
	return nil, nil
}

func (c *schedulerSnapshotGroupEventCache) GetOutboxWatermark(context.Context) (int64, error) {
	return 0, nil
}

func (c *schedulerSnapshotGroupEventCache) SetOutboxWatermark(context.Context, int64) error {
	return nil
}

type schedulerSnapshotGroupEventRepo struct {
	accountRepoStub
	account *Account
}

func (r *schedulerSnapshotGroupEventRepo) GetByID(context.Context, int64) (*Account, error) {
	return r.account, nil
}

func (r *schedulerSnapshotGroupEventRepo) ListSchedulableByGroupIDAndPlatform(context.Context, int64, string) ([]Account, error) {
	return []Account{}, nil
}

func (r *schedulerSnapshotGroupEventRepo) ListSchedulableByGroupIDAndPlatforms(context.Context, int64, []string) ([]Account, error) {
	return []Account{}, nil
}

func (r *schedulerSnapshotGroupEventRepo) ListSchedulableUngroupedByPlatform(context.Context, string) ([]Account, error) {
	return []Account{}, nil
}

func (r *schedulerSnapshotGroupEventRepo) ListSchedulableUngroupedByPlatforms(context.Context, []string) ([]Account, error) {
	return []Account{}, nil
}

func TestSchedulerSnapshotAccountGroupEventRebuildsUngroupedAfterUnbind(t *testing.T) {
	cache := &schedulerSnapshotGroupEventCache{}
	repo := &schedulerSnapshotGroupEventRepo{
		account: &Account{
			ID:          7,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			GroupIDs:    nil,
		},
	}
	svc := NewSchedulerSnapshotService(cache, nil, repo, nil, nil)

	accountID := int64(7)
	err := svc.handleAccountEvent(context.Background(), &accountID, map[string]any{
		"group_ids": []any{float64(12)},
	}, nil)

	require.NoError(t, err)
	require.Contains(t, cache.buckets, SchedulerBucket{GroupID: 12, Platform: PlatformAnthropic, Mode: SchedulerModeSingle})
	require.Contains(t, cache.buckets, SchedulerBucket{GroupID: 0, Platform: PlatformAnthropic, Mode: SchedulerModeSingle})
}

func TestSchedulerSnapshotAccountGroupEventKeepsExplicitUngroupedBucket(t *testing.T) {
	cache := &schedulerSnapshotGroupEventCache{}
	repo := &schedulerSnapshotGroupEventRepo{
		account: &Account{
			ID:          8,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			GroupIDs:    []int64{12},
		},
	}
	svc := NewSchedulerSnapshotService(cache, nil, repo, nil, nil)

	accountID := int64(8)
	err := svc.handleAccountEvent(context.Background(), &accountID, map[string]any{
		"group_ids": []any{float64(0), float64(12)},
	}, nil)

	require.NoError(t, err)
	require.Contains(t, cache.buckets, SchedulerBucket{GroupID: 0, Platform: PlatformAnthropic, Mode: SchedulerModeSingle})
	require.Contains(t, cache.buckets, SchedulerBucket{GroupID: 12, Platform: PlatformAnthropic, Mode: SchedulerModeSingle})
}
