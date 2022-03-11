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

import "encoding/xml"

// vmlDrawing directly maps the root element in the file
// xl/drawings/vmlDrawing%d.vml.
type vmlDrawing struct {
	XMLName     xml.Name         `xml:"xml"`
	XMLNSv      string           `xml:"xmlns:v,attr"`
	XMLNSo      string           `xml:"xmlns:o,attr"`
	XMLNSx      string           `xml:"xmlns:x,attr"`
	XMLNSmv     string           `xml:"xmlns:mv,attr"`
	Shapelayout *xlsxShapelayout `xml:"o:shapelayout"`
	Shapetype   *xlsxShapetype   `xml:"v:shapetype"`
	Shape       []xlsxShape      `xml:"v:shape"`
}

// xlsxShapelayout directly maps the shapelayout element. This element contains
// child elements that store information used in the editing and layout of
// shapes.
type xlsxShapelayout struct {
	Ext   string     `xml:"v:ext,attr"`
	IDmap *xlsxIDmap `xml:"o:idmap"`
}

// xlsxIDmap directly maps the idmap element.
type xlsxIDmap struct {
	Ext  string `xml:"v:ext,attr"`
	Data int    `xml:"data,attr"`
}

// xlsxShape directly maps the shape element.
type xlsxShape struct {
	XMLName   xml.Name `xml:"v:shape"`
	ID        string   `xml:"id,attr"`
	Type      string   `xml:"type,attr"`
	Style     string   `xml:"style,attr"`
	Fillcolor string   `xml:"fillcolor,attr"`
	// TODO: solve conflict
	// Insetmode   string   `xml:"urn:schemas-microsoft-com:office:office insetmode,attr,omitempty"`
	Insetmode   string `xml:"o:insetmode,attr"`
	Strokecolor string `xml:"strokecolor,attr,omitempty"`
	Button      string `xml:"o:button,attr"`
	Val         string `xml:",innerxml"`
}

type vFillButton struct {
	Color2           string `xml:"color2,attr"`
	Detectmouseclick string `xml:"o:detectmouseclick,attr"`
}

type oLockButton struct {
	Ext      string `xml:"v:ext,attr"`
	Rotation string `xml:"rotation,attr"`
}

type vTextboxButton struct {
	Style       string         `xml:"v:ext,attr"`
	Singleclick string         `xml:"o:singleclick,attr"`
	Div         *xlsxDivButton `xml:"div"`
}
type xlsxDivButton struct {
	Style string      `xml:"style,attr"`
	Font  *fontButton `xml:"font"`
}

type fontButton struct {
	Face    string `xml:"face,attr"`
	Size    string `xml:"size,attr"`
	Color   string `xml:"color,attr"`
	Caption string `xml:",chardata"`
}

type xClientDataButton struct {
	ObjectType  string `xml:"ObjectType,attr"`
	Anchor      string `xml:"x:Anchor"`
	PrintObject string `xml:"x:PrintObject"`
	AutoFill    string `xml:"x:AutoFill"`
	FmlaMacro   string `xml:"x:FmlaMacro"`
	TextHAlign  string `xml:"x:TextHAlign"`
	TextVAlign  string `xml:"x:TextVAlign"`
}

type encodeShapeButton struct {
	Fill       *vFillButton       `xml:"v:fill"`
	Lock       *oLockButton       `xml:"o:lock"`
	TextBox    *vTextboxButton    `xml:"v:textbox"`
	ClientData *xClientDataButton `xml:"x:ClientData"`
}

// xlsxShapetype directly maps the shapetype element.
type xlsxShapetype struct {
	ID        string      `xml:"id,attr"`
	Coordsize string      `xml:"coordsize,attr"`
	Spt       int         `xml:"o:spt,attr"`
	Path      string      `xml:"path,attr"`
	Stroke    *xlsxStroke `xml:"v:stroke"`
	VPath     *vPath      `xml:"v:path"`
}

// xlsxStroke directly maps the stroke element.
type xlsxStroke struct {
	Joinstyle string `xml:"joinstyle,attr"`
}

// vPath directly maps the v:path element.
type vPath struct {
	Gradientshapeok string `xml:"gradientshapeok,attr,omitempty"`
	Connecttype     string `xml:"o:connecttype,attr"`
}

// vFill directly maps the v:fill element. This element must be defined within a
// Shape element.
type vFill struct {
	Angle  int    `xml:"angle,attr,omitempty"`
	Color2 string `xml:"color2,attr"`
	Type   string `xml:"type,attr,omitempty"`
	Fill   *oFill `xml:"o:fill"`
}

// oFill directly maps the o:fill element.
type oFill struct {
	Ext  string `xml:"v:ext,attr"`
	Type string `xml:"type,attr,omitempty"`
}

// vShadow directly maps the v:shadow element. This element must be defined
// within a Shape element. In addition, the On attribute must be set to True.
type vShadow struct {
	On       string `xml:"on,attr"`
	Color    string `xml:"color,attr,omitempty"`
	Obscured string `xml:"obscured,attr"`
}

// vTextbox directly maps the v:textbox element. This element must be defined
// within a Shape element.
type vTextbox struct {
	Style string   `xml:"style,attr"`
	Div   *xlsxDiv `xml:"div"`
}

// xlsxDiv directly maps the div element.
type xlsxDiv struct {
	Style string `xml:"style,attr"`
}

// xClientData (Attached Object Data) directly maps the x:ClientData element.
// This element specifies data associated with objects attached to a
// spreadsheet. While this element might contain any of the child elements
// below, only certain combinations are meaningful. The ObjectType attribute
// determines the kind of object the element represents and which subset of
// child elements is appropriate. Relevant groups are identified for each child
// element.
type xClientData struct {
	ObjectType    string `xml:"ObjectType,attr"`
	MoveWithCells string `xml:"x:MoveWithCells,omitempty"`
	SizeWithCells string `xml:"x:SizeWithCells,omitempty"`
	Anchor        string `xml:"x:Anchor"`
	AutoFill      string `xml:"x:AutoFill"`
	Row           int    `xml:"x:Row"`
	Column        int    `xml:"x:Column"`
}

// decodeVmlDrawing defines the structure used to parse the file
// xl/drawings/vmlDrawing%d.vml.
type decodeVmlDrawing struct {
	Shape []decodeShape `xml:"urn:schemas-microsoft-com:vml shape"`
}

// decodeShape defines the structure used to parse the particular shape element.
type decodeShape struct {
	Val string `xml:",innerxml"`
}

// encodeShape defines the structure used to re-serialization shape element.
type encodeShape struct {
	Fill       *vFill       `xml:"v:fill"`
	Shadow     *vShadow     `xml:"v:shadow"`
	Path       *vPath       `xml:"v:path"`
	Textbox    *vTextbox    `xml:"v:textbox"`
	ClientData *xClientData `xml:"x:ClientData"`
}
