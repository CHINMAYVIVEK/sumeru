package module

import (
	"sync"
	"sumeru/core/engine/parser"
)

var (
	// LoadedAddons contains addons that are both discovered and fully loaded (metadata/rows).
	LoadedAddons     = map[string]*Addon{}
	// DiscoveredAddons contains all addons found on the filesystem during discovery.
	DiscoveredAddons = map[string]*Addon{}
	// MenuRegistry contains all menu items parsed from XML.
	MenuRegistry     = []parser.MenuItem{}
	// installMu ensures serial execution of module install/uninstall/update operations.
	installMu        sync.Mutex
)
