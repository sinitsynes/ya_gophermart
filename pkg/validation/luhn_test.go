package validation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sinitsynes/ya_gophermart/pkg/validation"
)

func TestLuhnChecksum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "ValidKnown", input: "4561261212345632", want: true},
		{name: "InvalidCheckDigit", input: "4561261212345631", want: false},
		{name: "NonDigit", input: "456126121234563a", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, validation.LuhnChecksum(test.input))
		})
	}
}
