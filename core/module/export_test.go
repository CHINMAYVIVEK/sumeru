package module

import "context"

// ExpandInstallModuleNamesForTest exposes expandInstallModuleNames for tests.
func ExpandInstallModuleNamesForTest(ctx context.Context, parts []string) ([]string, error) {
	return expandInstallModuleNames(ctx, parts)
}
