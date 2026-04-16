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

// NpcResource exchanges resources via the NPC merchant.
func NpcResource(ctx context.Context, b *browser.Browser, db *database.DB, bus *event.Bus, villageID int) error {
	// Navigate to marketplace NPC page
	if err := navigate.ToDorf(ctx, b, 2); err != nil {
		return fmt.Errorf("navigate to dorf2: %w", err)
	}
	if err := update.UpdateBuildings(b, db, bus, villageID); err != nil {
		return fmt.Errorf("update buildings: %w", err)
	}
	if err := navigate.ToBuildingByType(ctx, b, db, villageID, int(enum.BuildingMarketplace)); err != nil {
		return fmt.Errorf("navigate to marketplace: %w", err)
	}
	// Switch to NPC tab (tab 0 = exchange)
	if err := navigate.SwitchTab(ctx, b, 0); err != nil {
		return fmt.Errorf("switch to npc tab: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Open NPC dialog
	el, err := b.Element(parser.GetExchangeResourcesButtonSelector())
	if err != nil {
		return fmt.Errorf("find exchange button: %w", err)
	}
	if err := b.Click(el); err != nil {
		return fmt.Errorf("click exchange button: %w", err)
	}

	// Wait for NPC dialog
	if err := b.WaitElementVisible(ctx, "#npc"); err != nil {
		return fmt.Errorf("wait for npc dialog: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Get settings for resource ratio
	woodRatio, _ := service.GetVillageSettingValue(db, villageID, enum.VillageSettingAutoNPCWood)
	clayRatio, _ := service.GetVillageSettingValue(db, villageID, enum.VillageSettingAutoNPCClay)
	ironRatio, _ := service.GetVillageSettingValue(db, villageID, enum.VillageSettingAutoNPCIron)
	cropRatio, _ := service.GetVillageSettingValue(db, villageID, enum.VillageSettingAutoNPCCrop)

	ratio := [4]int64{int64(woodRatio), int64(clayRatio), int64(ironRatio), int64(cropRatio)}
	sumRatio := ratio[0] + ratio[1] + ratio[2] + ratio[3]
	if sumRatio == 0 {
		ratio = [4]int64{1, 1, 1, 1}
		sumRatio = 4
	}

	// Read total sum from dialog
	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get npc page html: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse npc html: %w", err)
	}
	sum := parser.GetNpcSum(doc)
	if sum <= 0 {
		return fmt.Errorf("invalid npc sum: %d", sum)
	}

	// Calculate desired values
	var values [4]int64
	for i := 0; i < 4; i++ {
		values[i] = sum * ratio[i] / sumRatio
	}
	// Put remainder on crop
	diff := sum - (values[0] + values[1] + values[2] + values[3])
	values[3] += diff

	// Check warehouse overflow
	overflowNPC, _ := service.GetVillageSettingValue(db, villageID, enum.VillageSettingAutoNPCOverflow)
	warehouseCap := parser.GetWarehouseCapacity(doc)
	if warehouseCap > 0 {
		for i := 0; i < 3; i++ {
			if values[i] > warehouseCap {
				if overflowNPC != 0 {
					return &errs.TaskError{
						Err:     errs.ErrStorageLimit,
						Message: fmt.Sprintf("warehouse overflow: %d > %d", values[i], warehouseCap),
					}
				}
				break
			}
		}
	}

	// Input amounts
	inputSelectors := parser.GetNpcInputSelectors()
	for i := 0; i < 4; i++ {
		inputEl, err := b.Element(inputSelectors[i])
		if err != nil {
			return fmt.Errorf("find npc input %d: %w", i, err)
		}
		if err := b.Input(inputEl, fmt.Sprintf("%d", values[i])); err != nil {
			return fmt.Errorf("input npc amount %d: %w", i, err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Distribute if overflow mode
	if overflowNPC != 0 {
		distEl, err := b.Element(parser.GetDistributeButtonSelector())
		if err != nil {
			return fmt.Errorf("find distribute button: %w", err)
		}
		if err := b.Click(distEl); err != nil {
			return fmt.Errorf("click distribute button: %w", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Click redeem
	redeemEl, err := b.Element(parser.GetRedeemButtonSelector())
	if err != nil {
		return fmt.Errorf("find redeem button: %w", err)
	}
	if err := b.Click(redeemEl); err != nil {
		return fmt.Errorf("click redeem button: %w", err)
	}

	// Wait for dialog to close
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		time.Sleep(500 * time.Millisecond)
		html, err = b.PageHTML()
		if err != nil {
			continue
		}
		doc, err = parser.DocFromHTML(html)
		if err != nil {
			continue
		}
		if !parser.IsNpcDialog(doc) {
			return nil
		}
	}

	return nil
}
