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
	"travian-bot/internal/event"
)

// UpdateFarmListTask navigates to the rally point farm list page and syncs
// farm lists to the database.
type UpdateFarmListTask struct {
	BaseTask
	bus *event.Bus
}

func NewUpdateFarmListTask(accountID int, bus *event.Bus) *UpdateFarmListTask {
	return &UpdateFarmListTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		bus: bus,
	}
}

func (t *UpdateFarmListTask) Description() string { return "Update farm list" }

func (t *UpdateFarmListTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
	// Find a village with a rally point
	var villageID int
	err := db.Get(&villageID,
		`SELECT v.id FROM villages v
		 JOIN buildings b ON b.village_id = v.id
		 WHERE v.account_id = ? AND b.type = ? AND b.level > 0
		 ORDER BY v.is_active DESC LIMIT 1`,
		t.accountID, int(enum.BuildingRallyPoint),
	)
	if err != nil {
		return fmt.Errorf("no village with rally point found: %w", err)
	}

	// Switch to that village
	if err := navigate.SwitchVillage(ctx, b, villageID); err != nil {
		return fmt.Errorf("switch to rallypoint village: %w", err)
	}

	// Navigate to dorf2 and update buildings
	if err := navigate.ToDorf(ctx, b, 2); err != nil {
		return fmt.Errorf("navigate to dorf2: %w", err)
	}
	if err := update.UpdateBuildings(b, db, t.bus, villageID); err != nil {
		return fmt.Errorf("update buildings: %w", err)
	}

	// Navigate to the rally point building
	if err := navigate.ToBuildingByType(ctx, b, db, villageID, int(enum.BuildingRallyPoint)); err != nil {
		return fmt.Errorf("navigate to rally point: %w", err)
	}

	// Switch to farm list tab (tab index 4)
	time.Sleep(500 * time.Millisecond)
	if err := navigate.SwitchTab(ctx, b, 4); err != nil {
		return fmt.Errorf("switch to farm list tab: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Parse and sync farm lists to the database
	if err := update.UpdateFarmList(b, db, t.bus, t.accountID); err != nil {
		return fmt.Errorf("update farm lists: %w", err)
	}

	return nil
}
