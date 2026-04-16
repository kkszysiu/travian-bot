package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/command/navigate"
	"travian-bot/internal/command/update"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/errs"
	"travian-bot/internal/event"
)

// UpdateBuildingTask navigates to both dorf pages and updates building data.
// This is a one-shot task, typically queued for villages missing building data.
type UpdateBuildingTask struct {
	BaseTask
	villageID int
	bus       *event.Bus
}

func NewUpdateBuildingTask(accountID, villageID int, bus *event.Bus) *UpdateBuildingTask {
	return &UpdateBuildingTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		villageID: villageID,
		bus:       bus,
	}
}

func (t *UpdateBuildingTask) Description() string { return "Update building" }
func (t *UpdateBuildingTask) VillageID() int      { return t.villageID }

func (t *UpdateBuildingTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
	url := b.CurrentURL()

	if strings.Contains(url, "dorf1") {
		// On dorf1: update fields, go to dorf2, update infrastructure
		if err := update.UpdateBuildings(b, db, t.bus, t.villageID); err != nil {
			return fmt.Errorf("update dorf1 buildings: %w", err)
		}
		if err := navigate.ToDorf(ctx, b, 2); err != nil {
			return &errs.TaskError{Err: errs.ErrRetry, Message: fmt.Sprintf("navigate to dorf2: %v", err)}
		}
		if err := update.UpdateBuildings(b, db, t.bus, t.villageID); err != nil {
			return fmt.Errorf("update dorf2 buildings: %w", err)
		}
	} else if strings.Contains(url, "dorf2") {
		// On dorf2: update infrastructure, go to dorf1, update fields
		if err := update.UpdateBuildings(b, db, t.bus, t.villageID); err != nil {
			return fmt.Errorf("update dorf2 buildings: %w", err)
		}
		if err := navigate.ToDorf(ctx, b, 1); err != nil {
			return &errs.TaskError{Err: errs.ErrRetry, Message: fmt.Sprintf("navigate to dorf1: %v", err)}
		}
		if err := update.UpdateBuildings(b, db, t.bus, t.villageID); err != nil {
			return fmt.Errorf("update dorf1 buildings: %w", err)
		}
	} else {
		// Not on any dorf: go to dorf2, update, go to dorf1, update
		if err := navigate.ToDorf(ctx, b, 2); err != nil {
			return &errs.TaskError{Err: errs.ErrRetry, Message: fmt.Sprintf("navigate to dorf2: %v", err)}
		}
		if err := update.UpdateBuildings(b, db, t.bus, t.villageID); err != nil {
			return fmt.Errorf("update dorf2 buildings: %w", err)
		}
		if err := navigate.ToDorf(ctx, b, 1); err != nil {
			return &errs.TaskError{Err: errs.ErrRetry, Message: fmt.Sprintf("navigate to dorf1: %v", err)}
		}
		if err := update.UpdateBuildings(b, db, t.bus, t.villageID); err != nil {
			return fmt.Errorf("update dorf1 buildings: %w", err)
		}
	}

	// Update storage (resource bar is visible on any dorf page)
	if err := update.UpdateStorage(b, db, t.bus, t.villageID); err != nil {
		// Non-fatal — just log
		_ = err
	}

	return nil
}
