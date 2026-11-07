package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/routing-engine/domain"
	"github.com/jurwaCST/Jurwa/services/routing-engine/ports"
)

// fakeRules is a RuleLookup test double: it answers from a map, or returns err.
type fakeRules struct {
	rules map[domain.Category]domain.Rule
	err   error
	calls int
}

func (f *fakeRules) GetRule(_ context.Context, c domain.Category, _ string) (domain.Rule, error) {
	f.calls++
	if f.err != nil {
		return domain.Rule{}, f.err
	}
	r, ok := f.rules[c]
	if !ok {
		return domain.Rule{}, domain.ErrRuleNotFound
	}
	return r, nil
}

// fakeHolders is a HolderLookup test double.
type fakeHolders struct {
	holders map[string]string
	err     error
}

func (f *fakeHolders) GetHolder(_ context.Context, position string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.holders[position], nil
}

// fakePublisher records every event.
type fakePublisher struct {
	events []ports.ReportRouted
	err    error
}

func (f *fakePublisher) PublishReportRouted(_ context.Context, ev ports.ReportRouted) error {
	f.events = append(f.events, ev)
	return f.err
}

func defaultRules() *fakeRules {
	return &fakeRules{rules: map[domain.Category]domain.Rule{
		domain.Hostel:        {ID: "rule-hostel", TargetPosition: "DSA"},
		domain.AcademicBlock: {ID: "rule-academic", TargetPosition: "DAA"},
		domain.Others:        {ID: "rule-others", TargetPosition: "SSO"},
	}}
}

func defaultHolders() *fakeHolders {
	return &fakeHolders{holders: map[string]string{"DSA": "user-dsa", "DAA": "user-daa", "SSO": "user-sso"}}
}

func TestRouteReport(t *testing.T) {
	tests := []struct {
		name       string
		category   domain.Category
		wantHead   string
		wantPos    string
		wantRuleID string
	}{
		{name: "hostel", category: domain.Hostel, wantHead: "user-dsa", wantPos: "DSA", wantRuleID: "rule-hostel"},
		{name: "academic block", category: domain.AcademicBlock, wantHead: "user-daa", wantPos: "DAA", wantRuleID: "rule-academic"},
		{name: "others", category: domain.Others, wantHead: "user-sso", wantPos: "SSO", wantRuleID: "rule-others"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pub := &fakePublisher{}
			s := NewService(defaultRules(), defaultHolders(), pub)

			got, err := s.RouteReport(context.Background(), domain.Request{ReportID: "r1", Category: tt.category, Building: "Hostel B"})
			require.NoError(t, err)
			assert.Equal(t, tt.wantHead, got.CategoryHeadID)
			assert.Equal(t, tt.wantPos, got.TargetPosition)
			assert.Equal(t, tt.wantRuleID, got.RuleID)

			// One events:report.routed event with the report id, category and head.
			require.Len(t, pub.events, 1)
			assert.Equal(t, ports.ReportRouted{
				ReportID: "r1", Category: tt.category, CategoryHeadID: tt.wantHead,
				TargetPosition: tt.wantPos, RuleID: tt.wantRuleID,
			}, pub.events[0])
		})
	}
}

func TestRouteReportDryRunPublishesNothing(t *testing.T) {
	pub := &fakePublisher{}
	got, err := NewService(defaultRules(), defaultHolders(), pub).
		RouteReport(context.Background(), domain.Request{Category: domain.Hostel, Building: "Hostel B"})
	require.NoError(t, err)
	assert.Equal(t, "DSA", got.TargetPosition)
	assert.Empty(t, pub.events)
}

func TestRouteReportErrors(t *testing.T) {
	errDown := errors.New("connection refused")
	tests := []struct {
		name      string
		req       domain.Request
		rules     *fakeRules
		holders   *fakeHolders
		pub       *fakePublisher
		wantErr   error
		wantCalls int
	}{
		{name: "unspecified category", req: domain.Request{Building: "B"}, wantErr: domain.ErrInvalidCategory},
		{name: "empty location", req: domain.Request{Category: domain.Hostel}, wantErr: domain.ErrInvalidLocation},
		{name: "no rule", req: domain.Request{Category: domain.Hostel, Building: "B"}, rules: &fakeRules{}, wantErr: domain.ErrRuleNotFound, wantCalls: 1},
		{name: "rules service down", req: domain.Request{Category: domain.Hostel, Building: "B"}, rules: &fakeRules{err: domain.ErrUnavailable}, wantErr: domain.ErrUnavailable, wantCalls: 1},
		{name: "directory down", req: domain.Request{Category: domain.Hostel, Building: "B"}, holders: &fakeHolders{err: domain.ErrUnavailable}, wantErr: domain.ErrUnavailable, wantCalls: 1},
		{name: "position has no holder", req: domain.Request{Category: domain.Hostel, Building: "B"}, holders: &fakeHolders{}, wantErr: domain.ErrNoHolder, wantCalls: 1},
		{name: "publish fails", req: domain.Request{ReportID: "r1", Category: domain.Hostel, Building: "B"}, pub: &fakePublisher{err: errDown}, wantErr: errDown, wantCalls: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.rules == nil {
				tt.rules = defaultRules()
			}
			if tt.holders == nil {
				tt.holders = defaultHolders()
			}
			if tt.pub == nil {
				tt.pub = &fakePublisher{}
			}
			_, err := NewService(tt.rules, tt.holders, tt.pub).RouteReport(context.Background(), tt.req)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantCalls, tt.rules.calls, "invalid input never reaches routing-rules")
			if !errors.Is(tt.wantErr, errDown) {
				assert.Empty(t, tt.pub.events, "no event after a failed route")
			}
		})
	}
}

// A report routed because of report.submitted carries that event's id, so the
// publisher can give a redelivered event the same report.routed id.
func TestRouteSubmittedKeepsTheSourceEventID(t *testing.T) {
	pub := &fakePublisher{}
	s := NewService(defaultRules(), defaultHolders(), pub)

	d, err := s.RouteSubmitted(context.Background(), "ev-submitted-1", domain.Request{ReportID: "r1", Category: domain.Hostel, Building: "Hostel B"})
	require.NoError(t, err)
	assert.Equal(t, "DSA", d.TargetPosition)
	require.Len(t, pub.events, 1)
	assert.Equal(t, "ev-submitted-1", pub.events[0].SourceEventID)
	assert.Equal(t, "r1", pub.events[0].ReportID)

	_, err = s.RouteReport(context.Background(), domain.Request{ReportID: "r2", Category: domain.Hostel, Building: "Hostel B"})
	require.NoError(t, err)
	assert.Empty(t, pub.events[1].SourceEventID, "a call through the RPC has no source event")
}

func TestRouteSubmittedNeedsAReportID(t *testing.T) {
	pub := &fakePublisher{}
	_, err := NewService(defaultRules(), defaultHolders(), pub).
		RouteSubmitted(context.Background(), "ev-1", domain.Request{Category: domain.Hostel, Building: "Hostel B"})
	assert.ErrorIs(t, err, domain.ErrMissingReportID)
	assert.Empty(t, pub.events)
}
