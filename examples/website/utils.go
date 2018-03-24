/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package main

import (
	"database/sql/driver"
	"fmt"
	"github.com/bigpigeon/toyorm"
	"github.com/gin-gonic/gin"
)

func ResultProcess(result *toyorm.Result, err error, ctx *gin.Context) bool {
	if err != nil {
		fmt.Printf("context error: %s", err)
		ctx.AbortWithError(500, err)
		return false
	}

	if err := result.Err(); err != nil {
		fmt.Printf("sql error: \n%s\n", err)
		ctx.AbortWithError(500, err)
		return false
	}
	fmt.Printf("report: \n %s\n", result.Report())
	return true
}

// password need hide when print in Debug or Report
type Password string

func (p Password) MarshalJSON() ([]byte, error) {
	return []byte(`"***"`), nil
}

func (p *Password) UnmarshalJSON(data []byte) error {
	*p = Password(data)
	return nil
}

func (p Password) Value() (driver.Value, error) {
	return string(p), nil
}
