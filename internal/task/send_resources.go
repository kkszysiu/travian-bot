package task

import (
	"context"
	"fmt"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/command/feature"
	"travian-bot/internal/command/navigate"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/domain/errs"
	"travian-bot/internal/event"
	"travian-bot/internal/service"
)

// SendResourcesTask automatically sends excess resources to villages that need them.
type SendResourcesTask struct {
	BaseTask
	villageID int
	bus       *event.Bus
}

func NewSendResourcesTask(accountID, villageID int, bus *event.Bus) *SendResourcesTask {
	return &SendResourcesTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		villageID: villageID,
		bus:       bus,
	}
}

func (t *SendResourcesTask) Description() string { return "SendResources" }
func (t *SendResourcesTask) VillageID() int      { return t.villageID }

type villageStorage struct {
	VillageID int   `db:"village_id"`
	Wood      int64 `db:"wood"`
	Clay      int64 `db:"clay"`
	Iron      int64 `db:"iron"`
	Crop      int64 `db:"crop"`
	Warehouse int64 `db:"warehouse"`
	Granary   int64 `db:"granary"`
	X         int   `db:"x"`
	Y         int   `db:"y"`
}

func (t *SendResourcesTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
	enabled, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingAutoSendResourceEnable)
	if enabled == 0 {
		return errs.NewSkipError("send resources disabled", time.Time{})
	}

	threshold, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingAutoSendResourceThreshold)
	if threshold <= 0 {
		threshold = 95
	}

	// Get current village storage
	var cur villageStorage
	err := db.Get(&cur,
		"SELECT s.village_id, s.wood, s.clay, s.iron, s.crop, s.warehouse, s.granary, v.x, v.y FROM storages s JOIN villages v ON v.id = s.village_id WHERE s.village_id = ?",
		t.villageID)
	if err != nil || cur.Warehouse <= 0 {
		t.reschedule(db)
		return nil
	}

	// Check which resources are above threshold
	thresholdWood := cur.Warehouse * int64(threshold) / 100
	thresholdClay := cur.Warehouse * int64(threshold) / 100
	thresholdIron := cur.Warehouse * int64(threshold) / 100
	thresholdCrop := cur.Granary * int64(threshold) / 100

	woodOver := cur.Wood > thresholdWood
	clayOver := cur.Clay > thresholdClay
	ironOver := cur.Iron > thresholdIron
	cropOver := cur.Crop > thresholdCrop && cur.Granary > 0

	if !woodOver && !clayOver && !ironOver && !cropOver {
		t.reschedule(db)
		return nil
	}

	// Calculate excess — send down to 50% of capacity
	halfWarehouse := cur.Warehouse / 2
	halfGranary := cur.Granary / 2

	var sendWood, sendClay, sendIron, sendCrop int
	if woodOver {
		sendWood = int(cur.Wood - halfWarehouse)
	}
	if clayOver {
		sendClay = int(cur.Clay - halfWarehouse)
	}
	if ironOver {
		sendIron = int(cur.Iron - halfWarehouse)
	}
	if cropOver {
		sendCrop = int(cur.Crop - halfGranary)
	}

	if sendWood <= 0 && sendClay <= 0 && sendIron <= 0 && sendCrop <= 0 {
		t.reschedule(db)
		return nil
	}

	// Get account ID to find sibling villages
	var accountID int
	if err := db.Get(&accountID, "SELECT account_id FROM villages WHERE id = ?", t.villageID); err != nil {
		t.reschedule(db)
		return nil
	}

	// Get all other villages' storage
	var others []villageStorage
	err = db.Select(&others, `
		SELECT s.village_id, s.wood, s.clay, s.iron, s.crop, s.warehouse, s.granary, v.x, v.y
		FROM storages s
		JOIN villages v ON v.id = s.village_id
		WHERE v.account_id = ? AND s.village_id != ? AND s.warehouse > 0`,
		accountID, t.villageID)
	if err != nil || len(others) == 0 {
		t.reschedule(db)
		return nil
	}

	// Find the village with the most need (lowest average fill %)
	bestIdx := -1
	bestScore := float64(999)
	for i, o := range others {
		// Calculate how full this village is for the resources we want to send
		score := float64(0)
		count := 0
		if sendWood > 0 && o.Warehouse > 0 {
			score += float64(o.Wood) / float64(o.Warehouse)
			count++
		}
		if sendClay > 0 && o.Warehouse > 0 {
			score += float64(o.Clay) / float64(o.Warehouse)
			count++
		}
		if sendIron > 0 && o.Warehouse > 0 {
			score += float64(o.Iron) / float64(o.Warehouse)
			count++
		}
		if sendCrop > 0 && o.Granary > 0 {
			score += float64(o.Crop) / float64(o.Granary)
			count++
		}
		if count == 0 {
			continue
		}
		avg := score / float64(count)
		if avg < bestScore {
			bestScore = avg
			bestIdx = i
		}
	}

	if bestIdx < 0 {
		t.reschedule(db)
		return nil
	}

	target := others[bestIdx]

	// Cap send amounts to target's available space
	if sendWood > 0 {
		space := int(target.Warehouse - target.Wood)
		if space <= 0 {
			sendWood = 0
		} else if sendWood > space {
			sendWood = space
		}
	}
	if sendClay > 0 {
		space := int(target.Warehouse - target.Clay)
		if space <= 0 {
			sendClay = 0
		} else if sendClay > space {
			sendClay = space
		}
	}
	if sendIron > 0 {
		space := int(target.Warehouse - target.Iron)
		if space <= 0 {
			sendIron = 0
		} else if sendIron > space {
			sendIron = space
		}
	}
	if sendCrop > 0 {
		space := int(target.Granary - target.Crop)
		if space <= 0 {
			sendCrop = 0
		} else if sendCrop > space {
			sendCrop = space
		}
	}

	if sendWood <= 0 && sendClay <= 0 && sendIron <= 0 && sendCrop <= 0 {
		t.reschedule(db)
		return nil
	}

	// Switch to village and send
	if err := navigate.SwitchVillage(ctx, b, t.villageID); err != nil {
		return fmt.Errorf("switch village: %w", err)
	}

	err = feature.SendResources(ctx, b, db, t.bus,
		t.villageID, target.X, target.Y,
		sendWood, sendClay, sendIron, sendCrop)
	if err != nil {
		// Log but don't fail — reschedule
		t.reschedule(db)
		return nil
	}

	t.reschedule(db)
	return nil
}

func (t *SendResourcesTask) reschedule(db *database.DB) {
	minVal, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingAutoSendResourceRepeatMin)
	maxVal, _ := service.GetVillageSettingValue(db, t.villageID, enum.VillageSettingAutoSendResourceRepeatMax)
	seconds := service.RandomBetween(minVal, maxVal)
	t.SetExecuteAt(time.Now().Add(time.Duration(seconds) * time.Second))
}
