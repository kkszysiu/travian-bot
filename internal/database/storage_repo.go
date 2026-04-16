package database

import "travian-bot/internal/domain/model"

type StorageDTO struct {
	Wood      int `json:"wood"`
	Clay      int `json:"clay"`
	Iron      int `json:"iron"`
	Crop      int `json:"crop"`
	Warehouse int `json:"warehouse"`
	Granary   int `json:"granary"`
	FreeCrop  int `json:"freeCrop"`
}

func (db *DB) GetStorage(villageID int) (*StorageDTO, error) {
	var s model.Storage
	err := db.Get(&s,
		"SELECT id, village_id, wood, clay, iron, crop, warehouse, granary, free_crop FROM storages WHERE village_id = ?",
		villageID,
	)
	if err != nil {
		return &StorageDTO{}, nil // Return empty if not found
	}
	return &StorageDTO{
		Wood: s.Wood, Clay: s.Clay, Iron: s.Iron, Crop: s.Crop,
		Warehouse: s.Warehouse, Granary: s.Granary, FreeCrop: s.FreeCrop,
	}, nil
}
