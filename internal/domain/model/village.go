package model

type Village struct {
	ID                     int    `db:"id" json:"id"`
	AccountID              int    `db:"account_id" json:"accountId"`
	Name                   string `db:"name" json:"name"`
	X                      int    `db:"x" json:"x"`
	Y                      int    `db:"y" json:"y"`
	IsActive               bool   `db:"is_active" json:"isActive"`
	IsUnderAttack          bool   `db:"is_under_attack" json:"isUnderAttack"`
	EvasionState           int    `db:"evasion_state" json:"evasionState"`
	EvasionTargetVillageID *int   `db:"evasion_target_village_id" json:"evasionTargetVillageId"`
}

type Building struct {
	ID                int  `db:"id" json:"id"`
	VillageID         int  `db:"village_id" json:"villageId"`
	Type              int  `db:"type" json:"type"`
	Level             int  `db:"level" json:"level"`
	IsUnderConstruction bool `db:"is_under_construction" json:"isUnderConstruction"`
	Location          int  `db:"location" json:"location"`
}

type QueueBuilding struct {
	ID           int       `db:"id" json:"id"`
	VillageID    int       `db:"village_id" json:"villageId"`
	Position     int       `db:"position" json:"position"`
	Location     int       `db:"location" json:"location"`
	Type         int       `db:"type" json:"type"`
	Level        int       `db:"level" json:"level"`
	CompleteTime SQLiteTime `db:"complete_time" json:"completeTime"`
}

type Storage struct {
	ID        int `db:"id" json:"id"`
	VillageID int `db:"village_id" json:"villageId"`
	Wood      int `db:"wood" json:"wood"`
	Clay      int `db:"clay" json:"clay"`
	Iron      int `db:"iron" json:"iron"`
	Crop      int `db:"crop" json:"crop"`
	Warehouse int `db:"warehouse" json:"warehouse"`
	Granary   int `db:"granary" json:"granary"`
	FreeCrop  int `db:"free_crop" json:"freeCrop"`
}

type VillageSettingRow struct {
	ID        int `db:"id" json:"id"`
	VillageID int `db:"village_id" json:"villageId"`
	Setting   int `db:"setting" json:"setting"`
	Value     int `db:"value" json:"value"`
}
