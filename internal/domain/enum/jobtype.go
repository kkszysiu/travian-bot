package enum

type JobType int

const (
	JobTypeNormalBuild   JobType = 0
	JobTypeResourceBuild JobType = 1
)

func (j JobType) String() string {
	switch j {
	case JobTypeNormalBuild:
		return "Normal Build"
	case JobTypeResourceBuild:
		return "Resource Build"
	default:
		return "Unknown"
	}
}
