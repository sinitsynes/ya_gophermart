package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		password string
		wantErr error
	}{
		{
			name:     "OK",
			password: "secret",
		},
		{
			name:     "TooLong",
			password: strings.Repeat("a", 73), // bcrypt limit is 72 bytes
			wantErr:  ErrPasswordTooLong,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := hashPassword(test.password)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				assert.Empty(t, got)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, got)
			assert.NotEqual(t, test.password, got)
			require.NoError(t, bcrypt.CompareHashAndPassword([]byte(got), []byte(test.password)))
		})
	}
}

func TestCheckPassword(t *testing.T) {
	t.Parallel()

	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	require.NoError(t, err)

	tests := []struct {
		name     string
		hash     string
		password string
		wantErr  error
	}{
		{
			name:     "OK",
			hash:     string(hash),
			password: "secret",
		},
		{
			name:     "Mismatch",
			hash:     string(hash),
			password: "wrong",
			wantErr:  ErrInvalidPassword,
		},
		{
			name:     "HashTooShort",
			hash:     "short",
			password: "secret",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := checkPassword(test.hash, test.password)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				return
			}
			if test.name == "HashTooShort" {
				require.Error(t, err)
				assert.ErrorContains(t, err, "hash too short")
				return
			}
			require.NoError(t, err)
		})
	}
}
