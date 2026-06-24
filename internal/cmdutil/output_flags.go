package cmdutil

import (
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/output"
)

// AddOutputFlags is a re-export of output.AddOutputFlags so command
// files can import only cmdutil.
func AddOutputFlags(cmd *cobra.Command, target *string) {
	output.AddOutputFlags(cmd, target)
}
