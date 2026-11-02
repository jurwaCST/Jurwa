package dispatcher

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeCounts returns fixed counts per category, or an error.
type fakeCounts struct {
	counts map[Category]Counts
	err    error
}

func (f fakeCounts) CategoryCounts(context.Context) (map[Category]Counts, error) {
	return f.counts, f.err
}

// fakeHeads maps a position code to a user id. failFor makes one lookup fail.
type fakeHeads struct {
	failFor string
}

var headIDs = map[string]string{
	"DSA": "00000000-0000-4000-8000-000000000001",
	"DAA": "00000000-0000-4000-8000-000000000002",
	"SSO": "00000000-0000-4000-8000-000000000003",
}

func (f fakeHeads) HolderOf(_ context.Context, position string) (string, error) {
	if position == f.failFor {
		return "", errors.New("directory down")
	}
	return headIDs[position], nil
}

// sent is one call to the fake sender.
type sent struct {
	userID, subject, body string
}

// fakeSender records every message. failFor makes sends to one user fail.
type fakeSender struct {
	sent    []sent
	failFor string
}

func (f *fakeSender) Send(_ context.Context, userID, subject, body string) error {
	if userID == f.failFor {
		return errors.New("notification down")
	}
	f.sent = append(f.sent, sent{userID, subject, body})
	return nil
}

func TestNoOpenOrEscalatedReportsSendsNothing(t *testing.T) {
	tests := []struct {
		name   string
		counts map[Category]Counts
	}{
		{"no reports at all", map[Category]Counts{}},
		{"every report resolved", map[Category]Counts{
			Hostel:        {Total: 3, Resolved: 3},
			AcademicBlock: {Total: 1, Resolved: 1},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &fakeSender{}
			res, err := New(fakeCounts{counts: tt.counts}, fakeHeads{}, sender).Run(context.Background())
			require.NoError(t, err)
			assert.Equal(t, Result{Digests: 0, Errors: 0}, res)
			assert.Empty(t, sender.sent)
		})
	}
}

func TestOneDigestPerCategoryHeadWithTheRightCounts(t *testing.T) {
	sender := &fakeSender{}
	counts := map[Category]Counts{
		Hostel:        {Total: 5, Open: 2, Assigned: 1, InProgress: 1, Resolved: 1, Escalated: 1},
		AcademicBlock: {Total: 2, Open: 1, Escalated: 1, Resolved: 0, Assigned: 0, InProgress: 0},
		Others:        {Total: 1, Assigned: 1},
	}
	res, err := New(fakeCounts{counts: counts}, fakeHeads{}, sender).Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, Result{Digests: 3, Errors: 0}, res)
	require.Len(t, sender.sent, 3)

	byUser := map[string]sent{}
	for _, s := range sender.sent {
		byUser[s.userID] = s
	}
	dsa := byUser[headIDs["DSA"]]
	assert.Equal(t, "Daily digest: Hostel", dsa.subject)
	assert.Contains(t, dsa.body, "4 unresolved", "5 total minus 1 resolved")
	assert.Contains(t, dsa.body, "2 open, 1 assigned, 1 in progress")
	assert.Contains(t, dsa.body, "1 escalated")

	daa := byUser[headIDs["DAA"]]
	assert.Equal(t, "Daily digest: Academic Block", daa.subject)
	assert.Contains(t, daa.body, "2 unresolved")

	sso := byUser[headIDs["SSO"]]
	assert.Equal(t, "Daily digest: Others", sso.subject)
	assert.Contains(t, sso.body, "1 unresolved")
	assert.Contains(t, sso.body, "0 escalated")
}

func TestOnlyHeadsWithSomethingToReportGetADigest(t *testing.T) {
	sender := &fakeSender{}
	counts := map[Category]Counts{
		Hostel: {Total: 2, Open: 2},
		Others: {Total: 4, Resolved: 4},
	}
	res, err := New(fakeCounts{counts: counts}, fakeHeads{}, sender).Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, Result{Digests: 1}, res)
	require.Len(t, sender.sent, 1)
	assert.Equal(t, headIDs["DSA"], sender.sent[0].userID)
}

func TestAFailureForOneHeadDoesNotStopTheOthers(t *testing.T) {
	all := map[Category]Counts{
		Hostel:        {Total: 1, Open: 1},
		AcademicBlock: {Total: 1, Open: 1},
		Others:        {Total: 1, Open: 1},
	}

	t.Run("send fails for DAA", func(t *testing.T) {
		sender := &fakeSender{failFor: headIDs["DAA"]}
		res, err := New(fakeCounts{counts: all}, fakeHeads{}, sender).Run(context.Background())
		require.NoError(t, err)
		assert.Equal(t, Result{Digests: 2, Errors: 1}, res)
		assert.Len(t, sender.sent, 2)
	})

	t.Run("directory lookup fails for SSO", func(t *testing.T) {
		sender := &fakeSender{}
		res, err := New(fakeCounts{counts: all}, fakeHeads{failFor: "SSO"}, sender).Run(context.Background())
		require.NoError(t, err)
		assert.Equal(t, Result{Digests: 2, Errors: 1}, res)
	})

	t.Run("position has no holder yet", func(t *testing.T) {
		sender := &fakeSender{}
		heads := emptyHolder{}
		res, err := New(fakeCounts{counts: all}, heads, sender).Run(context.Background())
		require.NoError(t, err)
		assert.Equal(t, Result{Digests: 0, Errors: 3}, res)
	})
}

type emptyHolder struct{}

func (emptyHolder) HolderOf(context.Context, string) (string, error) { return "", nil }

func TestAnalyticsDownFailsTheWholeRun(t *testing.T) {
	sender := &fakeSender{}
	_, err := New(fakeCounts{err: errors.New("analytics down")}, fakeHeads{}, sender).Run(context.Background())
	assert.Error(t, err)
	assert.Empty(t, sender.sent)
}

func TestMessageIsAtMost200Characters(t *testing.T) {
	// The longest numbers int64 can print: 19 digits, or 20 with a minus sign
	// (Unresolved is negative when Resolved is bigger than Total).
	worst := []Counts{
		{Total: math.MaxInt64, Open: math.MaxInt64, Assigned: math.MaxInt64, InProgress: math.MaxInt64, Escalated: math.MaxInt64},
		{Total: math.MinInt64 + 1, Resolved: math.MaxInt64, Open: math.MinInt64, Assigned: math.MinInt64, InProgress: math.MinInt64, Escalated: math.MinInt64},
	}
	for _, n := range worst {
		for _, c := range Categories {
			body := Body(c, n)
			assert.LessOrEqual(t, len([]rune(body)), MaxBodyChars, "body for %s: %q", c, body)
			assert.NotEmpty(t, strings.TrimSpace(body))
		}
	}
}

func TestCategoryNamesAndHeads(t *testing.T) {
	assert.Equal(t, "Hostel", Hostel.String())
	assert.Equal(t, "Academic Block", AcademicBlock.String())
	assert.Equal(t, "Others", Others.String())
	assert.Equal(t, "Unknown", Category(9).String())
	assert.Equal(t, map[Category]string{Hostel: "DSA", AcademicBlock: "DAA", Others: "SSO"}, HeadPosition)
}
