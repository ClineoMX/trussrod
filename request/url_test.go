package request

import (
	"net/http/httptest"
	"testing"
)

type queryFilters struct {
	Status string `json:"status"`
	Author string `json:"author"`
	Sort   string `json:"sort"`
}

type typedFilters struct {
	Limit int `json:"limit"`
}

func TestGetQueryParamsAs_EmptyQuery(t *testing.T) {
	r := httptest.NewRequest("GET", "/notes", nil)

	got := GetQueryParamsAs[queryFilters](r)
	if got != (queryFilters{}) {
		t.Fatalf("expected zero-value struct, got %+v", got)
	}
}

func TestGetQueryParamsAs_ParsesStringFields(t *testing.T) {
	r := httptest.NewRequest("GET", "/notes?status=active&author=alice&sort=desc", nil)

	got := GetQueryParamsAs[queryFilters](r)
	if got.Status != "active" {
		t.Fatalf("expected status active, got %q", got.Status)
	}
	if got.Author != "alice" {
		t.Fatalf("expected author alice, got %q", got.Author)
	}
	if got.Sort != "desc" {
		t.Fatalf("expected sort desc, got %q", got.Sort)
	}
}

func TestGetQueryParamsAs_PartialParams(t *testing.T) {
	r := httptest.NewRequest("GET", "/notes?status=draft", nil)

	got := GetQueryParamsAs[queryFilters](r)
	if got.Status != "draft" {
		t.Fatalf("expected status draft, got %q", got.Status)
	}
	if got.Author != "" || got.Sort != "" {
		t.Fatalf("expected unset fields to remain empty, got %+v", got)
	}
}

func TestGetQueryParamsAs_UsesFirstValueWhenRepeated(t *testing.T) {
	r := httptest.NewRequest("GET", "/notes?status=active&status=archived", nil)

	got := GetQueryParamsAs[queryFilters](r)
	if got.Status != "active" {
		t.Fatalf("expected first status value active, got %q", got.Status)
	}
}

func TestGetQueryParamsAs_NonStringFieldReturnsZeroValue(t *testing.T) {
	r := httptest.NewRequest("GET", "/notes?limit=10", nil)

	got := GetQueryParamsAs[typedFilters](r)
	if got.Limit != 0 {
		t.Fatalf("expected zero value when query value cannot unmarshal into int, got %d", got.Limit)
	}
}

func TestGetQueryParamsAs_InvalidTypeReturnsZeroValue(t *testing.T) {
	r := httptest.NewRequest("GET", "/notes?limit=not-a-number", nil)

	got := GetQueryParamsAs[typedFilters](r)
	if got.Limit != 0 {
		t.Fatalf("expected zero value for invalid int query value, got %d", got.Limit)
	}
}

func TestGetQueryParamsAs_StringMap(t *testing.T) {
	r := httptest.NewRequest("GET", "/search?q=hello&sort=desc", nil)

	got := GetQueryParamsAs[map[string]string](r)
	if got["q"] != "hello" {
		t.Fatalf("expected q=hello, got %q", got["q"])
	}
	if got["sort"] != "desc" {
		t.Fatalf("expected sort=desc, got %q", got["sort"])
	}
}
