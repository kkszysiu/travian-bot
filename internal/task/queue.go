package task

import (
	"sort"
	"sync"
	"time"
)

// Queue is a sorted task queue.
type Queue struct {
	mu    sync.Mutex
	tasks []Task
}

// NewQueue creates an empty queue.
func NewQueue() *Queue {
	return &Queue{}
}

// Add inserts a task and re-sorts by ExecuteAt.
func (q *Queue) Add(t Task) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tasks = append(q.tasks, t)
	q.sort()
}

// Remove deletes a task from the queue.
func (q *Queue) Remove(t Task) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, task := range q.tasks {
		if task == t {
			q.tasks = append(q.tasks[:i], q.tasks[i+1:]...)
			return
		}
	}
}

// Peek returns the first task without removing it, or nil if empty.
func (q *Queue) Peek() Task {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.tasks) == 0 {
		return nil
	}
	return q.tasks[0]
}

// ReOrder re-sorts the queue by ExecuteAt.
func (q *Queue) ReOrder() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.sort()
}

// All returns a snapshot of all tasks.
func (q *Queue) All() []Task {
	q.mu.Lock()
	defer q.mu.Unlock()
	result := make([]Task, len(q.tasks))
	copy(result, q.tasks)
	return result
}

// Len returns the number of tasks.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.tasks)
}

// Clear removes all tasks.
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tasks = nil
}

// HasTaskOfType checks if a task with the given description exists.
func (q *Queue) HasTaskOfType(description string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, t := range q.tasks {
		if t.Description() == description {
			return true
		}
	}
	return false
}

// HasVillageTask checks if a task with the given description and village ID exists.
func (q *Queue) HasVillageTask(description string, villageID int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, t := range q.tasks {
		if t.Description() == description {
			if vt, ok := t.(interface{ VillageID() int }); ok {
				if vt.VillageID() == villageID {
					return true
				}
			}
		}
	}
	return false
}

// UpdateVillageTaskExecuteAt finds a task matching description+villageID and updates its ExecuteAt.
// Matches C# TaskManager.AddOrUpdate behavior where existing tasks get rescheduled to now.
func (q *Queue) UpdateVillageTaskExecuteAt(description string, villageID int, executeAt time.Time) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, t := range q.tasks {
		if t.Description() == description {
			if vt, ok := t.(interface{ VillageID() int }); ok {
				if vt.VillageID() == villageID {
					t.SetExecuteAt(executeAt)
					q.sort()
					return true
				}
			}
		}
	}
	return false
}

// RemoveByType removes the first task matching the description.
func (q *Queue) RemoveByType(description string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, t := range q.tasks {
		if t.Description() == description {
			q.tasks = append(q.tasks[:i], q.tasks[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveVillageTask removes the first task matching description and village ID.
func (q *Queue) RemoveVillageTask(description string, villageID int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, t := range q.tasks {
		if t.Description() == description {
			if vt, ok := t.(interface{ VillageID() int }); ok {
				if vt.VillageID() == villageID {
					q.tasks = append(q.tasks[:i], q.tasks[i+1:]...)
					return true
				}
			}
		}
	}
	return false
}

// UpdateTaskExecuteAt updates the execution time of the first task matching description.
func (q *Queue) UpdateTaskExecuteAt(description string, executeAt time.Time) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, t := range q.tasks {
		if t.Description() == description {
			t.SetExecuteAt(executeAt)
			q.sort()
			return true
		}
	}
	return false
}

func (q *Queue) sort() {
	sort.Slice(q.tasks, func(i, j int) bool {
		return q.tasks[i].ExecuteAt().Before(q.tasks[j].ExecuteAt())
	})
}
