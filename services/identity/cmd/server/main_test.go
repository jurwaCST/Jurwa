package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvOr(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		fallback string
		want     string
	}{
		{name: "uses value when set", value: "9100", fallback: "9000", want: "9100"},
		{name: "uses fallback when empty", value: "", fallback: "9000", want: "9000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JURWA_TEST_VAR", tt.value)
			assert.Equal(t, tt.want, envOr("JURWA_TEST_VAR", tt.fallback))
		})
	}
}

func TestNewTokenIssuer(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		ttl     string
		wantErr bool
	}{
		{name: "valid", secret: "a-secret-that-is-at-least-32-bytes!", ttl: "15m"},
		{name: "missing secret", secret: "", ttl: "15m", wantErr: true},
		{name: "short secret", secret: "short", ttl: "15m", wantErr: true},
		{name: "bad duration", secret: "a-secret-that-is-at-least-32-bytes!", ttl: "fifteen", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newTokenIssuer(tt.secret, tt.ttl)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
