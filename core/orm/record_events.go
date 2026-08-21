package orm

// Record lifecycle event names published after successful mutations.
const (
	EventRecordCreated = "record.created"
	EventRecordUpdated = "record.updated"
	EventRecordDeleted = "record.deleted"
)
