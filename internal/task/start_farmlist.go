package task

import (
	"context"
	"fmt"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/command/feature"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/event"
	"travian-bot/internal/service"
)

// StartFarmListTask starts farm lists on a schedule.
type StartFarmListTask struct {
	BaseTask
	bus *event.Bus
}

func NewStartFarmListTask(accountID int, bus *event.Bus) *StartFarmListTask {
	return &StartFarmListTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		bus: bus,
	}
}

func (t *StartFarmListTask) Description() string { return "Start farm list" }

func (t *StartFarmListTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
	if err := feature.StartFarmList(ctx, b, db, t.bus, t.accountID); err != nil {
		return fmt.Errorf("start farm list: %w", err)
	}

	// Reschedule based on farm interval settings
	// Default values are 540 and 660 (in seconds)
	minVal, _ := service.GetAccountSettingValue(db, t.accountID, enum.AccountSettingFarmIntervalMin)
	maxVal, _ := service.GetAccountSettingValue(db, t.accountID, enum.AccountSettingFarmIntervalMax)
	seconds := service.RandomBetween(minVal, maxVal)
	t.SetExecuteAt(time.Now().Add(time.Duration(seconds) * time.Second))

	return nil
}
