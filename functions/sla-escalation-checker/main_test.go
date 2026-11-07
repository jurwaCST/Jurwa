package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jurwaCST/Jurwa/functions/sla-escalation-checker/checker"
)

type fakeRunner struct {
	res checker.Result
	err error
}

func (f fakeRunner) Run(context.Context) (checker.Result, error) { return f.res, f.err }

func call(h http.Handler, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestHandler_ReturnsCounts(t *testing.T) {
	h := newHandler(fakeRunner{res: checker.Result{Lapsed: 3, Escalated: 2, Skipped: 1}})
	for _, m := range []string{http.MethodGet, http.MethodPost} {
		rec := call(h, m, "/")
		assert.Equal(t, http.StatusOK, rec.Code, m)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"lapsed":3,"escalated":2,"skipped":1,"errors":0}`, rec.Body.String())
	}
}

func TestHandler_SlaTrackingDown(t *testing.T) {
	h := newHandler(fakeRunner{res: checker.Result{Lapsed: 100, Escalated: 100}, err: errors.New("unavailable")})
	rec := call(h, http.MethodPost, "/")
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.JSONEq(t, `{"lapsed":100,"escalated":100,"skipped":0,"errors":1}`, rec.Body.String(),
		"work done before the failure is still reported")
}

func TestHandler_BadRequests(t *testing.T) {
	h := newHandler(fakeRunner{})
	assert.Equal(t, http.StatusNotFound, call(h, http.MethodGet, "/other").Code)

	rec := call(h, http.MethodDelete, "/")
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Equal(t, "GET, POST", rec.Header().Get("Allow"))
}

func TestEnvOr(t *testing.T) {
	t.Setenv("JURWA_TEST_VAR", "9100")
	assert.Equal(t, "9100", envOr("JURWA_TEST_VAR", "9000"))
	t.Setenv("JURWA_TEST_VAR", "")
	assert.Equal(t, "9000", envOr("JURWA_TEST_VAR", "9000"))
}

func TestRunOnce_PrintsTheCountsAndExitsZero(t *testing.T) {
	var out strings.Builder
	code := runOnce(context.Background(), fakeRunner{res: checker.Result{Lapsed: 2, Escalated: 1, Skipped: 1}}, &out)
	assert.Equal(t, 0, code)
	assert.JSONEq(t, `{"lapsed":2,"escalated":1,"skipped":1,"errors":0}`, out.String())
}

func TestRunOnce_SlaTrackingDownExitsOne(t *testing.T) {
	var out strings.Builder
	code := runOnce(context.Background(), fakeRunner{err: errors.New("unavailable")}, &out)
	assert.Equal(t, 1, code, "the Kubernetes Job is marked failed")
	assert.JSONEq(t, `{"lapsed":0,"escalated":0,"skipped":0,"errors":1}`, out.String())
}

func TestRunOnceEnabled(t *testing.T) {
	for value, want := range map[string]bool{"true": true, "1": true, "TRUE": true, "": false, "false": false, "no": false} {
		t.Setenv("RUN_ONCE", value)
		assert.Equal(t, want, runOnceEnabled(), value)
	}
}
