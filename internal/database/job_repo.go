package database

import (
	"encoding/json"
	"fmt"

	"travian-bot/internal/domain/enum"
	"travian-bot/internal/domain/model"
)

type JobItem struct {
	ID       int    `json:"id"`
	Position int    `json:"position"`
	Type     int    `json:"type"`
	Content  string `json:"content"`
	Display  string `json:"display"`
}

// NormalBuildInput is the DTO for adding a normal building job.
type NormalBuildInput struct {
	VillageID int `json:"villageId"`
	Type      int `json:"type"`
	Level     int `json:"level"`
	Location  int `json:"location"`
}

// ResourceBuildInput is the DTO for adding a resource building job.
type ResourceBuildInput struct {
	VillageID int `json:"villageId"`
	Plan      int `json:"plan"`
	Level     int `json:"level"`
}

func (db *DB) GetJobs(villageID int) ([]JobItem, error) {
	var jobs []model.Job
	if err := db.Select(&jobs,
		"SELECT id, village_id, position, type, content FROM jobs WHERE village_id = ? ORDER BY position",
		villageID,
	); err != nil {
		return nil, err
	}

	items := make([]JobItem, len(jobs))
	for i, j := range jobs {
		items[i] = JobItem{
			ID:       j.ID,
			Position: j.Position,
			Type:     j.Type,
			Content:  j.Content,
			Display:  formatJobDisplay(j),
		}
	}
	return items, nil
}

func formatJobDisplay(j model.Job) string {
	switch enum.JobType(j.Type) {
	case enum.JobTypeNormalBuild:
		var data struct {
			Type     int `json:"type"`
			Level    int `json:"level"`
			Location int `json:"location"`
		}
		if err := json.Unmarshal([]byte(j.Content), &data); err == nil {
			if data.Location > 0 {
				return fmt.Sprintf("[%d] %s level %d", data.Location, enum.Building(data.Type).String(), data.Level)
			}
			return fmt.Sprintf("%s level %d", enum.Building(data.Type).String(), data.Level)
		}
	case enum.JobTypeResourceBuild:
		var data struct {
			Plan  int `json:"plan"`
			Level int `json:"level"`
		}
		if err := json.Unmarshal([]byte(j.Content), &data); err == nil {
			return fmt.Sprintf("%s to level %d", enum.ResourcePlan(data.Plan).String(), data.Level)
		}
	}
	return j.Content
}

func (db *DB) AddNormalBuildJob(input NormalBuildInput) error {
	location := input.Location

	// Auto-resolve location if not provided (matches C# NormalBuildCommand behavior)
	if location == 0 && input.Type > 0 {
		location = db.resolveNormalBuildLocation(input.VillageID, input.Type)
	}

	content, err := json.Marshal(map[string]int{"type": input.Type, "level": input.Level, "location": location})
	if err != nil {
		return err
	}

	var maxPos int
	db.Get(&maxPos, "SELECT COALESCE(MAX(position), 0) FROM jobs WHERE village_id = ?", input.VillageID)

	_, err = db.Exec(
		"INSERT INTO jobs (village_id, position, type, content) VALUES (?, ?, ?, ?)",
		input.VillageID, maxPos+1, int(enum.JobTypeNormalBuild), string(content),
	)
	return err
}

// resolveNormalBuildLocation finds the location for a building type in a village.
// Walls are always at location 40. For other types, picks the existing building
// with the highest level (matching C# NormalBuildCommand auto-adjustment).
func (db *DB) resolveNormalBuildLocation(villageID, buildingType int) int {
	bt := enum.Building(buildingType)
	if bt.IsWall() {
		return 40
	}

	var buildings []struct {
		Location int `db:"location"`
		Level    int `db:"level"`
	}
	if err := db.Select(&buildings,
		"SELECT location, level FROM buildings WHERE village_id = ? AND type = ? ORDER BY level DESC LIMIT 1",
		villageID, buildingType,
	); err != nil || len(buildings) == 0 {
		return 0
	}
	return buildings[0].Location
}

func (db *DB) AddResourceBuildJob(input ResourceBuildInput) error {
	content, err := json.Marshal(map[string]int{"plan": input.Plan, "level": input.Level})
	if err != nil {
		return err
	}

	var maxPos int
	db.Get(&maxPos, "SELECT COALESCE(MAX(position), 0) FROM jobs WHERE village_id = ?", input.VillageID)

	_, err = db.Exec(
		"INSERT INTO jobs (village_id, position, type, content) VALUES (?, ?, ?, ?)",
		input.VillageID, maxPos+1, int(enum.JobTypeResourceBuild), string(content),
	)
	return err
}

func (db *DB) DeleteJob(jobID int) error {
	_, err := db.Exec("DELETE FROM jobs WHERE id = ?", jobID)
	return err
}

func (db *DB) DeleteAllJobs(villageID int) error {
	_, err := db.Exec("DELETE FROM jobs WHERE village_id = ?", villageID)
	return err
}

func (db *DB) MoveJob(jobID int, direction string) error {
	var job model.Job
	if err := db.Get(&job, "SELECT id, village_id, position, type, content FROM jobs WHERE id = ?", jobID); err != nil {
		return err
	}

	var swapPos int
	switch direction {
	case "up":
		err := db.Get(&swapPos,
			"SELECT position FROM jobs WHERE village_id = ? AND position < ? ORDER BY position DESC LIMIT 1",
			job.VillageID, job.Position,
		)
		if err != nil {
			return nil // Already at top
		}
	case "down":
		err := db.Get(&swapPos,
			"SELECT position FROM jobs WHERE village_id = ? AND position > ? ORDER BY position ASC LIMIT 1",
			job.VillageID, job.Position,
		)
		if err != nil {
			return nil // Already at bottom
		}
	case "top":
		var minPos int
		if err := db.Get(&minPos, "SELECT COALESCE(MIN(position), 0) FROM jobs WHERE village_id = ?", job.VillageID); err != nil {
			return err
		}
		if minPos >= job.Position {
			return nil
		}
		_, err := db.Exec(
			"UPDATE jobs SET position = position + 1 WHERE village_id = ? AND position < ?",
			job.VillageID, job.Position,
		)
		if err != nil {
			return err
		}
		_, err = db.Exec("UPDATE jobs SET position = ? WHERE id = ?", minPos, jobID)
		return err
	case "bottom":
		var maxPos int
		if err := db.Get(&maxPos, "SELECT COALESCE(MAX(position), 0) FROM jobs WHERE village_id = ?", job.VillageID); err != nil {
			return err
		}
		if maxPos <= job.Position {
			return nil
		}
		_, err := db.Exec(
			"UPDATE jobs SET position = position - 1 WHERE village_id = ? AND position > ?",
			job.VillageID, job.Position,
		)
		if err != nil {
			return err
		}
		_, err = db.Exec("UPDATE jobs SET position = ? WHERE id = ?", maxPos, jobID)
		return err
	default:
		return fmt.Errorf("invalid direction: %s", direction)
	}

	// Swap positions for up/down
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE jobs SET position = ? WHERE village_id = ? AND position = ?",
		-1, job.VillageID, swapPos)
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE jobs SET position = ? WHERE id = ?", swapPos, jobID)
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE jobs SET position = ? WHERE village_id = ? AND position = ?",
		job.Position, job.VillageID, -1)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DB) ImportJobs(villageID int, jsonData string) error {
	var jobs []struct {
		Type    int    `json:"type"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(jsonData), &jobs); err != nil {
		return fmt.Errorf("invalid job data: %w", err)
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var maxPos int
	db.Get(&maxPos, "SELECT COALESCE(MAX(position), 0) FROM jobs WHERE village_id = ?", villageID)

	for i, j := range jobs {
		_, err := tx.Exec(
			"INSERT INTO jobs (village_id, position, type, content) VALUES (?, ?, ?, ?)",
			villageID, maxPos+i+1, j.Type, j.Content,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) ExportJobs(villageID int) (string, error) {
	var jobs []model.Job
	if err := db.Select(&jobs,
		"SELECT id, village_id, position, type, content FROM jobs WHERE village_id = ? ORDER BY position",
		villageID,
	); err != nil {
		return "", err
	}

	type exportJob struct {
		Type    int    `json:"type"`
		Content string `json:"content"`
	}
	exported := make([]exportJob, len(jobs))
	for i, j := range jobs {
		exported[i] = exportJob{Type: j.Type, Content: j.Content}
	}

	data, err := json.MarshalIndent(exported, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
