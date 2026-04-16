package task

import (
	"context"
	"log"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/domain/model"
	"travian-bot/internal/event"
	"travian-bot/internal/parser"
	"travian-bot/internal/service"
)

// CheckAttacksTask periodically checks all villages for incoming attacks
// and triggers evasion or recall tasks as needed.
type CheckAttacksTask struct {
	BaseTask
	bus       *event.Bus
	scheduler *Scheduler
}

func NewCheckAttacksTask(accountID int, bus *event.Bus, scheduler *Scheduler) *CheckAttacksTask {
	return &CheckAttacksTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		bus:       bus,
		scheduler: scheduler,
	}
}

func (t *CheckAttacksTask) Description() string { return "Check attacks" }

func (t *CheckAttacksTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
	// Parse current page for village attack status
	html, err := b.PageHTML()
	if err != nil {
		t.reschedule(db)
		return nil
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		t.reschedule(db)
		return nil
	}

	parsed := parser.GetVillages(doc)
	if len(parsed) == 0 {
		t.reschedule(db)
		return nil
	}

	// Build map of parsed attack status
	attackMap := make(map[int]bool, len(parsed))
	for _, p := range parsed {
		attackMap[p.ID] = p.IsUnderAttack
	}

	// Get all villages from DB
	var villages []model.Village
	if err := db.Select(&villages,
		"SELECT id, account_id, name, x, y, is_active, is_under_attack, evasion_state, evasion_target_village_id FROM villages WHERE account_id = ?",
		t.accountID,
	); err != nil {
		t.reschedule(db)
		return nil
	}

	// Update is_under_attack in DB
	for _, v := range villages {
		if underAttack, ok := attackMap[v.ID]; ok {
			isAttacked := 0
			if underAttack {
				isAttacked = 1
			}
			db.Exec("UPDATE villages SET is_under_attack = ? WHERE id = ?", isAttacked, v.ID)
		}
	}

	// Build a set of villages currently under attack (for safe village checks)
	underAttackSet := make(map[int]bool)
	for _, v := range villages {
		if isAttacked, ok := attackMap[v.ID]; ok && isAttacked {
			underAttackSet[v.ID] = true
		}
	}

	// Process each village
	for _, v := range villages {
		isAttacked := attackMap[v.ID]

		// Check evasion settings
		evasionEnabled, _ := service.GetVillageSettingValue(db, v.ID, enum.VillageSettingAttackEvasionEnable)
		if evasionEnabled == 0 {
			continue
		}

		safeVillageID, _ := service.GetVillageSettingValue(db, v.ID, enum.VillageSettingAttackEvasionSafeVillageID)
		if safeVillageID == 0 {
			continue
		}

		if isAttacked && v.EvasionState == 0 {
			// Under attack and not yet evacuated — trigger evasion

			// Check if safe village is also under attack
			if underAttackSet[safeVillageID] {
				log.Printf("[CheckAttacks] Safe village %d is also under attack, skipping evasion for village %d (%s)",
					safeVillageID, v.ID, v.Name)
				continue
			}

			// Queue EvadeTroopsTask
			if !t.scheduler.HasVillageTask(t.accountID, "Evade troops", v.ID) {
				log.Printf("[CheckAttacks] Village %d (%s) under attack! Queueing troop evasion to village %d",
					v.ID, v.Name, safeVillageID)
				t.scheduler.AddTask(t.accountID, NewEvadeTroopsTask(t.accountID, v.ID, t.bus))
			}

			// Queue EvacuateResourcesTask if enabled
			evacResources, _ := service.GetVillageSettingValue(db, v.ID, enum.VillageSettingAttackEvasionEvacResources)
			if evacResources != 0 {
				if !t.scheduler.HasVillageTask(t.accountID, "Evacuate resources", v.ID) {
					// Delay resource evacuation slightly so troops go first
					evacTask := NewEvacuateResourcesTask(t.accountID, v.ID, t.bus)
					evacTask.SetExecuteAt(time.Now().Add(30 * time.Second))
					t.scheduler.AddTask(t.accountID, evacTask)
				}
			}
		} else if !isAttacked && v.EvasionState > 0 {
			// Attack is over and troops are evacuated — trigger recall
			if v.EvasionState&1 != 0 { // Troops were evacuated
				if !t.scheduler.HasVillageTask(t.accountID, "Recall troops", v.ID) {
					log.Printf("[CheckAttacks] Attack over for village %d (%s), queueing troop recall",
						v.ID, v.Name)
					t.scheduler.AddTask(t.accountID, NewRecallTroopsTask(t.accountID, v.ID, t.bus))
				}
			} else {
				// Only resources were sent (no troops to recall), just clear state
				db.ClearEvasionState(v.ID)
			}
		}
	}

	t.bus.Emit(event.VillagesModified, t.accountID)
	t.reschedule(db)
	return nil
}

func (t *CheckAttacksTask) reschedule(db *database.DB) {
	// Use minimum check interval across all evasion-enabled villages
	minInterval := 360 // default 6 minutes
	maxInterval := 360

	var villageIDs []int
	db.Select(&villageIDs, "SELECT id FROM villages WHERE account_id = ?", t.accountID)

	anyEnabled := false
	for _, vid := range villageIDs {
		enabled, _ := service.GetVillageSettingValue(db, vid, enum.VillageSettingAttackEvasionEnable)
		if enabled == 0 {
			continue
		}
		anyEnabled = true
		min, _ := service.GetVillageSettingValue(db, vid, enum.VillageSettingAttackEvasionCheckIntervalMin)
		max, _ := service.GetVillageSettingValue(db, vid, enum.VillageSettingAttackEvasionCheckIntervalMax)
		if min > 0 && min < minInterval {
			minInterval = min
		}
		if max > 0 && max < maxInterval {
			maxInterval = max
		}
	}

	if !anyEnabled {
		// No villages have evasion enabled — don't reschedule
		return
	}

	if maxInterval < minInterval {
		maxInterval = minInterval
	}

	seconds := service.RandomBetween(minInterval, maxInterval)
	t.SetExecuteAt(time.Now().Add(time.Duration(seconds) * time.Second))
}
