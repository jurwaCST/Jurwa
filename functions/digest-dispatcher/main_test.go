package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jurwaCST/Jurwa/functions/digest-dispatcher/dispatcher"
)

type fakeCounts struct {
	counts map[dispatcher.Category]dispatcher.Counts
	err    error
}

func (f fakeCounts) CategoryCounts(context.Context) (map[dispatcher.Category]dispatcher.Counts, error) {
	return f.counts, f.err
}

type fakeHeads struct{}

func (fakeHeads) HolderOf(_ context.Context, position string) (string, error) {
	return "user-" + position, nil
}

type fakeSender struct{}

func (fakeSender) Send(context.Context, string, string, string) error { return nil }

func TestRunReturnsJSON(t *testing.T) {
	d := dispatcher.New(fakeCounts{counts: map[dispatcher.Category]dispatcher.Counts{
		dispatcher.Hostel:        {Total: 2, Open: 2},
		dispatcher.AcademicBlock: {Total: 1, Open: 1},
		dispatcher.Others:        {Total: 1, Assigned: 1},
	}}, fakeHeads{}, fakeSender{})

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		rec := httptest.NewRecorder()
		newHandler(d).ServeHTTP(rec, httptest.NewRequest(method, "/", nil))
		assert.Equal(t, http.StatusOK, rec.Code, method)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"digests":3,"errors":0}`, rec.Body.String())
	}
}

func TestAnalyticsDownIs503(t *testing.T) {
	d := dispatcher.New(fakeCounts{err: errors.New("down")}, fakeHeads{}, fakeSender{})
	rec := httptest.NewRecorder()
	newHandler(d).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.JSONEq(t, `{"digests":0,"errors":1}`, rec.Body.String())
}

func TestOtherPathsAre404(t *testing.T) {
	d := dispatcher.New(fakeCounts{}, fakeHeads{}, fakeSender{})
	rec := httptest.NewRecorder()
	newHandler(d).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/other", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestOtherMethodsAre405(t *testing.T) {
	d := dispatcher.New(fakeCounts{}, fakeHeads{}, fakeSender{})
	rec := httptest.NewRecorder()
	newHandler(d).ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/", nil))
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestEnvOr(t *testing.T) {
	t.Setenv("JURWA_TEST_VAR", "")
	assert.Equal(t, "fallback", envOr("JURWA_TEST_VAR", "fallback"))
	t.Setenv("JURWA_TEST_VAR", "set")
	assert.Equal(t, "set", envOr("JURWA_TEST_VAR", "fallback"))
}
