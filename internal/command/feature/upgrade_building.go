package feature

import (
	"context"
	"fmt"

	"travian-bot/internal/browser"
	"travian-bot/internal/command/navigate"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/domain/errs"
	"travian-bot/internal/parser"
	"travian-bot/internal/service"
)

// NormalBuildPlan represents a specific building upgrade plan.
type NormalBuildPlan struct {
	Type     int // Building enum value
	Level    int // Target level
	Location int // Slot location in the village
}

// UpgradeBuilding performs the full upgrade cycle for one building:
// navigate to the building page, validate resources, then click upgrade.
func UpgradeBuilding(ctx context.Context, b *browser.Browser, db *database.DB, plan NormalBuildPlan, villageID int) error {
	// Navigate to the building location
	if err := navigate.ToBuilding(ctx, b, plan.Location); err != nil {
		return fmt.Errorf("navigate to building location %d: %w", plan.Location, err)
	}

	building := enum.Building(plan.Type)

	// If the building has multiple tabs, switch to the correct category tab
	if building.HasMultipleTabs() {
		category := building.GetBuildingsCategory()
		if err := navigate.SwitchTab(ctx, b, category); err != nil {
			return fmt.Errorf("switch to tab %d for %s: %w", category, building, err)
		}
	}

	// Parse required resources from the page
	required, err := GetRequiredResource(b, plan.Type)
	if err != nil {
		return fmt.Errorf("get required resources for %s: %w", building, err)
	}

	// Validate that the village has enough resources and storage
	if err := ValidateResources(db, villageID, required); err != nil {
		return err
	}

	// Perform the upgrade action
	if err := HandleUpgrade(ctx, b, db, plan, villageID); err != nil {
		return err
	}

	return nil
}

// HandleUpgrade clicks the appropriate button (upgrade/construct/special-upgrade)
// based on the building state and village settings.
func HandleUpgrade(ctx context.Context, b *browser.Browser, db *database.DB, plan NormalBuildPlan, villageID int) error {
	building := enum.Building(plan.Type)

	// Determine which button to click
	var selector string

	isSite := plan.Type == int(enum.BuildingSite) || plan.Level == -1
	if isSite {
		// Empty site: use the construct button for the building type
		selector = parser.GetConstructButtonSelector(plan.Type)
	} else {
		// Existing building: check if special upgrade is enabled
		useSpecial, err := service.GetVillageSettingValue(db, villageID, enum.VillageSettingUseSpecialUpgrade)
		if err != nil {
			return fmt.Errorf("get special upgrade setting: %w", err)
		}

		if useSpecial != 0 {
			selector = parser.GetSpecialUpgradeButtonSelector()
		} else {
			selector = parser.GetUpgradeButtonSelector()
		}
	}

	el, err := b.Element(selector)
	if err != nil {
		return fmt.Errorf("find upgrade button (%s) for %s: %w", selector, building, err)
	}

	if err := b.Click(el); err != nil {
		return fmt.Errorf("click upgrade button for %s: %w", building, err)
	}

	// Wait for navigation back to a dorf page (build action triggers redirect)
	if err := b.WaitPageContains(ctx, "dorf"); err != nil {
		return fmt.Errorf("wait for dorf after upgrading %s: %w", building, err)
	}

	return nil
}

// ValidateResources checks if the village has enough resources and storage
// capacity for the given requirements.
// required is [wood, clay, iron, crop, freeCrop(upkeep)].
func ValidateResources(db *database.DB, villageID int, required [5]int64) error {
	storage, err := db.GetStorage(villageID)
	if err != nil {
		return fmt.Errorf("get storage for village %d: %w", villageID, err)
	}

	// Check warehouse capacity against the maximum of wood/clay/iron required
	maxResource := required[0]
	if required[1] > maxResource {
		maxResource = required[1]
	}
	if required[2] > maxResource {
		maxResource = required[2]
	}
	if int64(storage.Warehouse) < maxResource {
		return &errs.TaskError{
			Err:     errs.ErrStorageLimit,
			Message: fmt.Sprintf("warehouse capacity %d is less than required %d", storage.Warehouse, maxResource),
		}
	}

	// Check granary capacity against crop required
	if int64(storage.Granary) < required[3] {
		return &errs.TaskError{
			Err:     errs.ErrStorageLimit,
			Message: fmt.Sprintf("granary capacity %d is less than required crop %d", storage.Granary, required[3]),
		}
	}

	// Check current resources against requirements
	if int64(storage.Wood) < required[0] ||
		int64(storage.Clay) < required[1] ||
		int64(storage.Iron) < required[2] ||
		int64(storage.Crop) < required[3] {
		return &errs.TaskError{
			Err: errs.ErrMissingResource,
			Message: fmt.Sprintf(
				"not enough resources: have [%d/%d/%d/%d], need [%d/%d/%d/%d]",
				storage.Wood, storage.Clay, storage.Iron, storage.Crop,
				required[0], required[1], required[2], required[3],
			),
		}
	}

	// Check free crop (upkeep) if specified
	if required[4] > 0 && int64(storage.FreeCrop) < required[4] {
		return &errs.TaskError{
			Err: errs.ErrLackOfFreeCrop,
			Message: fmt.Sprintf(
				"not enough free crop: have %d, need %d",
				storage.FreeCrop, required[4],
			),
		}
	}

	return nil
}

// GetRequiredResource parses the required resources from the current build page.
// Returns a [5]int64 array: [wood, clay, iron, crop, freeCrop].
func GetRequiredResource(b *browser.Browser, buildingType int) ([5]int64, error) {
	var result [5]int64

	html, err := b.PageHTML()
	if err != nil {
		return result, fmt.Errorf("get page HTML: %w", err)
	}

	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return result, fmt.Errorf("parse HTML: %w", err)
	}

	resources := parser.GetRequiredResource(doc, buildingType)
	if resources == nil || len(resources) < 5 {
		return result, fmt.Errorf("could not parse required resources for building type %d", buildingType)
	}

	for i := 0; i < 5; i++ {
		result[i] = resources[i]
	}
	return result, nil
}
