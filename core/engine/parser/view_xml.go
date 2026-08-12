package parser

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

type View struct {
	XMLName xml.Name `xml:"view"`
	ID      string   `xml:"id,attr"`
	Model   string   `xml:"model,attr"`
	Type    string   `xml:"type,attr"`
	Title   string   `xml:"title,attr"`
	// TreeOpenAttr is the raw <tree open="..."/> or <view type="tree" open="..."/> attribute (false/0/off disables row→form).
	TreeOpenAttr string `xml:"open,attr"`
	// TreeNoRowOpen is derived from TreeOpenAttr by the arch parser for type tree/list.
	TreeNoRowOpen bool     `xml:"-"`
	Header        *Header  `xml:"header"`
	Sheet         *Sheet   `xml:"sheet"`
	Footer        *Footer  `xml:"footer"`
	Chatter       *Chatter `xml:"chatter"`
	Field         []Field  `xml:"field"`
	Group         []Group  `xml:"group"`
	// Form is a nested <form> under <view>; promoted onto View then cleared.
	Form *archFormRoot `xml:"form"`
}

type Header struct {
	Button []Button `xml:"button"`
	Field  []Field  `xml:"field"`
}

type Sheet struct {
	Div       []Div       `xml:"div"`
	Group     []Group     `xml:"group"`
	Notebook  []Notebook  `xml:"notebook"`
	Field     []Field     `xml:"field"`
	Separator []Separator `xml:"separator"`
	Label     []Label     `xml:"label"`
}

type Notebook struct {
	Page []Page `xml:"page"`
}

type Page struct {
	Title     string      `xml:"string,attr"`
	Field     []Field     `xml:"field"`
	Group     []Group     `xml:"group"`
	Separator []Separator `xml:"separator"`
	Label     []Label     `xml:"label"`
}

type Button struct {
	Name   string `xml:"name,attr"`
	String string `xml:"string,attr"`
	Type   string `xml:"type,attr"`
	Class  string `xml:"class,attr"`
}

type Chatter struct {
	Field []Field `xml:"field"`
}

type Div struct {
	Class string  `xml:"class,attr"`
	Field []Field `xml:"field"`
	H1    []H1    `xml:"h1"`
}

type H1 struct {
	Field []Field `xml:"field"`
}

type Field struct {
	Name        string `xml:"name,attr"`
	Label       string `xml:"string,attr"`
	Widget      string `xml:"widget,attr"`
	Placeholder string `xml:"placeholder,attr"`
	Options     string `xml:"options,attr"`
	Groups      string `xml:"groups,attr"`
}

type Group struct {
	Title     string      `xml:"string,attr"`
	Field     []Field     `xml:"field"`
	Group     []Group     `xml:"group"`
	Separator []Separator `xml:"separator"`
	Label     []Label     `xml:"label"`
}

// Footer is the form footer (action buttons).
type Footer struct {
	Button []Button `xml:"button"`
}

// Separator is a visual section break (<separator/>).
type Separator struct {
	String string `xml:"string,attr"`
}

// Label is a static label optionally tied to a field (<label for="..."/>).
type Label struct {
	For    string `xml:"for,attr"`
	String string `xml:"string,attr"`
}

type Menu struct {
	ID       string `xml:"id,attr"`
	Name     string `xml:"name,attr"`
	ParentID string `xml:"parent,attr"`
	Action   string `xml:"action,attr"`
	Sequence int    `xml:"sequence,attr"`
}

type Record struct {
	ID    string        `xml:"id,attr"`
	Model string        `xml:"model,attr"`
	Field []RecordField `xml:"field"`
}

type MenuItem struct {
	ID           string `xml:"id,attr"`
	Name         string `xml:"name,attr"`
	ParentID     string `xml:"parent,attr"`
	Action       string `xml:"action,attr"`
	Sequence     int    `xml:"sequence,attr"`
	WebIcon      string `xml:"web_icon,attr"`
	AccessGroups string `xml:"groups,attr"`
	Module       string `xml:"-"` // set from DB rows only (not in XML menu files)
}

type ViewList struct {
	XMLName   xml.Name           `xml:"sumeru"`
	NoUpdate  bool               `xml:"-"`
	Data      *moduleDataWrapper `xml:"data"`
	Records   []Record           `xml:"record"`
	Views     []View             `xml:"view"`
	MenuItems []MenuItem         `xml:"menuitem"`
	Actions   []Action           `xml:"action"`
}

type MenuList struct {
	XMLName   xml.Name           `xml:"sumeru"`
	NoUpdate  bool               `xml:"-"`
	Data      *moduleDataWrapper `xml:"data"`
	MenuItems []MenuItem         `xml:"menuitem"`
	Records   []Record           `xml:"record"`
	Actions   []Action           `xml:"action"`
}

func ParseViewList(filePath string) (*ViewList, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	root, err := PeekModuleXMLRootName(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filePath, err)
	}
	if err := ValidateModuleRoot(xml.Name{Local: root}); err != nil {
		return nil, fmt.Errorf("%s: %w", filePath, err)
	}
	var viewList ViewList
	if err := xml.Unmarshal(data, &viewList); err != nil {
		return nil, err
	}
	viewList.MergeViewListData()
	for i := range viewList.Views {
		promoteNestedForm(&viewList.Views[i])
	}
	return &viewList, nil
}

func ParseViewFromArch(arch string) (*View, error) {
	v, err := parseViewFromArchInternal(strings.TrimSpace(arch))
	if err != nil {
		return nil, err
	}
	v.Type = strings.ToLower(strings.TrimSpace(v.Type))
	return v, nil
}

func ParseMenuList(filePath string) (*MenuList, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	root, err := PeekModuleXMLRootName(data)
	if err != nil {
		return nil, err
	}
	if err := ValidateModuleRoot(xml.Name{Local: root}); err != nil {
		return nil, err
	}
	var menuList MenuList
	if err := xml.Unmarshal(data, &menuList); err != nil {
		return nil, err
	}
	menuList.MergeMenuListData()
	return &menuList, nil
}
