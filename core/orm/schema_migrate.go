package orm

// Legacy schema fixes that are not expressed as additive columns on Model.Fields():
// - widen sys.view.arch to TEXT
// - composite index on mail.message
// - backfill sys.menu.module from sys.model_data
//
// Registered model columns are added by SyncRegistrySchema (schema_sync.go), invoked on
// server startup and on module install (-i) / update (-u).

// BackfillSysMenuModule sets sys.menu.module from sys.model_data for rows missing it.
func BackfillSysMenuModule() error {
	if DB == nil {
		return nil
	}
	_, err := DB.Exec(`
		UPDATE sys.menu m
		SET module = d.module
		FROM sys.model_data d
		WHERE d.model = 'sys.menu' AND d.core_id = m.id
		  AND (m.module IS NULL OR TRIM(m.module) = '')
	`)
	return err
}

// EnsureSysViewArchText widens sys.view.arch for large inherited views.
func EnsureSysViewArchText() error {
	if DB == nil {
		return nil
	}
	tn := GetTableName("sys.view")
	_, err := DB.Exec(`ALTER TABLE ` + tn + ` ALTER COLUMN arch TYPE TEXT USING arch::text`)
	return err
}

// EnsureMailMessageModelResIndex adds a composite index for chatter and activity queries.
func EnsureMailMessageModelResIndex() error {
	if DB == nil {
		return nil
	}
	tn := GetTableName("mail.message")
	q := `CREATE INDEX IF NOT EXISTS idx_` + tn + `_model_core_created ON ` + tn + ` (model, core_id, create_date DESC)`
	_, err := DB.Exec(q)
	return err
}
