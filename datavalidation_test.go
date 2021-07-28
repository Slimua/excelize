// Copyright 2016 - 2021 The excelize Authors. All rights reserved. Use of
// this source code is governed by a BSD-style license that can be found in
// the LICENSE file.
//
// Package excelize providing a set of functions that allow you to write to
// and read from XLSX files. Support reads and writes XLSX file generated by
// Microsoft Excel™ 2007 and later. Support save file without losing original
// charts of XLSX. This library needs Go version 1.15 or later.

package excelize

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataValidation(t *testing.T) {
	resultFile := filepath.Join("test", "TestDataValidation.xlsx")

	f := NewFile()

	dvRange := NewDataValidation(true)
	dvRange.Sqref = "A1:B2"
	assert.NoError(t, dvRange.SetRange(10, 20, DataValidationTypeWhole, DataValidationOperatorBetween))
	dvRange.SetError(DataValidationErrorStyleStop, "error title", "error body")
	dvRange.SetError(DataValidationErrorStyleWarning, "error title", "error body")
	dvRange.SetError(DataValidationErrorStyleInformation, "error title", "error body")
	assert.NoError(t, f.AddDataValidation("Sheet1", dvRange))
	assert.NoError(t, f.SaveAs(resultFile))

	dvRange = NewDataValidation(true)
	dvRange.Sqref = "A3:B4"
	assert.NoError(t, dvRange.SetRange(10, 20, DataValidationTypeWhole, DataValidationOperatorGreaterThan))
	dvRange.SetInput("input title", "input body")
	assert.NoError(t, f.AddDataValidation("Sheet1", dvRange))
	assert.NoError(t, f.SaveAs(resultFile))

	dvRange = NewDataValidation(true)
	dvRange.Sqref = "A5:B6"
	for _, listValid := range [][]string{
		{"1", "2", "3"},
		{strings.Repeat("&", 255)},
		{strings.Repeat("\u4E00", 255)},
		{strings.Repeat("\U0001F600", 100), strings.Repeat("\u4E01", 50), "<&>"},
		{`A<`, `B>`, `C"`, "D\t", `E'`, `F`},
	} {
		dvRange.Formula1 = ""
		assert.NoError(t, dvRange.SetDropList(listValid),
			"SetDropList failed for valid input %v", listValid)
		assert.NotEqual(t, "", dvRange.Formula1,
			"Formula1 should not be empty for valid input %v", listValid)
	}
	assert.Equal(t, dvRange.Formula1, `<formula1>"A&lt;,B&gt;,C"",D	,E',F"</formula1>`)
	assert.NoError(t, f.AddDataValidation("Sheet1", dvRange))
	assert.NoError(t, f.SaveAs(resultFile))
}

func TestDataValidationError(t *testing.T) {
	resultFile := filepath.Join("test", "TestDataValidationError.xlsx")

	f := NewFile()
	assert.NoError(t, f.SetCellStr("Sheet1", "E1", "E1"))
	assert.NoError(t, f.SetCellStr("Sheet1", "E2", "E2"))
	assert.NoError(t, f.SetCellStr("Sheet1", "E3", "E3"))

	dvRange := NewDataValidation(true)
	dvRange.SetSqref("A7:B8")
	dvRange.SetSqref("A7:B8")
	assert.NoError(t, dvRange.SetSqrefDropList("$E$1:$E$3", true))

	err := dvRange.SetSqrefDropList("$E$1:$E$3", false)
	assert.EqualError(t, err, "cross-sheet sqref cell are not supported")

	assert.NoError(t, f.AddDataValidation("Sheet1", dvRange))
	assert.NoError(t, f.SaveAs(resultFile))

	dvRange = NewDataValidation(true)

	assert.NoError(t, dvRange.SetRange(10, 20, DataValidationTypeWhole, DataValidationOperatorGreaterThan))
	dvRange.SetSqref("A9:B10")

	assert.NoError(t, f.AddDataValidation("Sheet1", dvRange))
	assert.NoError(t, f.SaveAs(resultFile))

	// Test width invalid data validation formula.
	prevFormula1 := dvRange.Formula1
	for _, listTooLong := range [][]string{
		make([]string, 257),
		{strings.Repeat("s", 256)},
		{strings.Repeat("\u4E00", 256)},
		{strings.Repeat("\U0001F600", 128)},
		{strings.Repeat("\U0001F600", 127), "s"},
	} {
		err = dvRange.SetDropList(listTooLong)
		assert.Equal(t, prevFormula1, dvRange.Formula1,
			"Formula1 should be unchanged for invalid input %v", listTooLong)
		assert.EqualError(t, err, "data validation must be 0-255 characters")
	}

	assert.NoError(t, f.AddDataValidation("Sheet1", dvRange))
	assert.NoError(t, f.SaveAs(resultFile))

	dvRange.Formula1 = strings.Repeat("s", dataValidationFormulaStrLen+22)
	assert.EqualError(t, dvRange.SetRange(10, 20, DataValidationTypeWhole, DataValidationOperatorGreaterThan), "data validation must be 0-255 characters")

	// Test add data validation on no exists worksheet.
	f = NewFile()
	assert.EqualError(t, f.AddDataValidation("SheetN", nil), "sheet SheetN is not exist")
}

func TestDeleteDataValidation(t *testing.T) {
	f := NewFile()
	assert.NoError(t, f.DeleteDataValidation("Sheet1", "A1:B2"))

	dvRange := NewDataValidation(true)
	dvRange.Sqref = "A1:B2"
	assert.NoError(t, dvRange.SetRange(10, 20, DataValidationTypeWhole, DataValidationOperatorBetween))
	dvRange.SetInput("input title", "input body")
	assert.NoError(t, f.AddDataValidation("Sheet1", dvRange))

	assert.NoError(t, f.DeleteDataValidation("Sheet1", "A1:B2"))
	assert.NoError(t, f.SaveAs(filepath.Join("test", "TestDeleteDataValidation.xlsx")))

	// Test delete data validation on no exists worksheet.
	assert.EqualError(t, f.DeleteDataValidation("SheetN", "A1:B2"), "sheet SheetN is not exist")
}
