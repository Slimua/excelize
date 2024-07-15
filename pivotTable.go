// Copyright 2016 - 2024 The excelize Authors. All rights reserved. Use of
// this source code is governed by a BSD-style license that can be found in
// the LICENSE file.
//
// Package excelize providing a set of functions that allow you to write to and
// read from XLAM / XLSM / XLSX / XLTM / XLTX files. Supports reading and
// writing spreadsheet documents generated by Microsoft Excel™ 2007 and later.
// Supports complex components by high compatibility, and provided streaming
// API for generating or reading data from a worksheet with huge amounts of
// data. This library needs Go version 1.18 or later.

package excelize

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// PivotTableOptions directly maps the format settings of the pivot table.
//
// PivotTableStyleName: The built-in pivot table style names
//
//	PivotStyleLight1 - PivotStyleLight28
//	PivotStyleMedium1 - PivotStyleMedium28
//	PivotStyleDark1 - PivotStyleDark28
type PivotTableOptions struct {
	pivotTableXML       string
	pivotCacheXML       string
	pivotSheetName      string
	pivotDataRange      string
	namedDataRange      bool
	DataRange           string
	PivotTableRange     string
	Name                string
	Rows                []PivotTableField
	Columns             []PivotTableField
	Data                []PivotTableField
	Filter              []PivotTableField
	RowGrandTotals      bool
	ColGrandTotals      bool
	ShowDrill           bool
	UseAutoFormatting   bool
	PageOverThenDown    bool
	MergeItem           bool
	CompactData         bool
	ShowError           bool
	ShowRowHeaders      bool
	ShowColHeaders      bool
	ShowRowStripes      bool
	ShowColStripes      bool
	ShowLastColumn      bool
	PivotTableStyleName string
}

// PivotTableField directly maps the field settings of the pivot table.
// Subtotal specifies the aggregation function that applies to this data
// field. The default value is sum. The possible values for this attribute
// are:
//
//	Average
//	Count
//	CountNums
//	Max
//	Min
//	Product
//	StdDev
//	StdDevp
//	Sum
//	Var
//	Varp
//
// Name specifies the name of the data field. Maximum 255 characters
// are allowed in data field name, excess characters will be truncated.
type PivotTableField struct {
	Compact         bool
	Data            string
	Name            string
	Outline         bool
	Subtotal        string
	DefaultSubtotal bool
}

// AddPivotTable provides the method to add pivot table by given pivot table
// options. Note that the same fields can not in Columns, Rows and Filter
// fields at the same time.
//
// For example, create a pivot table on the range reference Sheet1!G2:M34 with
// the range reference Sheet1!A1:E31 as the data source, summarize by sum for
// sales:
//
//	package main
//
//	import (
//	    "fmt"
//	    "math/rand"
//
//	    "github.com/xuri/excelize/v2"
//	)
//
//	func main() {
//	    f := excelize.NewFile()
//	    defer func() {
//	        if err := f.Close(); err != nil {
//	            fmt.Println(err)
//	        }
//	    }()
//	    // Create some data in a sheet
//	    month := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
//	    year := []int{2017, 2018, 2019}
//	    types := []string{"Meat", "Dairy", "Beverages", "Produce"}
//	    region := []string{"East", "West", "North", "South"}
//	    f.SetSheetRow("Sheet1", "A1", &[]string{"Month", "Year", "Type", "Sales", "Region"})
//	    for row := 2; row < 32; row++ {
//	        f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), month[rand.Intn(12)])
//	        f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), year[rand.Intn(3)])
//	        f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), types[rand.Intn(4)])
//	        f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), rand.Intn(5000))
//	        f.SetCellValue("Sheet1", fmt.Sprintf("E%d", row), region[rand.Intn(4)])
//	    }
//	    if err := f.AddPivotTable(&excelize.PivotTableOptions{
//	        DataRange:       "Sheet1!A1:E31",
//	        PivotTableRange: "Sheet1!G2:M34",
//	        Rows:            []excelize.PivotTableField{{Data: "Month", DefaultSubtotal: true}, {Data: "Year"}},
//	        Filter:          []excelize.PivotTableField{{Data: "Region"}},
//	        Columns:         []excelize.PivotTableField{{Data: "Type", DefaultSubtotal: true}},
//	        Data:            []excelize.PivotTableField{{Data: "Sales", Name: "Summarize", Subtotal: "Sum"}},
//	        RowGrandTotals:  true,
//	        ColGrandTotals:  true,
//	        ShowDrill:       true,
//	        ShowRowHeaders:  true,
//	        ShowColHeaders:  true,
//	        ShowLastColumn:  true,
//	    }); err != nil {
//	        fmt.Println(err)
//	    }
//	    if err := f.SaveAs("Book1.xlsx"); err != nil {
//	        fmt.Println(err)
//	    }
//	}
func (f *File) AddPivotTable(opts *PivotTableOptions) error {
	// parameter validation
	_, pivotTableSheetPath, err := f.parseFormatPivotTableSet(opts)
	if err != nil {
		return err
	}

	pivotTableID := f.countPivotTables() + 1
	pivotCacheID := f.countPivotCache() + 1

	sheetRelationshipsPivotTableXML := "../pivotTables/pivotTable" + strconv.Itoa(pivotTableID) + ".xml"
	opts.pivotTableXML = strings.ReplaceAll(sheetRelationshipsPivotTableXML, "..", "xl")
	opts.pivotCacheXML = "xl/pivotCache/pivotCacheDefinition" + strconv.Itoa(pivotCacheID) + ".xml"
	if err = f.addPivotCache(opts); err != nil {
		return err
	}

	// workbook pivot cache
	workBookPivotCacheRID := f.addRels(f.getWorkbookRelsPath(), SourceRelationshipPivotCache, strings.TrimPrefix(opts.pivotCacheXML, "xl/"), "")
	cacheID := f.addWorkbookPivotCache(workBookPivotCacheRID)

	pivotCacheRels := "xl/pivotTables/_rels/pivotTable" + strconv.Itoa(pivotTableID) + ".xml.rels"
	// rId not used
	_ = f.addRels(pivotCacheRels, SourceRelationshipPivotCache, fmt.Sprintf("../pivotCache/pivotCacheDefinition%d.xml", pivotCacheID), "")
	if err = f.addPivotTable(cacheID, pivotTableID, opts); err != nil {
		return err
	}
	pivotTableSheetRels := "xl/worksheets/_rels/" + strings.TrimPrefix(pivotTableSheetPath, "xl/worksheets/") + ".rels"
	f.addRels(pivotTableSheetRels, SourceRelationshipPivotTable, sheetRelationshipsPivotTableXML, "")
	if err = f.addContentTypePart(pivotTableID, "pivotTable"); err != nil {
		return err
	}
	return f.addContentTypePart(pivotCacheID, "pivotCache")
}

// parseFormatPivotTableSet provides a function to validate pivot table
// properties.
func (f *File) parseFormatPivotTableSet(opts *PivotTableOptions) (*xlsxWorksheet, string, error) {
	if opts == nil {
		return nil, "", ErrParameterRequired
	}
	pivotTableSheetName, _, err := f.adjustRange(opts.PivotTableRange)
	if err != nil {
		return nil, "", newPivotTableRangeError(err.Error())
	}
	if len(opts.Name) > MaxFieldLength {
		return nil, "", ErrNameLength
	}
	opts.pivotSheetName = pivotTableSheetName
	if err = f.getPivotTableDataRange(opts); err != nil {
		return nil, "", err
	}
	dataSheetName, _, err := f.adjustRange(opts.pivotDataRange)
	if err != nil {
		return nil, "", newPivotTableDataRangeError(err.Error())
	}
	dataSheet, err := f.workSheetReader(dataSheetName)
	if err != nil {
		return dataSheet, "", err
	}
	pivotTableSheetPath, ok := f.getSheetXMLPath(pivotTableSheetName)
	if !ok {
		return dataSheet, pivotTableSheetPath, ErrSheetNotExist{pivotTableSheetName}
	}
	return dataSheet, pivotTableSheetPath, err
}

// adjustRange adjust range, for example: adjust Sheet1!$E$31:$A$1 to Sheet1!$A$1:$E$31
func (f *File) adjustRange(rangeStr string) (string, []int, error) {
	if len(rangeStr) < 1 {
		return "", []int{}, ErrParameterRequired
	}
	rng := strings.Split(rangeStr, "!")
	if len(rng) != 2 {
		return "", []int{}, ErrParameterInvalid
	}
	trimRng := strings.ReplaceAll(rng[1], "$", "")
	coordinates, err := rangeRefToCoordinates(trimRng)
	if err != nil {
		return rng[0], []int{}, err
	}
	x1, y1, x2, y2 := coordinates[0], coordinates[1], coordinates[2], coordinates[3]
	if x1 == x2 && y1 == y2 {
		return rng[0], []int{}, ErrParameterInvalid
	}

	// Correct the range, such correct C1:B3 to B1:C3.
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	if y2 < y1 {
		y1, y2 = y2, y1
	}
	return rng[0], []int{x1, y1, x2, y2}, nil
}

// getTableFieldsOrder provides a function to get order list of pivot table
// fields.
func (f *File) getTableFieldsOrder(opts *PivotTableOptions) ([]string, error) {
	var order []string
	if err := f.getPivotTableDataRange(opts); err != nil {
		return order, err
	}
	dataSheet, coordinates, err := f.adjustRange(opts.pivotDataRange)
	if err != nil {
		return order, newPivotTableDataRangeError(err.Error())
	}
	for col := coordinates[0]; col <= coordinates[2]; col++ {
		coordinate, _ := CoordinatesToCellName(col, coordinates[1])
		name, err := f.GetCellValue(dataSheet, coordinate)
		if err != nil {
			return order, err
		}
		if name == "" {
			return order, ErrParameterInvalid
		}
		order = append(order, name)
	}
	return order, nil
}

// addPivotCache provides a function to create a pivot cache by given properties.
func (f *File) addPivotCache(opts *PivotTableOptions) error {
	// validate data range
	dataSheet, coordinates, err := f.adjustRange(opts.pivotDataRange)
	if err != nil {
		return newPivotTableDataRangeError(err.Error())
	}
	order, err := f.getTableFieldsOrder(opts)
	if err != nil {
		return newPivotTableDataRangeError(err.Error())
	}
	topLeftCell, _ := CoordinatesToCellName(coordinates[0], coordinates[1])
	bottomRightCell, _ := CoordinatesToCellName(coordinates[2], coordinates[3])
	pc := xlsxPivotCacheDefinition{
		SaveData:              false,
		RefreshOnLoad:         true,
		CreatedVersion:        pivotTableVersion,
		RefreshedVersion:      pivotTableRefreshedVersion,
		MinRefreshableVersion: pivotTableVersion,
		CacheSource: &xlsxCacheSource{
			Type: "worksheet",
			WorksheetSource: &xlsxWorksheetSource{
				Ref:   topLeftCell + ":" + bottomRightCell,
				Sheet: dataSheet,
			},
		},
		CacheFields: &xlsxCacheFields{},
	}
	if opts.namedDataRange {
		pc.CacheSource.WorksheetSource = &xlsxWorksheetSource{Name: opts.DataRange}
	}
	for _, name := range order {
		pc.CacheFields.CacheField = append(pc.CacheFields.CacheField, &xlsxCacheField{
			Name:        name,
			SharedItems: &xlsxSharedItems{ContainsBlank: true, M: []xlsxMissing{{}}},
		})
	}
	pc.CacheFields.Count = len(pc.CacheFields.CacheField)
	pivotCache, err := xml.Marshal(pc)
	f.saveFileList(opts.pivotCacheXML, pivotCache)
	return err
}

// addPivotTable provides a function to create a pivot table by given pivot
// table ID and properties.
func (f *File) addPivotTable(cacheID, pivotTableID int, opts *PivotTableOptions) error {
	// validate pivot table range
	_, coordinates, err := f.adjustRange(opts.PivotTableRange)
	if err != nil {
		return newPivotTableRangeError(err.Error())
	}

	topLeftCell, _ := CoordinatesToCellName(coordinates[0], coordinates[1])
	bottomRightCell, _ := CoordinatesToCellName(coordinates[2], coordinates[3])

	pivotTableStyle := func() string {
		if opts.PivotTableStyleName == "" {
			return "PivotStyleLight16"
		}
		return opts.PivotTableStyleName
	}
	pt := xlsxPivotTableDefinition{
		Name:                  opts.Name,
		CacheID:               cacheID,
		RowGrandTotals:        &opts.RowGrandTotals,
		ColGrandTotals:        &opts.ColGrandTotals,
		UpdatedVersion:        pivotTableRefreshedVersion,
		MinRefreshableVersion: pivotTableVersion,
		ShowDrill:             &opts.ShowDrill,
		UseAutoFormatting:     &opts.UseAutoFormatting,
		PageOverThenDown:      &opts.PageOverThenDown,
		MergeItem:             &opts.MergeItem,
		CreatedVersion:        pivotTableVersion,
		CompactData:           &opts.CompactData,
		ShowError:             &opts.ShowError,
		DataCaption:           "Values",
		Location: &xlsxLocation{
			Ref:            topLeftCell + ":" + bottomRightCell,
			FirstDataCol:   1,
			FirstDataRow:   1,
			FirstHeaderRow: 1,
		},
		PivotFields: &xlsxPivotFields{},
		RowItems: &xlsxRowItems{
			Count: 1,
			I: []*xlsxI{
				{
					[]*xlsxX{{}, {}},
				},
			},
		},
		ColItems: &xlsxColItems{
			Count: 1,
			I:     []*xlsxI{{}},
		},
		PivotTableStyleInfo: &xlsxPivotTableStyleInfo{
			Name:           pivotTableStyle(),
			ShowRowHeaders: opts.ShowRowHeaders,
			ShowColHeaders: opts.ShowColHeaders,
			ShowRowStripes: opts.ShowRowStripes,
			ShowColStripes: opts.ShowColStripes,
			ShowLastColumn: opts.ShowLastColumn,
		},
	}
	if pt.Name == "" {
		pt.Name = fmt.Sprintf("PivotTable%d", pivotTableID)
	}
	// pivot fields
	_ = f.addPivotFields(&pt, opts)

	// count pivot fields
	pt.PivotFields.Count = len(pt.PivotFields.PivotField)

	// data range has been checked
	_ = f.addPivotRowFields(&pt, opts)
	_ = f.addPivotColFields(&pt, opts)
	_ = f.addPivotPageFields(&pt, opts)
	_ = f.addPivotDataFields(&pt, opts)

	pivotTable, err := xml.Marshal(pt)
	f.saveFileList(opts.pivotTableXML, pivotTable)
	return err
}

// addPivotRowFields provides a method to add row fields for pivot table by
// given pivot table options.
func (f *File) addPivotRowFields(pt *xlsxPivotTableDefinition, opts *PivotTableOptions) error {
	// row fields
	rowFieldsIndex, err := f.getPivotFieldsIndex(opts.Rows, opts)
	if err != nil {
		return err
	}
	for _, fieldIdx := range rowFieldsIndex {
		if pt.RowFields == nil {
			pt.RowFields = &xlsxRowFields{}
		}
		pt.RowFields.Field = append(pt.RowFields.Field, &xlsxField{
			X: fieldIdx,
		})
	}

	// count row fields
	if pt.RowFields != nil {
		pt.RowFields.Count = len(pt.RowFields.Field)
	}
	return err
}

// addPivotPageFields provides a method to add page fields for pivot table by
// given pivot table options.
func (f *File) addPivotPageFields(pt *xlsxPivotTableDefinition, opts *PivotTableOptions) error {
	// page fields
	pageFieldsIndex, err := f.getPivotFieldsIndex(opts.Filter, opts)
	if err != nil {
		return err
	}
	pageFieldsName := f.getPivotTableFieldsName(opts.Filter)
	for idx, pageField := range pageFieldsIndex {
		if pt.PageFields == nil {
			pt.PageFields = &xlsxPageFields{}
		}
		pt.PageFields.PageField = append(pt.PageFields.PageField, &xlsxPageField{
			Name: pageFieldsName[idx],
			Fld:  pageField,
		})
	}

	// count page fields
	if pt.PageFields != nil {
		pt.PageFields.Count = len(pt.PageFields.PageField)
	}
	return err
}

// addPivotDataFields provides a method to add data fields for pivot table by
// given pivot table options.
func (f *File) addPivotDataFields(pt *xlsxPivotTableDefinition, opts *PivotTableOptions) error {
	// data fields
	dataFieldsIndex, err := f.getPivotFieldsIndex(opts.Data, opts)
	if err != nil {
		return err
	}
	dataFieldsSubtotals := f.getPivotTableFieldsSubtotal(opts.Data)
	dataFieldsName := f.getPivotTableFieldsName(opts.Data)
	for idx, dataField := range dataFieldsIndex {
		if pt.DataFields == nil {
			pt.DataFields = &xlsxDataFields{}
		}
		pt.DataFields.DataField = append(pt.DataFields.DataField, &xlsxDataField{
			Name:     dataFieldsName[idx],
			Fld:      dataField,
			Subtotal: dataFieldsSubtotals[idx],
		})
	}

	// count data fields
	if pt.DataFields != nil {
		pt.DataFields.Count = len(pt.DataFields.DataField)
	}
	return err
}

// inPivotTableField provides a method to check if an element is present in
// pivot table fields list, and return the index of its location, otherwise
// return -1.
func inPivotTableField(a []PivotTableField, x string) int {
	for idx, n := range a {
		if x == n.Data {
			return idx
		}
	}
	return -1
}

// addPivotColFields create pivot column fields by given pivot table
// definition and option.
func (f *File) addPivotColFields(pt *xlsxPivotTableDefinition, opts *PivotTableOptions) error {
	if len(opts.Columns) == 0 {
		if len(opts.Data) <= 1 {
			return nil
		}
		pt.ColFields = &xlsxColFields{}
		// in order to create pivot table in case there is no input from Columns
		pt.ColFields.Count = 1
		pt.ColFields.Field = append(pt.ColFields.Field, &xlsxField{
			X: -2,
		})
		return nil
	}

	pt.ColFields = &xlsxColFields{}

	// col fields
	colFieldsIndex, err := f.getPivotFieldsIndex(opts.Columns, opts)
	if err != nil {
		return err
	}
	for _, fieldIdx := range colFieldsIndex {
		pt.ColFields.Field = append(pt.ColFields.Field, &xlsxField{
			X: fieldIdx,
		})
	}

	// in order to create pivot in case there is many Columns and Data
	if len(opts.Data) > 1 {
		pt.ColFields.Field = append(pt.ColFields.Field, &xlsxField{
			X: -2,
		})
	}

	// count col fields
	pt.ColFields.Count = len(pt.ColFields.Field)
	return err
}

// addPivotFields create pivot fields based on the column order of the first
// row in the data region by given pivot table definition and option.
func (f *File) addPivotFields(pt *xlsxPivotTableDefinition, opts *PivotTableOptions) error {
	order, err := f.getTableFieldsOrder(opts)
	if err != nil {
		return err
	}
	x := 0
	for _, name := range order {
		if inPivotTableField(opts.Rows, name) != -1 {
			rowOptions, ok := f.getPivotTableFieldOptions(name, opts.Rows)
			var items []*xlsxItem
			if !ok || !rowOptions.DefaultSubtotal {
				items = append(items, &xlsxItem{X: &x})
			} else {
				items = append(items, &xlsxItem{T: "default"})
			}

			pt.PivotFields.PivotField = append(pt.PivotFields.PivotField, &xlsxPivotField{
				Name:            f.getPivotTableFieldName(name, opts.Rows),
				Axis:            "axisRow",
				DataField:       inPivotTableField(opts.Data, name) != -1,
				Compact:         &rowOptions.Compact,
				Outline:         &rowOptions.Outline,
				DefaultSubtotal: &rowOptions.DefaultSubtotal,
				Items: &xlsxItems{
					Count: len(items),
					Item:  items,
				},
			})
			continue
		}
		if inPivotTableField(opts.Filter, name) != -1 {
			pt.PivotFields.PivotField = append(pt.PivotFields.PivotField, &xlsxPivotField{
				Axis:      "axisPage",
				DataField: inPivotTableField(opts.Data, name) != -1,
				Name:      f.getPivotTableFieldName(name, opts.Columns),
				Items: &xlsxItems{
					Count: 1,
					Item: []*xlsxItem{
						{T: "default"},
					},
				},
			})
			continue
		}
		if inPivotTableField(opts.Columns, name) != -1 {
			columnOptions, ok := f.getPivotTableFieldOptions(name, opts.Columns)
			var items []*xlsxItem
			if !ok || !columnOptions.DefaultSubtotal {
				items = append(items, &xlsxItem{X: &x})
			} else {
				items = append(items, &xlsxItem{T: "default"})
			}
			pt.PivotFields.PivotField = append(pt.PivotFields.PivotField, &xlsxPivotField{
				Name:            f.getPivotTableFieldName(name, opts.Columns),
				Axis:            "axisCol",
				DataField:       inPivotTableField(opts.Data, name) != -1,
				Compact:         &columnOptions.Compact,
				Outline:         &columnOptions.Outline,
				DefaultSubtotal: &columnOptions.DefaultSubtotal,
				Items: &xlsxItems{
					Count: len(items),
					Item:  items,
				},
			})
			continue
		}
		if inPivotTableField(opts.Data, name) != -1 {
			pt.PivotFields.PivotField = append(pt.PivotFields.PivotField, &xlsxPivotField{
				DataField: true,
			})
			continue
		}
		pt.PivotFields.PivotField = append(pt.PivotFields.PivotField, &xlsxPivotField{})
	}
	return err
}

// countPivotTables provides a function to get pivot table files count storage
// in the folder xl/pivotTables.
func (f *File) countPivotTables() int {
	count := 0
	f.Pkg.Range(func(k, v interface{}) bool {
		if strings.Contains(k.(string), "xl/pivotTables/pivotTable") {
			count++
		}
		return true
	})
	return count
}

// countPivotCache provides a function to get pivot table cache definition files
// count storage in the folder xl/pivotCache.
func (f *File) countPivotCache() int {
	count := 0
	f.Pkg.Range(func(k, v interface{}) bool {
		if strings.Contains(k.(string), "xl/pivotCache/pivotCacheDefinition") {
			count++
		}
		return true
	})
	return count
}

// getPivotFieldsIndex convert the column of the first row in the data region
// to a sequential index by given fields and pivot option.
func (f *File) getPivotFieldsIndex(fields []PivotTableField, opts *PivotTableOptions) ([]int, error) {
	var pivotFieldsIndex []int
	orders, err := f.getTableFieldsOrder(opts)
	if err != nil {
		return pivotFieldsIndex, err
	}
	for _, field := range fields {
		if pos := inStrSlice(orders, field.Data, true); pos != -1 {
			pivotFieldsIndex = append(pivotFieldsIndex, pos)
		}
	}
	return pivotFieldsIndex, nil
}

// getPivotTableFieldsSubtotal prepare fields subtotal by given pivot table fields.
func (f *File) getPivotTableFieldsSubtotal(fields []PivotTableField) []string {
	field := make([]string, len(fields))
	enums := []string{"average", "count", "countNums", "max", "min", "product", "stdDev", "stdDevp", "sum", "var", "varp"}
	inEnums := func(enums []string, val string) string {
		for _, enum := range enums {
			if strings.EqualFold(enum, val) {
				return enum
			}
		}
		return "sum"
	}
	for idx, fld := range fields {
		field[idx] = inEnums(enums, fld.Subtotal)
	}
	return field
}

// getPivotTableFieldsName prepare fields name list by given pivot table
// fields.
func (f *File) getPivotTableFieldsName(fields []PivotTableField) []string {
	field := make([]string, len(fields))
	for idx, fld := range fields {
		if len(fld.Name) > MaxFieldLength {
			field[idx] = fld.Name[:MaxFieldLength]
			continue
		}
		field[idx] = fld.Name
	}
	return field
}

// getPivotTableFieldName prepare field name by given pivot table fields.
func (f *File) getPivotTableFieldName(name string, fields []PivotTableField) string {
	fieldsName := f.getPivotTableFieldsName(fields)
	for idx, field := range fields {
		if field.Data == name {
			return fieldsName[idx]
		}
	}
	return ""
}

// getPivotTableFieldOptions return options for specific field by given field name.
func (f *File) getPivotTableFieldOptions(name string, fields []PivotTableField) (options PivotTableField, ok bool) {
	for _, field := range fields {
		if field.Data == name {
			options, ok = field, true
			return
		}
	}
	return
}

// addWorkbookPivotCache add the association ID of the pivot cache in workbook.xml.
func (f *File) addWorkbookPivotCache(RID int) int {
	wb, _ := f.workbookReader()
	if wb.PivotCaches == nil {
		wb.PivotCaches = &xlsxPivotCaches{}
	}
	cacheID := 1
	for _, pivotCache := range wb.PivotCaches.PivotCache {
		if pivotCache.CacheID > cacheID {
			cacheID = pivotCache.CacheID
		}
	}
	cacheID++
	wb.PivotCaches.PivotCache = append(wb.PivotCaches.PivotCache, xlsxPivotCache{
		CacheID: cacheID,
		RID:     fmt.Sprintf("rId%d", RID),
	})
	return cacheID
}

// GetPivotTables returns all pivot table definitions in a worksheet by given
// worksheet name.
func (f *File) GetPivotTables(sheet string) ([]PivotTableOptions, error) {
	var pivotTables []PivotTableOptions
	name, ok := f.getSheetXMLPath(sheet)
	if !ok {
		return pivotTables, ErrSheetNotExist{sheet}
	}
	rels := "xl/worksheets/_rels/" + strings.TrimPrefix(name, "xl/worksheets/") + ".rels"
	sheetRels, err := f.relsReader(rels)
	if err != nil {
		return pivotTables, err
	}
	if sheetRels == nil {
		sheetRels = &xlsxRelationships{}
	}
	for _, v := range sheetRels.Relationships {
		if v.Type == SourceRelationshipPivotTable {
			pivotTableXML := strings.ReplaceAll(v.Target, "..", "xl")
			pivotCacheRels := "xl/pivotTables/_rels/" + filepath.Base(v.Target) + ".rels"
			pivotTable, err := f.getPivotTable(sheet, pivotTableXML, pivotCacheRels)
			if err != nil {
				return pivotTables, err
			}
			pivotTables = append(pivotTables, pivotTable)
		}
	}
	return pivotTables, nil
}

// getPivotTableDataRange checking given if data range is a cell reference or
// named reference (defined name or table name), and set pivot table data range.
func (f *File) getPivotTableDataRange(opts *PivotTableOptions) error {
	if opts.DataRange == "" {
		return newPivotTableDataRangeError(ErrParameterRequired.Error())
	}
	if opts.pivotDataRange != "" {
		return nil
	}
	if strings.Contains(opts.DataRange, "!") {
		opts.pivotDataRange = opts.DataRange
		return nil
	}
	for _, sheetName := range f.GetSheetList() {
		tables, err := f.GetTables(sheetName)
		e := ErrSheetNotExist{sheetName}
		if err != nil && err.Error() != newNotWorksheetError(sheetName).Error() && err.Error() != e.Error() {
			return err
		}
		for _, table := range tables {
			if table.Name == opts.DataRange {
				opts.pivotDataRange, opts.namedDataRange = fmt.Sprintf("%s!%s", sheetName, table.Range), true
				return err
			}
		}
	}
	if !opts.namedDataRange {
		opts.pivotDataRange = f.getDefinedNameRefTo(opts.DataRange, opts.pivotSheetName)
		if opts.pivotDataRange != "" {
			opts.namedDataRange = true
			return nil
		}
	}
	return newPivotTableDataRangeError(ErrParameterInvalid.Error())
}

// getPivotTable provides a function to get a pivot table definition by given
// worksheet name, pivot table XML path and pivot cache relationship XML path.
func (f *File) getPivotTable(sheet, pivotTableXML, pivotCacheRels string) (PivotTableOptions, error) {
	var opts PivotTableOptions
	rels, err := f.relsReader(pivotCacheRels)
	if err != nil {
		return opts, err
	}
	var pivotCacheXML string
	for _, v := range rels.Relationships {
		if v.Type == SourceRelationshipPivotCache {
			pivotCacheXML = strings.ReplaceAll(v.Target, "..", "xl")
			break
		}
	}
	pc, err := f.pivotCacheReader(pivotCacheXML)
	if err != nil {
		return opts, err
	}
	pt, err := f.pivotTableReader(pivotTableXML)
	if err != nil {
		return opts, err
	}
	opts = PivotTableOptions{
		pivotTableXML:   pivotTableXML,
		pivotCacheXML:   pivotCacheXML,
		pivotSheetName:  sheet,
		DataRange:       fmt.Sprintf("%s!%s", pc.CacheSource.WorksheetSource.Sheet, pc.CacheSource.WorksheetSource.Ref),
		PivotTableRange: fmt.Sprintf("%s!%s", sheet, pt.Location.Ref),
		Name:            pt.Name,
	}
	if pc.CacheSource.WorksheetSource.Name != "" {
		opts.DataRange = pc.CacheSource.WorksheetSource.Name
		_ = f.getPivotTableDataRange(&opts)
	}
	fields := []string{"RowGrandTotals", "ColGrandTotals", "ShowDrill", "UseAutoFormatting", "PageOverThenDown", "MergeItem", "CompactData", "ShowError"}
	immutable, mutable := reflect.ValueOf(*pt), reflect.ValueOf(&opts).Elem()
	for _, field := range fields {
		immutableField := immutable.FieldByName(field)
		if immutableField.Kind() == reflect.Ptr && !immutableField.IsNil() && immutableField.Elem().Kind() == reflect.Bool {
			mutable.FieldByName(field).SetBool(immutableField.Elem().Bool())
		}
	}
	if si := pt.PivotTableStyleInfo; si != nil {
		opts.ShowRowHeaders = si.ShowRowHeaders
		opts.ShowColHeaders = si.ShowColHeaders
		opts.ShowRowStripes = si.ShowRowStripes
		opts.ShowColStripes = si.ShowColStripes
		opts.ShowLastColumn = si.ShowLastColumn
		opts.PivotTableStyleName = si.Name
	}
	order, err := f.getTableFieldsOrder(&opts)
	if err != nil {
		return opts, err
	}
	f.extractPivotTableFields(order, pt, &opts)
	return opts, err
}

// pivotTableReader provides a function to get the pointer to the structure
// after deserialization of xl/pivotTables/pivotTable%d.xml.
func (f *File) pivotTableReader(path string) (*xlsxPivotTableDefinition, error) {
	content, ok := f.Pkg.Load(path)
	pivotTable := &xlsxPivotTableDefinition{}
	if ok && content != nil {
		if err := f.xmlNewDecoder(bytes.NewReader(namespaceStrictToTransitional(content.([]byte)))).
			Decode(pivotTable); err != nil && err != io.EOF {
			return nil, err
		}
	}
	return pivotTable, nil
}

// pivotCacheReader provides a function to get the pointer to the structure
// after deserialization of xl/pivotCache/pivotCacheDefinition%d.xml.
func (f *File) pivotCacheReader(path string) (*xlsxPivotCacheDefinition, error) {
	content, ok := f.Pkg.Load(path)
	pivotCache := &xlsxPivotCacheDefinition{}
	if ok && content != nil {
		if err := f.xmlNewDecoder(bytes.NewReader(namespaceStrictToTransitional(content.([]byte)))).
			Decode(pivotCache); err != nil && err != io.EOF {
			return nil, err
		}
	}
	return pivotCache, nil
}

// extractPivotTableFields provides a function to extract all pivot table fields
// settings by given pivot table fields.
func (f *File) extractPivotTableFields(order []string, pt *xlsxPivotTableDefinition, opts *PivotTableOptions) {
	for fieldIdx, field := range pt.PivotFields.PivotField {
		if field.Axis == "axisRow" {
			opts.Rows = append(opts.Rows, extractPivotTableField(order[fieldIdx], field))
		}
		if field.Axis == "axisCol" {
			opts.Columns = append(opts.Columns, extractPivotTableField(order[fieldIdx], field))
		}
		if field.Axis == "axisPage" {
			opts.Filter = append(opts.Filter, extractPivotTableField(order[fieldIdx], field))
		}
	}
	if pt.DataFields != nil {
		for _, field := range pt.DataFields.DataField {
			opts.Data = append(opts.Data, PivotTableField{
				Data:     order[field.Fld],
				Name:     field.Name,
				Subtotal: cases.Title(language.English).String(field.Subtotal),
			})
		}
	}
}

// extractPivotTableField provides a function to extract pivot table field
// settings by given pivot table fields.
func extractPivotTableField(data string, fld *xlsxPivotField) PivotTableField {
	pivotTableField := PivotTableField{
		Data: data,
	}
	fields := []string{"Compact", "Name", "Outline", "Subtotal", "DefaultSubtotal"}
	immutable, mutable := reflect.ValueOf(*fld), reflect.ValueOf(&pivotTableField).Elem()
	for _, field := range fields {
		immutableField := immutable.FieldByName(field)
		if immutableField.Kind() == reflect.String {
			mutable.FieldByName(field).SetString(immutableField.String())
		}
		if immutableField.Kind() == reflect.Ptr && !immutableField.IsNil() && immutableField.Elem().Kind() == reflect.Bool {
			mutable.FieldByName(field).SetBool(immutableField.Elem().Bool())
		}
	}
	return pivotTableField
}

// genPivotCacheDefinitionID generates a unique pivot table cache definition ID.
func (f *File) genPivotCacheDefinitionID() int {
	var (
		ID                            int
		decodeExtLst                  = new(decodeExtLst)
		decodeX14PivotCacheDefinition = new(decodeX14PivotCacheDefinition)
	)
	f.Pkg.Range(func(k, v interface{}) bool {
		if strings.Contains(k.(string), "xl/pivotCache/pivotCacheDefinition") {
			pc, err := f.pivotCacheReader(k.(string))
			if err != nil {
				return true
			}
			if pc.ExtLst != nil {
				_ = f.xmlNewDecoder(strings.NewReader("<extLst>" + pc.ExtLst.Ext + "</extLst>")).Decode(decodeExtLst)
				for _, ext := range decodeExtLst.Ext {
					if ext.URI == ExtURIPivotCacheDefinition {
						_ = f.xmlNewDecoder(strings.NewReader(ext.Content)).Decode(decodeX14PivotCacheDefinition)
						if ID < decodeX14PivotCacheDefinition.PivotCacheID {
							ID = decodeX14PivotCacheDefinition.PivotCacheID
						}
					}
				}
			}
		}
		return true
	})
	return ID + 1
}

// deleteWorkbookPivotCache remove workbook pivot cache and pivot cache
// relationships.
func (f *File) deleteWorkbookPivotCache(opt PivotTableOptions) error {
	rID, err := f.deleteWorkbookRels(SourceRelationshipPivotCache, strings.TrimPrefix(strings.TrimPrefix(opt.pivotCacheXML, "/"), "xl/"))
	if err != nil {
		return err
	}
	wb, err := f.workbookReader()
	if err != nil {
		return err
	}
	if wb.PivotCaches != nil {
		for i, pivotCache := range wb.PivotCaches.PivotCache {
			if pivotCache.RID == rID {
				wb.PivotCaches.PivotCache = append(wb.PivotCaches.PivotCache[:i], wb.PivotCaches.PivotCache[i+1:]...)
			}
		}
		if len(wb.PivotCaches.PivotCache) == 0 {
			wb.PivotCaches = nil
		}
	}
	return err
}

// DeletePivotTable delete a pivot table by giving the worksheet name and pivot
// table name. Note that this function does not clean cell values in the pivot
// table range.
func (f *File) DeletePivotTable(sheet, name string) error {
	sheetXML, ok := f.getSheetXMLPath(sheet)
	if !ok {
		return ErrSheetNotExist{sheet}
	}
	rels := "xl/worksheets/_rels/" + strings.TrimPrefix(sheetXML, "xl/worksheets/") + ".rels"
	sheetRels, err := f.relsReader(rels)
	if err != nil {
		return err
	}
	if sheetRels == nil {
		sheetRels = &xlsxRelationships{}
	}
	opts, err := f.GetPivotTables(sheet)
	if err != nil {
		return err
	}
	pivotTableCaches := map[string]int{}
	for _, sheetName := range f.GetSheetList() {
		sheetPivotTables, _ := f.GetPivotTables(sheetName)
		for _, sheetPivotTable := range sheetPivotTables {
			pivotTableCaches[sheetPivotTable.pivotCacheXML]++
		}
	}
	for _, v := range sheetRels.Relationships {
		for _, opt := range opts {
			if v.Type == SourceRelationshipPivotTable {
				pivotTableXML := strings.ReplaceAll(v.Target, "..", "xl")
				if opt.Name == name && opt.pivotTableXML == pivotTableXML {
					if pivotTableCaches[opt.pivotCacheXML] == 1 {
						err = f.deleteWorkbookPivotCache(opt)
					}
					f.deleteSheetRelationships(sheet, v.ID)
					return err
				}
			}
		}
	}
	return newNoExistTableError(name)
}
