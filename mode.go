package toyorm

type Mode string

const (
	ModeDefault   Mode = "Default"
	ModeInsert         = "Insert"
	ModeReplace        = "Replace"
	ModeUpdate         = "Update"
	ModeCondition      = "Condition"
	ModeScan           = "Scan"
	ModeSelect         = "Select"
)
