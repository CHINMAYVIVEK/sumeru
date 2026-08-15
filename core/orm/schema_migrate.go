package orm

import (
	"errors"
	"fmt"
	"strings"
)

// One-off database repairs invoked at server startup (see run.go).
//
// SyncRegistrySchema handles the normal path: add columns declared on models.
// This file covers everything else — widen a column type, fix corrupt rows,
// add an index, or upgrade data left over from an older convention.
//
// Every function here is idempotent and safe to run on each boot.

// --- Menu data ---

// RunMenuDataFixes repairs sys.menu rows that break navigation if left alone.
func RunMenuDataFixes() error {
	var errs []error
	if err := backfillSysMenuModule(); err != nil {
		errs = append(errs, fmt.Errorf("backfill sys.menu.module: %w", err))
	}
	if err := fixSysMenuSelfParent(); err != nil {
		errs = append(errs, fmt.Errorf("fix sys.menu self-parent: %w", err))
	}
	return errors.Join(errs...)
}

// Older installs never stored which addon owns a menu row. The XML-id registry
// (sys.model.data) has that mapping — copy it across where module is blank.
func backfillSysMenuModule() error {
	if DB == nil {
		return nil
	}
	ok, err := tableExists(MustModelToTableName("sys.model.data"))
	if err != nil || !ok {
		return err
	}
	menuTbl := MustQuotedTableName("sys.menu")
	dataTbl := MustQuotedTableName("sys.model.data")
	_, err = DB.Exec(`UPDATE ` + menuTbl + ` m
		SET module = d.module
		FROM ` + dataTbl + ` d
		WHERE d.model = 'sys.menu' AND d.core_id = m.id
		  AND (m.module IS NULL OR TRIM(m.module) = '')`)
	return err
}

// A menu pointing at itself as parent never surfaces in the top bar. Clear it.
func fixSysMenuSelfParent() error {
	if DB == nil {
		return nil
	}
	menuTbl := MustQuotedTableName("sys.menu")
	_, err := DB.Exec(`UPDATE ` + menuTbl + ` SET parent_id = NULL WHERE id = parent_id AND parent_id IS NOT NULL`)
	return err
}

// --- Column widenings ---

// EnsureSysViewArchText promotes sys.view.arch to TEXT so large inherited views fit.
func EnsureSysViewArchText() error {
	return widenColumnToText("sys.view", "arch")
}

// EnsureCoreUserImageText promotes core.user.image to TEXT (Char is VARCHAR(255),
// which is too small for typical base64 avatar payloads).
func EnsureCoreUserImageText() error {
	if DB == nil {
		return nil
	}
	if ok, err := tableExists(MustModelToTableName("core.user")); err != nil || !ok {
		return err
	}
	return widenColumnToText("core.user", "image")
}

func widenColumnToText(modelName, column string) error {
	if DB == nil {
		return nil
	}
	table := MustQuotedTableName(modelName)
	_, err := DB.Exec(`ALTER TABLE ` + table + ` ALTER COLUMN ` + column + ` TYPE TEXT USING ` + column + `::text`)
	return err
}

// --- View type: tree → list ---

// Replacements applied when upgrading stored view arch from the old "tree"
// convention to "list". MigrateSysViewTreeToList uses the same rules in SQL.
var legacyTreeToListArchReplacements = [][2]string{
	{"<tree", "<list"},
	{"</tree>", "</list>"},
	{`type="tree"`, `type="list"`},
	{`type='tree'`, `type='list'`},
}

// RewriteViewArchTreeToList applies legacyTreeToListArchReplacements to a single arch string.
func RewriteViewArchTreeToList(arch string) string {
	for _, pair := range legacyTreeToListArchReplacements {
		arch = strings.ReplaceAll(arch, pair[0], pair[1])
	}
	return arch
}

// MigrateSysViewTreeToList upgrades sys.view rows still on the pre-list view type.
// Fresh installs and module syncs already write type=list; this is for existing DBs.
func MigrateSysViewTreeToList() error {
	if DB == nil {
		return nil
	}
	if ok, err := tableExists(MustModelToTableName("sys.view")); err != nil || !ok {
		return err
	}
	viewTable := MustQuotedTableName("sys.view")

	if _, err := DB.Exec(`UPDATE ` + viewTable + ` SET type = 'list' WHERE type = 'tree'`); err != nil {
		return err
	}

	// Nested REPLACE mirrors legacyTreeToListArchReplacements (innermost first).
	archExpr := "arch"
	for i := len(legacyTreeToListArchReplacements) - 1; i >= 0; i-- {
		old, new := legacyTreeToListArchReplacements[i][0], legacyTreeToListArchReplacements[i][1]
		archExpr = fmt.Sprintf("REPLACE(%s, '%s', '%s')", archExpr, quoteSQLString(old), quoteSQLString(new))
	}

	_, err := DB.Exec(`UPDATE ` + viewTable + ` SET arch = ` + archExpr + `
		WHERE arch LIKE '%<tree%' OR arch LIKE '%</tree>%'
		   OR arch LIKE '%type="tree"%' OR arch LIKE '%type=''tree''%'`)
	return err
}

// --- Indexes ---

// EnsureMailMessageModelResIndex speeds up chatter and activity lookups by (model, record, time).
func EnsureMailMessageModelResIndex() error {
	if DB == nil {
		return nil
	}
	tablePhysical := MustModelToTableName("mail.message")
	tableQuoted := MustQuotedTableName("mail.message")
	indexName := "idx_" + tablePhysical + "_model_core_created"
	_, err := DB.Exec(`CREATE INDEX IF NOT EXISTS ` + indexName + ` ON ` + tableQuoted + ` (model, core_id, create_date DESC)`)
	return err
}
