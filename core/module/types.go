package module

import (
	"sumeru/core/engine/parser"
)

const KernelModule = "base"

type Manifest struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"` // optional; Apps / shell label
	Version     string   `json:"version"`
	Depends     []string `json:"depends"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Data        []string `json:"data"`        // XML files to load
	Application *bool    `json:"application"` // nil = true (show in Apps)
}

func (manifest *Manifest) IsApplication() bool {
	if manifest.Application == nil {
		return true
	}
	return *manifest.Application
}

type Addon struct {
	Manifest Manifest
	Path     string
	Menus    []parser.MenuItem
}
