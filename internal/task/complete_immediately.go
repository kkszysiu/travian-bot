package task

import (
	"context"
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

// unskippableBuildings are buildings that cannot be completed immediately.
var unskippableBuildings = map[enum.Building]bool{
	enum.BuildingResidence:     true,
	enum.BuildingPalace:        true,
	enum.BuildingCommandCenter: true,
}

// CompleteImmediatelyTask uses gold to instantly complete building constructions.
type CompleteImmediatelyTask struct {
	BaseTask
	villageID int
	bus       *event.Bus
	scheduler *Scheduler
}

func NewCompleteImmediatelyTask(accountID, villageID int, bus *event.Bus, scheduler *Scheduler) *CompleteImmediatelyTask {
	return &CompleteImmediatelyTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		villageID: villageID,
		bus:       bus,
		scheduler: scheduler,
	}
}

func (t *CompleteImmediatelyTask) Description() string { return "Complete immediately" }
func (t *CompleteImmediatelyTask) VillageID() int      { return t.villageID }

func (t *CompleteImmediatelyTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
	// Check if enabled
	enabled, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingCompleteImmediately)
	if enabled == 0 {
		return errs.NewSkipError("complete immediately disabled", time.Time{})
	}

	// Check queue buildings - must have buildings and none unskippable
	type queueItem struct {
		Type         int       `db:"type"`
		CompleteTime time.Time `db:"complete_time"`
	}
	var queueBuildings []queueItem
	if err := db.Select(&queueBuildings,
		"SELECT type, complete_time FROM queue_buildings WHERE village_id = ?",
		t.villageID,
	); err != nil {
		return fmt.Errorf("get queue buildings: %w", err)
	}

	if len(queueBuildings) == 0 {
		return errs.NewSkipError("no queue buildings", time.Time{})
	}

	// Check for unskippable buildings
	for _, qb := range queueBuildings {
		if unskippableBuildings[enum.Building(qb.Type)] {
			return errs.NewSkipError("unskippable building in queue", time.Time{})
		}
	}

	// Check minimum time threshold
	completeImmediatelyTime, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingCompleteImmediatelyTime)
	requiredTime := time.Now().Add(time.Duration(completeImmediatelyTime) * time.Minute)
	anyEligible := false
	for _, qb := range queueBuildings {
		if qb.CompleteTime.After(requiredTime) {
			anyEligible = true
			break
		}
	}
	if !anyEligible {
		return errs.NewSkipError("no buildings eligible for immediate completion", time.Time{})
	}

	// Switch to village and complete
	if err := navigate.SwitchVillage(ctx, b, t.villageID); err != nil {
		return fmt.Errorf("switch village: %w", err)
	}

	if err := feature.CompleteImmediately(ctx, b); err != nil {
		return err
	}

	// Queue an upgrade building task to continue building
	t.scheduler.AddTask(t.accountID, NewUpgradeBuildingTask(t.accountID, t.villageID, t.bus))

	// One-shot task
	return nil
}
