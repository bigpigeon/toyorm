package toyorm

type Mode string

const (
	ModeDefault   = Mode("Default")
	ModeInsert    = Mode("Insert")
	ModeReplace   = Mode("Replace")
	ModeUpdate    = Mode("Update")
	ModeCondition = Mode("Condition")
	ModeScan      = Mode("Scan")
	ModeSelect    = Mode("Select")
)
