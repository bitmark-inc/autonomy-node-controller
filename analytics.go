package main

import (
	"sync"
	"time"
)

// DailyUsage is a map for requests counts of each hour
type DailyUsage map[int]int

type Usage struct {
	Weekly  WeeklyUsage  `json:"weekly"`
	Monthly MonthlyUsage `json:"Monthly"`
}

// CountRequests counts requests for usage analytic
func (u *Usage) CountRequests(n int) {
	// Increment the hourly usage of last week
	u.Weekly.CountRequests(n)
	u.Monthly.CountRequests(n)
}

// WeeklyUsage is a structure keeps the usage data for a week
type WeeklyUsage struct {
	sync.Mutex
	Today time.Time                   `json:"today"`
	Data  map[time.Weekday]DailyUsage `json:"data"`
}

// CountRequests counts requests for weekly usage data
func (w *WeeklyUsage) CountRequests(count int) {
	w.Lock()
	defer w.Unlock()

	now := time.Now()

	if now.Sub(w.Today) > 24*time.Hour {
		w.Data[time.Now().Weekday()] = DailyUsage{}
		w.Today = time.Unix(86400*(now.Unix()/86400), 0)
	}

	if _, ok := w.Data[time.Now().Weekday()]; !ok {
		w.Data[time.Now().Weekday()] = DailyUsage{}
	}

	w.Data[time.Now().Weekday()][time.Now().Hour()] += count
}

// MonthlyUsage is a structure keeps the usage data for the past 30 days
type MonthlyUsage struct {
	sync.Mutex
	Today time.Time    `json:"today"`
	Data  []DailyUsage `json:"data"`
}

// CountRequests counts requests for monthly usage data
func (u *MonthlyUsage) CountRequests(count int) {
	u.Lock()
	defer u.Unlock()

	now := time.Now()
	// push a new row if it is a new day
	if now.Sub(u.Today) > 24*time.Hour {
		if len(u.Data) >= 30 {
			u.Data = u.Data[1:]
		}
		u.Data = append(u.Data, DailyUsage{})
		u.Today = time.Unix(86400*(now.Unix()/86400), 0)
	}

	lastDayUsage := u.Data[len(u.Data)-1]
	lastDayUsage[now.Hour()] += count
}
