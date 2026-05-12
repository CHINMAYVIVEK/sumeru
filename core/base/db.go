package base

import "sumeru/core/orm"

// InitDBInput opens the PostgreSQL pool using a libpq connection string.
type InitDBInput struct {
	DSN string
}

// InitDB connects to the database (fatal on failure; matches ORM behavior).
func InitDB(in InitDBInput) {
	orm.InitDB(in.DSN)
}

// SyncModelsInput is a placeholder for future options; pass zero value.
type SyncModelsInput struct{}

// SyncModels creates or updates tables for all registered models.
func SyncModels(_ SyncModelsInput) error {
	return orm.SyncModels()
}
