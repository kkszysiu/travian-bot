package database

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	"travian-bot/internal/domain/enum"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB wraps sqlx.DB with application-specific operations.
type DB struct {
	*sqlx.DB
}

// DefaultDBPath returns the platform-specific database path.
func DefaultDBPath() (string, error) {
	var dir string
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, "Library", "Application Support", "travian-bot")
	case "linux":
		configDir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(configDir, "travian-bot")
	default: // windows
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		dir = filepath.Join(appData, "travian-bot")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "TBS.db"), nil
}

// Open creates or opens the SQLite database, runs migrations, and fills default settings.
func Open(dbPath string) (*DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)", dbPath)
	conn, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Single connection for SQLite
	conn.SetMaxOpenConns(1)

	db := &DB{conn}

	if err := db.runMigrations(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	if err := db.fillAccountSettings(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("fill account settings: %w", err)
	}

	if err := db.fillVillageSettings(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("fill village settings: %w", err)
	}

	return db, nil
}

func (db *DB) runMigrations() error {
	// Create migrations tracking table
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	if err != nil {
		return err
	}

	migrations := []struct {
		version  int
		filename string
	}{
		{1, "migrations/001_initial.sql"},
		{2, "migrations/002_transfer_rules.sql"},
		{3, "migrations/003_attack_evasion.sql"},
	}

	for _, m := range migrations {
		var applied int
		err = db.Get(&applied, "SELECT COUNT(*) FROM schema_migrations WHERE version = ?", m.version)
		if err != nil {
			return err
		}
		if applied > 0 {
			continue
		}

		sql, err := migrationsFS.ReadFile(m.filename)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", m.filename, err)
		}
		if _, err := db.Exec(string(sql)); err != nil {
			return fmt.Errorf("apply migration %s: %w", m.filename, err)
		}
		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", m.version); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) fillAccountSettings() error {
	// For each account, insert any missing default settings
	var accountIDs []int
	if err := db.Select(&accountIDs, "SELECT id FROM accounts"); err != nil {
		return err
	}

	for _, accountID := range accountIDs {
		for setting, value := range enum.DefaultAccountSettings {
			_, err := db.Exec(
				"INSERT OR IGNORE INTO accounts_setting (account_id, setting, value) VALUES (?, ?, ?)",
				accountID, int(setting), value,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DB) fillVillageSettings() error {
	var villageIDs []int
	if err := db.Select(&villageIDs, "SELECT id FROM villages"); err != nil {
		return err
	}

	for _, villageID := range villageIDs {
		for setting, value := range enum.DefaultVillageSettings {
			_, err := db.Exec(
				"INSERT OR IGNORE INTO villages_setting (village_id, setting, value) VALUES (?, ?, ?)",
				villageID, int(setting), value,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// FillVillageSettingsForNew fills default village settings for a newly added village,
// copying the tribe setting from the account settings.
func (db *DB) FillVillageSettingsForNew(accountID, villageID int) error {
	var tribe int
	err := db.Get(&tribe,
		"SELECT value FROM accounts_setting WHERE account_id = ? AND setting = ?",
		accountID, int(enum.AccountSettingTribe),
	)
	if err != nil {
		tribe = 0
	}

	for setting, value := range enum.DefaultVillageSettings {
		v := value
		if setting == enum.VillageSettingTribe {
			v = tribe
		}
		_, err := db.Exec(
			"INSERT OR IGNORE INTO villages_setting (village_id, setting, value) VALUES (?, ?, ?)",
			villageID, int(setting), v,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
