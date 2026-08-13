package orm

import (
	"errors"
	"fmt"
)

// Legacy schema fixes that are not expressed as additive columns on Model.Fields():
// - widen sys.view.arch to TEXT
// - composite index on mail.message
// - backfill sys.menu.module from the XML-id registry table (sys_model_data)
//
// Registered model columns are added by SyncRegistrySchema (schema_sync.go), invoked on
// server startup and on module install (-i) / update (-u).

// BackfillSysMenuModule sets sys.menu.module from sys_model_data for rows missing it.
func BackfillSysMenuModule() error {
	if DB == nil {
		return nil
	}
	ok, err := tableExists(MustModelToTableName("sys.model.data"))
	if err != nil || !ok {
		return err
	}
	menuTbl := MustQuotedTableName("sys.menu")
	dataTbl := MustQuotedTableName("sys.model.data")
	q := `UPDATE ` + menuTbl + ` m
		SET module = d.module
		FROM ` + dataTbl + ` d
		WHERE d.model = 'sys.menu' AND d.core_id = m.id
		  AND (m.module IS NULL OR TRIM(m.module) = '')`
	_, err = DB.Exec(q)
	return err
}

// FixSysMenuSelfParent clears parent_id when it incorrectly equals the row id (bad data hides
// roots from the shell top bar). Safe to run on every startup.
func FixSysMenuSelfParent() error {
	if DB == nil {
		return nil
	}
	menuTbl := MustQuotedTableName("sys.menu")
	_, err := DB.Exec(`UPDATE ` + menuTbl + ` SET parent_id = NULL WHERE id = parent_id AND parent_id IS NOT NULL`)
	return err
}

// RunMenuDataFixes applies startup/setup menu repairs (module backfill + self-parent cleanup).
func RunMenuDataFixes() error {
	var errs []error
	if err := BackfillSysMenuModule(); err != nil {
		errs = append(errs, fmt.Errorf("backfill sys.menu.module: %w", err))
	}
	if err := FixSysMenuSelfParent(); err != nil {
		errs = append(errs, fmt.Errorf("fix sys.menu self-parent: %w", err))
	}
	return errors.Join(errs...)
}

// EnsureSysViewArchText widens sys.view.arch for large inherited views.
func EnsureSysViewArchText() error {
	if DB == nil {
		return nil
	}
	tn := MustQuotedTableName("sys.view")
	_, err := DB.Exec(`ALTER TABLE ` + tn + ` ALTER COLUMN arch TYPE TEXT USING arch::text`)
	return err
}

// EnsureCoreUserImageText widens core.user.image so data-URL avatars can be stored
// (Char maps to VARCHAR(255), which rejects typical base64 payloads).
func EnsureCoreUserImageText() error {
	if DB == nil {
		return nil
	}
	tnPhysical := MustModelToTableName("core.user")
	ok, err := tableExists(tnPhysical)
	if err != nil || !ok {
		return err
	}
	tn := MustQuotedTableName("core.user")
	_, err = DB.Exec(`ALTER TABLE ` + tn + ` ALTER COLUMN image TYPE TEXT USING image::text`)
	return err
}

// EnsureMailMessageModelResIndex adds a composite index for chatter and activity queries.
func EnsureMailMessageModelResIndex() error {
	if DB == nil {
		return nil
	}
	tnPhysical := MustModelToTableName("mail.message") // physical name for index id; quote only in ON clause
	tn := MustQuotedTableName("mail.message")
	q := `CREATE INDEX IF NOT EXISTS idx_` + tnPhysical + `_model_core_created ON ` + tn + ` (model, core_id, create_date DESC)`
	_, err := DB.Exec(q)
	return err
}
