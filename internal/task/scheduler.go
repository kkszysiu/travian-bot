package task

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/command/navigate"
	"travian-bot/internal/command/update"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/domain/errs"
	"travian-bot/internal/event"
	"travian-bot/internal/service"
)

const maxRetries = 3
const taskTimeout = 3 * time.Minute
const longTaskTimeout = 5 * time.Minute // For Sleep and Login tasks

// Scheduler manages per-account task execution goroutines.
type Scheduler struct {
	mu         sync.RWMutex
	queues     map[int]*Queue          // accountID -> task queue
	statuses   map[int]enum.Status     // accountID -> status
	cancels    map[int]context.CancelFunc
	db         *database.DB
	browsers   *browser.Manager
	bus        *event.Bus
	logger     *slog.Logger
	wg         sync.WaitGroup
}

// NewScheduler creates a new scheduler.
func NewScheduler(db *database.DB, browsers *browser.Manager, bus *event.Bus, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		queues:   make(map[int]*Queue),
		statuses: make(map[int]enum.Status),
		cancels:  make(map[int]context.CancelFunc),
		db:       db,
		browsers: browsers,
		bus:      bus,
		logger:   logger,
	}
}

// Start begins the scheduler loop for an account.
func (s *Scheduler) Start(accountID int) {
	s.mu.Lock()
	if _, ok := s.queues[accountID]; !ok {
		s.queues[accountID] = NewQueue()
	}
	s.statuses[accountID] = enum.StatusOnline
	s.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	s.mu.Lock()
	if old, ok := s.cancels[accountID]; ok {
		old() // cancel previous
	}
	s.cancels[accountID] = cancel
	s.mu.Unlock()

	s.emitStatus(accountID)
	s.emitTasks(accountID)

	s.wg.Add(1)
	go s.loop(ctx, accountID)
}

// Stop cancels the scheduler loop for an account.
func (s *Scheduler) Stop(accountID int) {
	s.mu.Lock()
	if cancel, ok := s.cancels[accountID]; ok {
		cancel()
		delete(s.cancels, accountID)
	}
	s.statuses[accountID] = enum.StatusOffline
	s.mu.Unlock()
	s.emitStatus(accountID)
}

// Pause sets the account to paused state.
func (s *Scheduler) Pause(accountID int) {
	s.mu.Lock()
	s.statuses[accountID] = enum.StatusPaused
	s.mu.Unlock()
	s.emitStatus(accountID)
}

// Resume sets the account back to online.
func (s *Scheduler) Resume(accountID int) {
	s.mu.Lock()
	s.statuses[accountID] = enum.StatusOnline
	s.mu.Unlock()
	s.emitStatus(accountID)
}

// ClearTasks removes all tasks from an account's queue.
func (s *Scheduler) ClearTasks(accountID int) {
	s.mu.RLock()
	q, ok := s.queues[accountID]
	s.mu.RUnlock()
	if ok {
		q.Clear()
		s.emitTasks(accountID)
	}
}

// AddTask adds a task to an account's queue.
func (s *Scheduler) AddTask(accountID int, t Task) {
	s.mu.RLock()
	q, ok := s.queues[accountID]
	s.mu.RUnlock()
	if !ok {
		s.mu.Lock()
		q = NewQueue()
		s.queues[accountID] = q
		s.mu.Unlock()
	}
	q.Add(t)
	s.emitTasks(accountID)
}

// GetTasks returns current tasks for an account.
func (s *Scheduler) GetTasks(accountID int) []Task {
	s.mu.RLock()
	q, ok := s.queues[accountID]
	s.mu.RUnlock()
	if !ok {
		return nil
	}
	return q.All()
}

// GetStatus returns an account's status.
func (s *Scheduler) GetStatus(accountID int) enum.Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.statuses[accountID]
}

// HasTaskOfType checks if a task with the given description exists for an account.
func (s *Scheduler) HasTaskOfType(accountID int, description string) bool {
	s.mu.RLock()
	q, ok := s.queues[accountID]
	s.mu.RUnlock()
	if !ok {
		return false
	}
	return q.HasTaskOfType(description)
}

// HasVillageTask checks if a task with the given description exists for a specific village.
func (s *Scheduler) HasVillageTask(accountID int, description string, villageID int) bool {
	s.mu.RLock()
	q, ok := s.queues[accountID]
	s.mu.RUnlock()
	if !ok {
		return false
	}
	return q.HasVillageTask(description, villageID)
}

// AddOrUpdateVillageTask adds a new task or updates the ExecuteAt of an existing one.
// Matches C# TaskManager.AddOrUpdate behavior.
func (s *Scheduler) AddOrUpdateVillageTask(accountID int, t Task, villageID int) {
	s.mu.RLock()
	q, ok := s.queues[accountID]
	s.mu.RUnlock()
	if !ok {
		s.mu.Lock()
		q = NewQueue()
		s.queues[accountID] = q
		s.mu.Unlock()
	}
	if !q.UpdateVillageTaskExecuteAt(t.Description(), villageID, t.ExecuteAt()) {
		q.Add(t)
	}
	s.emitTasks(accountID)
}

// RemoveTaskOfType removes the first task matching the description from an account's queue.
func (s *Scheduler) RemoveTaskOfType(accountID int, description string) {
	s.mu.RLock()
	q, ok := s.queues[accountID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	if q.RemoveByType(description) {
		s.emitTasks(accountID)
	}
}

// RemoveVillageTask removes a village-specific task from an account's queue.
func (s *Scheduler) RemoveVillageTask(accountID int, description string, villageID int) {
	s.mu.RLock()
	q, ok := s.queues[accountID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	if q.RemoveVillageTask(description, villageID) {
		s.emitTasks(accountID)
	}
}

// RescheduleTask updates the execution time of a task by description.
func (s *Scheduler) RescheduleTask(accountID int, description string, executeAt time.Time) {
	s.mu.RLock()
	q, ok := s.queues[accountID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	if q.UpdateTaskExecuteAt(description, executeAt) {
		s.emitTasks(accountID)
	}
}

// Shutdown stops all scheduler loops.
func (s *Scheduler) Shutdown() {
	s.mu.Lock()
	for id, cancel := range s.cancels {
		cancel()
		delete(s.cancels, id)
	}
	s.mu.Unlock()
	s.wg.Wait()
}

func (s *Scheduler) loop(ctx context.Context, accountID int) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		status := s.GetStatus(accountID)
		if status != enum.StatusOnline {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		s.mu.RLock()
		q := s.queues[accountID]
		s.mu.RUnlock()
		if q == nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		t := q.Peek()
		if t == nil || t.ExecuteAt().After(time.Now()) {
			// Sleep until next task or 1 second
			sleepDur := 1 * time.Second
			if t != nil {
				until := time.Until(t.ExecuteAt())
				if until < sleepDur {
					sleepDur = until
				}
			}
			timer := time.NewTimer(sleepDur)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			continue
		}

		// Work window enforcement: postpone non-essential tasks outside the window
		ww := service.GetWorkWindow(s.db, accountID)
		if ww.IsOutsideWindow(time.Now()) {
			desc := t.Description()
			if desc != "Sleep" && desc != "Login" {
				nextStart := ww.NextWorkStart(time.Now())
				t.SetExecuteAt(nextStart)
				q.ReOrder()
				s.logger.Info("task postponed (outside work window)",
					"accountId", accountID, "task", desc,
					"nextStart", nextStart.Format("2006-01-02 15:04:05"))
				s.emitTasks(accountID)
				continue
			}
		}

		// Execute task with retry
		s.executeTask(ctx, accountID, t, q)

		// Inter-task delay
		delayMin, _ := service.GetAccountSettingValue(s.db, accountID, enum.AccountSettingTaskDelayMin)
		delayMax, _ := service.GetAccountSettingValue(s.db, accountID, enum.AccountSettingTaskDelayMax)
		delay := service.RandomDelay(delayMin, delayMax)

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (s *Scheduler) executeTask(ctx context.Context, accountID int, t Task, q *Queue) {
	t.SetStage(StageExecuting)
	s.emitTasks(accountID)

	cacheExecuteAt := t.ExecuteAt()

	s.logger.Info("task starting", "accountId", accountID, "task", t.Description())
	s.emitLog(accountID, "info", "Starting: "+t.Description())

	b := s.browsers.Get(accountID)
	if b == nil {
		s.logger.Warn("no browser for account, attempting relogin", "accountId", accountID)
		s.emitLog(accountID, "warn", "No browser — attempting relogin...")
		if err := s.relogin(accountID, q); err != nil {
			s.logger.Error("relogin failed", "accountId", accountID, "error", err.Error())
			s.emitLog(accountID, "error", "Relogin failed: "+err.Error())
			s.Pause(accountID)
		}
		return
	}

	// VillageTaskBehavior: switch to the correct village before executing village tasks
	// This matches the C# VillageTaskBehavior middleware pattern.
	if vt, ok := t.(villageTask); ok {
		if err := s.preVillageTask(ctx, accountID, vt.VillageID(), b); err != nil {
			s.logger.Warn("pre-village task failed", "accountId", accountID, "task", t.Description(), "error", err.Error())
			// Non-fatal: continue with task execution anyway
		}
	}

	var lastErr error
	timedOut := false
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if ctx.Err() != nil {
			s.logger.Info("task cancelled", "accountId", accountID, "task", t.Description())
			s.emitLog(accountID, "warn", "Cancelled: "+t.Description())
			s.Pause(accountID)
			return
		}

		// Execute with timeout to prevent tasks from blocking indefinitely
		// Sleep and Login need much longer — they close/reopen browsers and wait
		timeout := taskTimeout
		desc := t.Description()
		if desc == "Sleep" || desc == "Login" {
			timeout = longTaskTimeout
		}
		taskCtx, taskCancel := context.WithTimeout(ctx, timeout)
		err := t.Execute(taskCtx, b, s.db)
		taskCancel()

		if err != nil && taskCtx.Err() == context.DeadlineExceeded && ctx.Err() == nil {
			// Task timed out but the scheduler itself wasn't cancelled
			s.logger.Warn("task timed out", "accountId", accountID, "task", desc)
			s.emitLog(accountID, "warn", fmt.Sprintf("Timed out: %s (exceeded %s)", desc, timeout))
			timedOut = true
			lastErr = nil
			break
		}
		if err == nil {
			lastErr = nil
			break
		}

		// Check error type
		var taskErr *errs.TaskError
		if errors.As(err, &taskErr) {
			if errors.Is(taskErr.Err, errs.ErrSkip) {
				// Skip: reschedule or remove
				s.emitLog(accountID, "info", fmt.Sprintf("Skipped: %s - %s", t.Description(), taskErr.Message))
				if !taskErr.NextExecute.IsZero() {
					t.SetExecuteAt(taskErr.NextExecute)
				}
				lastErr = nil
				break
			}
			if errors.Is(taskErr.Err, errs.ErrStop) {
				s.logger.Warn("task stopped", "accountId", accountID, "task", t.Description(), "msg", taskErr.Message)
				s.emitLog(accountID, "warn", "Stopped: "+t.Description()+" - "+taskErr.Message)
				b.Screenshot()
				s.Pause(accountID)
				t.SetStage(StageWaiting)
				s.emitTasks(accountID)
				return
			}
			if errors.Is(taskErr.Err, errs.ErrCancel) {
				s.emitLog(accountID, "warn", "Cancelled: "+t.Description())
				s.Pause(accountID)
				t.SetStage(StageWaiting)
				s.emitTasks(accountID)
				return
			}
		}

		// Detect dead browser connection — retrying won't help, relogin instead
		if isConnectionError(err) {
			s.logger.Warn("browser connection lost, attempting relogin", "accountId", accountID, "task", t.Description())
			s.emitLog(accountID, "warn", "Browser connection lost — attempting relogin...")
			t.SetStage(StageWaiting)
			if reloginErr := s.relogin(accountID, q); reloginErr != nil {
				s.logger.Error("relogin failed", "accountId", accountID, "error", reloginErr.Error())
				s.emitLog(accountID, "error", "Relogin failed: "+reloginErr.Error())
				s.Pause(accountID)
			}
			s.emitTasks(accountID)
			return
		}

		// Retry with exponential backoff
		lastErr = err
		if attempt < maxRetries {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * 30 * time.Second
			jitter := time.Duration(rand.Intn(5000)) * time.Millisecond
			sleepTime := backoff + jitter
			s.logger.Warn("task retry",
				"accountId", accountID,
				"task", t.Description(),
				"attempt", attempt+1,
				"backoff", sleepTime,
				"error", err.Error(),
			)
			s.emitLog(accountID, "warn", fmt.Sprintf("Retry %d/%d: %s - %s", attempt+1, maxRetries, t.Description(), err.Error()))
			timer := time.NewTimer(sleepTime)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}

	t.SetStage(StageWaiting)

	if lastErr != nil {
		// All retries exhausted
		s.logger.Error("task failed after retries",
			"accountId", accountID,
			"task", t.Description(),
			"error", lastErr.Error(),
		)
		s.emitLog(accountID, "error", fmt.Sprintf("Failed: %s - %s", t.Description(), lastErr.Error()))
		b.Screenshot()
		s.Pause(accountID)
		s.emitTasks(accountID)
		return
	}

	// Success or skip: remove or reschedule
	if t.ExecuteAt() == cacheExecuteAt {
		if timedOut {
			// Task timed out without rescheduling itself — reschedule to retry in 5 minutes
			// instead of removing it permanently from the queue
			t.SetExecuteAt(time.Now().Add(5 * time.Minute))
			q.ReOrder()
			s.logger.Info("task rescheduled after timeout",
				"accountId", accountID,
				"task", t.Description(),
				"nextRun", t.ExecuteAt().Format("2006-01-02 15:04:05"),
			)
			s.emitLog(accountID, "info", fmt.Sprintf("Rescheduled after timeout: %s (next: %s)", t.Description(), t.ExecuteAt().Format("15:04:05")))
		} else {
			q.Remove(t)
			s.logger.Info("task completed", "accountId", accountID, "task", t.Description())
			s.emitLog(accountID, "info", "Completed: "+t.Description())
		}
	} else {
		q.ReOrder()
		s.logger.Info("task rescheduled",
			"accountId", accountID,
			"task", t.Description(),
			"nextRun", t.ExecuteAt().Format("2006-01-02 15:04:05"),
		)
		s.emitLog(accountID, "info", fmt.Sprintf("Rescheduled: %s (next: %s)", t.Description(), t.ExecuteAt().Format("15:04:05")))
	}

	// VillageTaskBehavior post-execution: update buildings and storage after task
	if vt, ok := t.(villageTask); ok {
		if err := s.postVillageTask(ctx, accountID, vt.VillageID(), b); err != nil {
			s.logger.Warn("post-village task failed", "accountId", accountID, "task", t.Description(), "error", err.Error())
		}
	}

	// Reactive triggers: auto-queue tasks based on page state after execution
	s.checkReactiveTriggers(accountID, t, b)

	s.emitTasks(accountID)
}

// villageTask is an optional interface for tasks that operate on a specific village.
type villageTask interface {
	VillageID() int
}

// preVillageTask implements the C# VillageTaskBehavior pre-execution:
// 1. Switch to the correct village
// 2. Update storage
// 3. Navigate to dorf1 and update buildings
func (s *Scheduler) preVillageTask(ctx context.Context, accountID, villageID int, b *browser.Browser) error {
	// Step 1: Switch to the correct village in-game
	if err := navigate.SwitchVillage(ctx, b, villageID); err != nil {
		return fmt.Errorf("switch village: %w", err)
	}

	// Step 2: Update storage (resource bar is always visible)
	if err := update.UpdateStorage(b, s.db, s.bus, villageID); err != nil {
		// Non-fatal
		s.logger.Warn("pre-task storage update failed", "villageId", villageID, "error", err.Error())
	}

	return nil
}

// postVillageTask implements the C# VillageTaskBehavior post-execution:
// 1. Navigate to dorf1
// 2. Update buildings
// 3. Update storage
func (s *Scheduler) postVillageTask(ctx context.Context, accountID, villageID int, b *browser.Browser) error {
	// Navigate to dorf1 to ensure we're on a known page
	if err := navigate.ToDorf(ctx, b, 1); err != nil {
		return fmt.Errorf("navigate to dorf1: %w", err)
	}

	// Update buildings from current dorf page
	if err := update.UpdateBuildings(b, s.db, s.bus, villageID); err != nil {
		s.logger.Warn("post-task building update failed", "villageId", villageID, "error", err.Error())
	}

	// Update storage
	if err := update.UpdateStorage(b, s.db, s.bus, villageID); err != nil {
		s.logger.Warn("post-task storage update failed", "villageId", villageID, "error", err.Error())
	}

	return nil
}

// checkReactiveTriggers inspects the current page state after task execution
// and auto-queues adventure/quest tasks if conditions are met.
func (s *Scheduler) checkReactiveTriggers(accountID int, t Task, b *browser.Browser) {
	// After any task: check if adventures are available
	if update.CheckAdventureAvailable(b) {
		if !s.HasTaskOfType(accountID, "Start adventure") {
			adventureEnabled, _ := service.GetAccountSettingValue(s.db, accountID, enum.AccountSettingEnableAutoStartAdventure)
			if adventureEnabled != 0 {
				s.AddTask(accountID, NewStartAdventureTask(accountID, s.bus))
				s.logger.Info("auto-queued adventure task", "accountId", accountID)
			}
		}
	}

	// After village tasks: check if quests are claimable
	if vt, ok := t.(villageTask); ok {
		if update.CheckQuestAvailable(b) {
			if !s.HasTaskOfType(accountID, "Claim quest") {
				questEnabled, _ := service.GetVillageSettingValue(s.db, vt.VillageID(), enum.VillageSettingAutoClaimQuestEnable)
				if questEnabled != 0 {
					s.AddTask(accountID, NewClaimQuestTask(accountID, vt.VillageID(), s.bus))
					s.logger.Info("auto-queued quest task", "accountId", accountID, "villageId", vt.VillageID())
				}
			}
		}
	}
}

// relogin closes the old browser, creates a new one, navigates to the server,
// clears the task queue, and queues a fresh LoginTask.
func (s *Scheduler) relogin(accountID int, q *Queue) error {
	// Close existing browser and wait for process to exit
	s.browsers.Close(accountID)
	time.Sleep(2 * time.Second)

	// Get account details for browser config
	detail, err := s.db.GetAccountDetail(accountID)
	if err != nil {
		return fmt.Errorf("get account detail: %w", err)
	}

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

	// Create new browser
	newBrowser, err := s.browsers.Create(accountID, cfg)
	if err != nil {
		return fmt.Errorf("create browser: %w", err)
	}

	// Navigate to server
	if err := newBrowser.Navigate(detail.Server); err != nil {
		s.browsers.Close(accountID)
		return fmt.Errorf("navigate to server: %w", err)
	}

	// Clear the queue and add a fresh login task
	q.Clear()
	q.Add(NewLoginTask(accountID, s.bus, s, s.browsers))
	s.emitLog(accountID, "info", "Relogin queued — will reconnect shortly")
	s.emitTasks(accountID)

	return nil
}

func (s *Scheduler) emitStatus(accountID int) {
	status := s.GetStatus(accountID)
	s.bus.Emit(event.StatusModified, event.StatusPayload{
		AccountID: accountID,
		Status:    int(status),
		Color:     status.Color(),
	})
}

func (s *Scheduler) emitTasks(accountID int) {
	s.bus.Emit(event.TasksModified, accountID)
}

func (s *Scheduler) emitLog(accountID int, level, message string) {
	s.bus.Emit(event.LogEmitted, event.LogPayload{
		AccountID: accountID,
		Message:   message,
		Level:     level,
	})
}

// isConnectionError checks if an error indicates the browser connection is dead.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "use of closed network connection") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset by peer") ||
		strings.Contains(msg, "websocket: close")
}
