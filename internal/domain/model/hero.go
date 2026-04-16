package model

type HeroItem struct {
	ID        int `db:"id" json:"id"`
	AccountID int `db:"account_id" json:"accountId"`
	Type      int `db:"type" json:"type"`
	Amount    int `db:"amount" json:"amount"`
}
