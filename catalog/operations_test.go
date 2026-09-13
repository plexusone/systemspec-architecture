package catalog

import (
	"testing"

	"github.com/plexusone/systemspec-architecture/sas"
)

func TestHTTPOperation(t *testing.T) {
	tests := []struct {
		method string
		want   sas.Operation
		wantOK bool
	}{
		{"GET", sas.OperationRead, true},
		{"get", sas.OperationRead, true}, // case-insensitive
		{"POST", sas.OperationCreate, true},
		{"PUT", sas.OperationUpdate, true},
		{"PATCH", sas.OperationUpdate, true},
		{"DELETE", sas.OperationDelete, true},
		{"TRACE", "", false},
	}
	for _, tt := range tests {
		got, ok := HTTPOperation(tt.method)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("HTTPOperation(%q) = (%q, %v), want (%q, %v)", tt.method, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestSQLOperation(t *testing.T) {
	tests := []struct {
		statement string
		want      sas.Operation
		wantOK    bool
	}{
		{"SELECT", sas.OperationRead, true},
		{"select", sas.OperationRead, true},
		{"INSERT", sas.OperationCreate, true},
		{"UPDATE", sas.OperationUpdate, true},
		{"DELETE", sas.OperationDelete, true},
		{"GRANT", sas.OperationAdminister, true},
		{"EXPLAIN", "", false},
	}
	for _, tt := range tests {
		got, ok := SQLOperation(tt.statement)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("SQLOperation(%q) = (%q, %v), want (%q, %v)", tt.statement, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestMCPOperation(t *testing.T) {
	tests := []struct {
		action string
		want   sas.Operation
		wantOK bool
	}{
		{"tools/call", sas.OperationExecute, true},
		{"tools/list", sas.OperationDiscover, true},
		{"resources/read", sas.OperationRead, true},
		{"prompts/get", "", false},
	}
	for _, tt := range tests {
		got, ok := MCPOperation(tt.action)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("MCPOperation(%q) = (%q, %v), want (%q, %v)", tt.action, got, ok, tt.want, tt.wantOK)
		}
	}
}
