package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestAddWeeklyUsageWithoutAnyData tests adding a new daily usage into the weekly usage if
// it is not existed
func TestAddWeeklyUsageWithoutAnyData(t *testing.T) {
	now := time.Now()
	u := WeeklyUsage{
		Data: map[time.Weekday]DailyUsage{},
	}
	u.CountRequests(5)

	assert.Equal(t, 5, u.Data[now.Weekday()][now.Hour()])
}

// TestAddWeeklyUsageWithOldData tests if the weekly usage contains the data of
// the same weekday in the same week, it will be added up
func TestAddWeeklyUsageWithOldData(t *testing.T) {
	now := time.Now()
	u := WeeklyUsage{
		Today: now,
		Data:  map[time.Weekday]DailyUsage{},
	}
	u.Data[now.Weekday()] = DailyUsage{
		now.Hour(): 3,
	}
	u.CountRequests(5)

	assert.Equal(t, 8, u.Data[now.Weekday()][now.Hour()])
}

// TestAddWeeklyUsageAcrossNextDay tests that the daily usage data
// will be removed if the usage data comes from a new day
func TestAddWeeklyUsageAcrossNextDay(t *testing.T) {
	now := time.Now()
	u := WeeklyUsage{
		Today: now.Add(-24 * time.Hour),
		Data: map[time.Weekday]DailyUsage{
			now.Weekday(): {
				now.Hour(): 3,
			},
		},
	}

	u.CountRequests(1)
	assert.Less(t, now.Sub(u.Today), 24*time.Hour)
	assert.Equal(t, 1, u.Data[now.Weekday()][now.Hour()])
}

// TestAddMonthlyUsageWithoutAnyData tests adding data to
// an empty monthly usage
func TestAddMonthlyUsageWithoutAnyData(t *testing.T) {
	now := time.Now()
	u := MonthlyUsage{}

	newDayTimestamp := 86400 * (now.Unix() / 86400)

	u.CountRequests(5)

	assert.Equal(t, 1, len(u.Data))
	assert.Equal(t, newDayTimestamp, u.Today.Unix())
	assert.Equal(t, 5, u.Data[len(u.Data)-1][now.Hour()])
}

// TestAddMonthlyUsageAcrossNextDay tests that a new row will be pushed
// into the usage data if a new day comes
func TestAddMonthlyUsageAcrossNextDay(t *testing.T) {
	now := time.Now()
	u := MonthlyUsage{
		Today: now.Add(-24 * time.Hour),
	}

	u.Data = []DailyUsage{
		{
			1: 1,
			2: 2,
		},
	}

	u.CountRequests(5)

	assert.Equal(t, 2, len(u.Data))
	assert.Less(t, now.Sub(u.Today), 24*time.Hour)
	assert.Equal(t, 5, u.Data[len(u.Data)-1][now.Hour()])
}

// TestAddMonthlyUsageCountWithinADay tests the requests data will be
// accmulated if the data is from the same day
func TestAddMonthlyUsageCountWithinADay(t *testing.T) {
	now := time.Now()
	u := MonthlyUsage{
		Today: now,
	}

	u.Data = []DailyUsage{
		{
			1: 1,
			2: 2,
		},
	}

	u.CountRequests(5)

	assert.Equal(t, 1, len(u.Data))
	assert.GreaterOrEqual(t, 5, u.Data[len(u.Data)-1][now.Hour()])
}

// TestAddMonthlyUsageAcrossNextDayRotate tests if the monthly data is more than
// 30 days, the oldest one will be removed.
func TestAddMonthlyUsageAcrossNextDayRotate(t *testing.T) {
	now := time.Now()
	u := MonthlyUsage{
		Today: now.Add(-24 * time.Hour),
	}

	u.Data = []DailyUsage{}

	for i := 0; i < 30; i++ {
		u.Data = append(u.Data, DailyUsage{1: i + 1})
	}

	u.CountRequests(5)

	assert.Equal(t, 30, len(u.Data))
	assert.Equal(t, 2, u.Data[0][1])
	assert.Equal(t, 5, u.Data[len(u.Data)-1][now.Hour()])
}
