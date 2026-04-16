package enum

type Status int

const (
	StatusOffline  Status = 0
	StatusStarting Status = 1
	StatusOnline   Status = 2
	StatusPausing  Status = 3
	StatusPaused   Status = 4
	StatusStopping Status = 5
)

func (s Status) String() string {
	switch s {
	case StatusOffline:
		return "Offline"
	case StatusStarting:
		return "Starting"
	case StatusOnline:
		return "Online"
	case StatusPausing:
		return "Pausing"
	case StatusPaused:
		return "Paused"
	case StatusStopping:
		return "Stopping"
	default:
		return "Unknown"
	}
}

// Color returns a CSS color string for UI display.
func (s Status) Color() string {
	switch s {
	case StatusOnline:
		return "green"
	case StatusStarting, StatusPausing, StatusStopping:
		return "orange"
	case StatusPaused:
		return "red"
	default:
		return "black"
	}
}
