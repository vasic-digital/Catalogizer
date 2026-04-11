package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// newParseTestHandler returns an ErrorReportingHandler with nil deps.
// The parse*Filters methods are pure receivers that don't touch the
// service interfaces — nil is safe here.
func newParseTestHandler() *ErrorReportingHandler {
	return &ErrorReportingHandler{}
}

func TestParseErrorReportFilters_Empty(t *testing.T) {
	h := newParseTestHandler()
	req := httptest.NewRequest("GET", "/api/v1/errors", nil)
	f := h.parseErrorReportFilters(req)
	require.NotNil(t, f)
	require.Empty(t, f.Level)
	require.Empty(t, f.Component)
	require.Nil(t, f.StartDate)
	require.Nil(t, f.EndDate)
}

func TestParseErrorReportFilters_AllFields(t *testing.T) {
	h := newParseTestHandler()
	req := httptest.NewRequest("GET",
		"/api/v1/errors?level=error&component=api&status=open&start_date=2026-01-01&end_date=2026-04-11&limit=50&offset=100",
		nil)
	f := h.parseErrorReportFilters(req)
	require.Equal(t, "error", f.Level)
	require.Equal(t, "api", f.Component)
	require.Equal(t, "open", f.Status)
	require.NotNil(t, f.StartDate)
	require.NotNil(t, f.EndDate)
	require.Equal(t, 50, f.Limit)
	require.Equal(t, 100, f.Offset)
}

func TestParseErrorReportFilters_InvalidDatesIgnored(t *testing.T) {
	h := newParseTestHandler()
	req := httptest.NewRequest("GET",
		"/api/v1/errors?start_date=not-a-date&end_date=also-bad",
		nil)
	f := h.parseErrorReportFilters(req)
	require.Nil(t, f.StartDate, "invalid start_date should remain nil")
	require.Nil(t, f.EndDate, "invalid end_date should remain nil")
}

func TestParseErrorReportFilters_InvalidLimitOffsetIgnored(t *testing.T) {
	h := newParseTestHandler()
	req := httptest.NewRequest("GET",
		"/api/v1/errors?limit=abc&offset=-5",
		nil)
	f := h.parseErrorReportFilters(req)
	require.Equal(t, 0, f.Limit, "non-numeric limit should remain 0")
	require.Equal(t, 0, f.Offset, "negative offset should remain 0")
}

func TestParseErrorReportFilters_ZeroLimitIgnored(t *testing.T) {
	h := newParseTestHandler()
	req := httptest.NewRequest("GET", "/api/v1/errors?limit=0", nil)
	f := h.parseErrorReportFilters(req)
	require.Equal(t, 0, f.Limit, "zero limit is treated as default")
}

func TestParseCrashReportFilters_AllFields(t *testing.T) {
	h := newParseTestHandler()
	req := httptest.NewRequest("GET",
		"/api/v1/crashes?signal=SIGSEGV&status=resolved&start_date=2026-03-01&end_date=2026-04-11&limit=25&offset=50",
		nil)
	f := h.parseCrashReportFilters(req)
	require.Equal(t, "SIGSEGV", f.Signal)
	require.Equal(t, "resolved", f.Status)
	require.NotNil(t, f.StartDate)
	require.NotNil(t, f.EndDate)
	require.Equal(t, 25, f.Limit)
	require.Equal(t, 50, f.Offset)
}

func TestParseCrashReportFilters_Empty(t *testing.T) {
	h := newParseTestHandler()
	req := httptest.NewRequest("GET", "/api/v1/crashes", nil)
	f := h.parseCrashReportFilters(req)
	require.NotNil(t, f)
	require.Empty(t, f.Signal)
	require.Empty(t, f.Status)
	require.Nil(t, f.StartDate)
}

func TestParseCrashReportFilters_InvalidDatesIgnored(t *testing.T) {
	h := newParseTestHandler()
	req := httptest.NewRequest("GET",
		"/api/v1/crashes?start_date=bad&end_date=bad",
		nil)
	f := h.parseCrashReportFilters(req)
	require.Nil(t, f.StartDate)
	require.Nil(t, f.EndDate)
}

func TestParseExportFilters_Defaults(t *testing.T) {
	h := newParseTestHandler()
	req := httptest.NewRequest("GET", "/api/v1/export", nil)
	f := h.parseExportFilters(req)
	require.Equal(t, "json", f.Format, "default format should be json")
	require.True(t, f.IncludeErrors, "default IncludeErrors should be true")
	require.True(t, f.IncludeCrashes, "default IncludeCrashes should be true")
}

func TestParseExportFilters_CustomFormat(t *testing.T) {
	h := newParseTestHandler()
	req := httptest.NewRequest("GET",
		"/api/v1/export?format=csv&level=error&component=api&signal=SIGSEGV&start_date=2026-01-01",
		nil)
	f := h.parseExportFilters(req)
	require.Equal(t, "csv", f.Format)
	require.Equal(t, "error", f.Level)
	require.Equal(t, "api", f.Component)
	require.Equal(t, "SIGSEGV", f.Signal)
	require.NotNil(t, f.StartDate)
}
