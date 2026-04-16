package enum

type VillageSetting int

const (
	VillageSettingUseHeroResourceForBuilding      VillageSetting = 1
	VillageSettingApplyRomanQueueLogicWhenBuilding VillageSetting = 2
	VillageSettingUseSpecialUpgrade               VillageSetting = 3
	VillageSettingCompleteImmediately             VillageSetting = 4
	VillageSettingTribe                           VillageSetting = 5
	VillageSettingTrainTroopEnable                VillageSetting = 6
	VillageSettingTrainTroopRepeatTimeMin         VillageSetting = 7
	VillageSettingTrainTroopRepeatTimeMax         VillageSetting = 8
	VillageSettingTrainWhenLowResource            VillageSetting = 9
	VillageSettingBarrackTroop                    VillageSetting = 10
	VillageSettingBarrackAmountMin                VillageSetting = 11
	VillageSettingBarrackAmountMax                VillageSetting = 12
	VillageSettingStableTroop                     VillageSetting = 13
	VillageSettingStableAmountMin                 VillageSetting = 14
	VillageSettingStableAmountMax                 VillageSetting = 15
	VillageSettingGreatBarrackTroop               VillageSetting = 16
	VillageSettingGreatBarrackAmountMin           VillageSetting = 17
	VillageSettingGreatBarrackAmountMax           VillageSetting = 18
	VillageSettingGreatStableTroop                VillageSetting = 19
	VillageSettingGreatStableAmountMin            VillageSetting = 20
	VillageSettingGreatStableAmountMax            VillageSetting = 21
	VillageSettingWorkshopTroop                   VillageSetting = 22
	VillageSettingWorkshopAmountMin               VillageSetting = 23
	VillageSettingWorkshopAmountMax               VillageSetting = 24
	VillageSettingAutoNPCEnable                   VillageSetting = 25
	VillageSettingAutoNPCOverflow                 VillageSetting = 26
	VillageSettingAutoNPCGranaryPercent           VillageSetting = 27
	VillageSettingAutoNPCWood                     VillageSetting = 28
	VillageSettingAutoNPCClay                     VillageSetting = 29
	VillageSettingAutoNPCIron                     VillageSetting = 30
	VillageSettingAutoNPCCrop                     VillageSetting = 31
	VillageSettingAutoRefreshEnable               VillageSetting = 32
	VillageSettingAutoRefreshMin                  VillageSetting = 33
	VillageSettingAutoRefreshMax                  VillageSetting = 34
	VillageSettingAutoClaimQuestEnable            VillageSetting = 35
	VillageSettingCompleteImmediatelyTime         VillageSetting = 36
	VillageSettingAutoSendResourceEnable          VillageSetting = 37
	VillageSettingAutoSendResourceRepeatMin       VillageSetting = 38
	VillageSettingAutoSendResourceRepeatMax       VillageSetting = 39
	VillageSettingAutoSendResourceThreshold       VillageSetting = 40

	VillageSettingAttackEvasionEnable           VillageSetting = 41
	VillageSettingAttackEvasionSafeVillageID    VillageSetting = 42
	VillageSettingAttackEvasionEvacResources    VillageSetting = 43
	VillageSettingAttackEvasionCheckIntervalMin VillageSetting = 44
	VillageSettingAttackEvasionCheckIntervalMax VillageSetting = 45
)

// DefaultVillageSettings maps each village setting to its default value.
var DefaultVillageSettings = map[VillageSetting]int{
	VillageSettingUseHeroResourceForBuilding:      0,
	VillageSettingApplyRomanQueueLogicWhenBuilding: 0,
	VillageSettingUseSpecialUpgrade:               0,
	VillageSettingCompleteImmediately:             0,
	VillageSettingTribe:                           0,
	VillageSettingTrainTroopEnable:                0,
	VillageSettingTrainTroopRepeatTimeMin:         120,
	VillageSettingTrainTroopRepeatTimeMax:         180,
	VillageSettingTrainWhenLowResource:            0,
	VillageSettingBarrackTroop:                    0,
	VillageSettingBarrackAmountMin:                1,
	VillageSettingBarrackAmountMax:                10,
	VillageSettingStableTroop:                     0,
	VillageSettingStableAmountMin:                 1,
	VillageSettingStableAmountMax:                 10,
	VillageSettingGreatBarrackTroop:               0,
	VillageSettingGreatBarrackAmountMin:           1,
	VillageSettingGreatBarrackAmountMax:           10,
	VillageSettingGreatStableTroop:                0,
	VillageSettingGreatStableAmountMin:            1,
	VillageSettingGreatStableAmountMax:            10,
	VillageSettingWorkshopTroop:                   0,
	VillageSettingWorkshopAmountMin:               1,
	VillageSettingWorkshopAmountMax:               10,
	VillageSettingAutoNPCEnable:                   0,
	VillageSettingAutoNPCOverflow:                 0,
	VillageSettingAutoNPCGranaryPercent:           95,
	VillageSettingAutoNPCWood:                     1,
	VillageSettingAutoNPCClay:                     1,
	VillageSettingAutoNPCIron:                     1,
	VillageSettingAutoNPCCrop:                     0,
	VillageSettingAutoRefreshEnable:               0,
	VillageSettingAutoRefreshMin:                  45,
	VillageSettingAutoRefreshMax:                  75,
	VillageSettingAutoClaimQuestEnable:            0,
	VillageSettingCompleteImmediatelyTime:         20,
	VillageSettingAutoSendResourceEnable:          0,
	VillageSettingAutoSendResourceRepeatMin:       300,
	VillageSettingAutoSendResourceRepeatMax:       600,
	VillageSettingAutoSendResourceThreshold:       95,
	VillageSettingAttackEvasionEnable:             0,
	VillageSettingAttackEvasionSafeVillageID:      0,
	VillageSettingAttackEvasionEvacResources:      1,
	VillageSettingAttackEvasionCheckIntervalMin:   240,
	VillageSettingAttackEvasionCheckIntervalMax:   360,
}

// AllVillageSettings returns all village setting keys in order.
var AllVillageSettings []VillageSetting

func init() {
	for i := VillageSettingUseHeroResourceForBuilding; i <= VillageSettingAttackEvasionCheckIntervalMax; i++ {
		AllVillageSettings = append(AllVillageSettings, i)
	}
}
