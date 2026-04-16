package model

type Job struct {
	ID        int    `db:"id" json:"id"`
	VillageID int    `db:"village_id" json:"villageId"`
	Position  int    `db:"position" json:"position"`
	Type      int    `db:"type" json:"type"`
	Content   string `db:"content" json:"content"`
}
