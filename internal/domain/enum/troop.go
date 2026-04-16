package enum

type Troop int

const (
	TroopNone Troop = 0

	// Romans
	TroopLegionnaire         Troop = 1
	TroopPraetorian          Troop = 2
	TroopImperian            Troop = 3
	TroopEquitesLegati       Troop = 4
	TroopEquitesImperatoris  Troop = 5
	TroopEquitesCaesaris     Troop = 6
	TroopRomanRam            Troop = 7
	TroopRomanCatapult       Troop = 8
	TroopRomanChief          Troop = 9
	TroopRomanSettler        Troop = 10

	// Teutons
	TroopClubswinger    Troop = 11
	TroopSpearman       Troop = 12
	TroopAxeman         Troop = 13
	TroopScout          Troop = 14
	TroopPaladin        Troop = 15
	TroopTeutonicKnight Troop = 16
	TroopTeutonRam      Troop = 17
	TroopTeutonCatapult Troop = 18
	TroopTeutonChief    Troop = 19
	TroopTeutonSettler  Troop = 20

	// Gauls
	TroopPhalanx          Troop = 21
	TroopSwordsman        Troop = 22
	TroopPathfinder       Troop = 23
	TroopTheutatesThunder Troop = 24
	TroopDruidrider       Troop = 25
	TroopHaeduan          Troop = 26
	TroopGaulRam          Troop = 27
	TroopGaulCatapult     Troop = 28
	TroopGaulChief        Troop = 29
	TroopGaulSettler      Troop = 30

	// Nature
	TroopRat       Troop = 31
	TroopSpider    Troop = 32
	TroopSnake     Troop = 33
	TroopBat       Troop = 34
	TroopWildBoar  Troop = 35
	TroopWolf      Troop = 36
	TroopBear      Troop = 37
	TroopCrocodile Troop = 38
	TroopTiger     Troop = 39
	TroopElephant  Troop = 40

	// Natars
	TroopPikeman          Troop = 41
	TroopThornedWarrior   Troop = 42
	TroopGuardsman        Troop = 43
	TroopBirdsOfPrey      Troop = 44
	TroopAxerider         Troop = 45
	TroopNatarianKnight   Troop = 46
	TroopWarelephant      Troop = 47
	TroopBallista         Troop = 48
	TroopNatarianEmperor  Troop = 49
	TroopNatarSettler     Troop = 50

	// Egyptians
	TroopSlaveMilitia    Troop = 51
	TroopAshWarden       Troop = 52
	TroopKhopeshWarrior  Troop = 53
	TroopSopduExplorer   Troop = 54
	TroopAnhurGuard      Troop = 55
	TroopReshephChariot  Troop = 56
	TroopEgyptianRam     Troop = 57
	TroopEgyptianCatapult Troop = 58
	TroopEgyptianChief   Troop = 59
	TroopEgyptianSettler Troop = 60

	// Huns
	TroopMercenary    Troop = 61
	TroopBowman       Troop = 62
	TroopSpotter      Troop = 63
	TroopSteppeRider  Troop = 64
	TroopMarksman     Troop = 65
	TroopMarauder     Troop = 66
	TroopHunRam       Troop = 67
	TroopHunCatapult  Troop = 68
	TroopHunChief     Troop = 69
	TroopHunSettler   Troop = 70

	// Hero
	TroopHero Troop = 71
)

func (t Troop) String() string {
	names := map[Troop]string{
		TroopNone: "None",
		// Romans
		TroopLegionnaire: "Legionnaire", TroopPraetorian: "Praetorian",
		TroopImperian: "Imperian", TroopEquitesLegati: "Equites Legati",
		TroopEquitesImperatoris: "Equites Imperatoris", TroopEquitesCaesaris: "Equites Caesaris",
		TroopRomanRam: "Battering Ram", TroopRomanCatapult: "Fire Catapult",
		TroopRomanChief: "Senator", TroopRomanSettler: "Settler",
		// Teutons
		TroopClubswinger: "Clubswinger", TroopSpearman: "Spearman",
		TroopAxeman: "Axeman", TroopScout: "Scout",
		TroopPaladin: "Paladin", TroopTeutonicKnight: "Teutonic Knight",
		TroopTeutonRam: "Ram", TroopTeutonCatapult: "Catapult",
		TroopTeutonChief: "Chief", TroopTeutonSettler: "Settler",
		// Gauls
		TroopPhalanx: "Phalanx", TroopSwordsman: "Swordsman",
		TroopPathfinder: "Pathfinder", TroopTheutatesThunder: "Theutates Thunder",
		TroopDruidrider: "Druidrider", TroopHaeduan: "Haeduan",
		TroopGaulRam: "Ram", TroopGaulCatapult: "Trebuchet",
		TroopGaulChief: "Chieftain", TroopGaulSettler: "Settler",
		// Nature
		TroopRat: "Rat", TroopSpider: "Spider", TroopSnake: "Snake",
		TroopBat: "Bat", TroopWildBoar: "Wild Boar", TroopWolf: "Wolf",
		TroopBear: "Bear", TroopCrocodile: "Crocodile",
		TroopTiger: "Tiger", TroopElephant: "Elephant",
		// Natars
		TroopPikeman: "Pikeman", TroopThornedWarrior: "Thorned Warrior",
		TroopGuardsman: "Guardsman", TroopBirdsOfPrey: "Birds of Prey",
		TroopAxerider: "Axerider", TroopNatarianKnight: "Natarian Knight",
		TroopWarelephant: "War Elephant", TroopBallista: "Ballista",
		TroopNatarianEmperor: "Natarian Emperor", TroopNatarSettler: "Settler",
		// Egyptians
		TroopSlaveMilitia: "Slave Militia", TroopAshWarden: "Ash Warden",
		TroopKhopeshWarrior: "Khopesh Warrior", TroopSopduExplorer: "Sopdu Explorer",
		TroopAnhurGuard: "Anhur Guard", TroopReshephChariot: "Resheph Chariot",
		TroopEgyptianRam: "Ram", TroopEgyptianCatapult: "Catapult",
		TroopEgyptianChief: "Nomarch", TroopEgyptianSettler: "Settler",
		// Huns
		TroopMercenary: "Mercenary", TroopBowman: "Bowman",
		TroopSpotter: "Spotter", TroopSteppeRider: "Steppe Rider",
		TroopMarksman: "Marksman", TroopMarauder: "Marauder",
		TroopHunRam: "Ram", TroopHunCatapult: "Catapult",
		TroopHunChief: "Logades", TroopHunSettler: "Settler",
		// Hero
		TroopHero: "Hero",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return "Unknown"
}

// GetTrainBuilding returns the base building type where this troop is trained.
// Great Barracks/Great Stable train the same troops as their base counterparts.
func (t Troop) GetTrainBuilding() Building {
	switch t {
	// Romans: Barracks
	case TroopLegionnaire, TroopPraetorian, TroopImperian:
		return BuildingBarracks
	// Romans: Stable
	case TroopEquitesLegati, TroopEquitesImperatoris, TroopEquitesCaesaris:
		return BuildingStable
	// Romans: Workshop
	case TroopRomanRam, TroopRomanCatapult:
		return BuildingWorkshop

	// Teutons: Barracks
	case TroopClubswinger, TroopSpearman, TroopAxeman:
		return BuildingBarracks
	// Teutons: Stable
	case TroopScout, TroopPaladin, TroopTeutonicKnight:
		return BuildingStable
	// Teutons: Workshop
	case TroopTeutonRam, TroopTeutonCatapult:
		return BuildingWorkshop

	// Gauls: Barracks
	case TroopPhalanx, TroopSwordsman:
		return BuildingBarracks
	// Gauls: Stable
	case TroopPathfinder, TroopTheutatesThunder, TroopDruidrider, TroopHaeduan:
		return BuildingStable
	// Gauls: Workshop
	case TroopGaulRam, TroopGaulCatapult:
		return BuildingWorkshop

	// Egyptians: Barracks
	case TroopSlaveMilitia, TroopAshWarden, TroopKhopeshWarrior:
		return BuildingBarracks
	// Egyptians: Stable
	case TroopSopduExplorer, TroopAnhurGuard, TroopReshephChariot:
		return BuildingStable
	// Egyptians: Workshop
	case TroopEgyptianRam, TroopEgyptianCatapult:
		return BuildingWorkshop

	// Huns: Barracks
	case TroopMercenary, TroopBowman:
		return BuildingBarracks
	// Huns: Stable
	case TroopSpotter, TroopSteppeRider, TroopMarksman, TroopMarauder:
		return BuildingStable
	// Huns: Workshop
	case TroopHunRam, TroopHunCatapult:
		return BuildingWorkshop
	}
	return BuildingSite // chief/settler/nature/natars — not trainable
}

// GetTribe returns the tribe that the troop belongs to.
func (t Troop) GetTribe() Tribe {
	v := int(t)
	switch {
	case v >= int(TroopLegionnaire) && v <= int(TroopRomanSettler):
		return TribeRomans
	case v >= int(TroopClubswinger) && v <= int(TroopTeutonSettler):
		return TribeTeutons
	case v >= int(TroopPhalanx) && v <= int(TroopGaulSettler):
		return TribeGauls
	case v >= int(TroopRat) && v <= int(TroopElephant):
		return TribeNature
	case v >= int(TroopPikeman) && v <= int(TroopNatarSettler):
		return TribeNatars
	case v >= int(TroopSlaveMilitia) && v <= int(TroopEgyptianSettler):
		return TribeEgyptians
	case v >= int(TroopMercenary) && v <= int(TroopHunSettler):
		return TribeHuns
	}
	return TribeAny
}
