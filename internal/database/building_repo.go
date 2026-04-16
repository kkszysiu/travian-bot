package database

import (
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/domain/model"
)

type BuildingItem struct {
	ID                  int    `json:"id"`
	Type                int    `json:"type"`
	TypeName            string `json:"typeName"`
	Level               int    `json:"level"`
	MaxLevel            int    `json:"maxLevel"`
	IsUnderConstruction bool   `json:"isUnderConstruction"`
	Location            int    `json:"location"`
	Color               string `json:"color"`
}

type QueueBuildingItem struct {
	Position     int    `json:"position"`
	Location     int    `json:"location"`
	TypeName     string `json:"typeName"`
	Level        int    `json:"level"`
	CompleteTime string `json:"completeTime"`
}

func (db *DB) GetBuildings(villageID int) ([]BuildingItem, error) {
	var buildings []model.Building
	if err := db.Select(&buildings,
		"SELECT id, village_id, type, level, is_under_construction, location FROM buildings WHERE village_id = ? ORDER BY location",
		villageID,
	); err != nil {
		return nil, err
	}

	items := make([]BuildingItem, len(buildings))
	for i, b := range buildings {
		bt := enum.Building(b.Type)
		items[i] = BuildingItem{
			ID:                  b.ID,
			Type:                b.Type,
			TypeName:            bt.String(),
			Level:               b.Level,
			MaxLevel:            bt.GetMaxLevel(),
			IsUnderConstruction: b.IsUnderConstruction,
			Location:            b.Location,
			Color:               bt.GetColor(),
		}
	}
	return items, nil
}

func (db *DB) GetQueueBuildings(villageID int) ([]QueueBuildingItem, error) {
	var queue []model.QueueBuilding
	if err := db.Select(&queue,
		"SELECT id, village_id, position, location, type, level, complete_time FROM queue_buildings WHERE village_id = ? ORDER BY position",
		villageID,
	); err != nil {
		return nil, err
	}

	items := make([]QueueBuildingItem, len(queue))
	for i, q := range queue {
		items[i] = QueueBuildingItem{
			Position:     q.Position,
			Location:     q.Location,
			TypeName:     enum.Building(q.Type).String(),
			Level:        q.Level,
			CompleteTime: q.CompleteTime.Format("15:04:05"),
		}
	}
	return items, nil
}
