package task

import (
	"context"
	"fmt"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/command/feature"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/domain/errs"
	"travian-bot/internal/domain/model"
	"travian-bot/internal/event"
	"travian-bot/internal/service"
)

// EvadeTroopsTask sends all troops from a village to a configured safe village as reinforcement.
type EvadeTroopsTask struct {
	BaseTask
	villageID int
	bus       *event.Bus
}

func NewEvadeTroopsTask(accountID, villageID int, bus *event.Bus) *EvadeTroopsTask {
	return &EvadeTroopsTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		villageID: villageID,
		bus:       bus,
	}
}

func (t *EvadeTroopsTask) Description() string { return "Evade troops" }
func (t *EvadeTroopsTask) VillageID() int      { return t.villageID }

func (t *EvadeTroopsTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
	// Get safe village ID from settings
	safeVillageID, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingAttackEvasionSafeVillageID)
	if safeVillageID == 0 {
		return errs.NewSkipError("no safe village configured", time.Time{})
	}

	// Look up safe village coordinates
	var safeVillage model.Village
	err := db.Get(&safeVillage,
		"SELECT id, account_id, name, x, y, is_active, is_under_attack, evasion_state, evasion_target_village_id FROM villages WHERE id = ?",
		safeVillageID,
	)
	if err != nil {
		return fmt.Errorf("get safe village: %w", err)
	}

	// Send all troops as reinforcement
	if err := feature.EvadeTroops(ctx, b, db, t.bus, t.villageID, safeVillage.X, safeVillage.Y); err != nil {
		return fmt.Errorf("evade troops: %w", err)
	}

	// Update evasion state: set bit 1 (troops evacuated)
	var currentState int
	db.Get(&currentState, "SELECT evasion_state FROM villages WHERE id = ?", t.villageID)
	newState := currentState | 1
	targetID := safeVillageID
	if err := db.SetEvasionState(t.villageID, newState, &targetID); err != nil {
		return fmt.Errorf("update evasion state: %w", err)
	}

	t.bus.Emit(event.EvasionStateModified, t.villageID)
	t.bus.Emit(event.VillagesModified, t.accountID)

	return nil
}
