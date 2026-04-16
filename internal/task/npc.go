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

// NPCTask exchanges resources via the NPC merchant.
type NPCTask struct {
	BaseTask
	villageID int
	bus       *event.Bus
	scheduler *Scheduler
}

func NewNPCTask(accountID, villageID int, bus *event.Bus, scheduler *Scheduler) *NPCTask {
	return &NPCTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		villageID: villageID,
		bus:       bus,
		scheduler: scheduler,
	}
}

func (t *NPCTask) Description() string { return "NPC" }
func (t *NPCTask) VillageID() int      { return t.villageID }

func (t *NPCTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
	// Check if enabled
	enabled, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingAutoNPCEnable)
	if enabled == 0 {
		return errs.NewSkipError("npc disabled", time.Time{})
	}

	// Check gold (minimum 3)
	var gold int
	if err := db.Get(&gold,
		"SELECT gold FROM accounts_info WHERE account_id = ?",
		t.accountID,
	); err != nil || gold < 3 {
		return errs.NewSkipError("not enough gold for NPC", time.Time{})
	}

	// Check granary percent
	type storageData struct {
		Crop    int64 `db:"crop"`
		Granary int64 `db:"granary"`
	}
	var storage storageData
	if err := db.Get(&storage,
		"SELECT crop, granary FROM storages WHERE village_id = ?",
		t.villageID,
	); err != nil || storage.Granary == 0 {
		return errs.NewSkipError("no storage data", time.Time{})
	}

	granaryPercent := int(storage.Crop * 100 / storage.Granary)
	autoNPCGranaryPercent, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingAutoNPCGranaryPercent)
	if granaryPercent < autoNPCGranaryPercent {
		return errs.NewSkipError(
			fmt.Sprintf("granary %d%% < threshold %d%%", granaryPercent, autoNPCGranaryPercent),
			time.Time{},
		)
	}

	// Switch to village
	if err := navigate.SwitchVillage(ctx, b, t.villageID); err != nil {
		return fmt.Errorf("switch village: %w", err)
	}

	err := feature.NpcResource(ctx, b, db, t.bus, t.villageID)
	if err != nil {
		// Handle storage limit error - reschedule in 5 hours
		var taskErr *errs.TaskError
		if errors.As(err, &taskErr) && errors.Is(taskErr.Err, errs.ErrStorageLimit) {
			t.SetExecuteAt(time.Now().Add(5 * time.Hour))
			return nil
		}
		return err
	}

	// Queue upgrade building task
	t.scheduler.AddTask(t.accountID, NewUpgradeBuildingTask(t.accountID, t.villageID, t.bus))

	// One-shot task
	return nil
}
