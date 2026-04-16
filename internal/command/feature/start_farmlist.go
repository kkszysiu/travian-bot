package feature

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
	"travian-bot/internal/parser"
	"travian-bot/internal/service"
)

// StartFarmList navigates to the rally point farm list page and starts farm lists.
func StartFarmList(ctx context.Context, b *browser.Browser, db *database.DB, bus *event.Bus, accountID int) error {
	// Find a village with a rally point
	rallypointVillageID, err := findRallypointVillage(db, accountID)
	if err != nil {
		return errs.NewSkipError("no village with rally point found — skipping farm list", time.Time{})
	}

	// Switch to that village
	if err := navigate.SwitchVillage(ctx, b, rallypointVillageID); err != nil {
		return errs.NewSkipError(fmt.Sprintf("cannot switch to rally point village %d — skipping farm list", rallypointVillageID), time.Time{})
	}

	// Navigate to dorf2 and update buildings
	if err := navigate.ToDorf(ctx, b, 2); err != nil {
		return fmt.Errorf("navigate to dorf2: %w", err)
	}
	if err := update.UpdateBuildings(b, db, bus, rallypointVillageID); err != nil {
		return fmt.Errorf("update buildings: %w", err)
	}

	// Navigate to the rally point building
	if err := navigate.ToBuildingByType(ctx, b, db, rallypointVillageID, int(enum.BuildingRallyPoint)); err != nil {
		return fmt.Errorf("navigate to rally point: %w", err)
	}

	// Switch to farm list tab (tab index 4)
	time.Sleep(500 * time.Millisecond)
	if err := navigate.SwitchTab(ctx, b, 4); err != nil {
		return fmt.Errorf("switch to farm list tab: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Update farm lists in the database
	if err := update.UpdateFarmList(b, db, bus, accountID); err != nil {
		return fmt.Errorf("update farm lists: %w", err)
	}

	// Check if we should use "Start All" button
	useStartAll, _ := service.GetAccountSettingValue(db, accountID, enum.AccountSettingUseStartAllButton)
	if useStartAll != 0 {
		return startAllFarmLists(b)
	}
	return startActiveFarmLists(b, db, accountID)
}

func findRallypointVillage(db *database.DB, accountID int) (int, error) {
	var villageID int
	err := db.Get(&villageID,
		`SELECT v.id FROM villages v
		 JOIN buildings b ON b.village_id = v.id
		 WHERE v.account_id = ? AND b.type = ? AND b.level > 0
		 ORDER BY v.is_active DESC LIMIT 1`,
		accountID, int(enum.BuildingRallyPoint),
	)
	if err != nil {
		return 0, fmt.Errorf("no village with rally point found: %w", err)
	}
	return villageID, nil
}

func startAllFarmLists(b *browser.Browser) error {
	selector := parser.GetStartAllFarmListButtonSelector()
	el, err := b.Element(selector)
	if err != nil {
		return fmt.Errorf("find start all button: %w", err)
	}
	return b.Click(el)
}

func startActiveFarmLists(b *browser.Browser, db *database.DB, accountID int) error {
	// Get active farm lists from DB
	farms, err := db.GetFarmLists(accountID)
	if err != nil {
		return fmt.Errorf("get farm lists: %w", err)
	}
	if len(farms) == 0 {
		return errs.NewSkipError("no farm lists found", time.Time{})
	}

	started := 0
	for _, farm := range farms {
		if !farm.IsActive {
			continue
		}

		selector := parser.GetStartFarmListButtonSelector(farm.ID)
		el, err := b.Element(selector)
		if err != nil {
			continue // skip if not found on page
		}

		if err := b.Click(el); err != nil {
			continue
		}
		started++
		time.Sleep(300 * time.Millisecond) // delay between farm list starts
	}

	if started == 0 {
		return errs.NewSkipError("no active farm lists started", time.Time{})
	}
	return nil
}
