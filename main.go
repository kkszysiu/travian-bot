package main

import (
	"embed"
	"log"
	"log/slog"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"travian-bot/internal/app"
	"travian-bot/internal/browser"
	"travian-bot/internal/database"
	"travian-bot/internal/event"
	"travian-bot/internal/task"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Open database
	dbPath, err := database.DefaultDBPath()
	if err != nil {
		log.Fatalf("failed to determine database path: %v", err)
	}
	logger.Info("database path", "path", dbPath)

	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	// Create event bus
	bus := event.NewBus()

	// Create browser manager and scheduler
	browsers := browser.NewManager(logger)
	scheduler := task.NewScheduler(db, browsers, bus, logger)

	// Create application
	application := app.NewApp(db, bus, browsers, scheduler, logger)

	err = wails.Run(&options.App{
		Title:  "Travian Bot Go",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        application.Startup,
		OnShutdown:       application.Shutdown,
		Bind: []interface{}{
			application,
		},
	})
	if err != nil {
		log.Fatalf("failed to run application: %v", err)
	}
}
