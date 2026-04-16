package service

import (
	"testing"
	"time"
)

func makeTime(hour, minute int) time.Time {
	return time.Date(2026, 4, 2, hour, minute, 0, 0, time.Local)
}

func TestIsOutsideWindow_NormalWindow(t *testing.T) {
	// 06:00 - 22:00
	ww := WorkWindow{StartHour: 6, StartMinute: 0, EndHour: 22, EndMinute: 0}

	tests := []struct {
		name    string
		time    time.Time
		outside bool
	}{
		{"before start", makeTime(5, 30), true},
		{"at start", makeTime(6, 0), false},
		{"midday", makeTime(12, 0), false},
		{"just before end", makeTime(21, 59), false},
		{"at end", makeTime(22, 0), true},
		{"after end", makeTime(23, 0), true},
		{"midnight", makeTime(0, 0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ww.IsOutsideWindow(tt.time)
			if got != tt.outside {
				t.Errorf("IsOutsideWindow(%s) = %v, want %v", tt.time.Format("15:04"), got, tt.outside)
			}
		})
	}
}

func TestIsOutsideWindow_MidnightCrossing(t *testing.T) {
	// 22:00 - 06:00
	ww := WorkWindow{StartHour: 22, StartMinute: 0, EndHour: 6, EndMinute: 0}

	tests := []struct {
		name    string
		time    time.Time
		outside bool
	}{
		{"before end (early morning)", makeTime(3, 0), false},
		{"at end", makeTime(6, 0), true},
		{"midday", makeTime(12, 0), true},
		{"afternoon", makeTime(15, 0), true},
		{"just before start", makeTime(21, 59), true},
		{"at start", makeTime(22, 0), false},
		{"late night", makeTime(23, 30), false},
		{"midnight", makeTime(0, 0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ww.IsOutsideWindow(tt.time)
			if got != tt.outside {
				t.Errorf("IsOutsideWindow(%s) = %v, want %v", tt.time.Format("15:04"), got, tt.outside)
			}
		})
	}
}

func TestNextWorkStart(t *testing.T) {
	ww := WorkWindow{StartHour: 6, StartMinute: 30}

	tests := []struct {
		name     string
		now      time.Time
		wantHour int
		wantDay  int
	}{
		{"before start", makeTime(5, 0), 6, 2},
		{"at start", makeTime(6, 30), 6, 3}, // at or after -> next day
		{"after start", makeTime(12, 0), 6, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ww.NextWorkStart(tt.now)
			if got.Hour() != tt.wantHour || got.Day() != tt.wantDay {
				t.Errorf("NextWorkStart(%s) = %s, want hour=%d day=%d",
					tt.now.Format("15:04"), got.Format("2006-01-02 15:04"), tt.wantHour, tt.wantDay)
			}
			if got.Minute() != 30 {
				t.Errorf("NextWorkStart minute = %d, want 30", got.Minute())
			}
		})
	}
}

func TestNextWorkEnd_NoJitter(t *testing.T) {
	ww := WorkWindow{StartHour: 6, StartMinute: 0, EndHour: 22, EndMinute: 0, RandomMinute: 60}

	tests := []struct {
		name     string
		now      time.Time
		wantHour int
		wantDay  int
	}{
		{"before end", makeTime(12, 0), 22, 2},
		{"at end", makeTime(22, 0), 22, 3},
		{"after end", makeTime(23, 0), 22, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ww.NextWorkEnd(tt.now, false)
			if got.Hour() != tt.wantHour || got.Day() != tt.wantDay {
				t.Errorf("NextWorkEnd(%s) = %s, want hour=%d day=%d",
					tt.now.Format("15:04"), got.Format("2006-01-02 15:04"), tt.wantHour, tt.wantDay)
			}
		})
	}
}

func TestNextWorkEnd_MidnightCrossing(t *testing.T) {
	// 22:00 - 06:00: end is 06:00 which is before start 22:00
	ww := WorkWindow{StartHour: 22, StartMinute: 0, EndHour: 6, EndMinute: 0, RandomMinute: 0}

	now := makeTime(23, 0) // inside window
	got := ww.NextWorkEnd(now, false)

	// End should be tomorrow at 06:00
	if got.Hour() != 6 || got.Day() != 3 {
		t.Errorf("NextWorkEnd(%s) = %s, want 2026-04-03 06:00",
			now.Format("15:04"), got.Format("2006-01-02 15:04"))
	}
}

func TestSleepDurationCapping(t *testing.T) {
	// Simulate: next work start in 30 minutes, but random sleep would be 480 minutes
	// The capping should reduce it to 30
	ww := WorkWindow{StartHour: 6, StartMinute: 0, EndHour: 22, EndMinute: 0}

	now := makeTime(5, 30) // 30 minutes before 06:00 start
	nextStart := ww.NextWorkStart(now)
	minutesUntilStart := int(nextStart.Sub(now).Minutes())

	sleepMinutes := 480 // random sleep duration
	if minutesUntilStart > 0 && sleepMinutes > minutesUntilStart {
		sleepMinutes = minutesUntilStart
	}

	if sleepMinutes != 30 {
		t.Errorf("capped sleep = %d, want 30", sleepMinutes)
	}
}
