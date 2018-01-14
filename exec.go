package toyorm

import (
	"encoding/json"
)

//type Exec struct {
//	Value   []ExecValue
//	Preload map[*ModelField]*Exec
//}

type ExecValue struct {
	Query string
	Args  []interface{}
}

func (e ExecValue) JsonArgs() string {
	s, err := json.Marshal(e.Args)
	if err != nil {
		panic(err)
	}
	return string(s)
}
