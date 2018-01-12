package toyorm

import (
	"database/sql"
	"fmt"
)

type Result struct {
	Records        ModelRecords
	ActionFlow     []SqlAction
	RecordsActions map[int][]SqlAction
	Preload        map[*ModelField]*Result
	// in many-to-many model, have a middle model query need to record
	MiddleModelPreload map[*ModelField]*Result
}

func (r *Result) AllErrors() []error {
	var errs []error
	for _, action := range r.ActionFlow {
		switch t := action.sqlAction.(type) {
		case QueryAction:
			errs = append(errs, t.Error...)
		case ExecAction:
			if t.Error != nil {
				errs = append(errs, t.Error)
			}
		}
	}
	for _, preload := range r.Preload {
		errs = append(errs, preload.AllErrors()...)
	}
	for _, middlePreload := range r.MiddleModelPreload {
		errs = append(errs, middlePreload.AllErrors()...)
	}
	return errs
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
