package repository

import (
	"context"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestRFC3339TimestampPatternAllowsValidLeapDay(t *testing.T) {
	re := regexp.MustCompile(rfc3339TimestampPattern())

	require.True(t, re.MatchString("2028-02-29T00:00:00Z"))
	require.True(t, re.MatchString("2026-02-28T23:59:59+08:00"))
	require.False(t, re.MatchString("2026-02-29T00:00:00Z"))
	require.False(t, re.MatchString("2026-04-31T00:00:00Z"))
}

func TestSetSchedulableTrueClearsExpiredAutoPauseSQL(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := newAccountRepositoryWithSQL(nil, db, nil)

	mock.ExpectExec(`(?s)UPDATE accounts\s+SET schedulable = \$1,\s+auto_pause_on_expired = CASE\s+WHEN \$1 = TRUE AND expires_at IS NOT NULL AND expires_at <= NOW\(\) THEN FALSE\s+ELSE auto_pause_on_expired\s+END,\s+updated_at = NOW\(\)\s+WHERE id = \$2\s+AND deleted_at IS NULL`).
		WithArgs(true, int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)INSERT INTO scheduler_outbox.*WHERE NOT EXISTS`).
		WithArgs(service.SchedulerOutboxEventAccountChanged, sqlmock.AnyArg(), nil, nil, schedulerOutboxDedupWindow.Seconds()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.SetSchedulable(context.Background(), 42, true))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkUpdateSchedulableTrueClearsExpiredAutoPauseSQL(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := newAccountRepositoryWithSQL(nil, db, nil)
	ids := []int64{1, 2}
	enable := true

	mock.ExpectExec(`(?s)UPDATE accounts SET schedulable = \$1, auto_pause_on_expired = CASE\s+WHEN expires_at IS NOT NULL AND expires_at <= NOW\(\) THEN FALSE\s+ELSE auto_pause_on_expired\s+END, updated_at = NOW\(\) WHERE id = ANY\(\$2\) AND deleted_at IS NULL`).
		WithArgs(true, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(`(?s)INSERT INTO scheduler_outbox`).
		WithArgs(service.SchedulerOutboxEventAccountBulkChanged, nil, nil, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rows, err := repo.BulkUpdate(context.Background(), ids, service.AccountBulkUpdate{Schedulable: &enable})
	require.NoError(t, err)
	require.Equal(t, int64(2), rows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestIncrementQuotaUsedDailyCrossingEnqueuesOutbox(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := newAccountRepositoryWithSQL(nil, db, nil)

	mock.ExpectQuery(`(?s)UPDATE accounts SET extra = .*RETURNING.*quota_daily_used.*quota_daily_limit.*quota_weekly_used.*quota_weekly_limit`).
		WithArgs(8.0, int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{
			"quota_used",
			"quota_limit",
			"quota_daily_used",
			"quota_daily_limit",
			"quota_weekly_used",
			"quota_weekly_limit",
		}).AddRow(12.0, 0.0, 12.0, 10.0, 0.0, 0.0))
	mock.ExpectExec(`(?s)INSERT INTO scheduler_outbox`).
		WithArgs(service.SchedulerOutboxEventAccountChanged, sqlmock.AnyArg(), nil, nil, schedulerOutboxDedupWindow.Seconds()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.IncrementQuotaUsed(context.Background(), 42, 8.0))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestIncrementQuotaUsedWeeklyCrossingEnqueuesOutbox(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := newAccountRepositoryWithSQL(nil, db, nil)

	mock.ExpectQuery(`(?s)UPDATE accounts SET extra = .*RETURNING.*quota_daily_used.*quota_daily_limit.*quota_weekly_used.*quota_weekly_limit`).
		WithArgs(15.0, int64(43)).
		WillReturnRows(sqlmock.NewRows([]string{
			"quota_used",
			"quota_limit",
			"quota_daily_used",
			"quota_daily_limit",
			"quota_weekly_used",
			"quota_weekly_limit",
		}).AddRow(15.0, 0.0, 0.0, 0.0, 15.0, 10.0))
	mock.ExpectExec(`(?s)INSERT INTO scheduler_outbox`).
		WithArgs(service.SchedulerOutboxEventAccountChanged, sqlmock.AnyArg(), nil, nil, schedulerOutboxDedupWindow.Seconds()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.IncrementQuotaUsed(context.Background(), 43, 15.0))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestIncrementQuotaUsedBelowLimitDoesNotEnqueueOutbox(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := newAccountRepositoryWithSQL(nil, db, nil)

	mock.ExpectQuery(`(?s)UPDATE accounts SET extra = .*RETURNING.*quota_daily_used.*quota_daily_limit.*quota_weekly_used.*quota_weekly_limit`).
		WithArgs(4.0, int64(44)).
		WillReturnRows(sqlmock.NewRows([]string{
			"quota_used",
			"quota_limit",
			"quota_daily_used",
			"quota_daily_limit",
			"quota_weekly_used",
			"quota_weekly_limit",
		}).AddRow(4.0, 0.0, 4.0, 10.0, 4.0, 20.0))

	require.NoError(t, repo.IncrementQuotaUsed(context.Background(), 44, 4.0))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMergeGroupIDsIncludesUngroupedTransitions(t *testing.T) {
	require.Equal(t, []int64{0, 12}, mergeGroupIDs(nil, []int64{12}))
	require.Equal(t, []int64{0, 12}, mergeGroupIDs([]int64{12}, nil))
	require.Equal(t, []int64{12, 13}, mergeGroupIDs([]int64{12}, []int64{13}))
}

func TestRemoveGroupID(t *testing.T) {
	require.Equal(t, []int64{12, 14}, removeGroupID([]int64{12, 13, 14}, 13))
	require.Empty(t, removeGroupID([]int64{13}, 13))
}
