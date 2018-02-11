package toyorm

type Mode int8

const (
	ModeDefault Mode = iota
	ModeInsert
	ModeReplace
	ModeUpdate
	ModeScan
	ModeSelect
	ModeCondition
	ModePreload
	ModeEnd
)
