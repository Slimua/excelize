// Copyright 2016 - 2022 The excelize Authors. All rights reserved. Use of
// this source code is governed by a BSD-style license that can be found in
// the LICENSE file.
//
// Package excelize providing a set of functions that allow you to write to and
// read from XLAM / XLSM / XLSX / XLTM / XLTX files. Supports reading and
// writing spreadsheet documents generated by Microsoft Excel™ 2007 and later.
// Supports complex components by high compatibility, and provided streaming
// API for generating or reading data from a worksheet with huge amounts of
// data. This library needs Go version 1.15 or later.

package excelize

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf16"
)

// DataValidationType defined the type of data validation.
type DataValidationType int

// Data validation types.
const (
	_DataValidationType = iota
	typeNone            // inline use
	DataValidationTypeCustom
	DataValidationTypeDate
	DataValidationTypeDecimal
	typeList // inline use
	DataValidationTypeTextLength
	DataValidationTypeTime
	// DataValidationTypeWhole Integer
	DataValidationTypeWhole
)

// DataValidationErrorStyle defined the style of data validation error alert.
type DataValidationErrorStyle int

// Data validation error styles.
const (
	_ DataValidationErrorStyle = iota
	DataValidationErrorStyleStop
	DataValidationErrorStyleWarning
	DataValidationErrorStyleInformation
)

// Data validation error styles.
const (
	styleStop        = "stop"
	styleWarning     = "warning"
	styleInformation = "information"
)

// DataValidationOperator operator enum.
type DataValidationOperator int

// Data validation operators.
const (
	_DataValidationOperator = iota
	DataValidationOperatorBetween
	DataValidationOperatorEqual
	DataValidationOperatorGreaterThan
	DataValidationOperatorGreaterThanOrEqual
	DataValidationOperatorLessThan
	DataValidationOperatorLessThanOrEqual
	DataValidationOperatorNotBetween
	DataValidationOperatorNotEqual
)

// formulaEscaper mimics the Excel escaping rules for data validation,
// which converts `"` to `""` instead of `&quot;`.
var formulaEscaper = strings.NewReplacer(
	`&`, `&amp;`,
	`<`, `&lt;`,
	`>`, `&gt;`,
	`"`, `""`,
)

// NewDataValidation return data validation struct.
func NewDataValidation(allowBlank bool) *DataValidation {
	return &DataValidation{
		AllowBlank:       allowBlank,
		ShowErrorMessage: false,
		ShowInputMessage: false,
	}
}

// SetError set error notice.
func (dd *DataValidation) SetError(style DataValidationErrorStyle, title, msg string) {
	dd.Error = &msg
	dd.ErrorTitle = &title
	strStyle := styleStop
	switch style {
	case DataValidationErrorStyleStop:
		strStyle = styleStop
	case DataValidationErrorStyleWarning:
		strStyle = styleWarning
	case DataValidationErrorStyleInformation:
		strStyle = styleInformation

	}
	dd.ShowErrorMessage = true
	dd.ErrorStyle = &strStyle
}

// SetInput set prompt notice.
func (dd *DataValidation) SetInput(title, msg string) {
	dd.ShowInputMessage = true
	dd.PromptTitle = &title
	dd.Prompt = &msg
}

// SetDropList data validation list.
func (dd *DataValidation) SetDropList(keys []string) error {
	formula := strings.Join(keys, ",")
	if MaxFieldLength < len(utf16.Encode([]rune(formula))) {
		return ErrDataValidationFormulaLength
	}
	dd.Formula1 = fmt.Sprintf(`<formula1>"%s"</formula1>`, formulaEscaper.Replace(formula))
	dd.Type = convDataValidationType(typeList)
	return nil
}

// SetRange provides function to set data validation range in drop list, only
// accepts int, float64, or string data type formula argument.
func (dd *DataValidation) SetRange(f1, f2 interface{}, t DataValidationType, o DataValidationOperator) error {
	var formula1, formula2 string
	switch v := f1.(type) {
	case int:
		formula1 = fmt.Sprintf("<formula1>%d</formula1>", v)
	case float64:
		if math.Abs(v) > math.MaxFloat32 {
			return ErrDataValidationRange
		}
		formula1 = fmt.Sprintf("<formula1>%.17g</formula1>", v)
	case string:
		formula1 = fmt.Sprintf("<formula1>%s</formula1>", v)
	default:
		return ErrParameterInvalid
	}
	switch v := f2.(type) {
	case int:
		formula2 = fmt.Sprintf("<formula2>%d</formula2>", v)
	case float64:
		if math.Abs(v) > math.MaxFloat32 {
			return ErrDataValidationRange
		}
		formula2 = fmt.Sprintf("<formula2>%.17g</formula2>", v)
	case string:
		formula2 = fmt.Sprintf("<formula2>%s</formula2>", v)
	default:
		return ErrParameterInvalid
	}
	dd.Formula1, dd.Formula2 = formula1, formula2
	dd.Type = convDataValidationType(t)
	dd.Operator = convDataValidationOperator(o)
	return nil
}

// SetSqrefDropList provides set data validation on a range with source
// reference range of the worksheet by given data validation object and
// worksheet name. The data validation object can be created by
// NewDataValidation function. For example, set data validation on
// Sheet1!A7:B8 with validation criteria source Sheet1!E1:E3 settings, create
// in-cell dropdown by allowing list source:
//
//	dvRange := excelize.NewDataValidation(true)
//	dvRange.Sqref = "A7:B8"
//	dvRange.SetSqrefDropList("$E$1:$E$3")
//	f.AddDataValidation("Sheet1", dvRange)
func (dd *DataValidation) SetSqrefDropList(sqref string) {
	dd.Formula1 = fmt.Sprintf("<formula1>%s</formula1>", sqref)
	dd.Type = convDataValidationType(typeList)
}

// SetSqref provides function to set data validation range in drop list.
func (dd *DataValidation) SetSqref(sqref string) {
	if dd.Sqref == "" {
		dd.Sqref = sqref
	} else {
		dd.Sqref = fmt.Sprintf("%s %s", dd.Sqref, sqref)
	}
}

// convDataValidationType get excel data validation type.
func convDataValidationType(t DataValidationType) string {
	typeMap := map[DataValidationType]string{
		typeNone:                     "none",
		DataValidationTypeCustom:     "custom",
		DataValidationTypeDate:       "date",
		DataValidationTypeDecimal:    "decimal",
		typeList:                     "list",
		DataValidationTypeTextLength: "textLength",
		DataValidationTypeTime:       "time",
		DataValidationTypeWhole:      "whole",
	}

	return typeMap[t]
}

// convDataValidationOperator get excel data validation operator.
func convDataValidationOperator(o DataValidationOperator) string {
	typeMap := map[DataValidationOperator]string{
		DataValidationOperatorBetween:            "between",
		DataValidationOperatorEqual:              "equal",
		DataValidationOperatorGreaterThan:        "greaterThan",
		DataValidationOperatorGreaterThanOrEqual: "greaterThanOrEqual",
		DataValidationOperatorLessThan:           "lessThan",
		DataValidationOperatorLessThanOrEqual:    "lessThanOrEqual",
		DataValidationOperatorNotBetween:         "notBetween",
		DataValidationOperatorNotEqual:           "notEqual",
	}

	return typeMap[o]
}

// AddDataValidation provides set data validation on a range of the worksheet
// by given data validation object and worksheet name. The data validation
// object can be created by NewDataValidation function.
//
// Example 1, set data validation on Sheet1!A1:B2 with validation criteria
// settings, show error alert after invalid data is entered with "Stop" style
// and custom title "error body":
//
//	dvRange := excelize.NewDataValidation(true)
//	dvRange.Sqref = "A1:B2"
//	dvRange.SetRange(10, 20, excelize.DataValidationTypeWhole, excelize.DataValidationOperatorBetween)
//	dvRange.SetError(excelize.DataValidationErrorStyleStop, "error title", "error body")
//	err := f.AddDataValidation("Sheet1", dvRange)
//
// Example 2, set data validation on Sheet1!A3:B4 with validation criteria
// settings, and show input message when cell is selected:
//
//	dvRange = excelize.NewDataValidation(true)
//	dvRange.Sqref = "A3:B4"
//	dvRange.SetRange(10, 20, excelize.DataValidationTypeWhole, excelize.DataValidationOperatorGreaterThan)
//	dvRange.SetInput("input title", "input body")
//	err = f.AddDataValidation("Sheet1", dvRange)
//
// Example 3, set data validation on Sheet1!A5:B6 with validation criteria
// settings, create in-cell dropdown by allowing list source:
//
//	dvRange = excelize.NewDataValidation(true)
//	dvRange.Sqref = "A5:B6"
//	dvRange.SetDropList([]string{"1", "2", "3"})
//	err = f.AddDataValidation("Sheet1", dvRange)
func (f *File) AddDataValidation(sheet string, dv *DataValidation) error {
	ws, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	if nil == ws.DataValidations {
		ws.DataValidations = new(xlsxDataValidations)
	}
	ws.DataValidations.DataValidation = append(ws.DataValidations.DataValidation, dv)
	ws.DataValidations.Count = len(ws.DataValidations.DataValidation)
	return err
}

// GetDataValidation retrieves all comments and returns a map of worksheet name to
// the worksheet data validations.
func (f *File) GetDataValidation() (map[string][]*DataValidation, error) {
	resultDVs := map[string][]*DataValidation{}
	for sheet := range f.sheetMap {
		ws, err := f.workSheetReader(sheet)
		if err != nil {
			return nil, err
		}

		// No data validaion on this sheet
		if ws.DataValidations == nil {
			return resultDVs, nil
		}

		for _, item := range ws.DataValidations.DataValidation {
			dv := *item
			resultDVs[sheet] = append(resultDVs[sheet], &dv)
		}
	}
	return resultDVs, nil
}

// DeleteDataValidation delete data validation by given worksheet name and
// reference sequence. All data validations in the worksheet will be deleted
// if not specify reference sequence parameter.
func (f *File) DeleteDataValidation(sheet string, sqref ...string) error {
	ws, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	if ws.DataValidations == nil {
		return nil
	}
	if sqref == nil {
		ws.DataValidations = nil
		return nil
	}
	delCells, err := f.flatSqref(sqref[0])
	if err != nil {
		return err
	}
	dv := ws.DataValidations
	for i := 0; i < len(dv.DataValidation); i++ {
		var applySqref []string
		colCells, err := f.flatSqref(dv.DataValidation[i].Sqref)
		if err != nil {
			return err
		}
		for col, cells := range delCells {
			for _, cell := range cells {
				idx := inCoordinates(colCells[col], cell)
				if idx != -1 {
					colCells[col] = append(colCells[col][:idx], colCells[col][idx+1:]...)
				}
			}
		}
		for _, col := range colCells {
			applySqref = append(applySqref, f.squashSqref(col)...)
		}
		dv.DataValidation[i].Sqref = strings.Join(applySqref, " ")
		if len(applySqref) == 0 {
			dv.DataValidation = append(dv.DataValidation[:i], dv.DataValidation[i+1:]...)
			i--
		}
	}
	dv.Count = len(dv.DataValidation)
	if dv.Count == 0 {
		ws.DataValidations = nil
	}
	return nil
}

// squashSqref generates cell reference sequence by given cells coordinates list.
func (f *File) squashSqref(cells [][]int) []string {
	if len(cells) == 1 {
		cell, _ := CoordinatesToCellName(cells[0][0], cells[0][1])
		return []string{cell}
	} else if len(cells) == 0 {
		return []string{}
	}
	var res []string
	l, r := 0, 0
	for i := 1; i < len(cells); i++ {
		if cells[i][0] == cells[r][0] && cells[i][1]-cells[r][1] > 1 {
			curr, _ := f.coordinatesToAreaRef(append(cells[l], cells[r]...))
			if l == r {
				curr, _ = CoordinatesToCellName(cells[l][0], cells[l][1])
			}
			res = append(res, curr)
			l, r = i, i
		} else {
			r++
		}
	}
	curr, _ := f.coordinatesToAreaRef(append(cells[l], cells[r]...))
	if l == r {
		curr, _ = CoordinatesToCellName(cells[l][0], cells[l][1])
	}
	return append(res, curr)
}
