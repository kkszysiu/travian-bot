package task

import (
	"context"
	"fmt"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/command/feature"
	"travian-bot/internal/command/navigate"
	"travian-bot/internal/command/update"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/domain/errs"
	"travian-bot/internal/domain/model"
	"travian-bot/internal/event"
	"travian-bot/internal/parser"
	"travian-bot/internal/service"
)

// LoginTask logs into the Travian server, updates account/village data,
// and queues follow-up tasks (UpdateBuilding, UpgradeBuilding per village).
type LoginTask struct {
	BaseTask
	bus       *event.Bus
	scheduler *Scheduler
	browsers  *browser.Manager
}

func NewLoginTask(accountID int, bus *event.Bus, scheduler *Scheduler, browsers *browser.Manager) *LoginTask {
	return &LoginTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		bus:       bus,
		scheduler: scheduler,
		browsers:  browsers,
	}
}

func (t *LoginTask) Description() string { return "Login" }

func (t *LoginTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
	// Step 1: Login
	if err := feature.Login(ctx, b, db, t.accountID); err != nil {
		return err
	}

	// Step 2: Check if contextual help needs to be disabled
	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get page html: %w", err)
	}

	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse html: %w", err)
	}

	if parser.IsContextualHelpEnabled(doc) {
		if err := feature.DisableContextualHelp(ctx, b); err != nil {
			return fmt.Errorf("disable contextual help: %w", err)
		}

		if err := navigate.ToDorf(ctx, b, 0); err != nil {
			return &errs.TaskError{Err: errs.ErrRetry, Message: fmt.Sprintf("navigate to dorf: %v", err)}
		}
	}

	// Step 3: Update account info
	if err := update.UpdateAccountInfo(b, db, t.accountID); err != nil {
		// Non-fatal, just log
		_ = err
	}

	// Step 4: Update village list
	if err := update.UpdateVillageList(b, db, t.bus, t.accountID); err != nil {
		return fmt.Errorf("update village list: %w", err)
	}

	// Step 4b: Update storage for the active village (resource bar is visible)
	var activeVillageID int
	if err := db.Get(&activeVillageID,
		"SELECT id FROM villages WHERE account_id = ? AND is_active = 1 LIMIT 1",
		t.accountID,
	); err == nil && activeVillageID > 0 {
		if err := update.UpdateStorage(b, db, t.bus, activeVillageID); err != nil {
			_ = err // Non-fatal
		}
	}

	// Step 5: Queue follow-up tasks for each village
	var villages []model.Village
	if err := db.Select(&villages,
		"SELECT id, account_id, name, x, y, is_active, is_under_attack, evasion_state, evasion_target_village_id FROM villages WHERE account_id = ?",
		t.accountID,
	); err != nil {
		return fmt.Errorf("get villages: %w", err)
	}

	autoLoadBuildings, _ := service.GetAccountSettingValue(db, t.accountID, enum.AccountSettingEnableAutoLoadVillageBuilding)

	for _, v := range villages {
		// Queue UpdateBuildingTask for each village (if enabled)
		if autoLoadBuildings != 0 {
			t.scheduler.AddTask(t.accountID, NewUpdateBuildingTask(t.accountID, v.ID, t.bus))
		}

		// Queue UpgradeBuildingTask for each village that has jobs
		var jobCount int
		db.Get(&jobCount, "SELECT COUNT(*) FROM jobs WHERE village_id = ?", v.ID)
		if jobCount > 0 {
			t.scheduler.AddTask(t.accountID, NewUpgradeBuildingTask(t.accountID, v.ID, t.bus))
		}

		// Queue TrainTroopTask for each village with training enabled
		trainEnabled, _ := service.GetVillageSettingValue(db, v.ID, enum.VillageSettingTrainTroopEnable)
		if trainEnabled != 0 {
			t.scheduler.AddTask(t.accountID, NewTrainTroopTask(t.accountID, v.ID, t.bus))
		}

		// Queue ClaimQuestTask for each village with quest claiming enabled
		questEnabled, _ := service.GetVillageSettingValue(db, v.ID, enum.VillageSettingAutoClaimQuestEnable)
		if questEnabled != 0 {
			t.scheduler.AddTask(t.accountID, NewClaimQuestTask(t.accountID, v.ID, t.bus))
		}

		// Queue CompleteImmediatelyTask for each village with instant completion enabled
		completeEnabled, _ := service.GetVillageSettingValue(db, v.ID, enum.VillageSettingCompleteImmediately)
		if completeEnabled != 0 {
			t.scheduler.AddTask(t.accountID, NewCompleteImmediatelyTask(t.accountID, v.ID, t.bus, t.scheduler))
		}

		// Queue NPCTask for each village with auto NPC enabled
		npcEnabled, _ := service.GetVillageSettingValue(db, v.ID, enum.VillageSettingAutoNPCEnable)
		if npcEnabled != 0 {
			t.scheduler.AddTask(t.accountID, NewNPCTask(t.accountID, v.ID, t.bus, t.scheduler))
		}

		// Queue SendResourcesTask for each village with auto send enabled
		sendEnabled, _ := service.GetVillageSettingValue(db, v.ID, enum.VillageSettingAutoSendResourceEnable)
		if sendEnabled != 0 {
			t.scheduler.AddTask(t.accountID, NewSendResourcesTask(t.accountID, v.ID, t.bus))
		}

		// Queue UpdateVillageTask for each village with auto-refresh enabled
		refreshEnabled, _ := service.GetVillageSettingValue(db, v.ID, enum.VillageSettingAutoRefreshEnable)
		if refreshEnabled != 0 {
			t.scheduler.AddTask(t.accountID, NewUpdateVillageTask(t.accountID, v.ID, t.bus))
		}
	}

	// Queue CheckAttacksTask if any village has attack evasion enabled
	for _, v := range villages {
		evasionEnabled, _ := service.GetVillageSettingValue(db, v.ID, enum.VillageSettingAttackEvasionEnable)
		if evasionEnabled != 0 {
			t.scheduler.AddTask(t.accountID, NewCheckAttacksTask(t.accountID, t.bus, t.scheduler))
			break
		}
	}

	// Queue account-level tasks: UpdateFarmList + StartFarmList
	t.scheduler.AddTask(t.accountID, NewUpdateFarmListTask(t.accountID, t.bus))
	t.scheduler.AddTask(t.accountID, NewStartFarmListTask(t.accountID, t.bus))

	// Queue StartAdventureTask if enabled
	adventureEnabled, _ := service.GetAccountSettingValue(db, t.accountID, enum.AccountSettingEnableAutoStartAdventure)
	if adventureEnabled != 0 {
		t.scheduler.AddTask(t.accountID, NewStartAdventureTask(t.accountID, t.bus))
	}

	// Queue SleepTask based on work window
	sleepTask := NewSleepTask(t.accountID, t.bus, t.scheduler, t.browsers)
	ww := service.GetWorkWindow(db, t.accountID)
	now := time.Now()
	if ww.IsOutsideWindow(now) {
		// Outside window: schedule sleep at next work start (will wake then)
		sleepTask.SetExecuteAt(ww.NextWorkStart(now))
	} else {
		// Inside window: schedule sleep at work end + jitter
		sleepTask.SetExecuteAt(ww.NextWorkEnd(now, true))
	}
	t.scheduler.AddTask(t.accountID, sleepTask)

	return nil
}
