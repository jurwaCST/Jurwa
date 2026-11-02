package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClock struct{ t time.Time }

func (c *fakeClock) Now() time.Time          { return c.t }
func (c *fakeClock) Advance(d time.Duration) { c.t = c.t.Add(d) }

func TestLimiterAllowsBurstThenBlocks(t *testing.T) {
	clock := &fakeClock{t: time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)}
	l := NewLimiter(20, clock.Now)

	for i := 1; i <= 20; i++ {
		assert.True(t, l.Allow("10.0.0.1"), "request %d", i)
	}
	assert.False(t, l.Allow("10.0.0.1"), "the 21st request in the same second is refused")
	assert.True(t, l.Allow("10.0.0.2"), "each IP has its own bucket")
}

func TestLimiterRefills(t *testing.T) {
	clock := &fakeClock{t: time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)}
	l := NewLimiter(20, clock.Now)
	for i := 0; i < 20; i++ {
		l.Allow("ip")
	}
	require.False(t, l.Allow("ip"))

	clock.Advance(50 * time.Millisecond) // 20 per second = one token every 50 ms
	assert.True(t, l.Allow("ip"))
	assert.False(t, l.Allow("ip"))

	clock.Advance(time.Hour)
	for i := 0; i < 20; i++ {
		assert.True(t, l.Allow("ip"), "bucket never holds more than 20")
	}
	assert.False(t, l.Allow("ip"))
}

func TestBearerToken(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr error
	}{
		{name: "valid", header: "Bearer abc.def.ghi", want: "abc.def.ghi"},
		{name: "scheme is case-insensitive", header: "bearer abc", want: "abc"},
		{name: "extra spaces", header: "Bearer   abc  ", want: "abc"},
		{name: "missing", header: "", wantErr: ErrMissingToken},
		{name: "wrong scheme", header: "Basic dXNlcjpwYXNz", wantErr: ErrMalformedToken},
		{name: "no token", header: "Bearer ", wantErr: ErrMalformedToken},
		{name: "no space", header: "Bearerabc", wantErr: ErrMalformedToken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BearerToken(tt.header)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrincipalHasRole(t *testing.T) {
	p := Principal{UserID: "u1", Roles: []string{"STUDENT", "STAFF"}}
	assert.True(t, p.HasRole("STUDENT"))
	assert.True(t, p.HasRole("STAFF"))
	assert.False(t, p.HasRole("SYSTEM_ADMIN"))
	assert.False(t, Principal{}.HasRole("STUDENT"))
}

// A student may read only their own report; staff roles may read any.
func TestCanReadReport(t *testing.T) {
	student := Principal{UserID: "u1", Roles: []string{"STUDENT"}}
	head := Principal{UserID: "u9", Roles: []string{"CATEGORY_HEAD"}}
	assert.True(t, CanReadReport(student, "u1"))
	assert.False(t, CanReadReport(student, "u2"))
	assert.True(t, CanReadReport(head, "u2"))
	assert.True(t, CanReadReport(Principal{UserID: "a", Roles: []string{"SYSTEM_ADMIN"}}, "u2"))
	assert.True(t, CanReadReport(Principal{UserID: "s", Roles: []string{"STAFF"}}, "u2"))
	assert.False(t, CanReadReport(Principal{UserID: "p", Roles: []string{"PRESIDENT"}}, "u2"))
}

func TestLimiterDropsFullBucketsWhenCrowded(t *testing.T) {
	clock := &fakeClock{t: time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)}
	l := NewLimiter(20, clock.Now)
	for i := 0; i < maxBuckets; i++ {
		l.Allow(time.Duration(i).String())
	}
	clock.Advance(time.Second) // every bucket is full again
	assert.True(t, l.Allow("new-ip"))
	assert.Len(t, l.buckets, 1, "full buckets were dropped, only the new one is left")
}
