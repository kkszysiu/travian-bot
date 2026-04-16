package enum

type AccountSetting int

const (
	AccountSettingClickDelayMin                AccountSetting = 1
	AccountSettingClickDelayMax                AccountSetting = 2
	AccountSettingTaskDelayMin                 AccountSetting = 3
	AccountSettingTaskDelayMax                 AccountSetting = 4
	AccountSettingEnableAutoLoadVillageBuilding AccountSetting = 5
	AccountSettingUseStartAllButton            AccountSetting = 6
	AccountSettingFarmIntervalMin              AccountSetting = 7
	AccountSettingFarmIntervalMax              AccountSetting = 8
	AccountSettingTribe                        AccountSetting = 9
	AccountSettingWorkTimeMin                  AccountSetting = 10
	AccountSettingWorkTimeMax                  AccountSetting = 11
	AccountSettingSleepTimeMin                 AccountSetting = 12
	AccountSettingSleepTimeMax                 AccountSetting = 13
	AccountSettingHeadlessChrome               AccountSetting = 14
	AccountSettingEnableAutoStartAdventure     AccountSetting = 15
	AccountSettingWorkStartHour               AccountSetting = 16
	AccountSettingWorkStartMinute             AccountSetting = 17
	AccountSettingWorkEndHour                 AccountSetting = 18
	AccountSettingWorkEndMinute               AccountSetting = 19
	AccountSettingSleepRandomMinute           AccountSetting = 20
)

func (a AccountSetting) String() string {
	switch a {
	case AccountSettingClickDelayMin:
		return "ClickDelayMin"
	case AccountSettingClickDelayMax:
		return "ClickDelayMax"
	case AccountSettingTaskDelayMin:
		return "TaskDelayMin"
	case AccountSettingTaskDelayMax:
		return "TaskDelayMax"
	case AccountSettingEnableAutoLoadVillageBuilding:
		return "EnableAutoLoadVillageBuilding"
	case AccountSettingUseStartAllButton:
		return "UseStartAllButton"
	case AccountSettingFarmIntervalMin:
		return "FarmIntervalMin"
	case AccountSettingFarmIntervalMax:
		return "FarmIntervalMax"
	case AccountSettingTribe:
		return "Tribe"
	case AccountSettingWorkTimeMin:
		return "WorkTimeMin"
	case AccountSettingWorkTimeMax:
		return "WorkTimeMax"
	case AccountSettingSleepTimeMin:
		return "SleepTimeMin"
	case AccountSettingSleepTimeMax:
		return "SleepTimeMax"
	case AccountSettingHeadlessChrome:
		return "HeadlessChrome"
	case AccountSettingEnableAutoStartAdventure:
		return "EnableAutoStartAdventure"
	case AccountSettingWorkStartHour:
		return "WorkStartHour"
	case AccountSettingWorkStartMinute:
		return "WorkStartMinute"
	case AccountSettingWorkEndHour:
		return "WorkEndHour"
	case AccountSettingWorkEndMinute:
		return "WorkEndMinute"
	case AccountSettingSleepRandomMinute:
		return "SleepRandomMinute"
	default:
		return "Unknown"
	}
}

// DefaultAccountSettings maps each account setting to its default value.
var DefaultAccountSettings = map[AccountSetting]int{
	AccountSettingClickDelayMin:                500,
	AccountSettingClickDelayMax:                900,
	AccountSettingTaskDelayMin:                 1000,
	AccountSettingTaskDelayMax:                 1500,
	AccountSettingEnableAutoLoadVillageBuilding: 1,
	AccountSettingUseStartAllButton:            0,
	AccountSettingFarmIntervalMin:              540,
	AccountSettingFarmIntervalMax:              660,
	AccountSettingTribe:                        0,
	AccountSettingSleepTimeMin:                 480,
	AccountSettingSleepTimeMax:                 600,
	AccountSettingHeadlessChrome:               0,
	AccountSettingEnableAutoStartAdventure:     0,
	AccountSettingWorkStartHour:               6,
	AccountSettingWorkStartMinute:             0,
	AccountSettingWorkEndHour:                 22,
	AccountSettingWorkEndMinute:               0,
	AccountSettingSleepRandomMinute:           60,
}

// AllAccountSettings returns all account setting keys in order.
var AllAccountSettings = []AccountSetting{
	AccountSettingClickDelayMin, AccountSettingClickDelayMax,
	AccountSettingTaskDelayMin, AccountSettingTaskDelayMax,
	AccountSettingEnableAutoLoadVillageBuilding,
	AccountSettingUseStartAllButton,
	AccountSettingFarmIntervalMin, AccountSettingFarmIntervalMax,
	AccountSettingTribe,
	AccountSettingSleepTimeMin, AccountSettingSleepTimeMax,
	AccountSettingHeadlessChrome,
	AccountSettingEnableAutoStartAdventure,
	AccountSettingWorkStartHour, AccountSettingWorkStartMinute,
	AccountSettingWorkEndHour, AccountSettingWorkEndMinute,
	AccountSettingSleepRandomMinute,
}
