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
)

// UseHeroResource navigates to the hero inventory and uses resource items
// (wood, clay, iron, crop) to supplement a building upgrade. The resource
// array is [wood, clay, iron, crop] representing the amounts needed.
//
// The function:
// 1. Navigates to the hero inventory page
// 2. Updates the inventory database from the page
// 3. Rounds each resource amount up to the nearest 100
// 4. Validates the hero has enough of each resource item
// 5. Uses each resource item with the required amount
func UseHeroResource(ctx context.Context, b *browser.Browser, db *database.DB, accountID int, resource [4]int64) error {
	// Navigate to hero inventory
	if err := navigate.ToHeroInventory(ctx, b); err != nil {
		return fmt.Errorf("navigate to hero inventory: %w", err)
	}

	// Update inventory from the page so DB is current
	if err := update.UpdateInventory(b, db, accountID); err != nil {
		return fmt.Errorf("update inventory: %w", err)
	}

	// Round each resource up to the nearest 100
	for i := 0; i < 4; i++ {
		resource[i] = roundUpTo100(resource[i])
	}

	// Validate hero has enough of each resource item
	if err := validateHeroResources(db, accountID, resource); err != nil {
		return err
	}

	// Map resource index to hero item type
	resourceItems := [4]enum.HeroItem{
		enum.HeroItemWood,
		enum.HeroItemClay,
		enum.HeroItemIron,
		enum.HeroItemCrop,
	}

	// Use each resource item
	for i, item := range resourceItems {
		amount := resource[i]
		if amount == 0 {
			continue
		}
		if err := UseHeroItem(ctx, b, item, amount); err != nil {
			return fmt.Errorf("use hero item %d (amount %d): %w", item, amount, err)
		}
	}

	time.Sleep(500 * time.Millisecond)
	return nil
}

// validateHeroResources checks that the hero's inventory contains enough of
// each resource item type to cover the required amounts.
func validateHeroResources(db *database.DB, accountID int, resource [4]int64) error {
	resourceItems := [4]enum.HeroItem{
		enum.HeroItemWood,
		enum.HeroItemClay,
		enum.HeroItemIron,
		enum.HeroItemCrop,
	}

	for i, itemType := range resourceItems {
		if resource[i] == 0 {
			continue
		}
		amount, err := db.GetHeroItemAmount(accountID, int(itemType))
		if err != nil {
			return fmt.Errorf("get hero item amount for %d: %w", itemType, err)
		}
		// Hero resource items are stored in units of 100 in the game.
		// The amount in DB is the number of item stacks, each worth 100 resources.
		available := int64(amount) * 100
		if available < resource[i] {
			return &errs.TaskError{
				Err: errs.ErrMissingResource,
				Message: fmt.Sprintf(
					"hero inventory missing resource: type %d, have %d (x100=%d), need %d",
					itemType, amount, available, resource[i],
				),
			}
		}
	}
	return nil
}

// roundUpTo100 rounds a value up to the next multiple of 100.
// Returns 0 if the input is 0.
func roundUpTo100(res int64) int64 {
	if res == 0 {
		return 0
	}
	remainder := res % 100
	if remainder == 0 {
		return res
	}
	return res + (100 - remainder)
}
