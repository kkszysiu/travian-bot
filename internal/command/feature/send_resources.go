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
	"travian-bot/internal/event"
	"travian-bot/internal/parser"
)

// SendResources sends resources from the current village to a target village via the marketplace.
func SendResources(ctx context.Context, b *browser.Browser, db *database.DB, bus *event.Bus,
	villageID int, targetX, targetY int, wood, clay, iron, crop int) error {

	// Navigate to marketplace
	if err := navigate.ToDorf(ctx, b, 2); err != nil {
		return fmt.Errorf("navigate to dorf2: %w", err)
	}
	if err := update.UpdateBuildings(b, db, bus, villageID); err != nil {
		return fmt.Errorf("update buildings: %w", err)
	}
	if err := navigate.ToBuildingByType(ctx, b, db, villageID, int(enum.BuildingMarketplace)); err != nil {
		return fmt.Errorf("navigate to marketplace: %w", err)
	}
	// Switch to send resources tab (tab 0)
	if err := navigate.SwitchTab(ctx, b, 0); err != nil {
		return fmt.Errorf("switch to send tab: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Parse page for merchant info and current resources
	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get marketplace html: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse marketplace html: %w", err)
	}

	available := parser.GetAvailableMerchants(doc)
	if available <= 0 {
		return fmt.Errorf("no available merchants")
	}
	capacity := parser.GetMerchantCapacity(doc)
	if capacity <= 0 {
		return fmt.Errorf("invalid merchant capacity")
	}

	// Cap each resource to what's available in storage
	curWood := int(parser.GetWood(doc))
	curClay := int(parser.GetClay(doc))
	curIron := int(parser.GetIron(doc))
	curCrop := int(parser.GetCrop(doc))

	sendWood := min(wood, curWood)
	sendClay := min(clay, curClay)
	sendIron := min(iron, curIron)
	sendCrop := min(crop, curCrop)

	if sendWood < 0 {
		sendWood = 0
	}
	if sendClay < 0 {
		sendClay = 0
	}
	if sendIron < 0 {
		sendIron = 0
	}
	if sendCrop < 0 {
		sendCrop = 0
	}

	total := sendWood + sendClay + sendIron + sendCrop
	if total == 0 {
		return nil // Nothing to send
	}

	// Cap total to available merchant capacity
	maxCarry := available * capacity
	if total > maxCarry {
		// Proportionally reduce
		ratio := float64(maxCarry) / float64(total)
		sendWood = int(float64(sendWood) * ratio)
		sendClay = int(float64(sendClay) * ratio)
		sendIron = int(float64(sendIron) * ratio)
		sendCrop = maxCarry - sendWood - sendClay - sendIron
		if sendCrop < 0 {
			sendCrop = 0
		}
	}

	total = sendWood + sendClay + sendIron + sendCrop
	if total == 0 {
		return nil
	}

	// Input resource amounts
	amounts := [4]int{sendWood, sendClay, sendIron, sendCrop}
	inputSelectors := parser.GetSendResourceInputSelectors()
	for i := 0; i < 4; i++ {
		if amounts[i] <= 0 {
			continue
		}
		inputEl, err := b.Element(inputSelectors[i])
		if err != nil {
			return fmt.Errorf("find resource input %d: %w", i, err)
		}
		if err := b.Input(inputEl, fmt.Sprintf("%d", amounts[i])); err != nil {
			return fmt.Errorf("input resource amount %d: %w", i, err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Input target coordinates
	xEl, err := b.Element(parser.GetCoordXInputSelector())
	if err != nil {
		return fmt.Errorf("find coord X input: %w", err)
	}
	if err := b.Input(xEl, fmt.Sprintf("%d", targetX)); err != nil {
		return fmt.Errorf("input coord X: %w", err)
	}
	time.Sleep(200 * time.Millisecond)

	yEl, err := b.Element(parser.GetCoordYInputSelector())
	if err != nil {
		return fmt.Errorf("find coord Y input: %w", err)
	}
	if err := b.Input(yEl, fmt.Sprintf("%d", targetY)); err != nil {
		return fmt.Errorf("input coord Y: %w", err)
	}
	time.Sleep(200 * time.Millisecond)

	// Click send button
	sendBtn, err := b.Element(parser.GetSendButtonSelector())
	if err != nil {
		return fmt.Errorf("find send button: %w", err)
	}
	if err := b.Click(sendBtn); err != nil {
		return fmt.Errorf("click send button: %w", err)
	}

	// Wait for confirmation / page reload
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
		// Check if merchants decreased (send was successful)
		newAvailable := parser.GetAvailableMerchants(doc)
		if newAvailable < available {
			return nil
		}
	}

	return nil
}
