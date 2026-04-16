package feature

import (
	"context"
	"fmt"

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

// troopSettings maps training building types to their troop setting enum.
var troopSettings = map[enum.Building]enum.VillageSetting{
	enum.BuildingBarracks:      enum.VillageSettingBarrackTroop,
	enum.BuildingStable:        enum.VillageSettingStableTroop,
	enum.BuildingGreatBarracks: enum.VillageSettingGreatBarrackTroop,
	enum.BuildingGreatStable:   enum.VillageSettingGreatStableTroop,
	enum.BuildingWorkshop:      enum.VillageSettingWorkshopTroop,
}

// amountSettings maps building types to their (min, max) amount settings.
var amountSettings = map[enum.Building][2]enum.VillageSetting{
	enum.BuildingBarracks:      {enum.VillageSettingBarrackAmountMin, enum.VillageSettingBarrackAmountMax},
	enum.BuildingStable:        {enum.VillageSettingStableAmountMin, enum.VillageSettingStableAmountMax},
	enum.BuildingGreatBarracks: {enum.VillageSettingGreatBarrackAmountMin, enum.VillageSettingGreatBarrackAmountMax},
	enum.BuildingGreatStable:   {enum.VillageSettingGreatStableAmountMin, enum.VillageSettingGreatStableAmountMax},
	enum.BuildingWorkshop:      {enum.VillageSettingWorkshopAmountMin, enum.VillageSettingWorkshopAmountMax},
}

// GetTrainTroopBuildings returns the list of training buildings that have a troop configured.
func GetTrainTroopBuildings(db *database.DB, villageID int) []enum.Building {
	var buildings []enum.Building
	for building, setting := range troopSettings {
		val, err := service.GetVillageSettingValue(db, villageID, setting)
		if err == nil && val != 0 {
			buildings = append(buildings, building)
		}
	}
	return buildings
}

// TrainTroop navigates to a training building and trains troops.
func TrainTroop(ctx context.Context, b *browser.Browser, db *database.DB, bus *event.Bus, villageID int, building enum.Building) error {
	// Navigate to dorf2 where military buildings are located
	if err := navigate.ToDorf(ctx, b, 2); err != nil {
		return fmt.Errorf("navigate to dorf2: %w", err)
	}
	if err := update.UpdateBuildings(b, db, bus, villageID); err != nil {
		return fmt.Errorf("update buildings: %w", err)
	}
	if err := navigate.ToBuildingByType(ctx, b, db, villageID, int(building)); err != nil {
		return fmt.Errorf("navigate to %s: %w", building, err)
	}

	// Get configured troop
	troopSetting := troopSettings[building]
	troopID, _ := service.GetVillageSettingValue(db, villageID, troopSetting)
	if troopID == 0 {
		return nil // no troop configured
	}

	// Parse current page to check trainable amount
	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get page html: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse html: %w", err)
	}

	maxAmount := parser.GetMaxTrainAmount(doc, troopID)
	if maxAmount <= 0 {
		return &errs.TaskError{
			Err:     errs.ErrMissingResource,
			Message: fmt.Sprintf("cannot train troop %d (max=0)", troopID),
		}
	}

	// Get configured amount range
	amountRange := amountSettings[building]
	minAmount, _ := service.GetVillageSettingValue(db, villageID, amountRange[0])
	maxAmountSetting, _ := service.GetVillageSettingValue(db, villageID, amountRange[1])
	amount := service.RandomBetween(minAmount, maxAmountSetting)
	if amount <= 0 {
		amount = 1
	}

	if amount > maxAmount {
		trainWhenLow, _ := service.GetVillageSettingValue(db, villageID, enum.VillageSettingTrainWhenLowResource)
		if trainWhenLow == 0 {
			return &errs.TaskError{
				Err:     errs.ErrMissingResource,
				Message: fmt.Sprintf("want %d but max %d for troop %d", amount, maxAmount, troopID),
			}
		}
		amount = maxAmount
	}

	// Find the input element for this troop
	inputSelector := parser.GetTroopInputSelector(doc, troopID)
	if inputSelector == "" {
		return fmt.Errorf("cannot find input for troop %d", troopID)
	}

	inputEl, err := b.Element(inputSelector)
	if err != nil {
		return fmt.Errorf("find troop input: %w", err)
	}
	if err := b.Input(inputEl, fmt.Sprintf("%d", amount)); err != nil {
		return fmt.Errorf("input troop amount: %w", err)
	}

	// Click train button
	trainBtn, err := b.Element(parser.GetTrainButtonSelector())
	if err != nil {
		return fmt.Errorf("find train button: %w", err)
	}
	if err := b.Click(trainBtn); err != nil {
		return fmt.Errorf("click train button: %w", err)
	}

	return nil
}
