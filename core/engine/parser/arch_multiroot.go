package parser

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// archFormRoot matches ir.ui.view arch rooted at <form>.
type archFormRoot struct {
	XMLName xml.Name `xml:"form"`
	String  string   `xml:"string,attr"`
	Header  *Header  `xml:"header"`
	Sheet   *Sheet   `xml:"sheet"`
	Footer  *Footer  `xml:"footer"`
	Chatter *Chatter `xml:"chatter"`
	Field   []Field  `xml:"field"`
	Group   []Group  `xml:"group"`
}

type archTreeRoot struct {
	XMLName xml.Name `xml:"tree"`
	String  string   `xml:"string,attr"`
	Field   []Field  `xml:"field"`
}

type archKanbanRoot struct {
	XMLName xml.Name `xml:"kanban"`
	Field   []Field  `xml:"field"`
}

func parseViewFromArchInternal(arch string) (*View, error) {
	if arch == "" {
		return nil, fmt.Errorf("empty view arch")
	}

	var v View
	if err := xml.Unmarshal([]byte(arch), &v); err == nil && viewLooksPopulated(&v) {
		return &v, nil
	}

	var f archFormRoot
	if err := xml.Unmarshal([]byte(arch), &f); err == nil && formArchHasContent(&f) {
		return &View{
			Type:    "form",
			Header:  f.Header,
			Sheet:   f.Sheet,
			Footer:  f.Footer,
			Chatter: f.Chatter,
			Field:   f.Field,
			Group:   f.Group,
		}, nil
	}

	var t archTreeRoot
	if err := xml.Unmarshal([]byte(arch), &t); err == nil && strings.HasPrefix(strings.ToLower(strings.TrimSpace(arch)), "<tree") {
		return &View{Type: "tree", Field: t.Field}, nil
	}

	var k archKanbanRoot
	if err := xml.Unmarshal([]byte(arch), &k); err == nil && strings.HasPrefix(strings.ToLower(strings.TrimSpace(arch)), "<kanban") {
		return &View{Type: "kanban", Field: k.Field}, nil
	}

	if err := xml.Unmarshal([]byte(arch), &v); err != nil {
		return nil, fmt.Errorf("parse view arch: %w", err)
	}
	if !viewLooksPopulated(&v) {
		return nil, fmt.Errorf("parse view arch: unsupported or empty root (use <view>, <form>, <tree>, or <kanban>)")
	}
	return &v, nil
}

func viewLooksPopulated(v *View) bool {
	if v == nil {
		return false
	}
	return v.Type != "" || v.Header != nil || v.Sheet != nil || v.Footer != nil || v.Chatter != nil ||
		len(v.Field) > 0 || len(v.Group) > 0
}

func formArchHasContent(f *archFormRoot) bool {
	return f != nil && (f.Header != nil || f.Sheet != nil || f.Footer != nil || f.Chatter != nil ||
		len(f.Field) > 0 || len(f.Group) > 0)
}
