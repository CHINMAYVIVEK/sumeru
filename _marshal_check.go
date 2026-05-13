package main

import (
	"encoding/xml"
	"fmt"
	"sumeru/core/engine/parser"
)

func main() {
	v := parser.View{Model: "sale.order", Type: "tree", Field: []parser.Field{{Name: "state", Label: "Status"}}}
	b, _ := xml.Marshal(v)
	fmt.Println(string(b))
}
