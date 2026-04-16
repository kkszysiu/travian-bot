package feature

import (
	"context"
	"math"

	"travian-bot/internal/browser"
	"travian-bot/internal/database"
	"travian-bot/internal/event"
)

// EvacuateResources sends all available resources from the current village to a target village.
// It delegates to SendResources with max values — the existing function caps to actual storage
// and merchant capacity.
func EvacuateResources(ctx context.Context, b *browser.Browser, db *database.DB, bus *event.Bus,
	villageID int, targetX, targetY int) error {

	maxRes := math.MaxInt32
	return SendResources(ctx, b, db, bus, villageID, targetX, targetY, maxRes, maxRes, maxRes, maxRes)
}
