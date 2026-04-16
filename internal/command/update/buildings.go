package update

import (
	"fmt"
	"strings"
	"time"

	"travian-bot/internal/browser"
	"travian-bot/internal/database"
	"travian-bot/internal/domain/enum"
	"travian-bot/internal/domain/model"
	"travian-bot/internal/event"
	"travian-bot/internal/parser"
)

// UpdateBuildings parses buildings and queue from the current page HTML,
// syncs them to the database, and emits a BuildingsModified event.
// The page must be on dorf1 (resource fields) or dorf2 (infrastructure).
func UpdateBuildings(b *browser.Browser, db *database.DB, bus *event.Bus, villageID int) error {
	html, err := b.PageHTML()
	if err != nil {
		return fmt.Errorf("get page html: %w", err)
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return fmt.Errorf("parse html: %w", err)
	}

	// Determine which page we're on from the URL
	currentURL := b.CurrentURL()
	var parsedBuildings []parser.BuildingDTO
	if strings.Contains(currentURL, "dorf1") {
		parsedBuildings = parser.GetFields(doc)
	} else if strings.Contains(currentURL, "dorf2") {
		parsedBuildings = parser.GetInfrastructures(doc)
	} else {
		return fmt.Errorf("not on dorf1 or dorf2 page, url: %s", currentURL)
	}

	// Parse queue buildings
	parsedQueue := parser.GetQueueBuilding(doc)

	// Resolve wall type for location 40 from village tribe setting
	wallType := enum.BuildingSite
	if strings.Contains(currentURL, "dorf2") {
		wallType = getWallTypeForVillage(db, villageID)
	}

	// Get existing buildings and queue from DB
	var existingBuildings []model.Building
	if err := db.Select(&existingBuildings,
		"SELECT id, village_id, type, level, is_under_construction, location FROM buildings WHERE village_id = ?",
		villageID,
	); err != nil {
		return fmt.Errorf("get existing buildings: %w", err)
	}

	var existingQueue []model.QueueBuilding
	if err := db.Select(&existingQueue,
		"SELECT id, village_id, position, location, type, level, complete_time FROM queue_buildings WHERE village_id = ?",
		villageID,
	); err != nil {
		return fmt.Errorf("get existing queue: %w", err)
	}

	existingByLocation := make(map[int]model.Building, len(existingBuildings))
	for _, eb := range existingBuildings {
		existingByLocation[eb.Location] = eb
	}

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// --- Sync buildings ---
	for _, pb := range parsedBuildings {
		buildingType := pb.Type

		// Wall special case: location 40, resolve type from tribe
		if pb.Location == 40 && wallType != enum.BuildingSite {
			buildingType = int(wallType)
		}

		if existing, found := existingByLocation[pb.Location]; found {
			// Update existing building
			_, err := tx.Exec(
				"UPDATE buildings SET type = ?, level = ?, is_under_construction = ? WHERE id = ?",
				buildingType, pb.Level, boolToInt(pb.IsUnderConstruction), existing.ID,
			)
			if err != nil {
				return fmt.Errorf("update building at location %d: %w", pb.Location, err)
			}
		} else {
			// Insert new building
			_, err := tx.Exec(
				"INSERT INTO buildings (village_id, type, level, is_under_construction, location) VALUES (?, ?, ?, ?, ?)",
				villageID, buildingType, pb.Level, boolToInt(pb.IsUnderConstruction), pb.Location,
			)
			if err != nil {
				return fmt.Errorf("insert building at location %d: %w", pb.Location, err)
			}
		}
	}

	// --- Sync queue buildings ---

	// Delete completed queue buildings (complete_time in the past)
	now := time.Now()
	for _, eq := range existingQueue {
		if !eq.CompleteTime.IsZero() && eq.CompleteTime.Before(now) {
			_, err := tx.Exec("DELETE FROM queue_buildings WHERE id = ?", eq.ID)
			if err != nil {
				return fmt.Errorf("delete completed queue building %d: %w", eq.ID, err)
			}
		}
	}

	// Reload existing queue after deletions (within the same transaction)
	var remainingQueue []model.QueueBuilding
	if err := tx.Select(&remainingQueue,
		"SELECT id, village_id, position, location, type, level, complete_time FROM queue_buildings WHERE village_id = ?",
		villageID,
	); err != nil {
		return fmt.Errorf("get remaining queue: %w", err)
	}

	// Build a set of matched existing queue IDs to track which ones are still valid
	matchedQueueIDs := make(map[int]bool)

	// Build under-construction location lookup from parsed buildings for filling missing locations
	underConstructionByType := make(map[int]int) // type -> location
	for _, pb := range parsedBuildings {
		if pb.IsUnderConstruction {
			underConstructionByType[pb.Type] = pb.Location
		}
	}

	// For matching: group remaining queue entries by (level, type) to handle duplicates
	type queueKey struct {
		level int
		typ   int
	}
	remainingByKey := make(map[queueKey][]model.QueueBuilding)
	for _, rq := range remainingQueue {
		key := queueKey{level: rq.Level, typ: rq.Type}
		remainingByKey[key] = append(remainingByKey[key], rq)
	}

	// Build lookup of under-construction buildings by target level (level+1)
	// for fallback when name-based type resolution fails (non-English locales)
	type ucBuilding struct {
		typ      int
		location int
		used     bool
	}
	var underConstructionList []*ucBuilding
	for _, pb := range parsedBuildings {
		if pb.IsUnderConstruction {
			underConstructionList = append(underConstructionList, &ucBuilding{
				typ: pb.Type, location: pb.Location,
			})
		}
	}

	// Process each parsed queue building
	for i, pq := range parsedQueue {
		position := i + 1
		completeTime := time.Now().Add(pq.Duration)

		// Resolve building type from name (works for English locale)
		resolvedType := resolveBuildingTypeFromName(pq.Type)

		// Fallback: match against under-construction buildings by target level
		if resolvedType == 0 {
			for _, uc := range underConstructionList {
				if !uc.used {
					resolvedType = uc.typ
					uc.used = true
					break
				}
			}
		}

		// Resolve location from under-construction map or from the matched building
		location := pq.Location
		if location < 0 && resolvedType > 0 {
			if loc, ok := underConstructionByType[resolvedType]; ok {
				location = loc
			}
		}

		// Try to match with an existing queue entry by level and resolved type
		matched := false
		if resolvedType > 0 {
			key := queueKey{level: pq.Level, typ: resolvedType}
			if candidates, ok := remainingByKey[key]; ok && len(candidates) > 0 {
				candidate := candidates[0]
				remainingByKey[key] = candidates[1:]
				matchedQueueIDs[candidate.ID] = true
				_, err := tx.Exec(
					"UPDATE queue_buildings SET position = ?, location = ?, complete_time = ? WHERE id = ?",
					position, location, completeTime.Format(time.RFC3339), candidate.ID,
				)
				if err != nil {
					return fmt.Errorf("update queue building %d: %w", candidate.ID, err)
				}
				matched = true
			}
		}

		// Fallback: if type resolution failed (non-English locale, or dorf1 page),
		// match existing queue entries by level alone and preserve their stored type.
		if !matched && resolvedType == 0 {
			for key, candidates := range remainingByKey {
				if key.level == pq.Level && len(candidates) > 0 {
					candidate := candidates[0]
					remainingByKey[key] = candidates[1:]
					matchedQueueIDs[candidate.ID] = true
					// Preserve existing type and location
					if location < 0 {
						location = candidate.Location
					}
					_, err := tx.Exec(
						"UPDATE queue_buildings SET position = ?, complete_time = ? WHERE id = ?",
						position, completeTime.Format(time.RFC3339), candidate.ID,
					)
					if err != nil {
						return fmt.Errorf("update queue building %d: %w", candidate.ID, err)
					}
					matched = true
					break
				}
			}
		}

		if !matched {
			buildingType := 0
			if resolvedType > 0 {
				buildingType = resolvedType
			}
			_, err := tx.Exec(
				"INSERT INTO queue_buildings (village_id, position, location, type, level, complete_time) VALUES (?, ?, ?, ?, ?, ?)",
				villageID, position, location, buildingType, pq.Level, completeTime.Format(time.RFC3339),
			)
			if err != nil {
				return fmt.Errorf("insert queue building: %w", err)
			}
		}
	}

	// Remove stale queue buildings that were not matched by any parsed entry
	for _, rq := range remainingQueue {
		if !matchedQueueIDs[rq.ID] {
			_, err := tx.Exec("DELETE FROM queue_buildings WHERE id = ?", rq.ID)
			if err != nil {
				return fmt.Errorf("delete stale queue building %d: %w", rq.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	bus.Emit(event.BuildingsModified, villageID)
	return nil
}

// getWallTypeForVillage looks up the tribe setting for a village and returns the wall building type.
func getWallTypeForVillage(db *database.DB, villageID int) enum.Building {
	var tribeValue int
	err := db.Get(&tribeValue,
		"SELECT value FROM villages_setting WHERE village_id = ? AND setting = ?",
		villageID, int(enum.VillageSettingTribe),
	)
	if err != nil || tribeValue == 0 {
		return enum.BuildingSite
	}
	return enum.GetWallForTribe(enum.Tribe(tribeValue))
}

// resolveBuildingTypeFromName maps a building type name string back to its enum int value.
// This is used to match parsed queue building names to building types.
func resolveBuildingTypeFromName(name string) int {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0
	}

	nameToType := map[string]enum.Building{
		"Woodcutter":             enum.BuildingWoodcutter,
		"Clay Pit":              enum.BuildingClayPit,
		"Iron Mine":             enum.BuildingIronMine,
		"Cropland":              enum.BuildingCropland,
		"Sawmill":               enum.BuildingSawmill,
		"Brickyard":             enum.BuildingBrickyard,
		"Iron Foundry":          enum.BuildingIronFoundry,
		"Grain Mill":            enum.BuildingGrainMill,
		"Bakery":                enum.BuildingBakery,
		"Warehouse":             enum.BuildingWarehouse,
		"Granary":               enum.BuildingGranary,
		"Blacksmith":            enum.BuildingBlacksmith,
		"Smithy":                enum.BuildingSmithy,
		"Tournament Square":     enum.BuildingTournamentSquare,
		"Main Building":         enum.BuildingMainBuilding,
		"Rally Point":           enum.BuildingRallyPoint,
		"Marketplace":           enum.BuildingMarketplace,
		"Embassy":               enum.BuildingEmbassy,
		"Barracks":              enum.BuildingBarracks,
		"Stable":                enum.BuildingStable,
		"Workshop":              enum.BuildingWorkshop,
		"Academy":               enum.BuildingAcademy,
		"Cranny":                enum.BuildingCranny,
		"Town Hall":             enum.BuildingTownHall,
		"Residence":             enum.BuildingResidence,
		"Palace":                enum.BuildingPalace,
		"Treasury":              enum.BuildingTreasury,
		"Trade Office":          enum.BuildingTradeOffice,
		"Great Barracks":        enum.BuildingGreatBarracks,
		"Great Stable":          enum.BuildingGreatStable,
		"City Wall":             enum.BuildingCityWall,
		"Earth Wall":            enum.BuildingEarthWall,
		"Palisade":              enum.BuildingPalisade,
		"Stonemason's Lodge":    enum.BuildingStonemasonsLodge,
		"Brewery":               enum.BuildingBrewery,
		"Trapper":               enum.BuildingTrapper,
		"Hero's Mansion":        enum.BuildingHerosMansion,
		"Great Warehouse":       enum.BuildingGreatWarehouse,
		"Great Granary":         enum.BuildingGreatGranary,
		"Wonder of the World":   enum.BuildingWW,
		"Horse Drinking Trough": enum.BuildingHorseDrinkingTrough,
		"Stone Wall":            enum.BuildingStoneWall,
		"Makeshift Wall":        enum.BuildingMakeshiftWall,
		"Command Center":        enum.BuildingCommandCenter,
		"Waterworks":            enum.BuildingWaterworks,
		"Hospital":              enum.BuildingHospital,
	}

	if bt, ok := nameToType[name]; ok {
		return int(bt)
	}

	// Case-insensitive fallback
	nameLower := strings.ToLower(name)
	for k, v := range nameToType {
		if strings.ToLower(k) == nameLower {
			return int(v)
		}
	}

	return 0
}

// boolToInt converts a bool to 0 or 1 for SQLite storage.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
