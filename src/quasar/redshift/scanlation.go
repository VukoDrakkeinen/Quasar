package redshift

type ScanlationStatus int

const (
	ScanlationStatusInvalid ScanlationStatus = iota
	ScanlationComplete
	ScanlationOngoing
	ScanlationOnHiatus
	ScanlationDropped
	ScanlationInDesperateNeedOfMoreStaff
)
