package cmdutil

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/amoconst/germinator/internal/core"
	"gitlab.com/amoconst/germinator/internal/iostreams"
	"gitlab.com/amoconst/germinator/internal/output"
)

// TestPartialSuccessCrossPackage guards against the three packages
// (core, output, cmdutil) drifting out of sync. It constructs a
// *core.PartialSuccessError, asserts cmdutil.ExitCodeFor maps it to
// ExitCodeSuccess per design Decision 12, and asserts output.FormatError
// writes the expected partial-success string to io.ErrOut.
func TestPartialSuccessCrossPackage(t *testing.T) {
	t.Parallel()

	io := iostreams.Test()
	ie := core.NewInitializeError("skill/missing", "/lib/skills/missing.md", "/out/.opencode/skills/missing/SKILL.md", errors.New("file not found"))
	psErr := core.NewPartialSuccessError(3, 1, []core.InitializeError{*ie})

	assert.Equal(t, ExitCodeSuccess, ExitCodeFor(psErr), "PartialSuccessError with Succeeded>0 should map to ExitCodeSuccess per design Decision 12")

	output.FormatError(io, psErr)

	buf, ok := io.ErrOut.(*bytes.Buffer)
	require.True(t, ok)
	got := buf.String()
	assert.Contains(t, got, "partial success: 3 succeeded, 1 failed")
	assert.Contains(t, got, "skill/missing")
}
