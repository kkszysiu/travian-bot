package database

import (
	"fmt"
	"time"

	"travian-bot/internal/domain/enum"
	"travian-bot/internal/domain/model"
)

// AccountListItem is a lightweight DTO for the sidebar account list.
type AccountListItem struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Server   string `json:"server"`
}

// AccountDetail is the full DTO used for add/edit account forms.
type AccountDetail struct {
	ID       int            `json:"id"`
	Username string         `json:"username"`
	Server   string         `json:"server"`
	Accesses []AccessDetail `json:"accesses"`
}

type AccessDetail struct {
	ID            int    `json:"id"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	ProxyHost     string `json:"proxyHost"`
	ProxyPort     int    `json:"proxyPort"`
	ProxyUsername string `json:"proxyUsername"`
	ProxyPassword string `json:"proxyPassword"`
	Useragent     string `json:"useragent"`
	LastUsed      string `json:"lastUsed"`
}

func (db *DB) GetAccounts() ([]AccountListItem, error) {
	var accounts []model.Account
	if err := db.Select(&accounts, "SELECT id, username, server FROM accounts ORDER BY id"); err != nil {
		return nil, err
	}

	items := make([]AccountListItem, len(accounts))
	for i, a := range accounts {
		items[i] = AccountListItem{ID: a.ID, Username: a.Username, Server: a.Server}
	}
	return items, nil
}

func (db *DB) GetAccountDetail(accountID int) (*AccountDetail, error) {
	var account model.Account
	if err := db.Get(&account, "SELECT id, username, server FROM accounts WHERE id = ?", accountID); err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	var accesses []model.Access
	if err := db.Select(&accesses, "SELECT * FROM accesses WHERE account_id = ? ORDER BY id", accountID); err != nil {
		return nil, err
	}

	detail := &AccountDetail{
		ID:       account.ID,
		Username: account.Username,
		Server:   account.Server,
		Accesses: make([]AccessDetail, len(accesses)),
	}
	for i, a := range accesses {
		detail.Accesses[i] = AccessDetail{
			ID:            a.ID,
			Username:      a.Username,
			Password:      a.Password,
			ProxyHost:     a.ProxyHost,
			ProxyPort:     a.ProxyPort,
			ProxyUsername: a.ProxyUsername,
			ProxyPassword: a.ProxyPassword,
			Useragent:     a.Useragent,
			LastUsed:      a.LastUsed.Format(time.RFC3339),
		}
	}
	return detail, nil
}

func (db *DB) AddAccount(detail AccountDetail) (int, error) {
	tx, err := db.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		"INSERT INTO accounts (username, server) VALUES (?, ?)",
		detail.Username, detail.Server,
	)
	if err != nil {
		return 0, err
	}

	accountID64, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	accountID := int(accountID64)

	// Insert account info record
	_, err = tx.Exec(
		"INSERT INTO accounts_info (account_id) VALUES (?)",
		accountID,
	)
	if err != nil {
		return 0, err
	}

	// Insert accesses
	for _, a := range detail.Accesses {
		_, err := tx.Exec(
			`INSERT INTO accesses (account_id, username, password, proxy_host, proxy_port, proxy_username, proxy_password, useragent)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			accountID, a.Username, a.Password, a.ProxyHost, a.ProxyPort, a.ProxyUsername, a.ProxyPassword, a.Useragent,
		)
		if err != nil {
			return 0, err
		}
	}

	// Insert default account settings
	for setting, value := range enum.DefaultAccountSettings {
		_, err := tx.Exec(
			"INSERT INTO accounts_setting (account_id, setting, value) VALUES (?, ?, ?)",
			accountID, int(setting), value,
		)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return accountID, nil
}

func (db *DB) UpdateAccount(detail AccountDetail) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"UPDATE accounts SET username = ?, server = ? WHERE id = ?",
		detail.Username, detail.Server, detail.ID,
	)
	if err != nil {
		return err
	}

	// Replace all accesses: delete old, insert new
	_, err = tx.Exec("DELETE FROM accesses WHERE account_id = ?", detail.ID)
	if err != nil {
		return err
	}

	for _, a := range detail.Accesses {
		_, err := tx.Exec(
			`INSERT INTO accesses (account_id, username, password, proxy_host, proxy_port, proxy_username, proxy_password, useragent)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			detail.ID, a.Username, a.Password, a.ProxyHost, a.ProxyPort, a.ProxyUsername, a.ProxyPassword, a.Useragent,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) DeleteAccount(accountID int) error {
	_, err := db.Exec("DELETE FROM accounts WHERE id = ?", accountID)
	return err
}
