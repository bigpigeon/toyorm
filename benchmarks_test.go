/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func getTestBenchmarkTable(time time.Time) *TestBenchmarkTable {
	return &TestBenchmarkTable{
		Key:     "key",
		Value:   "value",
		StrVal1: "it must be long",
		StrVal2: "it must be very long",
		StrVal3: "it must be very very long",
		StrVal4: "it must be very very very very long",
		StrVal5: "it must be very very very very long",
		IntVal1: 100,
		IntVal2: 200,
		IntVal3: 300,
		IntVal4: 400,
		IntVal5: 500,
		FloVal1: 100.0,
		FloVal2: 100.0,
		FloVal3: 100.0,
		FloVal4: 100.0,
		FloVal5: 100.0,
		TimVal1: time,
		TimVal2: time,
		TimVal3: time,
	}
}

func BenchmarkStandardInsert(b *testing.B) {
	brick := TestDB.Model(&TestBenchmarkTable{})
	createTableUnit(brick)(b)
	b.StartTimer()
	// get insert exec
	now := time.Now()
	result, err := brick.Insert(getTestBenchmarkTable(now))
	if err != nil {
		b.Error(err)
		b.FailNow()
	}
	if result.Err() != nil {
		b.Error(err)
		b.FailNow()
	}
	// insert sql action must be 1
	assert.Equal(b, len(result.ActionFlow), 1)
	exec := result.ActionFlow[0].(ExecAction)
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		result, err := TestDB.db.Exec(exec.Exec.Query(), exec.Exec.Args()...)
		result.LastInsertId()
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
	}
}

func BenchmarkInsert(b *testing.B) {
	brick := TestDB.Model(&TestBenchmarkTable{})
	createTableUnit(brick)(b)
	b.StartTimer()
	now := time.Now()
	for n := 0; n < b.N; n++ {
		result, err := brick.Insert(getTestBenchmarkTable(now))
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		if result.Err() != nil {
			b.Error(err)
			b.FailNow()
		}
	}
}

func BenchmarkStandardFind(b *testing.B) {
	brick := TestDB.Model(&TestBenchmarkTable{})
	createTableUnit(brick)(b)
	now := time.Now()
	// fill some data
	for i := 0; i < 1; i++ {
		result, err := brick.Insert(getTestBenchmarkTable(now))
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		if result.Err() != nil {
			b.Error(err)
			b.FailNow()
		}
	}
	b.StartTimer()
	// get find query
	result, err := brick.Find(&TestBenchmarkTable{})
	if err != nil {
		b.Error(err)
		b.FailNow()
	}
	if result.Err() != nil {
		b.Error(err)
		b.FailNow()
	}
	assert.Equal(b, len(result.ActionFlow), 1)
	query := result.ActionFlow[0].(QueryAction)
	for n := 0; n < b.N; n++ {
		var data []TestBenchmarkTable
		rows, err := TestDB.db.Query(query.Exec.Query(), query.Exec.Args()...)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		for rows.Next() {
			elem := TestBenchmarkTable{}
			err := rows.Scan(
				&elem.ID, &elem.Key, &elem.Value,
				&elem.StrVal1, &elem.StrVal2, &elem.StrVal3, &elem.StrVal4, &elem.StrVal5,
				&elem.IntVal1, &elem.IntVal2, &elem.IntVal3, &elem.IntVal4, &elem.IntVal5,
				&elem.FloVal1, &elem.FloVal2, &elem.FloVal3, &elem.FloVal4, &elem.FloVal5,
				&elem.TimVal1, &elem.TimVal2, &elem.TimVal3,
			)
			if err != nil {
				b.Error(err)
				b.FailNow()
			}
			data = append(data, elem)
		}
	}
}

func BenchmarkFind(b *testing.B) {
	brick := TestDB.Model(&TestBenchmarkTable{})
	createTableUnit(brick)(b)
	now := time.Now()
	// fill some data
	for i := 0; i < 1; i++ {
		result, err := brick.Insert(getTestBenchmarkTable(now))
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		if result.Err() != nil {
			b.Error(err)
			b.FailNow()
		}
	}
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		var data []TestBenchmarkTable
		result, err := brick.Find(&data)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		if result.Err() != nil {
			b.Error(err)
			b.FailNow()
		}
	}
}

func BenchmarkStandardUpdate(b *testing.B) {
	brick := TestDB.Model(&TestBenchmarkTable{})
	createTableUnit(brick)(b)
	now := time.Now()
	// fill some data
	for i := 0; i < 3; i++ {
		result, err := brick.Insert(getTestBenchmarkTable(now))
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		if result.Err() != nil {
			b.Error(err)
			b.FailNow()
		}
	}
	b.StartTimer()
	// get update exec
	result, err := brick.Update(&TestBenchmarkTable{
		Value: "value" + strconv.Itoa(1),
		Key:   "key2",
	})
	if err != nil {
		b.Error(err)
		b.FailNow()
	}
	if result.Err() != nil {
		b.Error(err)
		b.FailNow()
	}
	assert.Equal(b, len(result.ActionFlow), 1)
	exec := result.ActionFlow[0].(ExecAction)
	for n := 0; n < b.N; n++ {
		_, err := TestDB.db.Exec(exec.Exec.Query(), exec.Exec.Args()...)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
	}
}

func BenchmarkUpdate(b *testing.B) {
	brick := TestDB.Model(&TestBenchmarkTable{})
	createTableUnit(brick)(b)
	// fill some data
	for i := 0; i < 3; i++ {
		result, err := brick.Insert(&TestBenchmarkTable{
			Key:   "key" + strconv.Itoa(i),
			Value: "value",
		})
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		if result.Err() != nil {
			b.Error(err)
			b.FailNow()
		}
	}
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		result, err := brick.Update(&TestBenchmarkTable{
			Value: "value" + strconv.Itoa(n),
		})
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		if result.Err() != nil {
			b.Error(err)
			b.FailNow()
		}
	}
}
