package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/command/feature"
	"travian-bot/internal/command/navigate"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/domain/errs"
	"travian-bot/internal/event"
	"travian-bot/internal/service"
)

// TrainTroopTask trains troops in configured buildings for a village.
type TrainTroopTask struct {
	BaseTask
	villageID int
	bus       *event.Bus
}

func NewTrainTroopTask(accountID, villageID int, bus *event.Bus) *TrainTroopTask {
	return &TrainTroopTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		villageID: villageID,
		bus:       bus,
	}
}

func (t *TrainTroopTask) Description() string { return "Train troop" }
func (t *TrainTroopTask) VillageID() int      { return t.villageID }

func (t *TrainTroopTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
	// Check if training is enabled for this village
	enabled, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingTrainTroopEnable)
	if enabled == 0 {
		return errs.NewSkipError("troop training disabled", time.Time{})
	}

	// Switch to the village
	if err := navigate.SwitchVillage(ctx, b, t.villageID); err != nil {
		return fmt.Errorf("switch to village %d: %w", t.villageID, err)
	}

	// Get training buildings with configured troops
	buildings := feature.GetTrainTroopBuildings(db, t.villageID)
	if len(buildings) == 0 {
		return errs.NewSkipError("no training buildings configured", time.Time{})
	}

	// For each building, train troops
	for _, building := range buildings {
		if ctx.Err() != nil {
			return &errs.TaskError{Err: errs.ErrCancel, Message: "cancelled"}
		}

		err := feature.TrainTroop(ctx, b, db, t.bus, t.villageID, building)
		if err != nil {
			// If missing resource, stop training remaining buildings
			var taskErr *errs.TaskError
			if errors.As(err, &taskErr) && errors.Is(taskErr.Err, errs.ErrMissingResource) {
				break
			}
			return err
		}
	}

	// Reschedule based on repeat time settings
	// Default values are 120 and 180 (in seconds), multiplied by 60 to get the actual delay
	minVal, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingTrainTroopRepeatTimeMin)
	maxVal, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingTrainTroopRepeatTimeMax)
	seconds := service.RandomBetween(minVal, maxVal) * 60
	t.SetExecuteAt(time.Now().Add(time.Duration(seconds) * time.Second))

	return nil
}
