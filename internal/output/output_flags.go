package output

import (
	"github.com/spf13/cobra"
)

// ValidOutputFormats lists the accepted values for the --output flag.
var ValidOutputFormats = []string{"json", "table", "plain"}

// DefaultOutputFormat is the value used when --output is not provided.
const DefaultOutputFormat = "plain"

// AddOutputFlags wires the --output string flag onto cmd, writing the
// selected value through output. The flag is registered with a
// completion function that suggests the three valid formats.
func AddOutputFlags(cmd *cobra.Command, output *string) {
	cmd.Flags().StringVarP(output, "output", "o", DefaultOutputFormat, "Output format: json, table, plain")
	_ = cmd.RegisterFlagCompletionFunc("output", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return ValidOutputFormats, cobra.ShellCompDirectiveNoFileComp
	})
}
