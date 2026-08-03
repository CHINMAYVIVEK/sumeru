package parser

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

// moduleDataWrapper matches an optional <data noupdate="..."> wrapper inside the module root.
type moduleDataWrapper struct {
	NoUpdate  string     `xml:"noupdate,attr"`
	Records   []Record   `xml:"record"`
	Views     []View     `xml:"view"`
	MenuItems []MenuItem `xml:"menuitem"`
	Actions   []Action   `xml:"action"`
}

// Action represents a simplified <action> tag for actions like sys.action.window.
type Action struct {
	XMLName  xml.Name `xml:"action"`
	ID       string   `xml:"id,attr"`
	Type     string   `xml:"type,attr"`
	Model    string   `xml:"model,attr"`
	Name     string   `xml:"name,attr"`
	ViewMode string   `xml:"view_mode,attr"`
	Help     string   `xml:"help,innerxml"`
}

// ToRecord converts an Action to a Record for backward compatibility.
func (a Action) ToRecord() Record {
	fields := []RecordField{
		{Name: "name", Body: a.Name},
		{Name: "core_model", Body: a.Model},
		{Name: "view_mode", Body: a.ViewMode},
	}
	if a.Help != "" {
		fields = append(fields, RecordField{Name: "help", Type: "html", Body: a.Help})
	}
	return Record{
		ID:    a.ID,
		Model: "sys.action.window",
		Field: fields,
	}
}

// ValidateModuleRoot returns an error if the module data XML root element is not <sumeru>.
func ValidateModuleRoot(n xml.Name) error {
	switch strings.ToLower(strings.TrimSpace(n.Local)) {
	case "sumeru":
		return nil
	default:
		return fmt.Errorf("module XML root must be <sumeru>, got <%s>", n.Local)
	}
}

// MergeViewListData flattens <data> children into the top-level slices.
func (v *ViewList) MergeViewListData() {
	if v.Data == nil {
		return
	}
	v.NoUpdate = parseNoUpdateFlag(v.Data.NoUpdate)
	v.Records = append(v.Data.Records, v.Records...)
	v.Views = append(v.Data.Views, v.Views...)
	for _, a := range v.Data.Actions {
		v.Records = append(v.Records, a.ToRecord())
	}
	v.Actions = append(v.Data.Actions, v.Actions...)
	v.Data = nil
}

// MergeMenuListData flattens <data> menuitem children.
func (m *MenuList) MergeMenuListData() {
	if m.Data == nil {
		return
	}
	m.NoUpdate = parseNoUpdateFlag(m.Data.NoUpdate)
	m.MenuItems = append(m.Data.MenuItems, m.MenuItems...)
	m.Records = append(m.Data.Records, m.Records...)
	for _, a := range m.Data.Actions {
		m.Records = append(m.Records, a.ToRecord())
	}
	m.Actions = append(m.Data.Actions, m.Actions...)
	m.Data = nil
}

func parseNoUpdateFlag(raw string) bool {
	s := strings.TrimSpace(strings.ToLower(raw))
	return s == "1" || s == "true" || s == "yes"
}

// PeekModuleXMLRootName returns the local name of the first start element (e.g. sumeru).
func PeekModuleXMLRootName(data []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(bytes.TrimSpace(data)))
	for {
		t, err := dec.Token()
		if err != nil {
			return "", err
		}
		if se, ok := t.(xml.StartElement); ok {
			return se.Name.Local, nil
		}
	}
}
