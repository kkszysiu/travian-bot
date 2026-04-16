package app

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/event"
	"travian-bot/internal/service"
	"travian-bot/internal/task"
)

// App is the main application struct bound to Wails.
// All exported methods become available as frontend API calls.
type App struct {
	ctx       context.Context
	db        *database.DB
	bus       *event.Bus
	browsers  *browser.Manager
	scheduler *task.Scheduler
	log       *slog.Logger
	logs      *logStore

	// Per-account status tracking
	mu       sync.RWMutex
	statuses map[int]enum.Status
}

// NewApp creates a new App instance.
func NewApp(db *database.DB, bus *event.Bus, browsers *browser.Manager, scheduler *task.Scheduler, logger *slog.Logger) *App {
	return &App{
		db:        db,
		bus:       bus,
		browsers:  browsers,
		scheduler: scheduler,
		log:       logger,
		logs:      newLogStore(),
		statuses:  make(map[int]enum.Status),
	}
}

// Startup is called when the Wails app starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.bus.SetContext(ctx)

	// Subscribe to log events to populate the per-account log store.
	a.bus.On(event.LogEmitted, func(data interface{}) {
		if p, ok := data.(event.LogPayload); ok {
			a.logs.Add(p.AccountID, p.Level, p.Message)
		}
	})

	a.log.Info("application started")
}

// LogForAccount records a log entry for a specific account and emits it to the frontend.
func (a *App) LogForAccount(accountID int, level, message string) {
	a.logs.Add(accountID, level, message)
	a.bus.Emit(event.LogEmitted, event.LogPayload{
		AccountID: accountID,
		Message:   message,
		Level:     level,
	})
}

// Shutdown is called when the Wails app shuts down.
func (a *App) Shutdown(ctx context.Context) {
	a.log.Info("application shutting down")
	if a.scheduler != nil {
		a.scheduler.Shutdown()
	}
	if a.browsers != nil {
		a.browsers.Shutdown()
	}
	if a.db != nil {
		a.db.Close()
	}
}

// --- Account Management ---

func (a *App) GetAccounts() ([]database.AccountListItem, error) {
	return a.db.GetAccounts()
}

func (a *App) AddAccount(detail database.AccountDetail) error {
	id, err := a.db.AddAccount(detail)
	if err != nil {
		return fmt.Errorf("failed to add account: %w", err)
	}
	a.log.Info("account added", "id", id, "username", detail.Username)
	a.bus.Emit(event.AccountsModified, nil)
	return nil
}

func (a *App) UpdateAccount(detail database.AccountDetail) error {
	if err := a.db.UpdateAccount(detail); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}
	a.bus.Emit(event.AccountsModified, nil)
	return nil
}

func (a *App) DeleteAccount(accountID int) error {
	if err := a.db.DeleteAccount(accountID); err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}
	a.mu.Lock()
	delete(a.statuses, accountID)
	a.mu.Unlock()
	a.bus.Emit(event.AccountsModified, nil)
	return nil
}

func (a *App) GetAccountDetail(accountID int) (*database.AccountDetail, error) {
	return a.db.GetAccountDetail(accountID)
}

// --- Account Actions ---

func (a *App) Login(accountID int) error {
	a.setStatus(accountID, enum.StatusStarting)

	// Get account details for browser config
	detail, err := a.db.GetAccountDetail(accountID)
	if err != nil {
		a.setStatus(accountID, enum.StatusOffline)
		return fmt.Errorf("get account detail: %w", err)
	}

	// Build browser config from the first (most recent) access
	cfg := browser.Config{
		ProfilePath: strconv.Itoa(accountID),
	}
	if len(detail.Accesses) > 0 {
		access := detail.Accesses[0]
		cfg.ProxyHost = access.ProxyHost
		cfg.ProxyPort = access.ProxyPort
		cfg.ProxyUsername = access.ProxyUsername
		cfg.ProxyPassword = access.ProxyPassword
		cfg.UserAgent = access.Useragent
	}

	// Launch browser
	b, err := a.browsers.Create(accountID, cfg)
	if err != nil {
		a.setStatus(accountID, enum.StatusOffline)
		return fmt.Errorf("launch browser: %w", err)
	}

	// Navigate to server URL
	if err := b.Navigate(detail.Server); err != nil {
		a.browsers.Close(accountID)
		a.setStatus(accountID, enum.StatusOffline)
		return fmt.Errorf("navigate to server: %w", err)
	}

	// Clear stale tasks from any previous session and queue a fresh login
	a.scheduler.ClearTasks(accountID)
	loginTask := task.NewLoginTask(accountID, a.bus, a.scheduler, a.browsers)
	a.scheduler.AddTask(accountID, loginTask)
	a.scheduler.Start(accountID)

	a.setStatus(accountID, enum.StatusOnline)
	a.log.Info("account login started", "accountId", accountID)
	return nil
}

func (a *App) Logout(accountID int) error {
	a.setStatus(accountID, enum.StatusStopping)

	// Stop scheduler and close browser
	a.scheduler.Stop(accountID)
	a.browsers.Close(accountID)

	a.setStatus(accountID, enum.StatusOffline)
	a.log.Info("account logged out", "accountId", accountID)
	return nil
}

func (a *App) Pause(accountID int) error {
	a.scheduler.Pause(accountID)
	a.setStatus(accountID, enum.StatusPaused)
	return nil
}

func (a *App) Restart(accountID int) error {
	a.scheduler.Resume(accountID)
	a.setStatus(accountID, enum.StatusOnline)
	return nil
}

func (a *App) GetStatus(accountID int) int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return int(a.statuses[accountID])
}

func (a *App) setStatus(accountID int, status enum.Status) {
	a.mu.Lock()
	a.statuses[accountID] = status
	a.mu.Unlock()
	a.bus.Emit(event.StatusModified, event.StatusPayload{
		AccountID: accountID,
		Status:    int(status),
		Color:     status.Color(),
	})
}

// --- Village Data ---

func (a *App) RefreshVillage(accountID, villageID int) {
	a.scheduler.AddTask(accountID, task.NewUpdateVillageTask(accountID, villageID, a.bus))
}

func (a *App) GetVillages(accountID int) ([]database.VillageListItem, error) {
	return a.db.GetVillages(accountID)
}

func (a *App) GetBuildings(villageID int) ([]database.BuildingItem, error) {
	return a.db.GetBuildings(villageID)
}

func (a *App) GetStorage(villageID int) (*database.StorageDTO, error) {
	return a.db.GetStorage(villageID)
}

func (a *App) GetQueueBuildings(villageID int) ([]database.QueueBuildingItem, error) {
	return a.db.GetQueueBuildings(villageID)
}

func (a *App) GetJobs(villageID int) ([]database.JobItem, error) {
	return a.db.GetJobs(villageID)
}

// --- Job Management ---

func (a *App) AddNormalBuildJob(input database.NormalBuildInput) error {
	// Validate prerequisites for new construction on empty sites
	if input.Location > 0 && input.Type > 0 {
		var existingType int
		a.db.Get(&existingType,
			"SELECT COALESCE(type, 0) FROM buildings WHERE village_id = ? AND location = ?",
			input.VillageID, input.Location)
		if existingType == 0 {
			bt := enum.Building(input.Type)
			for _, p := range bt.GetPrerequisiteBuildings() {
				var maxLevel int
				a.db.Get(&maxLevel,
					"SELECT COALESCE(MAX(level), 0) FROM buildings WHERE village_id = ? AND type = ?",
					input.VillageID, int(p.Building))
				if maxLevel < p.Level {
					return fmt.Errorf("prerequisite not met: %s level %d required (have %d)",
						p.Building.String(), p.Level, maxLevel)
				}
			}
		}
	}

	if err := a.db.AddNormalBuildJob(input); err != nil {
		return err
	}
	a.bus.Emit(event.JobsModified, input.VillageID)
	a.ensureUpgradeTask(input.VillageID)
	return nil
}

func (a *App) AddResourceBuildJob(input database.ResourceBuildInput) error {
	if err := a.db.AddResourceBuildJob(input); err != nil {
		return err
	}
	a.bus.Emit(event.JobsModified, input.VillageID)
	a.ensureUpgradeTask(input.VillageID)
	return nil
}

func (a *App) DeleteJob(jobID int) error {
	if err := a.requirePausedForJob(jobID); err != nil {
		return err
	}
	if err := a.db.DeleteJob(jobID); err != nil {
		return err
	}
	a.bus.Emit(event.JobsModified, nil)
	return nil
}

func (a *App) DeleteAllJobs(villageID int) error {
	if err := a.requirePausedForVillage(villageID); err != nil {
		return err
	}
	if err := a.db.DeleteAllJobs(villageID); err != nil {
		return err
	}
	a.bus.Emit(event.JobsModified, villageID)
	return nil
}

func (a *App) MoveJob(jobID int, direction string) error {
	if err := a.requirePausedForJob(jobID); err != nil {
		return err
	}
	if err := a.db.MoveJob(jobID, direction); err != nil {
		return err
	}
	a.bus.Emit(event.JobsModified, nil)
	return nil
}

func (a *App) ImportJobs(villageID int, jsonData string) error {
	if err := a.requirePausedForVillage(villageID); err != nil {
		return err
	}
	if err := a.db.ImportJobs(villageID, jsonData); err != nil {
		return err
	}
	a.bus.Emit(event.JobsModified, villageID)
	a.ensureUpgradeTask(villageID)
	return nil
}

// requirePausedForVillage returns an error if the account owning the village is online.
func (a *App) requirePausedForVillage(villageID int) error {
	accountID, err := a.db.GetAccountIDForVillage(villageID)
	if err != nil {
		return nil // can't determine account, allow
	}
	if a.scheduler.GetStatus(accountID) == enum.StatusOnline {
		return fmt.Errorf("pause the account before making changes")
	}
	return nil
}

// requirePausedForJob returns an error if the account owning the job is online.
func (a *App) requirePausedForJob(jobID int) error {
	accountID, err := a.db.GetAccountIDForJob(jobID)
	if err != nil {
		return nil // can't determine account, allow
	}
	if a.scheduler.GetStatus(accountID) == enum.StatusOnline {
		return fmt.Errorf("pause the account before making changes")
	}
	return nil
}

func (a *App) ExportJobs(villageID int) (string, error) {
	return a.db.ExportJobs(villageID)
}

// ensureUpgradeTask queues an UpgradeBuildingTask for the village if the account
// is online and no such task already exists. Matches C# behavior where adding a
// job automatically triggers TaskManager.AddOrUpdate(UpgradeBuildingTask).
func (a *App) ensureUpgradeTask(villageID int) {
	accountID, err := a.db.GetAccountIDForVillage(villageID)
	if err != nil {
		return
	}
	status := a.scheduler.GetStatus(accountID)
	if status != enum.StatusOnline {
		return
	}
	// Matches C# TaskManager.AddOrUpdate: add if missing, update ExecuteAt if exists
	a.scheduler.AddOrUpdateVillageTask(accountID, task.NewUpgradeBuildingTask(accountID, villageID, a.bus), villageID)
}

// --- Settings ---

func (a *App) GetAccountSettings(accountID int) (map[string]int, error) {
	return a.db.GetAccountSettings(accountID)
}

func (a *App) SaveAccountSettings(accountID int, settings map[string]int) error {
	if err := a.db.SaveAccountSettings(accountID, settings); err != nil {
		return err
	}

	// Apply runtime side effects for running accounts (matches C# SaveAccountSettingCommand)
	a.applyAccountSettingSideEffects(accountID, settings)
	return nil
}

// applyAccountSettingSideEffects updates the running task queue when settings change.
func (a *App) applyAccountSettingSideEffects(accountID int, settings map[string]int) {
	if a.scheduler.GetStatus(accountID) != enum.StatusOnline {
		return
	}

	// EnableAutoStartAdventure: add/remove adventure task
	if val, ok := settings["EnableAutoStartAdventure"]; ok {
		if val != 0 {
			if !a.scheduler.HasTaskOfType(accountID, "Start adventure") {
				a.scheduler.AddTask(accountID, task.NewStartAdventureTask(accountID, a.bus))
			}
		} else {
			a.scheduler.RemoveTaskOfType(accountID, "Start adventure")
		}
	}

	// Work window changes: reschedule sleep task to new work end time
	ww := service.GetWorkWindow(a.db, accountID)
	now := time.Now()
	if ww.IsOutsideWindow(now) {
		// Outside window: sleep should fire at next work start
		a.scheduler.RescheduleTask(accountID, "Sleep", ww.NextWorkStart(now))
	} else {
		// Inside window: sleep should fire at work end + jitter
		a.scheduler.RescheduleTask(accountID, "Sleep", ww.NextWorkEnd(now, true))
	}
}

func (a *App) GetVillageSettings(villageID int) (map[string]int, error) {
	return a.db.GetVillageSettings(villageID)
}

func (a *App) SaveVillageSettings(villageID int, settings map[string]int) error {
	if err := a.db.SaveVillageSettings(villageID, settings); err != nil {
		return err
	}

	// Apply runtime side effects for running accounts (matches C# SaveVillageSettingCommand)
	a.applyVillageSettingSideEffects(villageID, settings)
	return nil
}

// applyVillageSettingSideEffects updates the running task queue when village settings change.
func (a *App) applyVillageSettingSideEffects(villageID int, settings map[string]int) {
	accountID, err := a.db.GetAccountIDForVillage(villageID)
	if err != nil {
		return
	}
	if a.scheduler.GetStatus(accountID) != enum.StatusOnline {
		return
	}

	// CompleteImmediately: add/remove task
	if val, ok := settings["CompleteImmediately"]; ok {
		if val != 0 {
			if !a.scheduler.HasVillageTask(accountID, "Complete immediately", villageID) {
				a.scheduler.AddTask(accountID, task.NewCompleteImmediatelyTask(accountID, villageID, a.bus, a.scheduler))
			}
		} else {
			a.scheduler.RemoveVillageTask(accountID, "Complete immediately", villageID)
		}
	}

	// TrainTroopEnable: add/remove task
	if val, ok := settings["TrainTroopEnable"]; ok {
		if val != 0 {
			if !a.scheduler.HasVillageTask(accountID, "Train troop", villageID) {
				a.scheduler.AddTask(accountID, task.NewTrainTroopTask(accountID, villageID, a.bus))
			}
		} else {
			a.scheduler.RemoveVillageTask(accountID, "Train troop", villageID)
		}
	}

	// AutoNPCEnable: add/remove task
	if val, ok := settings["AutoNPCEnable"]; ok {
		if val != 0 {
			if !a.scheduler.HasVillageTask(accountID, "NPC", villageID) {
				a.scheduler.AddTask(accountID, task.NewNPCTask(accountID, villageID, a.bus, a.scheduler))
			}
		} else {
			a.scheduler.RemoveVillageTask(accountID, "NPC", villageID)
		}
	}

	// AutoClaimQuestEnable: add/remove task
	if val, ok := settings["AutoClaimQuestEnable"]; ok {
		if val != 0 {
			if !a.scheduler.HasVillageTask(accountID, "Claim quest", villageID) {
				a.scheduler.AddTask(accountID, task.NewClaimQuestTask(accountID, villageID, a.bus))
			}
		} else {
			a.scheduler.RemoveVillageTask(accountID, "Claim quest", villageID)
		}
	}

	// AutoRefreshEnable: add/remove update village task
	if val, ok := settings["AutoRefreshEnable"]; ok {
		if val != 0 {
			if !a.scheduler.HasVillageTask(accountID, "Update village", villageID) {
				a.scheduler.AddTask(accountID, task.NewUpdateVillageTask(accountID, villageID, a.bus))
			}
		} else {
			a.scheduler.RemoveVillageTask(accountID, "Update village", villageID)
		}
	}

	// AutoSendResourceEnable: add/remove send resources task
	if val, ok := settings["AutoSendResourceEnable"]; ok {
		if val != 0 {
			if !a.scheduler.HasVillageTask(accountID, "SendResources", villageID) {
				a.scheduler.AddTask(accountID, task.NewSendResourcesTask(accountID, villageID, a.bus))
			}
		} else {
			a.scheduler.RemoveVillageTask(accountID, "SendResources", villageID)
		}
	}

	// AttackEvasionEnable: add check attacks task if not already running
	if val, ok := settings["AttackEvasionEnable"]; ok {
		if val != 0 {
			if !a.scheduler.HasTaskOfType(accountID, "Check attacks") {
				a.scheduler.AddTask(accountID, task.NewCheckAttacksTask(accountID, a.bus, a.scheduler))
			}
		}
	}
}

// --- Transfer Rules ---

func (a *App) GetTransferRules(villageID int) ([]database.TransferRuleDTO, error) {
	return a.db.GetTransferRules(villageID)
}

func (a *App) AddTransferRule(input database.TransferRuleInput) error {
	if err := a.db.AddTransferRule(input); err != nil {
		return err
	}
	a.bus.Emit(event.TransferRulesModified, input.VillageID)
	a.ensureSendResourcesTask(input.VillageID)
	return nil
}

func (a *App) DeleteTransferRule(ruleID int) error {
	if err := a.db.DeleteTransferRule(ruleID); err != nil {
		return err
	}
	a.bus.Emit(event.TransferRulesModified, nil)
	return nil
}

func (a *App) DeleteAllTransferRules(villageID int) error {
	if err := a.db.DeleteAllTransferRules(villageID); err != nil {
		return err
	}
	a.bus.Emit(event.TransferRulesModified, villageID)
	return nil
}

// ensureSendResourcesTask queues a SendResourcesTask for the village if the account is online.
func (a *App) ensureSendResourcesTask(villageID int) {
	accountID, err := a.db.GetAccountIDForVillage(villageID)
	if err != nil {
		return
	}
	status := a.scheduler.GetStatus(accountID)
	if status != enum.StatusOnline {
		return
	}
	a.scheduler.AddOrUpdateVillageTask(accountID, task.NewSendResourcesTask(accountID, villageID, a.bus), villageID)
}

// --- Farm Lists ---

func (a *App) GetFarmLists(accountID int) ([]database.FarmItem, error) {
	return a.db.GetFarmLists(accountID)
}

func (a *App) ToggleFarmList(farmID int, active bool) error {
	val := 0
	if active {
		val = 1
	}
	if _, err := a.db.Exec("UPDATE farm_lists SET is_active = ? WHERE id = ?", val, farmID); err != nil {
		return err
	}
	a.bus.Emit(event.FarmsModified, nil)
	return nil
}

// --- Static Game Data ---

type BuildingTypeInfo struct {
	Type int    `json:"type"`
	Name string `json:"name"`
}

func (a *App) GetBuildingTypes() []BuildingTypeInfo {
	types := make([]BuildingTypeInfo, 0, 47)
	for i := 0; i <= 46; i++ {
		b := enum.Building(i)
		types = append(types, BuildingTypeInfo{
			Type: i,
			Name: b.String(),
		})
	}
	return types
}

// GetAvailableNewBuildings returns building types that can be constructed on empty
// infrastructure sites in a village. Excludes buildings already present (unless
// they allow multiples), walls, resource fields, WW, and deprecated buildings.
func (a *App) GetAvailableNewBuildings(villageID int) []BuildingTypeInfo {
	// Collect existing building types in the village
	var buildings []struct {
		Type int `db:"type"`
	}
	a.db.Select(&buildings, "SELECT type FROM buildings WHERE village_id = ? AND type > 0", villageID)

	existingTypes := make(map[int]bool)
	for _, b := range buildings {
		existingTypes[b.Type] = true
	}

	var available []BuildingTypeInfo
	for i := 5; i <= 46; i++ { // skip resource fields (1-4) and Site (0)
		bt := enum.Building(i)
		if bt == enum.BuildingBlacksmith || bt == enum.BuildingWW {
			continue
		}
		if bt.IsWall() {
			continue
		}
		if existingTypes[i] && !bt.IsMultipleBuilding() {
			continue
		}
		available = append(available, BuildingTypeInfo{Type: i, Name: bt.String()})
	}
	return available
}

type TroopTypeInfo struct {
	Type int    `json:"type"`
	Name string `json:"name"`
}

func (a *App) GetTroopTypes(tribe int, building int) []TroopTypeInfo {
	// Normalize Great Barracks/Stable to their base building for filtering
	bld := enum.Building(building)
	if bld == enum.BuildingGreatBarracks {
		bld = enum.BuildingBarracks
	} else if bld == enum.BuildingGreatStable {
		bld = enum.BuildingStable
	}

	var troops []TroopTypeInfo
	for i := 1; i <= 71; i++ {
		t := enum.Troop(i)
		if tribe != 0 && int(t.GetTribe()) != tribe {
			continue
		}
		if building != 0 && t.GetTrainBuilding() != bld {
			continue
		}
		troops = append(troops, TroopTypeInfo{
			Type: i,
			Name: t.String(),
		})
	}
	return troops
}

// --- Debug (stubs for Phase 5) ---

type TaskItem struct {
	Task      string `json:"task"`
	ExecuteAt string `json:"executeAt"`
	Stage     string `json:"stage"`
}

type LogEntry struct {
	Message string `json:"message"`
	Level   string `json:"level"`
	Time    string `json:"time"`
}

func (a *App) GetTasks(accountID int) []TaskItem {
	tasks := a.scheduler.GetTasks(accountID)
	items := make([]TaskItem, len(tasks))
	for i, t := range tasks {
		items[i] = TaskItem{
			Task:      t.Description(),
			ExecuteAt: t.ExecuteAt().Format("2006-01-02 15:04:05"),
			Stage:     t.Stage().String(),
		}
	}
	return items
}

func (a *App) GetLogs(accountID int) []LogEntry {
	entries := a.logs.Get(accountID)
	result := make([]LogEntry, len(entries))
	for i, e := range entries {
		result[i] = LogEntry{Message: e.Message, Level: e.Level, Time: e.Time}
	}
	return result
}
