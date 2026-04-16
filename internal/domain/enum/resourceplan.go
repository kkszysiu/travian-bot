package enum

type ResourcePlan int

const (
	ResourcePlanAllResources ResourcePlan = 0
	ResourcePlanExcludeCrop  ResourcePlan = 1
	ResourcePlanOnlyCrop     ResourcePlan = 2
)

func (r ResourcePlan) String() string {
	switch r {
	case ResourcePlanAllResources:
		return "All Resources"
	case ResourcePlanExcludeCrop:
		return "Exclude Crop"
	case ResourcePlanOnlyCrop:
		return "Only Crop"
	default:
		return "Unknown"
	}
}
