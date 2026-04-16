package task

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"travian-bot/internal/browser"
	"travian-bot/internal/command/navigate"
	"travian-bot/internal/command/update"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/event"
	"travian-bot/internal/parser"
)

// RecallTroopsTask recalls reinforcements that were sent to a safe village during evasion.
type RecallTroopsTask struct {
	BaseTask
	villageID int
	bus       *event.Bus
}

func NewRecallTroopsTask(accountID, villageID int, bus *event.Bus) *RecallTroopsTask {
	return &RecallTroopsTask{
		BaseTask: BaseTask{
			accountID: accountID,
			executeAt: time.Now(),
		},
		villageID: villageID,
		bus:       bus,
	}
}

func (t *RecallTroopsTask) Description() string { return "Recall troops" }
func (t *RecallTroopsTask) VillageID() int      { return t.villageID }

func (t *RecallTroopsTask) Execute(ctx context.Context, b *browser.Browser, db *database.DB) error {
	// Check if this village has troops to recall
	var evasionState int
	if err := db.Get(&evasionState, "SELECT evasion_state FROM villages WHERE id = ?", t.villageID); err != nil {
		return fmt.Errorf("get evasion state: %w", err)
	}

	if evasionState == 0 {
		return nil // Nothing to recall
	}

	if evasionState&1 == 0 {
		// No troops were evacuated, just clear state
		db.ClearEvasionState(t.villageID)
		t.bus.Emit(event.VillagesModified, t.accountID)
		return nil
	}

	// Navigate to the source village's rally point overview tab (tt=1)
	if err := navigate.ToDorf(ctx, b, 2); err != nil {
		return fmt.Errorf("navigate to dorf2: %w", err)
	}
	if err := update.UpdateBuildings(b, db, t.bus, t.villageID); err != nil {
		return fmt.Errorf("update buildings: %w", err)
	}
	if err := navigate.ToBuildingByType(ctx, b, db, t.villageID, int(enum.BuildingRallyPoint)); err != nil {
		return fmt.Errorf("navigate to rally point: %w", err)
	}

	// Switch to overview tab (tt=1)
	if err := navigate.SwitchTab(ctx, b, 1); err != nil {
		return fmt.Errorf("switch to overview tab: %w", err)
	}
	time.Sleep(1 * time.Second)

	// Parse page and look for recall links (div.sback a.arrow)
	// There may be multiple — one for oasis nature troops ("zawróć") and one for reinforcements ("powrót").
	// We identify the correct one by checking href contains "from={villageID}" (our source village).
	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get rally point html: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse rally point html: %w", err)
	}

	// Find the correct recall link for our reinforcements
	fromParam := fmt.Sprintf("from=%d", t.villageID)
	found := false
	doc.Find(parser.GetRecallTroopsSelector()).Each(func(_ int, s *goquery.Selection) {
		if found {
			return
		}
		href, exists := s.Attr("href")
		if !exists {
			return
		}
		// Match recall links that originate from our village
		if strings.Contains(href, fromParam) {
			found = true
		}
	})

	if found {
		// Click the recall link that matches our village
		// Use a more specific selector with href matching
		selector := fmt.Sprintf("div.sback a.arrow[href*='from=%d']", t.villageID)
		el, err := b.Element(selector)
		if err != nil {
			// Fallback: click any recall link
			log.Printf("[RecallTroops] Specific selector failed, trying generic: %v", err)
			el, err = b.Element(parser.GetRecallTroopsSelector())
			if err != nil {
				log.Printf("[RecallTroops] Could not find any recall button for village %d: %v", t.villageID, err)
			} else {
				if err := b.Click(el); err != nil {
					log.Printf("[RecallTroops] Could not click recall button: %v", err)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		} else {
			if err := b.Click(el); err != nil {
				log.Printf("[RecallTroops] Could not click recall button: %v", err)
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	} else {
		log.Printf("[RecallTroops] No recall link found for village %d — troops may have already returned", t.villageID)
	}

	// Clear evasion state regardless — if troops already returned, that's fine
	db.ClearEvasionState(t.villageID)
	t.bus.Emit(event.EvasionStateModified, t.villageID)
	t.bus.Emit(event.VillagesModified, t.accountID)

	return nil
}
