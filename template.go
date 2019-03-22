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

func findWord(s string) int {
	if len(s) > 0 {
		if s[0] == '-' || s[0] == '_' {
			return 0
		}
	}
	linkChar := false
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			linkChar = false
		} else if c == '-' || c == '_' {
			linkChar = true
		} else {
			break
		}
	}
	if linkChar == true {
		i--
	}
	return i
}

func findLeftParenthesis(s string, start int, parenthesis byte) int {
	for i := start; i < len(s); i++ {
		switch s[i] {
		case parenthesis:
			return i
		case '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0:
			continue
		default:
			return -1
		}
	}
	return -1
}

func findRightParenthesis(s string, start int, parenthesisPair [2]byte) int {
	// count embed right parenthesis and skip corresponding left parenthesis
	rbc := 1
	for i := start; i < len(s); i++ {
		switch s[i] {
		case parenthesisPair[0]:
			rbc++
		case parenthesisPair[1]:
			rbc--
			if rbc == 0 {
				return i
			}
		}
	}
	return -1
}

func findConditionPlaceholder(s string) (string, string, string, int) {
	// find first '('
	if len(s) < 1 || s[0] != '(' {
		return "", "", "", 0
	}
	// frb,       slb,        srb,        tlb,       trb representing:
	// first ')', second '(', second ')', third '(', third ')'
	frb := findRightParenthesis(s, 1, [2]byte{'(', ')'})
	if frb == -1 {
		return "", "", "", 1
	}
	slb := findLeftParenthesis(s, frb+1, '{')
	if slb == -1 {
		return "", "", "", frb + 1
	}
	srb := findRightParenthesis(s, slb+1, [2]byte{'{', '}'})
	if srb == -1 {
		return "", "", "", slb + 1
	}
	tlb := findLeftParenthesis(s, srb+1, '{')
	if tlb == -1 {
		return s[1:frb], s[slb+1 : srb], "", srb + 1
	}
	trb := findRightParenthesis(s, tlb+1, [2]byte{'{', '}'})
	if trb == -1 {
		return "", "", "", tlb + 1
	}
	return s[1:frb], s[slb+1 : srb], s[tlb+1 : trb], trb + 1
}

func checkTempCondition(cond string, execs map[string]BasicExec) (bool, error) {
	expr, err := conditionToReversePolandExpr(cond, execs)
	if err != nil {
		return false, err
	}
	var stack CharStack
	for _, e := range expr {
		switch e {
		case '|':
			b1, b2 := stack.Pop(), stack.Pop()
			if b1 == '1' || b2 == '1' {
				stack.Push('1')
			} else {
				stack.Push('0')
			}
		case '&':
			b1, b2 := stack.Pop(), stack.Pop()
			if b1 == '1' && b2 == '1' {
				stack.Push('1')
			} else {
				stack.Push('0')
			}
		default:
			stack = append(stack, e)
		}
	}
	if stack[0] == '1' {
		return true, nil
	} else if stack[0] == '0' {
		return false, nil
	} else {
		return false, ErrInvalidConditionWord{cond, cond}
	}
}

func getTemplateExecAndArgsIndex(exec BasicExec, execs map[string]BasicExec) (BasicExec, int, error) {
	buff := bytes.Buffer{}
	var pre, i int
	var args []interface{}
	isEscaping := false

	qNum := 0
	for i < len(exec.query) {
		switch exec.query[i] {
		case '$':
			if isEscaping == false {
				buff.WriteString(exec.query[pre:i])
				i++
				if endOf := findWord(exec.query[i:]); endOf != 0 {
					word := exec.query[i : i+endOf]
					if match, ok := execs[word]; ok {
						buff.WriteString(match.query)
						args = append(args, match.args...)
					} else {
						//return BasicExec{}, 0, ErrTemplateExecInvalidWord{"$" + word}
					}
					i += endOf
					pre = i
				} else if cond, code1, code2, endOf := findConditionPlaceholder(exec.query[i:]); cond != "" {
					yes, err := checkTempCondition(cond, execs)
					if err != nil {
						return BasicExec{}, 0, err
					}
					if yes {
						retExec, retQNum, err := getTemplateExecAndArgsIndex(BasicExec{
							query: code1,
							args:  exec.args[qNum:],
						}, execs)
						if err != nil {
							return BasicExec{}, 0, err
						}
						_, err = buff.WriteString(retExec.query)
						if err != nil {
							return BasicExec{}, 0, err
						}
						args = append(args, retExec.args...)
						qNum += retQNum
					} else if code2 != "" {
						retExec, retQNum, err := getTemplateExecAndArgsIndex(BasicExec{
							query: code2,
							args:  exec.args[qNum:],
						}, execs)
						if err != nil {
							return BasicExec{}, 0, err
						}
						_, err = buff.WriteString(retExec.query)
						if err != nil {
							return BasicExec{}, 0, err
						}
						args = append(args, retExec.args...)
						qNum += retQNum
					}
					i += endOf
					pre = i
				} else {
					return BasicExec{}, 0, ErrTemplateExecInvalidWord{"$" + exec.query[i:i+endOf]}
				}

			} else {
				_, err := buff.WriteString(exec.query[pre : i-1])
				if err != nil {
					return BasicExec{}, 0, err
				}
				err = buff.WriteByte(exec.query[i])
				if err != nil {
					return BasicExec{}, 0, err
				}
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

	buff.WriteString(exec.query[pre:i])
	return BasicExec{buff.String(), args}, qNum, nil
}

// placeholder format:
// $(name){code} ok condition placeholder means if name is not zero render code
// $(name){code}{other code} ok condition placeholder means if name is zero render code else render other code
// $Name   ok
// $name   ok
// $1Name  ok
// $User-Name  ok
// $user_name  ok
// $User-  no, the placeholder are interception as $User
// $user_name_ no, the placeholder are interception as $user_name
// $-   error, placeholder is null
func GetTemplateExec(exec BasicExec, execs map[string]BasicExec) (*BasicExec, error) {
	newExec, qNum, err := getTemplateExecAndArgsIndex(exec, execs)
	if err != nil {
		return nil, err
	}
	newExec.args = append(newExec.args, exec.args[qNum:]...)
	return &newExec, nil
}

type TemplateBasic struct {
	Temp  BasicExec
	Model *Model
	Alias string
	Quote string
}

func (b *TemplateBasic) DefaultExecs() map[string]BasicExec {
	modelDef := b.Quote + b.Model.Name + b.Quote
	if b.Alias != "" {
		modelDef = b.Quote + b.Model.Name + b.Quote + " as " + b.Alias
	}
	result := map[string]BasicExec{
		"ModelName": {b.Model.Name, nil},
		"ModelDef":  {modelDef, nil},
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
	Conditions   BasicExec
	Cas          BasicExec
}

func (exec *SaveTemplate) Render() (*BasicExec, error) {
	execs := exec.DefaultExecs()
	execs["Columns"] = exec.Columns
	execs["Values"] = exec.Values
	execs["UpdateValues"] = exec.UpdateValues
	execs["Cas"] = exec.Cas
	execs["Conditions"] = exec.Conditions
	return GetTemplateExec(exec.Temp, execs)
}

type TempMode int8

const (
	TempDefault TempMode = iota
	TempInsert
	TempSave
	TempUSave
	TempUpdate
	TempFind
	TempCreateTable
	TempDropTable
	TempEnd
)
