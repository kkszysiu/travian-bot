package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
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

// UpgradeBuildingTask processes the build job queue for a village,
// upgrading or constructing buildings one at a time.
type UpgradeBuildingTask struct {
	BaseTask
	villageID int
	bus       *event.Bus
}

func NewUpgradeBuildingTask(accountID, villageID int, bus *event.Bus) *UpgradeBuildingTask {
	return &UpgradeBuildingTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		villageID: villageID,
		bus:       bus,
	}
}

func (t *UpgradeBuildingTask) Description() string { return "Upgrade building" }
func (t *UpgradeBuildingTask) VillageID() int      { return t.villageID }

func (t *UpgradeBuildingTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
	for {
		if ctx.Err() != nil {
			return &errs.TaskError{Err: errs.ErrCancel, Message: "cancelled"}
		}

		// Step 1: Get the next build plan from the job queue
		plan, err := t.getBuildPlan(ctx, b, db)
		if err != nil {
			return err
		}

		// Step 2: Navigate to build page
		if err := t.toBuildPage(ctx, b, db, plan); err != nil {
			return err
		}

		// Step 3: Update storage and validate resources
		if err := update.UpdateStorage(b, db, t.bus, t.villageID); err != nil {
			return fmt.Errorf("update storage: %w", err)
		}

		required, err := feature.GetRequiredResource(b, plan.Type)
		if err != nil {
			return fmt.Errorf("get required resources: %w", err)
		}

		if err := feature.ValidateResources(db, t.villageID, required); err != nil {
			var taskErr *errs.TaskError
			if errors.As(err, &taskErr) {
				if errors.Is(taskErr.Err, errs.ErrLackOfFreeCrop) {
					// Add cropland job and retry
					if addErr := t.addCroplandJob(db); addErr == nil {
						continue
					}
				}
				if errors.Is(taskErr.Err, errs.ErrStorageLimit) {
					return &errs.TaskError{Err: errs.ErrStop, Message: taskErr.Message}
				}
				if errors.Is(taskErr.Err, errs.ErrMissingResource) {
					// Get time when enough resources and reschedule
					wait := time.Duration(0)
					html, _ := b.PageHTML()
					if doc, parseErr := parser.DocFromHTML(html); parseErr == nil {
						wait = parser.GetTimeWhenEnoughResource(doc, plan.Type)
					}
					if wait <= 0 {
						// Default retry: 5 minutes if we can't determine exact wait time
						wait = 5 * time.Minute
					}
					t.SetExecuteAt(time.Now().Add(wait))
					return errs.NewSkipError(taskErr.Message, t.ExecuteAt())
				}
			}
			return err
		}

		// Step 4: Perform the upgrade
		if err := feature.HandleUpgrade(ctx, b, db, plan, t.villageID); err != nil {
			return err
		}

		// Step 5: Update buildings after upgrade
		if err := update.UpdateBuildings(b, db, t.bus, t.villageID); err != nil {
			return fmt.Errorf("update buildings after upgrade: %w", err)
		}
	}
}

// getBuildPlan retrieves and validates the next build plan from the job queue.
func (t *UpgradeBuildingTask) getBuildPlan(ctx context.Context, b *browser.Browser, db *database.DB) (feature.NormalBuildPlan, error) {
	for {
		if ctx.Err() != nil {
			return feature.NormalBuildPlan{}, &errs.TaskError{Err: errs.ErrCancel, Message: "cancelled"}
		}

		// Get the next job from queue, considering queue capacity and Roman logic
		job, err := t.getJob(db)
		if err != nil {
			return feature.NormalBuildPlan{}, err
		}

		if enum.JobType(job.Type) == enum.JobTypeResourceBuild {
			// Convert ResourceBuild to NormalBuildPlan by finding lowest resource field
			plan, err := t.convertResourceJob(db, job)
			if err != nil {
				// Delete the completed resource job and try next
				db.Exec("DELETE FROM jobs WHERE id = ?", job.ID)
				t.bus.Emit(event.JobsModified, t.villageID)
				continue
			}
			// Replace resource job with a specific normal build job at the front
			content, _ := json.Marshal(map[string]interface{}{
				"type":     plan.Type,
				"level":    plan.Level,
				"location": plan.Location,
			})
			db.Exec(
				"INSERT INTO jobs (village_id, position, type, content) VALUES (?, 0, ?, ?)",
				t.villageID, int(enum.JobTypeNormalBuild), string(content),
			)
			t.bus.Emit(event.JobsModified, t.villageID)
			continue
		}

		// Parse NormalBuildPlan from job content
		var planData struct {
			Type     int `json:"type"`
			Level    int `json:"level"`
			Location int `json:"location"`
		}
		if err := json.Unmarshal([]byte(job.Content), &planData); err != nil {
			return feature.NormalBuildPlan{}, fmt.Errorf("parse job content: %w", err)
		}

		plan := feature.NormalBuildPlan{
			Type:     planData.Type,
			Level:    planData.Level,
			Location: planData.Location,
		}

		building := enum.Building(plan.Type)

		// Navigate to correct dorf and update buildings before validation
		if building.IsResourceBonus() {
			// Resource bonus buildings need both dorf1 and dorf2 updated
			if err := navigate.ToDorf(ctx, b, 1); err != nil {
				return feature.NormalBuildPlan{}, &errs.TaskError{Err: errs.ErrRetry, Message: fmt.Sprintf("nav to dorf1: %v", err)}
			}
			if err := update.UpdateBuildings(b, db, t.bus, t.villageID); err != nil {
				return feature.NormalBuildPlan{}, fmt.Errorf("update dorf1: %w", err)
			}
			if err := navigate.ToDorf(ctx, b, 2); err != nil {
				return feature.NormalBuildPlan{}, &errs.TaskError{Err: errs.ErrRetry, Message: fmt.Sprintf("nav to dorf2: %v", err)}
			}
			if err := update.UpdateBuildings(b, db, t.bus, t.villageID); err != nil {
				return feature.NormalBuildPlan{}, fmt.Errorf("update dorf2: %w", err)
			}
		} else {
			dorf := 2
			if plan.Location > 0 && plan.Location < 19 {
				dorf = 1
			}
			if err := navigate.ToDorf(ctx, b, dorf); err != nil {
				return feature.NormalBuildPlan{}, &errs.TaskError{Err: errs.ErrRetry, Message: fmt.Sprintf("nav to dorf%d: %v", dorf, err)}
			}
			if err := update.UpdateBuildings(b, db, t.bus, t.villageID); err != nil {
				return feature.NormalBuildPlan{}, fmt.Errorf("update dorf%d: %w", dorf, err)
			}
		}

		// Validate plan is still needed
		complete, err := t.validatePlanComplete(db, plan)
		if err != nil {
			return feature.NormalBuildPlan{}, err
		}
		if !complete {
			// Plan already done, delete job and continue
			db.Exec("DELETE FROM jobs WHERE id = ?", job.ID)
			t.bus.Emit(event.JobsModified, t.villageID)
			continue
		}

		return plan, nil
	}
}

// getJob retrieves the next build job considering queue capacity and Roman queue logic.
func (t *UpgradeBuildingTask) getJob(db *database.DB) (model.Job, error) {
	var jobs []model.Job
	if err := db.Select(&jobs,
		"SELECT id, village_id, position, type, content FROM jobs WHERE village_id = ? AND type IN (?, ?) ORDER BY position",
		t.villageID, int(enum.JobTypeNormalBuild), int(enum.JobTypeResourceBuild),
	); err != nil {
		return model.Job{}, fmt.Errorf("get build jobs: %w", err)
	}
	if len(jobs) == 0 {
		return model.Job{}, errs.NewSkipError("no build jobs in queue", time.Time{})
	}

	// Get queue buildings to determine capacity
	var queueBuildings []model.QueueBuilding
	db.Select(&queueBuildings,
		"SELECT id, village_id, position, location, type, level, complete_time FROM queue_buildings WHERE village_id = ? ORDER BY complete_time",
		t.villageID,
	)

	// Clean up completed queue buildings
	now := time.Now()
	var activeQueue []model.QueueBuilding
	for _, qb := range queueBuildings {
		if qb.CompleteTime.After(now) {
			activeQueue = append(activeQueue, qb)
		}
	}

	if len(activeQueue) == 0 {
		return jobs[0], nil
	}

	// Check plus account
	var hasPlusAccount int
	db.Get(&hasPlusAccount,
		"SELECT has_plus_account FROM accounts_info WHERE account_id = ?",
		t.accountID,
	)

	applyRoman, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingApplyRomanQueueLogicWhenBuilding)

	if len(activeQueue) == 1 {
		if hasPlusAccount != 0 {
			return jobs[0], nil
		}
		if applyRoman != 0 {
			return t.getJobRomanLogic(activeQueue, jobs)
		}
		// Queue full, reschedule to when first item completes
		return model.Job{}, errs.NewSkipError("construction queue full", activeQueue[0].CompleteTime.Time)
	}

	if len(activeQueue) == 2 {
		if hasPlusAccount != 0 && applyRoman != 0 {
			return t.getJobRomanLogic(activeQueue, jobs)
		}
		return model.Job{}, errs.NewSkipError("construction queue full", activeQueue[0].CompleteTime.Time)
	}

	// 3+ queue items
	return model.Job{}, errs.NewSkipError("construction queue full", activeQueue[0].CompleteTime.Time)
}

// getJobRomanLogic applies Roman queue logic: alternate between resource and infrastructure.
func (t *UpgradeBuildingTask) getJobRomanLogic(queue []model.QueueBuilding, jobs []model.Job) (model.Job, error) {
	resourceCount := 0
	for _, qb := range queue {
		if enum.Building(qb.Type).IsResourceField() {
			resourceCount++
		}
	}
	infraCount := len(queue) - resourceCount

	if resourceCount > infraCount {
		// Need infrastructure job
		job := t.findInfrastructureJob(jobs)
		if job != nil {
			return *job, nil
		}
	} else {
		// Need resource job
		job := t.findResourceJob(jobs)
		if job != nil {
			return *job, nil
		}
	}

	return model.Job{}, errs.NewSkipError("construction queue full (roman logic)", queue[0].CompleteTime.Time)
}

func (t *UpgradeBuildingTask) findInfrastructureJob(jobs []model.Job) *model.Job {
	for _, j := range jobs {
		if enum.JobType(j.Type) != enum.JobTypeNormalBuild {
			continue
		}
		var data struct {
			Type int `json:"type"`
		}
		if json.Unmarshal([]byte(j.Content), &data) == nil {
			if !enum.Building(data.Type).IsResourceField() {
				return &j
			}
		}
	}
	return nil
}

func (t *UpgradeBuildingTask) findResourceJob(jobs []model.Job) *model.Job {
	for _, j := range jobs {
		if enum.JobType(j.Type) == enum.JobTypeResourceBuild {
			return &j
		}
		if enum.JobType(j.Type) != enum.JobTypeNormalBuild {
			continue
		}
		var data struct {
			Type int `json:"type"`
		}
		if json.Unmarshal([]byte(j.Content), &data) == nil {
			if enum.Building(data.Type).IsResourceField() {
				return &j
			}
		}
	}
	return nil
}

// convertResourceJob converts a ResourceBuild job to a NormalBuildPlan.
func (t *UpgradeBuildingTask) convertResourceJob(db *database.DB, job model.Job) (feature.NormalBuildPlan, error) {
	var planData struct {
		Plan  int `json:"plan"`
		Level int `json:"level"`
	}
	if err := json.Unmarshal([]byte(job.Content), &planData); err != nil {
		return feature.NormalBuildPlan{}, fmt.Errorf("parse resource plan: %w", err)
	}

	// Get resource fields from database
	var buildings []model.Building
	if err := db.Select(&buildings,
		"SELECT id, village_id, type, level, is_under_construction, location FROM buildings WHERE village_id = ?",
		t.villageID,
	); err != nil {
		return feature.NormalBuildPlan{}, fmt.Errorf("get buildings: %w", err)
	}

	resourcePlan := enum.ResourcePlan(planData.Plan)
	var candidates []model.Building

	for _, b := range buildings {
		bt := enum.Building(b.Type)
		if !bt.IsResourceField() {
			continue
		}
		if b.Level >= planData.Level {
			continue
		}

		switch resourcePlan {
		case enum.ResourcePlanExcludeCrop:
			if bt != enum.BuildingCropland {
				candidates = append(candidates, b)
			}
		case enum.ResourcePlanOnlyCrop:
			if bt == enum.BuildingCropland {
				candidates = append(candidates, b)
			}
		default: // AllResources
			candidates = append(candidates, b)
		}
	}

	if len(candidates) == 0 {
		return feature.NormalBuildPlan{}, fmt.Errorf("no resource fields below level %d", planData.Level)
	}

	// Find the lowest level field
	minLevel := candidates[0].Level
	for _, c := range candidates[1:] {
		if c.Level < minLevel {
			minLevel = c.Level
		}
	}

	// From lowest level fields, pick one randomly
	var lowest []model.Building
	for _, c := range candidates {
		if c.Level == minLevel {
			lowest = append(lowest, c)
		}
	}
	chosen := lowest[rand.Intn(len(lowest))]

	return feature.NormalBuildPlan{
		Type:     chosen.Type,
		Level:    chosen.Level + 1,
		Location: chosen.Location,
	}, nil
}

// toBuildPage navigates to the building page and switches to the correct tab.
func (t *UpgradeBuildingTask) toBuildPage(ctx context.Context, b *browser.Browser, db *database.DB, plan feature.NormalBuildPlan) error {
	// Navigate to the building location
	if err := navigate.ToBuilding(ctx, b, plan.Location); err != nil {
		return fmt.Errorf("navigate to building %d: %w", plan.Location, err)
	}

	// Switch management tab if needed
	building := enum.Building(plan.Type)

	// Check if this is an empty site
	var existingType int
	err := db.Get(&existingType,
		"SELECT type FROM buildings WHERE village_id = ? AND location = ?",
		t.villageID, plan.Location,
	)
	if err == nil && existingType == int(enum.BuildingSite) {
		// Empty site: switch to the category tab for the building we want to construct
		category := building.GetBuildingsCategory()
		if err := navigate.SwitchTab(ctx, b, category); err != nil {
			return fmt.Errorf("switch to category tab %d: %w", category, err)
		}
	} else if err == nil && building.HasMultipleTabs() {
		// Existing building with multiple tabs: switch to first tab
		if err := navigate.SwitchTab(ctx, b, 0); err != nil {
			return fmt.Errorf("switch to first tab: %w", err)
		}
	}

	return nil
}

// validatePlanComplete checks if the plan still needs to be executed.
// Returns true if the plan should proceed, false if already complete.
func (t *UpgradeBuildingTask) validatePlanComplete(db *database.DB, plan feature.NormalBuildPlan) (bool, error) {
	var buildings []model.Building
	if err := db.Select(&buildings,
		"SELECT id, village_id, type, level, is_under_construction, location FROM buildings WHERE village_id = ?",
		t.villageID,
	); err != nil {
		return false, fmt.Errorf("get buildings: %w", err)
	}

	var queueBuildings []model.QueueBuilding
	db.Select(&queueBuildings,
		"SELECT id, village_id, position, location, type, level, complete_time FROM queue_buildings WHERE village_id = ? ORDER BY complete_time",
		t.villageID,
	)

	// Find existing building at the plan location
	for _, b := range buildings {
		if b.Location == plan.Location && b.Type == plan.Type {
			if b.Level >= plan.Level {
				return false, nil // Already at or above target level
			}
			// Check if queued at the target level or above
			for _, qb := range queueBuildings {
				if qb.Location == plan.Location && qb.Level >= plan.Level {
					return false, nil
				}
			}
			return true, nil
		}
	}

	// Building doesn't exist at location yet (new construction)
	// Check prerequisites
	prereqs := enum.Building(plan.Type).GetPrerequisiteBuildings()
	for _, prereq := range prereqs {
		met := false
		for _, b := range buildings {
			if b.Type == int(prereq.Building) && b.Level >= prereq.Level {
				met = true
				break
			}
		}
		if !met {
			// Check if prerequisite is in queue
			nextExecute := time.Time{}
			for _, qb := range queueBuildings {
				if qb.Type == int(prereq.Building) && qb.Level >= prereq.Level {
					nextExecute = qb.CompleteTime.Time
					break
				}
			}
			if !nextExecute.IsZero() {
				return false, errs.NewSkipError(
					fmt.Sprintf("prerequisite %s level %d in queue", enum.Building(prereq.Building), prereq.Level),
					nextExecute,
				)
			}
			return false, errs.NewSkipError(
				fmt.Sprintf("missing prerequisite: %s level %d", enum.Building(prereq.Building), prereq.Level),
				time.Time{},
			)
		}
	}

	return true, nil
}

// addCroplandJob inserts a job to upgrade the lowest-level cropland.
func (t *UpgradeBuildingTask) addCroplandJob(db *database.DB) error {
	var buildings []model.Building
	if err := db.Select(&buildings,
		"SELECT id, village_id, type, level, is_under_construction, location FROM buildings WHERE village_id = ? AND type = ?",
		t.villageID, int(enum.BuildingCropland),
	); err != nil || len(buildings) == 0 {
		return fmt.Errorf("no croplands found")
	}

	// Find lowest level cropland
	lowest := buildings[0]
	for _, b := range buildings[1:] {
		if b.Level < lowest.Level {
			lowest = b
		}
	}

	content, _ := json.Marshal(map[string]interface{}{
		"type":     int(enum.BuildingCropland),
		"level":    lowest.Level + 1,
		"location": lowest.Location,
	})

	_, err := db.Exec(
		"INSERT INTO jobs (village_id, position, type, content) VALUES (?, 0, ?, ?)",
		t.villageID, int(enum.JobTypeNormalBuild), string(content),
	)
	if err == nil {
		t.bus.Emit(event.JobsModified, t.villageID)
	}
	return err
}
