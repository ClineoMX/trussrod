package request

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/clineomx/trussrod/audit"
)

type stubAuditor struct{}

func (stubAuditor) Log(context.Context, *audit.Log) {}
func (stubAuditor) WithFields(map[string]any) audit.Auditor { return stubAuditor{} }

func TestWithAuditor_FromContextReturnsSameInstance(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	auditor := stubAuditor{}
	r = WithAuditor(r, auditor)

	got := audit.FromContext(r.Context())
	if got == nil {
		t.Fatal("expected auditor from context, got nil")
	}
	if got != auditor {
		t.Fatal("expected same auditor instance from audit.FromContext")
	}
}

func TestGetAuditor(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	auditor := stubAuditor{}
	r = WithAuditor(r, auditor)

	got, ok := GetAuditor(r)
	if !ok {
		t.Fatal("expected GetAuditor ok=true")
	}
	if got != auditor {
		t.Fatal("expected same auditor instance from GetAuditor")
	}
}

func TestMustGetAuditor(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	auditor := stubAuditor{}
	r = WithAuditor(r, auditor)

	got := MustGetAuditor(r)
	if got != auditor {
		t.Fatal("expected same auditor instance from MustGetAuditor")
	}
}

func TestGetAuditor_Missing(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)

	_, ok := GetAuditor(r)
	if ok {
		t.Fatal("expected GetAuditor ok=false when auditor not in context")
	}
}
