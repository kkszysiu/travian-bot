package enum

import "math"

type Building int

const (
	BuildingUnknown            Building = -1
	BuildingSite               Building = 0
	BuildingWoodcutter         Building = 1
	BuildingClayPit            Building = 2
	BuildingIronMine           Building = 3
	BuildingCropland           Building = 4
	BuildingSawmill            Building = 5
	BuildingBrickyard          Building = 6
	BuildingIronFoundry        Building = 7
	BuildingGrainMill          Building = 8
	BuildingBakery             Building = 9
	BuildingWarehouse          Building = 10
	BuildingGranary            Building = 11
	BuildingBlacksmith         Building = 12 // Deprecated
	BuildingSmithy             Building = 13
	BuildingTournamentSquare   Building = 14
	BuildingMainBuilding       Building = 15
	BuildingRallyPoint         Building = 16
	BuildingMarketplace        Building = 17
	BuildingEmbassy            Building = 18
	BuildingBarracks           Building = 19
	BuildingStable             Building = 20
	BuildingWorkshop           Building = 21
	BuildingAcademy            Building = 22
	BuildingCranny             Building = 23
	BuildingTownHall           Building = 24
	BuildingResidence          Building = 25
	BuildingPalace             Building = 26
	BuildingTreasury           Building = 27
	BuildingTradeOffice        Building = 28
	BuildingGreatBarracks      Building = 29
	BuildingGreatStable        Building = 30
	BuildingCityWall           Building = 31
	BuildingEarthWall          Building = 32
	BuildingPalisade           Building = 33
	BuildingStonemasonsLodge   Building = 34
	BuildingBrewery            Building = 35
	BuildingTrapper            Building = 36
	BuildingHerosMansion       Building = 37
	BuildingGreatWarehouse     Building = 38
	BuildingGreatGranary       Building = 39
	BuildingWW                 Building = 40
	BuildingHorseDrinkingTrough Building = 41
	BuildingStoneWall          Building = 42
	BuildingMakeshiftWall      Building = 43
	BuildingCommandCenter      Building = 44
	BuildingWaterworks         Building = 45
	BuildingHospital           Building = 46
)

func (b Building) String() string {
	names := map[Building]string{
		BuildingUnknown: "Unknown", BuildingSite: "Site",
		BuildingWoodcutter: "Woodcutter", BuildingClayPit: "Clay Pit",
		BuildingIronMine: "Iron Mine", BuildingCropland: "Cropland",
		BuildingSawmill: "Sawmill", BuildingBrickyard: "Brickyard",
		BuildingIronFoundry: "Iron Foundry", BuildingGrainMill: "Grain Mill",
		BuildingBakery: "Bakery", BuildingWarehouse: "Warehouse",
		BuildingGranary: "Granary", BuildingBlacksmith: "Blacksmith",
		BuildingSmithy: "Smithy", BuildingTournamentSquare: "Tournament Square",
		BuildingMainBuilding: "Main Building", BuildingRallyPoint: "Rally Point",
		BuildingMarketplace: "Marketplace", BuildingEmbassy: "Embassy",
		BuildingBarracks: "Barracks", BuildingStable: "Stable",
		BuildingWorkshop: "Workshop", BuildingAcademy: "Academy",
		BuildingCranny: "Cranny", BuildingTownHall: "Town Hall",
		BuildingResidence: "Residence", BuildingPalace: "Palace",
		BuildingTreasury: "Treasury", BuildingTradeOffice: "Trade Office",
		BuildingGreatBarracks: "Great Barracks", BuildingGreatStable: "Great Stable",
		BuildingCityWall: "City Wall", BuildingEarthWall: "Earth Wall",
		BuildingPalisade: "Palisade", BuildingStonemasonsLodge: "Stonemason's Lodge",
		BuildingBrewery: "Brewery", BuildingTrapper: "Trapper",
		BuildingHerosMansion: "Hero's Mansion", BuildingGreatWarehouse: "Great Warehouse",
		BuildingGreatGranary: "Great Granary", BuildingWW: "Wonder of the World",
		BuildingHorseDrinkingTrough: "Horse Drinking Trough",
		BuildingStoneWall: "Stone Wall", BuildingMakeshiftWall: "Makeshift Wall",
		BuildingCommandCenter: "Command Center", BuildingWaterworks: "Waterworks",
		BuildingHospital: "Hospital",
	}
	if name, ok := names[b]; ok {
		return name
	}
	return "Unknown"
}

func (b Building) IsWall() bool {
	switch b {
	case BuildingCityWall, BuildingEarthWall, BuildingPalisade,
		BuildingStoneWall, BuildingMakeshiftWall:
		return true
	}
	return false
}

func GetWallForTribe(tribe Tribe) Building {
	switch tribe {
	case TribeRomans:
		return BuildingCityWall
	case TribeTeutons:
		return BuildingEarthWall
	case TribeGauls:
		return BuildingPalisade
	case TribeEgyptians:
		return BuildingStoneWall
	case TribeHuns:
		return BuildingMakeshiftWall
	}
	return BuildingSite
}

func (b Building) IsMultipleBuilding() bool {
	switch b {
	case BuildingWarehouse, BuildingGranary, BuildingGreatWarehouse,
		BuildingGreatGranary, BuildingTrapper, BuildingCranny:
		return true
	}
	return false
}

func (b Building) GetMaxLevel() int {
	switch b {
	case BuildingBakery, BuildingBrickyard, BuildingIronFoundry,
		BuildingGrainMill, BuildingSawmill:
		return 5
	case BuildingCranny:
		return 10
	}
	return 20
}

func (b Building) IsResourceBonus() bool {
	switch b {
	case BuildingSawmill, BuildingBrickyard, BuildingIronFoundry,
		BuildingBakery, BuildingGrainMill:
		return true
	}
	return false
}

func (b Building) IsResourceField() bool {
	return int(b) >= 1 && int(b) <= 4
}

func (b Building) GetColor() string {
	switch b {
	// Empty
	case BuildingSite:
		return "#9ca3af"
	// Resource fields
	case BuildingWoodcutter:
		return "#16a34a"
	case BuildingClayPit:
		return "#c2410c"
	case BuildingIronMine:
		return "#6b7280"
	case BuildingCropland:
		return "#ca8a04"
	// Resource bonus
	case BuildingSawmill, BuildingBrickyard, BuildingIronFoundry,
		BuildingGrainMill, BuildingBakery:
		return "#b45309"
	// Storage
	case BuildingWarehouse, BuildingGranary, BuildingGreatWarehouse, BuildingGreatGranary:
		return "#0369a1"
	// Military
	case BuildingBarracks, BuildingGreatBarracks, BuildingStable, BuildingGreatStable,
		BuildingWorkshop, BuildingAcademy, BuildingSmithy, BuildingHospital:
		return "#dc2626"
	// Defense
	case BuildingCityWall, BuildingEarthWall, BuildingPalisade,
		BuildingStoneWall, BuildingMakeshiftWall, BuildingTrapper, BuildingCranny:
		return "#7c3aed"
	// Infrastructure
	case BuildingMainBuilding, BuildingRallyPoint, BuildingTownHall,
		BuildingStonemasonsLodge, BuildingWaterworks:
		return "#4b5563"
	// Trade
	case BuildingMarketplace, BuildingTradeOffice:
		return "#0891b2"
	// Government
	case BuildingResidence, BuildingPalace, BuildingEmbassy,
		BuildingTreasury, BuildingCommandCenter:
		return "#9333ea"
	// Hero
	case BuildingHerosMansion, BuildingTournamentSquare:
		return "#be185d"
	// Special
	case BuildingBrewery, BuildingHorseDrinkingTrough:
		return "#65a30d"
	case BuildingWW:
		return "#d97706"
	}
	return "#0e7490"
}

func (b Building) GetBuildingsCategory() int {
	switch b {
	case BuildingGrainMill, BuildingSawmill, BuildingBrickyard,
		BuildingIronFoundry, BuildingBakery:
		return 2
	case BuildingAcademy, BuildingSmithy, BuildingBarracks,
		BuildingGreatBarracks, BuildingStable, BuildingGreatStable,
		BuildingWorkshop, BuildingHerosMansion, BuildingTournamentSquare,
		BuildingHospital, BuildingTrapper:
		return 1
	}
	return 0
}

func (b Building) HasMultipleTabs() bool {
	switch b {
	case BuildingRallyPoint, BuildingCommandCenter, BuildingResidence,
		BuildingPalace, BuildingMarketplace, BuildingTreasury:
		return true
	}
	return false
}

// PrerequisiteBuilding represents a building requirement.
type PrerequisiteBuilding struct {
	Building Building
	Level    int
}

func (b Building) GetPrerequisiteBuildings() []PrerequisiteBuilding {
	switch b {
	case BuildingSawmill:
		return []PrerequisiteBuilding{{BuildingWoodcutter, 10}, {BuildingMainBuilding, 5}}
	case BuildingBrickyard:
		return []PrerequisiteBuilding{{BuildingClayPit, 10}, {BuildingMainBuilding, 5}}
	case BuildingIronFoundry:
		return []PrerequisiteBuilding{{BuildingIronMine, 10}, {BuildingMainBuilding, 5}}
	case BuildingGrainMill:
		return []PrerequisiteBuilding{{BuildingCropland, 5}}
	case BuildingBakery:
		return []PrerequisiteBuilding{{BuildingCropland, 10}, {BuildingGrainMill, 5}, {BuildingMainBuilding, 5}}
	case BuildingWarehouse:
		return []PrerequisiteBuilding{{BuildingMainBuilding, 1}}
	case BuildingGranary:
		return []PrerequisiteBuilding{{BuildingMainBuilding, 1}}
	case BuildingSmithy:
		return []PrerequisiteBuilding{{BuildingMainBuilding, 3}, {BuildingAcademy, 1}}
	case BuildingTournamentSquare:
		return []PrerequisiteBuilding{{BuildingRallyPoint, 15}}
	case BuildingMarketplace:
		return []PrerequisiteBuilding{{BuildingWarehouse, 1}, {BuildingGranary, 1}, {BuildingMainBuilding, 3}}
	case BuildingEmbassy:
		return []PrerequisiteBuilding{{BuildingMainBuilding, 1}}
	case BuildingBarracks:
		return []PrerequisiteBuilding{{BuildingRallyPoint, 1}, {BuildingMainBuilding, 3}}
	case BuildingStable:
		return []PrerequisiteBuilding{{BuildingAcademy, 5}, {BuildingSmithy, 3}}
	case BuildingWorkshop:
		return []PrerequisiteBuilding{{BuildingAcademy, 10}, {BuildingMainBuilding, 5}}
	case BuildingAcademy:
		return []PrerequisiteBuilding{{BuildingBarracks, 3}, {BuildingMainBuilding, 3}}
	case BuildingTownHall:
		return []PrerequisiteBuilding{{BuildingAcademy, 10}, {BuildingMainBuilding, 10}}
	case BuildingResidence:
		return []PrerequisiteBuilding{{BuildingMainBuilding, 5}}
	case BuildingPalace:
		return []PrerequisiteBuilding{{BuildingMainBuilding, 5}, {BuildingEmbassy, 1}}
	case BuildingTreasury:
		return []PrerequisiteBuilding{{BuildingMainBuilding, 10}}
	case BuildingTradeOffice:
		return []PrerequisiteBuilding{{BuildingStable, 10}, {BuildingMarketplace, 20}}
	case BuildingGreatBarracks:
		return []PrerequisiteBuilding{{BuildingBarracks, 20}}
	case BuildingGreatStable:
		return []PrerequisiteBuilding{{BuildingStable, 20}}
	case BuildingStonemasonsLodge:
		return []PrerequisiteBuilding{{BuildingMainBuilding, 5}}
	case BuildingBrewery:
		return []PrerequisiteBuilding{{BuildingGranary, 20}, {BuildingRallyPoint, 10}}
	case BuildingTrapper:
		return []PrerequisiteBuilding{{BuildingRallyPoint, 1}}
	case BuildingHerosMansion:
		return []PrerequisiteBuilding{{BuildingMainBuilding, 3}, {BuildingRallyPoint, 1}}
	case BuildingGreatWarehouse:
		return []PrerequisiteBuilding{{BuildingMainBuilding, 10}}
	case BuildingGreatGranary:
		return []PrerequisiteBuilding{{BuildingMainBuilding, 10}}
	case BuildingHorseDrinkingTrough:
		return []PrerequisiteBuilding{{BuildingRallyPoint, 10}, {BuildingStable, 20}}
	case BuildingCommandCenter:
		return []PrerequisiteBuilding{{BuildingMainBuilding, 5}}
	case BuildingWaterworks:
		return []PrerequisiteBuilding{{BuildingHerosMansion, 10}}
	}
	return nil
}

func (b Building) getKBase() float64 {
	switch b {
	case BuildingWoodcutter, BuildingClayPit, BuildingIronMine, BuildingCropland:
		return 1.67
	case BuildingSawmill, BuildingBrickyard, BuildingIronFoundry,
		BuildingGrainMill, BuildingBakery:
		return 1.8
	case BuildingTreasury:
		return 1.26
	case BuildingBrewery:
		return 1.4
	case BuildingHerosMansion:
		return 1.33
	case BuildingWW:
		return 1.0275
	case BuildingCommandCenter:
		return 1.22
	case BuildingWaterworks:
		return 1.31
	case BuildingWarehouse, BuildingGranary, BuildingBlacksmith,
		BuildingSmithy, BuildingTournamentSquare, BuildingMainBuilding,
		BuildingRallyPoint, BuildingMarketplace, BuildingEmbassy,
		BuildingBarracks, BuildingStable, BuildingWorkshop,
		BuildingAcademy, BuildingCranny, BuildingTownHall,
		BuildingResidence, BuildingPalace, BuildingTradeOffice,
		BuildingGreatBarracks, BuildingGreatStable, BuildingCityWall,
		BuildingEarthWall, BuildingPalisade, BuildingStonemasonsLodge,
		BuildingTrapper, BuildingGreatWarehouse, BuildingGreatGranary,
		BuildingHorseDrinkingTrough, BuildingStoneWall, BuildingMakeshiftWall,
		BuildingHospital:
		return 1.28
	}
	return 0
}

func (b Building) getBaseCost() [5]int64 {
	costs := map[Building][5]int64{
		BuildingWoodcutter:          {40, 100, 50, 60, 2},
		BuildingClayPit:             {80, 40, 80, 50, 2},
		BuildingIronMine:            {100, 80, 30, 60, 3},
		BuildingCropland:            {70, 90, 70, 20, 0},
		BuildingSawmill:             {520, 380, 290, 90, 4},
		BuildingBrickyard:           {440, 480, 320, 50, 3},
		BuildingIronFoundry:         {200, 450, 510, 120, 6},
		BuildingGrainMill:           {500, 440, 380, 1240, 3},
		BuildingBakery:              {1200, 1480, 870, 1600, 4},
		BuildingWarehouse:           {130, 160, 90, 40, 1},
		BuildingGranary:             {80, 100, 70, 20, 1},
		BuildingBlacksmith:          {170, 200, 380, 130, 4},
		BuildingSmithy:              {130, 210, 410, 130, 4},
		BuildingTournamentSquare:    {1750, 2250, 1530, 240, 1},
		BuildingMainBuilding:        {70, 40, 60, 20, 2},
		BuildingRallyPoint:          {110, 160, 90, 70, 1},
		BuildingMarketplace:         {80, 70, 120, 70, 4},
		BuildingEmbassy:             {180, 130, 150, 80, 3},
		BuildingBarracks:            {210, 140, 260, 120, 4},
		BuildingStable:              {260, 140, 220, 100, 5},
		BuildingWorkshop:            {460, 510, 600, 320, 3},
		BuildingAcademy:             {220, 160, 90, 40, 4},
		BuildingCranny:              {40, 50, 30, 10, 0},
		BuildingTownHall:            {1250, 1110, 1260, 600, 4},
		BuildingResidence:           {580, 460, 350, 180, 1},
		BuildingPalace:              {550, 800, 750, 250, 1},
		BuildingTreasury:            {2880, 2740, 2580, 990, 4},
		BuildingTradeOffice:         {1400, 1330, 1200, 400, 3},
		BuildingGreatBarracks:       {630, 420, 780, 360, 4},
		BuildingGreatStable:         {780, 420, 660, 300, 5},
		BuildingCityWall:            {70, 90, 170, 70, 0},
		BuildingEarthWall:           {120, 200, 0, 80, 0},
		BuildingPalisade:            {160, 100, 80, 60, 0},
		BuildingStonemasonsLodge:    {155, 130, 125, 70, 2},
		BuildingBrewery:             {1460, 930, 1250, 1740, 6},
		BuildingTrapper:             {100, 100, 100, 100, 4},
		BuildingHerosMansion:        {700, 670, 700, 240, 2},
		BuildingGreatWarehouse:      {650, 800, 450, 200, 1},
		BuildingGreatGranary:        {400, 500, 350, 100, 1},
		BuildingWW:                  {66700, 69050, 72200, 13200, 1},
		BuildingHorseDrinkingTrough: {780, 420, 660, 540, 5},
		BuildingStoneWall:           {110, 160, 70, 60, 0},
		BuildingMakeshiftWall:       {50, 80, 40, 30, 0},
		BuildingCommandCenter:       {1600, 1250, 1050, 200, 1},
		BuildingWaterworks:          {910, 945, 910, 340, 1},
		BuildingHospital:            {320, 280, 420, 360, 3},
	}
	if c, ok := costs[b]; ok {
		return c
	}
	return [5]int64{0, 0, 0, 0, 0}
}

func roundMul(v, n float64) float64 {
	return math.Round(v/n) * n
}

// GetCost calculates the resource cost [wood, clay, iron, crop, upkeep] at a given level.
func (b Building) GetCost(level int) [5]int64 {
	k := b.getKBase()
	base := b.getBaseCost()
	var cost [5]int64
	for i, x := range base {
		val := roundMul(float64(x)*math.Pow(k, float64(level-1)), 5)
		if b == BuildingWW {
			val = math.Min(val, 1e6)
		}
		cost[i] = int64(val)
	}
	return cost
}
