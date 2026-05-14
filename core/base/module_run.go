package base

import "sumeru/core/module"

// LoadAddonPathsInput lists absolute addon root directories (order matters for overrides).
type LoadAddonPathsInput struct {
	Paths []string
}

// LoadAddonPaths discovers addons, syncs sys.module, and loads installed module data.
func LoadAddonPaths(in LoadAddonPathsInput) error {
	return module.LoadAddonPaths(in.Paths)
}

// RunModuleCLIInput carries -i / -u comma-separated module lists.
type RunModuleCLIInput struct {
	Install string
	Update  string
}

// RunModuleCLI runs install then update for the given module lists.
func RunModuleCLI(in RunModuleCLIInput) error {
	return module.RunModuleCLI(in.Install, in.Update)
}
