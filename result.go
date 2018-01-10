package toyorm

import (
	"database/sql"
	"fmt"
)

//const (
//	FindResult        = ResultType("Find")
//	InsertResult      = ResultType("Insert")
//	UpdateResult      = ResultType("Update")
//	ReplaceResult     = ResultType("Replace")
//	DropResult        = ResultType("Drop")
//	CreateTableResult = ResultType("CreateTable")
//	DropTableResult   = ResultType("DropTable")
//)

type Result struct {
	Records        ModelRecords
	ActionFlow     []SqlAction
	RecordsActions map[int][]SqlAction
	Preload        map[*ModelField]*Result
	// in many-to-many model, have a middle model query need to record
	MiddleModelPreload map[*ModelField]*Result
}

func (r *Result) AddExecRecord(e ExecAction, index ...int) {
	r.ActionFlow = append(r.ActionFlow, SqlAction{e})
	for _, i := range index {
		r.RecordsActions[i] = append(r.RecordsActions[i], SqlAction{e})
	}
}

// TODO a report log
func (r *Result) Report() string {
	return ""
}

func (r *Result) AddQueryRecord(q QueryAction, index ...int) {
	r.ActionFlow = append(r.ActionFlow, SqlAction{q})
	for _, i := range index {
		r.RecordsActions[i] = append(r.RecordsActions[i], SqlAction{q})
	}
}

type SqlActionType int

const (
	ResultActionExec = SqlActionType(iota)
	ResultActionQuery
)

type SqlAction struct {
	sqlAction
}

func (r SqlAction) ToExec() ExecAction {
	return r.sqlAction.(ExecAction)
}

func (r SqlAction) ToQuery() QueryAction {
	return r.sqlAction.(QueryAction)
}

type sqlAction interface {
	String() string
	Type() SqlActionType
}

type ExecAction struct {
	//Type   ResultType
	Exec   ExecValue
	Result sql.Result
	Error  error
}

func (r ExecAction) String() string {
	return fmt.Sprintf("%s %s error(%s)", r.Exec.Query, r.Exec.Args, r.Error)
}

func (r ExecAction) Type() SqlActionType {
	return ResultActionExec
}

type QueryAction struct {
	Exec  ExecValue
	Error []error
}

func (r QueryAction) String() string {
	return fmt.Sprintf("%s %s error(%s)", r.Exec.Query, r.Exec.Args, r.Error)
}

func (r QueryAction) Type() SqlActionType {
	return ResultActionQuery
}
