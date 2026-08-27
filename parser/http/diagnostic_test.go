package http

import (
	"errors"
	"testing"
)

func TestDiagnosticSeverityControlsExecution(t *testing.T) {
	warning := diagnosticAt(2, "unrecognized content")
	if warning.IsBlocking() {
		t.Fatal("warning unexpectedly blocks execution")
	}

	suite := HttpSuite{Diagnostics: []Diagnostic{warning}}
	if err := suite.ValidateForExecution(); err != nil {
		t.Fatalf("warning made the request unrunnable: %v", err)
	}

	suite.Diagnostics = append(suite.Diagnostics, blockingDiagnosticAt(4, "missing body"))
	if err := suite.ValidateForExecution(); !errors.Is(err, ErrRequestNotRunnable) {
		t.Fatalf("blocking diagnostic was not enforced: %v", err)
	}
}
