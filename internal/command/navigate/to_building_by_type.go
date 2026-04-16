package navigate

import (
	"context"
	"fmt"

	"travian-bot/internal/browser"
	"travian-bot/internal/database"
)

// ToBuildingByType looks up the location of a building by its type within a
// village, then navigates the browser to that building's page.
func ToBuildingByType(ctx context.Context, b *browser.Browser, db *database.DB, villageID int, buildingType int) error {
	var location int
	err := db.Get(&location,
		"SELECT location FROM buildings WHERE village_id = ? AND type = ? LIMIT 1",
		villageID, buildingType,
	)
	if err != nil {
		return fmt.Errorf("building type %d not found in village %d: %w", buildingType, villageID, err)
	}

	return ToBuilding(ctx, b, location)
}
