package event

// Event name constants used for Wails EventsEmit and internal subscriptions.
const (
	AccountsModified  = "accounts:modified"
	StatusModified    = "status:modified"
	TasksModified     = "tasks:modified"
	VillagesModified  = "villages:modified"
	BuildingsModified = "buildings:modified"
	JobsModified      = "jobs:modified"
	FarmsModified     = "farms:modified"
	StorageModified   = "storage:modified"
	LogEmitted             = "log:emitted"
	TransferRulesModified  = "transfer_rules:modified"
	EvasionStateModified   = "evasion:modified"
)

// StatusPayload is emitted when an account status changes.
type StatusPayload struct {
	AccountID int    `json:"accountId"`
	Status    int    `json:"status"`
	Color     string `json:"color"`
}

// LogPayload is emitted when a log message is produced.
type LogPayload struct {
	AccountID int    `json:"accountId"`
	Message   string `json:"message"`
	Level     string `json:"level"`
}
