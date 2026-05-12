package orm

// EnsureIrUiMenuModuleColumn adds the module column when upgrading an older DB
// where ir_ui_menu was created before the field existed.
func EnsureIrUiMenuModuleColumn() error {
	tn := GetTableName("ir.ui.menu")
	_, err := DB.Exec("ALTER TABLE " + tn + " ADD COLUMN IF NOT EXISTS module VARCHAR(255)")
	return err
}

// BackfillIrUiMenuModule sets ir_ui_menu.module from ir_model_data for rows missing it.
func BackfillIrUiMenuModule() error {
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
	tn := GetTableName("ir.ui.view")
	_, err := DB.Exec(`ALTER TABLE ` + tn + ` ALTER COLUMN arch TYPE TEXT USING arch::text`)
	return err
}

// EnsureResCompanyEnterpriseColumns adds res.company columns shipped with the
// enterprise-style company module (existing deployments: IF NOT EXISTS).
func EnsureResCompanyEnterpriseColumns() error {
	tn := GetTableName("res.company")
	stmts := []string{
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS street2 VARCHAR(255)`,
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS city VARCHAR(255)`,
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS zip VARCHAR(255)`,
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS state VARCHAR(255)`,
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS country VARCHAR(255)`,
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS mobile VARCHAR(255)`,
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS vat VARCHAR(255)`,
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS company_registry VARCHAR(255)`,
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS internal_notes TEXT`,
	}
	for _, q := range stmts {
		if _, err := DB.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// EnsureResUsersEnterpriseColumns adds res.users columns for the enhanced user module.
func EnsureResUsersEnterpriseColumns() error {
	tn := GetTableName("res.users")
	stmts := []string{
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS email VARCHAR(255)`,
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS phone VARCHAR(255)`,
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS mobile VARCHAR(255)`,
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS company_id BIGINT`,
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS lang VARCHAR(50)`,
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS tz VARCHAR(255)`,
		`ALTER TABLE ` + tn + ` ADD COLUMN IF NOT EXISTS signature TEXT`,
	}
	for _, q := range stmts {
		if _, err := DB.Exec(q); err != nil {
			return err
		}
	}
	idx := `CREATE INDEX IF NOT EXISTS idx_` + tn + `_company_id ON ` + tn + ` (company_id)`
	if _, err := DB.Exec(idx); err != nil {
		return err
	}
	return nil
}
