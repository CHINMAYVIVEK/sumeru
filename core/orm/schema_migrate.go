package orm

// Legacy schema fixes that are not expressed as additive columns on Model.Fields():
// - widen ir.ui.view.arch to TEXT
// - composite index on mail.message
// - backfill ir_ui_menu.module from ir_model_data
//
// Registered model columns are added by SyncRegistrySchema (schema_sync.go), invoked on
// server startup and on module install (-i) / update (-u).

// BackfillIrUiMenuModule sets ir_ui_menu.module from ir_model_data for rows missing it.
func BackfillIrUiMenuModule() error {
	if DB == nil {
		return nil
	}
	_, err := DB.Exec(`
		UPDATE ir_ui_menu m
		SET module = d.module
		FROM ir_model_data d
		WHERE d.model = 'ir.ui.menu' AND d.res_id = m.id
		  AND (m.module IS NULL OR TRIM(m.module) = '')
	`)
	return err
}

// EnsureIrUiViewArchText widens ir.ui.view.arch for large inherited views.
func EnsureIrUiViewArchText() error {
	if DB == nil {
		return nil
	}
	tn := GetTableName("ir.ui.view")
	_, err := DB.Exec(`ALTER TABLE ` + tn + ` ALTER COLUMN arch TYPE TEXT USING arch::text`)
	return err
}

// EnsureMailMessageModelResIndex adds a composite index for chatter and activity queries.
func EnsureMailMessageModelResIndex() error {
	if DB == nil {
		return nil
	}
	tn := GetTableName("mail.message")
	q := `CREATE INDEX IF NOT EXISTS idx_` + tn + `_model_res_created ON ` + tn + ` (model, res_id, create_date DESC)`
	_, err := DB.Exec(q)
	return err
}
