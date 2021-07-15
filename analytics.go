// SPDX-License-Identifier: ISC
// Copyright (c) 2019-2021 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

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

// CheckAndUpdateNewDay checks if it is a new day now and
// do corresponded updates.
func (u *Usage) CheckAndUpdateNewDay() {
	now := time.Now()
	u.Weekly.CheckAndUpdateNewDay(now)
	u.Monthly.CheckAndUpdateNewDay(now)
}

// WeeklyUsage is a structure keeps the usage data for a week
type WeeklyUsage struct {
	sync.Mutex
	Today time.Time                   `json:"today"`
	Data  map[time.Weekday]DailyUsage `json:"data"`
}

// CheckAndUpdateNewDay checks if now is a new day.
// If it is new day now, reset records of that day
func (w *WeeklyUsage) CheckAndUpdateNewDay(now time.Time) {
	if now.Sub(w.Today) > 24*time.Hour {
		w.Lock()
		defer w.Unlock()
		w.Data[time.Now().Weekday()] = DailyUsage{}
		w.Today = time.Unix(86400*(now.Unix()/86400), 0)
	}
}

// CountRequests counts requests for weekly usage data
func (w *WeeklyUsage) CountRequests(count int) {
	now := time.Now()
	w.CheckAndUpdateNewDay(now)

	w.Lock()
	defer w.Unlock()
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

// CheckAndUpdateNewDay checks if now is a new day.
// If it is new day now, push a new row of that day
func (u *MonthlyUsage) CheckAndUpdateNewDay(now time.Time) {
	// push a new row of a day if it is a new coming day
	if now.Sub(u.Today) > 24*time.Hour {
		u.Lock()
		defer u.Unlock()
		if len(u.Data) >= 30 {
			u.Data = u.Data[1:]
		}
		u.Data = append(u.Data, DailyUsage{})
		u.Today = time.Unix(86400*(now.Unix()/86400), 0)
	}
}

// CountRequests counts requests for monthly usage data
func (u *MonthlyUsage) CountRequests(count int) {
	now := time.Now()
	u.CheckAndUpdateNewDay(now)

	u.Lock()
	defer u.Unlock()
	lastDayUsage := u.Data[len(u.Data)-1]
	lastDayUsage[now.Hour()] += count
}
