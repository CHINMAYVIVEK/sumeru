// Package swcmeta serializes parsed views and ORM workspace data to JSON for SWC.
//
// Used by GET /web/swc/workspace. Always redacts records with orm.RedactRecordForRead
// before sending to the browser.
package swcmeta
