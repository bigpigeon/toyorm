package toyorm

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
)

type Pair struct {
	Main int
	Sub  int
}

type Result struct {
	Records        ModelRecords
	ActionFlow     []SqlAction
	RecordsActions map[int][]int
	Preload        map[string]*Result
	// container is simple object
	SimpleRelation map[string]map[int]int
	// container is slice object
	MultipleRelation map[string]map[int]Pair

	// in many-to-many model, have a middle model query need to record
	MiddleModelPreload map[string]*Result
}

func (r *Result) Err() error {
	var errStr string

	for _, action := range r.ActionFlow {
		switch t := action.(type) {
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

func (r *Result) AddRecord(q SqlAction) {
	last := len(r.ActionFlow)
	r.ActionFlow = append(r.ActionFlow, q)
	for _, i := range q.AffectData() {
		r.RecordsActions[i] = append(r.RecordsActions[i], last)
	}
}

type AffectNode struct {
	Val    int
	Ignore bool
	Next   *AffectNode
}

func NewAffectNode(val int, next *AffectNode) *AffectNode {
	return &AffectNode{val, false, next}
}

func IgnoreAffectNode(next *AffectNode) *AffectNode {
	return &AffectNode{0, true, next}
}

type ReportData struct {
	Depth      int
	AffectData []*AffectNode
	Str        string
}

func (r *Result) report() (reportData []ReportData) {
	for _, action := range r.ActionFlow {
		data := ReportData{0, nil, action.String()}
		for _, d := range action.AffectData() {
			data.AffectData = append(data.AffectData, NewAffectNode(d, nil))
		}
		reportData = append(reportData, data)
	}
	if len(r.Preload) != 0 {
		for name, preloadResults := range r.Preload {
			reportData = append(reportData, ReportData{1, nil, "preload " + name})
			for _, oldData := range preloadResults.report() {
				var data ReportData
				if oldData.AffectData == nil {
					data = ReportData{oldData.Depth + 1, nil, oldData.Str}
				} else {
					data = ReportData{oldData.Depth + 1, make([]*AffectNode, len(oldData.AffectData)), oldData.Str}
					if relation := r.SimpleRelation[name]; relation != nil {
						for i, node := range oldData.AffectData {
							a := NewAffectNode(relation[node.Val], IgnoreAffectNode(node.Next))
							data.AffectData[i] = a
						}
					} else if relation := r.MultipleRelation[name]; relation != nil {
						for i, node := range oldData.AffectData {
							pair := relation[node.Val]
							data.AffectData[i] = NewAffectNode(pair.Main, NewAffectNode(pair.Sub, node.Next))
						}
					} else {
						panic("relation map not found")
					}
				}
				reportData = append(reportData, data)
			}
		}
	}
	return reportData
}

// Rename to Log
func (r *Result) Report() string {
	reportData := r.report()
	var buf bytes.Buffer
	for _, report := range reportData {
		for i := 0; i < report.Depth; i++ {
			buf.WriteByte('\t')
		}
		if report.AffectData == nil {
			buf.WriteString(report.Str + "\n")
		} else {
			buf.WriteByte('[')
			for _, affect := range report.AffectData {
				if affect != nil {
					if affect.Ignore {
						buf.WriteString("-")
					} else {
						buf.WriteString(fmt.Sprintf("%d", affect.Val))
					}
					affect = affect.Next
				}
				for affect != nil {
					if affect.Ignore {
						buf.WriteString("-")
					} else {
						buf.WriteString(fmt.Sprintf("-%d", affect.Val))
					}
					affect = affect.Next
				}

				buf.WriteString(", ")
			}
			buf.WriteByte(']')
			buf.WriteString(" " + report.Str)
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

type SqlActionType int

const (
	ResultActionExec = SqlActionType(iota)
	ResultActionQuery
)

type SqlAction interface {
	String() string
	Type() SqlActionType
	AffectData() []int
	SetAffectData([]int)
}

type ExecAction struct {
	//Type   ResultType
	Exec       ExecValue
	Result     sql.Result
	affectData []int
	Error      error
}

func (r ExecAction) String() string {
	if r.Error != nil {
		return fmt.Sprintf("%s args:%s error(%v)", r.Exec.Query, r.Exec.JsonArgs(), r.Error)
	}
	return fmt.Sprintf("%s args:%s", r.Exec.Query, r.Exec.JsonArgs())
}

func (r ExecAction) Type() SqlActionType {
	return ResultActionExec
}

func (r ExecAction) AffectData() []int {
	return r.affectData
}

func (r ExecAction) SetAffectData(d []int) {
	r.affectData = d
}

type QueryAction struct {
	Exec       ExecValue
	affectData []int
	Error      []error
}

func (r QueryAction) String() string {
	var errors []error
	for _, err := range r.Error {
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) != 0 {
		return fmt.Sprintf("%s args:%s error(%v)", r.Exec.Query, r.Exec.JsonArgs(), errors)
	}
	return fmt.Sprintf("%s args:%s", r.Exec.Query, r.Exec.JsonArgs())
}

func (r QueryAction) Type() SqlActionType {
	return ResultActionQuery
}

func (r QueryAction) AffectData() []int {
	return r.affectData
}

func (r QueryAction) SetAffectData(d []int) {
	r.affectData = d
}
