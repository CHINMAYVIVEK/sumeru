package module

import (
	"fmt"
	"strings"
)

// SyncErrorClass distinguishes fatal install failures from recoverable warnings.
type SyncErrorClass int

const (
	SyncRecoverable SyncErrorClass = iota
	SyncFatal
)

// SyncError is a classified module sync failure.
type SyncError struct {
	Class   SyncErrorClass
	Module  string
	Message string
	Cause   error
}

func (e *SyncError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *SyncError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func FatalSync(module, msg string, cause error) error {
	return &SyncError{Class: SyncFatal, Module: module, Message: msg, Cause: cause}
}

func RecoverableSync(module, msg string, cause error) error {
	return &SyncError{Class: SyncRecoverable, Module: module, Message: msg, Cause: cause}
}

func IsFatalSync(err error) bool {
	if err == nil {
		return false
	}
	if se, ok := err.(*SyncError); ok {
		return se.Class == SyncFatal
	}
	// Default unknown errors from SyncToDB / schema to fatal.
	return true
}

func aggregateErrors(module string, errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	fatal := false
	var msgs []string
	for _, e := range errs {
		if e == nil {
			continue
		}
		if IsFatalSync(e) {
			fatal = true
		}
		msgs = append(msgs, e.Error())
	}
	if len(msgs) == 0 {
		return nil
	}
	joined := strings.Join(msgs, "; ")
	if fatal {
		return FatalSync(module, "module sync failed", fmt.Errorf("%s", joined))
	}
	return RecoverableSync(module, "module sync warnings", fmt.Errorf("%s", joined))
}
