/*
 * Copyright 2019. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"bytes"
	"fmt"
)

type Template interface {
	Render() BasicExec
}

func GetTemplateExec(exec BasicExec, execs map[string]BasicExec) (*BasicExec, error) {
	buff := bytes.Buffer{}
	var pre, i int
	var args []interface{}
	isEscaping := false
	qNum := 0
	query := exec.query
	for i < len(query) {
		switch query[i] {
		case '$':
			if isEscaping == false {
				buff.WriteString(query[pre:i])
				i++
				pre = i
				end := i
				for i < len(query) {
					c := query[i]
					i++
					if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
						end = i
					} else if c == '-' || c == '_' {

					} else {
						break
					}
				}
				word := query[pre:end]
				if word == "" {
					return nil, ErrTemplateExecInvalidWord{"$"}
				} else if match, ok := execs[word]; ok {
					buff.WriteString(match.query)
					args = append(args, match.args...)
				} else {
					// allow $word with nil
					//return nil, ErrTemplateExecInvalidWord{"$" + word}
				}

				pre, i = end, end
			} else {
				buff.WriteString(query[pre : i-1])
				buff.WriteByte(query[i])

				pre, i = i+1, i+1
			}
		case '?':
			args = append(args, exec.args[qNum])
			qNum++
			i++
		case '\\':
			i++
			isEscaping = true
			continue
		default:
			i++
		}
		isEscaping = false
	}
	args = append(args, exec.args[qNum:]...)

	buff.WriteString(query[pre:i])
	return &BasicExec{buff.String(), args}, nil
}

type TemplateBasic struct {
	Temp  BasicExec
	Model *Model
}

func (b *TemplateBasic) DefaultExecs() map[string]BasicExec {
	result := map[string]BasicExec{
		"ModelName": {b.Model.Name, nil},
	}
	for _, field := range b.Model.GetSqlFields() {
		// add field name placeholder exec
		result["FN-"+field.Name()] = BasicExec{field.Column(), nil}
		// add field offset placeholder exec
		result[fmt.Sprintf("0x%x", field.Offset())] = BasicExec{field.Column(), nil}
	}
	return result
}

type SaveTemplate struct {
	TemplateBasic
	Columns      BasicExec
	Values       BasicExec
	UpdateValues BasicExec
	Condition    BasicExec
	Cas          BasicExec
}

func (exec *SaveTemplate) Render() (*BasicExec, error) {
	execs := exec.DefaultExecs()
	execs["Columns"] = exec.Columns
	execs["Values"] = exec.Values
	execs["UpdateValues"] = exec.UpdateValues
	execs["Cas"] = exec.Cas
	execs["Condition"] = exec.Condition
	return GetTemplateExec(exec.Temp, execs)
}

type TempMode int8

const (
	TempDefault TempMode = iota
	TempInsert
	TempSave
	TempUpdate
	TempFind
	TempCreateTable
	TempDropTable
	TempEnd
)
