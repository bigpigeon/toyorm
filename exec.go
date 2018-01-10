package toyorm

//type Exec struct {
//	Value   []ExecValue
//	Preload map[*ModelField]*Exec
//}

type ExecValue struct {
	Query string
	Args  []interface{}
}
