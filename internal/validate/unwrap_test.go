package validate

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnwrapJoinedErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantLen int
	}{
		{name: "nil error returns nil", err: nil, wantLen: 0},
		{name: "single error", err: errors.New("boom"), wantLen: 1},
		{name: "two joined errors", err: errors.Join(errors.New("a"), errors.New("b")), wantLen: 2},
		{name: "three joined errors", err: errors.Join(errors.New("x"), errors.New("y"), errors.New("z")), wantLen: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs := unwrapJoinedErrors(tt.err)
			assert.Len(t, errs, tt.wantLen)
		})
	}
}
