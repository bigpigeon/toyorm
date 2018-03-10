package toyorm

import (
	"database/sql"
	"fmt"
)

type CollectionResult struct {
	Records        ModelRecords
	ActionFlow     []SqlAction
	RecordsActions map[int][]int
	Preload        map[string]*CollectionResult
	// container is simple object
	SimpleRelation map[string]map[int]int
	// container is slice object
	MultipleRelation map[string]map[int]Pair

	// in many-to-many model, have a middle model query need to record
	MiddleModelPreload map[string]*CollectionResult
}

func (r *CollectionResult) AddRecord(e SqlAction) {
	last := len(r.ActionFlow)
	r.ActionFlow = append(r.ActionFlow, e)
	for _, i := range e.AffectData() {
		r.RecordsActions[i] = append(r.RecordsActions[i], last)
	}
}

type CollectionExecAction struct {
	//Type   ResultType
	Exec       ExecValue
	Result     sql.Result
	dbIndex    int
	affectData []int
	Error      error
}

func (r CollectionExecAction) String() string {
	if r.Error != nil {
		return fmt.Sprintf("[%d]%s args:%s error(%v)", r.dbIndex, r.Exec.Query, r.Exec.JsonArgs(), r.Error)
	}
	return fmt.Sprintf("[%d]%s args:%s", r.dbIndex, r.Exec.Query, r.Exec.JsonArgs())
}

func (r CollectionExecAction) Type() SqlActionType {
	return ResultActionExec
}

func (r CollectionExecAction) AffectData() []int {
	return r.affectData
}

func (r CollectionExecAction) SetAffectData(d []int) {
	r.affectData = d
}

type CollectionQueryAction struct {
	Exec       ExecValue
	affectData []int
	Error      []error
	dbIndex    int
}

func (r CollectionQueryAction) String() string {
	var errors []error
	for _, err := range r.Error {
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) != 0 {
		return fmt.Sprintf("[%d]%s args:%s error(%v)", r.dbIndex, r.Exec.Query, r.Exec.JsonArgs(), errors)
	}
	return fmt.Sprintf("[%d]%s args:%s", r.dbIndex, r.Exec.Query, r.Exec.JsonArgs())
}

func (r CollectionQueryAction) Type() SqlActionType {
	return ResultActionQuery
}

func (r CollectionQueryAction) AffectData() []int {
	return r.affectData
}

func (r CollectionQueryAction) SetAffectData(d []int) {
	r.affectData = d
}
