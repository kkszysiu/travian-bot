package task

import (
	"context"
	"fmt"
	"log"
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

// EvacuateResourcesTask sends all resources from a village to a configured safe village.
type EvacuateResourcesTask struct {
	BaseTask
	villageID int
	bus       *event.Bus
}

func NewEvacuateResourcesTask(accountID, villageID int, bus *event.Bus) *EvacuateResourcesTask {
	return &EvacuateResourcesTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		villageID: villageID,
		bus:       bus,
	}
}

func (t *EvacuateResourcesTask) Description() string { return "Evacuate resources" }
func (t *EvacuateResourcesTask) VillageID() int      { return t.villageID }

func (t *EvacuateResourcesTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
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

	// Send all resources
	if err := feature.EvacuateResources(ctx, b, db, t.bus, t.villageID, safeVillage.X, safeVillage.Y); err != nil {
		log.Printf("[EvacuateResources] Failed for village %d: %v (may have no merchants)", t.villageID, err)
		// Don't fail the task — resources are best-effort
		return nil
	}

	// Update evasion state: set bit 2 (resources evacuated)
	var currentState int
	db.Get(&currentState, "SELECT evasion_state FROM villages WHERE id = ?", t.villageID)
	newState := currentState | 2
	// Preserve existing target village ID
	var targetVillageID *int
	db.Get(&targetVillageID, "SELECT evasion_target_village_id FROM villages WHERE id = ?", t.villageID)
	if targetVillageID == nil {
		targetVillageID = &safeVillageID
	}
	if err := db.SetEvasionState(t.villageID, newState, targetVillageID); err != nil {
		return fmt.Errorf("update evasion state: %w", err)
	}

	t.bus.Emit(event.EvasionStateModified, t.villageID)
	t.bus.Emit(event.VillagesModified, t.accountID)

	return nil
}
