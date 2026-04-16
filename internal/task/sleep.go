package task

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/domain/errs"
	"travian-bot/internal/event"
	"travian-bot/internal/service"
)

// SleepTask closes the browser, sleeps for the configured time, reopens the
// browser, and queues a fresh login task so the scheduler loop can continue.
type SleepTask struct {
	BaseTask
	bus       *event.Bus
	scheduler *Scheduler
	browsers  *browser.Manager
}

func NewSleepTask(accountID int, bus *event.Bus, scheduler *Scheduler, browsers *browser.Manager) *SleepTask {
	return &SleepTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		bus:       bus,
		scheduler: scheduler,
		browsers:  browsers,
	}
}

func (t *SleepTask) Description() string { return "Sleep" }

func (t *SleepTask) Execute(ctx context.Context, _ *browser.Browser, db *database.DB) error {
	// Close the browser
	t.browsers.Close(t.accountID)

	// Calculate sleep duration from settings, capped by next work start
	sleepMin, _ := service.GetAccountSettingValue(db, t.accountID, enum.AccountSettingSleepTimeMin)
	sleepMax, _ := service.GetAccountSettingValue(db, t.accountID, enum.AccountSettingSleepTimeMax)
	sleepMinutes := service.RandomBetween(sleepMin, sleepMax)

	ww := service.GetWorkWindow(db, t.accountID)
	nextStart := ww.NextWorkStart(time.Now())
	minutesUntilStart := int(time.Until(nextStart).Minutes())
	if minutesUntilStart > 0 && sleepMinutes > minutesUntilStart {
		sleepMinutes = minutesUntilStart
	}

	sleepDuration := time.Duration(sleepMinutes) * time.Minute

	// Sleep with cancellation support (check every second)
	sleepEnd := time.Now().Add(sleepDuration)
	for {
		if ctx.Err() != nil {
			return &errs.TaskError{Err: errs.ErrCancel, Message: "sleep cancelled"}
		}
		remaining := time.Until(sleepEnd)
		if remaining <= 0 {
			break
		}
		wait := time.Second
		if remaining < wait {
			wait = remaining
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return &errs.TaskError{Err: errs.ErrCancel, Message: "sleep cancelled"}
		case <-timer.C:
		}
	}

	// Reopen browser
	detail, err := db.GetAccountDetail(t.accountID)
	if err != nil {
		return fmt.Errorf("get account detail for reopen: %w", err)
	}

	cfg := browser.Config{
		ProfilePath: strconv.Itoa(t.accountID),
	}
	if len(detail.Accesses) > 0 {
		access := detail.Accesses[0]
		cfg.ProxyHost = access.ProxyHost
		cfg.ProxyPort = access.ProxyPort
		cfg.ProxyUsername = access.ProxyUsername
		cfg.ProxyPassword = access.ProxyPassword
		cfg.UserAgent = access.Useragent
	}

	newBrowser, err := t.browsers.Create(t.accountID, cfg)
	if err != nil {
		return fmt.Errorf("reopen browser: %w", err)
	}

	// Navigate to server
	if err := newBrowser.Navigate(detail.Server); err != nil {
		return fmt.Errorf("navigate to server: %w", err)
	}

	// Queue a fresh login task
	t.scheduler.AddTask(t.accountID, NewLoginTask(t.accountID, t.bus, t.scheduler, t.browsers))

	// Reschedule this sleep task based on work window
	now := time.Now()
	ww = service.GetWorkWindow(db, t.accountID)
	if ww.IsOutsideWindow(now) {
		// Outside window: wake at next work start, sleep again at end of that window
		t.SetExecuteAt(ww.NextWorkEnd(now, true))
	} else {
		// Inside window: schedule sleep at end of current window + jitter
		t.SetExecuteAt(ww.NextWorkEnd(now, true))
	}

	return nil
}
