package task

import (
	"context"
	"fmt"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/command/navigate"
	"travian-bot/internal/command/update"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/domain/errs"
	"travian-bot/internal/event"
	"travian-bot/internal/service"
)

// UpdateVillageTask refreshes village data (buildings, storage, village list)
// by navigating to both dorf pages and parsing the HTML.
type UpdateVillageTask struct {
	BaseTask
	villageID int
	bus       *event.Bus
}

func NewUpdateVillageTask(accountID, villageID int, bus *event.Bus) *UpdateVillageTask {
	return &UpdateVillageTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		villageID: villageID,
		bus:       bus,
	}
}

func (t *UpdateVillageTask) Description() string { return "Update village" }
func (t *UpdateVillageTask) VillageID() int      { return t.villageID }

func (t *UpdateVillageTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
	// Navigate to dorf1 to update resource fields
	if err := navigate.ToDorf(ctx, b, 1); err != nil {
		return &errs.TaskError{Err: errs.ErrRetry, Message: fmt.Sprintf("navigate to dorf1: %v", err)}
	}

	if err := update.UpdateBuildings(b, db, t.bus, t.villageID); err != nil {
		return fmt.Errorf("update dorf1 buildings: %w", err)
	}

	if err := update.UpdateStorage(b, db, t.bus, t.villageID); err != nil {
		return fmt.Errorf("update storage from dorf1: %w", err)
	}

	// Navigate to dorf2 to update infrastructure
	if err := navigate.ToDorf(ctx, b, 2); err != nil {
		return &errs.TaskError{Err: errs.ErrRetry, Message: fmt.Sprintf("navigate to dorf2: %v", err)}
	}

	if err := update.UpdateBuildings(b, db, t.bus, t.villageID); err != nil {
		return fmt.Errorf("update dorf2 buildings: %w", err)
	}

	// Reschedule based on auto-refresh settings
	enabled, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingAutoRefreshEnable)
	if enabled != 0 {
		minVal, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingAutoRefreshMin)
		maxVal, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingAutoRefreshMax)
		minutes := service.RandomBetween(minVal, maxVal)
		t.SetExecuteAt(time.Now().Add(time.Duration(minutes) * time.Minute))
	}

	return nil
}
