package model

type Farm struct {
	ID        int    `db:"id" json:"id"`
	AccountID int    `db:"account_id" json:"accountId"`
	Name      string `db:"name" json:"name"`
	IsActive  bool   `db:"is_active" json:"isActive"`
}
