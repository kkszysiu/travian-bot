package task

import (
	"context"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/database"
)

// Stage represents a task's execution state.
type Stage int

const (
	StageWaiting   Stage = 0
	StageExecuting Stage = 1
)

func (s Stage) String() string {
	if s == StageExecuting {
		return "Executing"
	}
	return "Waiting"
}

// Task is the interface all tasks implement.
type Task interface {
	Description() string
	AccountID() int
	ExecuteAt() time.Time
	SetExecuteAt(t time.Time)
	Stage() Stage
	SetStage(s Stage)
	Execute(ctx context.Context, b *browser.Browser, db *database.DB) error
}

// BaseTask provides common fields for all tasks.
type BaseTask struct {
	accountID int
	executeAt time.Time
	stage     Stage
}

func (t *BaseTask) AccountID() int          { return t.accountID }
func (t *BaseTask) ExecuteAt() time.Time    { return t.executeAt }
func (t *BaseTask) SetExecuteAt(tm time.Time) { t.executeAt = tm }
func (t *BaseTask) Stage() Stage            { return t.stage }
func (t *BaseTask) SetStage(s Stage)        { t.stage = s }
