package toyorm

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Result struct {
	Records        ModelRecords
	ActionFlow     []SqlAction
	RecordsActions map[int][]int
	Preload        map[string]*Result
	// in many-to-many model, have a middle model query need to record
	MiddleModelPreload map[string]*Result
}

func (r *Result) Err() error {
	var errStr string

	for _, action := range r.ActionFlow {
		switch t := action.sqlAction.(type) {
		case QueryAction:
			errs := ""
			for i, err := range t.Error {
				if err != nil {
					errs += fmt.Sprintf("\t[%d]%s\n", i, err)
				}
			}
			if errs != "" {
				errStr += fmt.Sprintf("%s args:%s errors(\n%s)\n", t.Exec.Query, t.Exec.JsonArgs(), errs)
			}
		case ExecAction:
			if t.Error != nil {
				errStr += fmt.Sprintf("%s args:%s errors(\n\t%s\n)\n", t.Exec.Query, t.Exec.JsonArgs(), t.Error)
			}
		}
	}
	for name, preload := range r.Preload {
		err := preload.Err()
		if err != nil {
			errStr += fmt.Sprintf("Preload %s errors:\n%s", name, err)
		}
	}
	for name, preload := range r.MiddleModelPreload {
		err := preload.Err()
		if err != nil {
			errStr += fmt.Sprintf("Middle Preload %s errors:\n%s", name, err)
		}
	}
	if errStr != "" {
		return errors.New(errStr)
	}
	return nil
}

func (r *Result) AddExecRecord(e ExecAction) {
	last := len(r.ActionFlow)
	r.ActionFlow = append(r.ActionFlow, SqlAction{e})
	for _, i := range e.affectData {
		r.RecordsActions[i] = append(r.RecordsActions[i], last)
	}
}

func (r *Result) AddQueryRecord(q QueryAction) {
	last := len(r.ActionFlow)
	r.ActionFlow = append(r.ActionFlow, SqlAction{q})
	for _, i := range q.affectData {
		r.RecordsActions[i] = append(r.RecordsActions[i], last)
	}
}

// TODO a report log
func (r *Result) Report() string {
	var flowStrList []string
	for _, action := range r.ActionFlow {
		flowStrList = append(flowStrList, fmt.Sprintf("%v %s", action.AffectData(), action.String()))
	}
	return strings.Join(flowStrList, "\n")
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
	AffectData() []int
}

type ExecAction struct {
	//Type   ResultType
	Exec       ExecValue
	Result     sql.Result
	affectData []int
	Error      error
}

func (r ExecAction) String() string {
	return fmt.Sprintf("%s %s error(%s)", r.Exec.Query, r.Exec.JsonArgs(), r.Error)
}

func (r ExecAction) Type() SqlActionType {
	return ResultActionExec
}

func (r ExecAction) AffectData() []int {
	return r.affectData
}

type QueryAction struct {
	Exec       ExecValue
	affectData []int
	Error      []error
}

func (r QueryAction) String() string {
	return fmt.Sprintf("%s %s error(%s)", r.Exec.Query, r.Exec.JsonArgs(), r.Error)
}

func (r QueryAction) Type() SqlActionType {
	return ResultActionQuery
}

func (r QueryAction) AffectData() []int {
	return r.affectData
}
