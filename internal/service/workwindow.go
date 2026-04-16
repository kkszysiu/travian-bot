package service

import (
	"math/rand"
	"time"

	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
)

// WorkWindow represents a daily time window during which the bot should operate.
type WorkWindow struct {
	StartHour, StartMinute int
	EndHour, EndMinute     int
	RandomMinute           int
}

// GetWorkWindow loads the work window settings for an account from the database.
func GetWorkWindow(db *database.DB, accountID int) WorkWindow {
	startHour, _ := GetAccountSettingValue(db, accountID, enum.AccountSettingWorkStartHour)
	startMinute, _ := GetAccountSettingValue(db, accountID, enum.AccountSettingWorkStartMinute)
	endHour, _ := GetAccountSettingValue(db, accountID, enum.AccountSettingWorkEndHour)
	endMinute, _ := GetAccountSettingValue(db, accountID, enum.AccountSettingWorkEndMinute)
	randomMinute, _ := GetAccountSettingValue(db, accountID, enum.AccountSettingSleepRandomMinute)

	// Validate ranges, fall back to defaults on invalid
	if startHour < 0 || startHour > 23 {
		startHour = 6
	}
	if startMinute < 0 || startMinute > 59 {
		startMinute = 0
	}
	if endHour < 0 || endHour > 23 {
		endHour = 22
	}
	if endMinute < 0 || endMinute > 59 {
		endMinute = 0
	}
	if randomMinute < 0 {
		randomMinute = 60
	}

	return WorkWindow{
		StartHour:    startHour,
		StartMinute:  startMinute,
		EndHour:      endHour,
		EndMinute:    endMinute,
		RandomMinute: randomMinute,
	}
}

// IsOutsideWindow returns true if the given time falls outside the work window.
// Handles midnight-crossing windows (e.g., 22:00-06:00) correctly.
func (w WorkWindow) IsOutsideWindow(now time.Time) bool {
	start := time.Date(now.Year(), now.Month(), now.Day(), w.StartHour, w.StartMinute, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), w.EndHour, w.EndMinute, 0, 0, now.Location())

	if end.Before(start) || end.Equal(start) {
		// Midnight-crossing window (e.g., 22:00-06:00)
		// Outside = now >= end AND now < start
		return !now.Before(end) && now.Before(start)
	}

	// Normal window (e.g., 06:00-22:00)
	// Outside = now < start OR now >= end
	return now.Before(start) || !now.Before(end)
}

// NextWorkStart returns the next work start time at or after the given time.
func (w WorkWindow) NextWorkStart(now time.Time) time.Time {
	startToday := time.Date(now.Year(), now.Month(), now.Day(), w.StartHour, w.StartMinute, 0, 0, now.Location())
	if now.Before(startToday) {
		return startToday
	}
	return startToday.Add(24 * time.Hour)
}

// NextWorkEnd returns the next work end time at or after the given time.
// If withJitter is true, adds a random offset of [0, RandomMinute) minutes.
func (w WorkWindow) NextWorkEnd(now time.Time, withJitter bool) time.Time {
	endToday := time.Date(now.Year(), now.Month(), now.Day(), w.EndHour, w.EndMinute, 0, 0, now.Location())
	start := time.Date(now.Year(), now.Month(), now.Day(), w.StartHour, w.StartMinute, 0, 0, now.Location())

	var end time.Time
	if endToday.Before(start) || endToday.Equal(start) {
		// Midnight-crossing: end is tomorrow
		end = endToday.Add(24 * time.Hour)
	} else {
		end = endToday
	}

	if now.Before(end) {
		if withJitter && w.RandomMinute > 0 {
			end = end.Add(time.Duration(rand.Intn(w.RandomMinute)) * time.Minute)
		}
		return end
	}

	// End already passed today, use tomorrow
	end = endToday.Add(24 * time.Hour)
	if withJitter && w.RandomMinute > 0 {
		end = end.Add(time.Duration(rand.Intn(w.RandomMinute)) * time.Minute)
	}
	return end
}
