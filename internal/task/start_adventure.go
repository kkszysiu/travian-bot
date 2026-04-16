package task

import (
	"context"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/command/feature"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/domain/errs"
	"travian-bot/internal/event"
	"travian-bot/internal/service"
)

// StartAdventureTask starts an available adventure for the hero.
type StartAdventureTask struct {
	BaseTask
	bus *event.Bus
}

func NewStartAdventureTask(accountID int, bus *event.Bus) *StartAdventureTask {
	return &StartAdventureTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		bus: bus,
	}
}

func (t *StartAdventureTask) Description() string { return "Start adventure" }

func (t *StartAdventureTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
	// Check if enabled
	enabled, _ := service.GetAccountSettingValue(db, t.accountID, enum.AccountSettingEnableAutoStartAdventure)
	if enabled == 0 {
		return errs.NewSkipError("adventure disabled", time.Time{})
	}

	duration, err := feature.StartAdventure(ctx, b)
	if err != nil {
		return err
	}

	// Reschedule: wait 2x the adventure duration (round trip)
	if duration > 0 {
		nextRun := time.Now().Add(time.Duration(duration*2) * time.Second)
		t.SetExecuteAt(nextRun)
	} else {
		// No adventure available, check again in 30 minutes
		t.SetExecuteAt(time.Now().Add(30 * time.Minute))
	}

	return nil
}
