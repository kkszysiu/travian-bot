package task

import (
	"context"
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

// ClaimQuestTask claims available quests for a village.
type ClaimQuestTask struct {
	BaseTask
	villageID int
	bus       *event.Bus
}

func NewClaimQuestTask(accountID, villageID int, bus *event.Bus) *ClaimQuestTask {
	return &ClaimQuestTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		villageID: villageID,
		bus:       bus,
	}
}

func (t *ClaimQuestTask) Description() string { return "Claim quest" }
func (t *ClaimQuestTask) VillageID() int      { return t.villageID }

func (t *ClaimQuestTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
	// Check if enabled
	enabled, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingAutoClaimQuestEnable)
	if enabled == 0 {
		return errs.NewSkipError("quest claiming disabled", time.Time{})
	}

	// Switch to the village
	if err := navigate.SwitchVillage(ctx, b, t.villageID); err != nil {
		return err
	}

	if err := feature.ClaimQuest(ctx, b); err != nil {
		return err
	}

	// One-shot task: return nil and let scheduler remove it
	return nil
}
