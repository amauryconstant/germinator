package cmd

import (
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

// completionShells contains all supported shell names.
var completionShells = []string{
	"bash",
	"zsh",
	"fish",
	"powershell",
	"elvish",
	"nushell",
	"oil",
	"tcsh",
	"xonsh",
	"cmd-clink",
}

// getShellInstructions returns installation instructions for a given shell.
func getShellInstructions(shell string) string {
	instructions := map[string]string{
		"bash": `Add the following to your ~/.bashrc:

    source <(germinator completion bash)

Or for dynamic loading:

    eval "$(germinator completion bash)"`,
		"zsh": `Add the following to your ~/.zshrc:

    source <(germinator completion zsh)

Or add to a completions directory:

    germinator completion zsh > ~/.zfunc/_germinator`,
		"fish": `Add the completion to fish:

    germinator completion fish | source

Or save to completions directory:

    germinator completion fish > ~/.config/fish/completions/germinator.fish`,
		"powershell": `Add to your PowerShell profile:

    germinator completion powershell | Out-String | Invoke-Expression

Also enable menu completion:

    Set-PSReadlineKeyHandler -Key Tab -Function MenuComplete`,
		"elvish": `Add to your rc.elv:

    eval (germinator completion elvish | slurp)`,
		"nushell": `Run the command and update your config.nu according to the output:

    germinator completion nushell`,
		"oil": `Add to your ~/.oilrc:

    source <(germinator completion oil)`,
		"tcsh": "Add to your ~/.tcshrc:\n\n    set autolist\n    eval `germinator completion tcsh`",
		"xonsh": `Add to your xonshrc:

    COMPLETIONS_CONFIRM=True
    exec($(germinator completion xonsh))`,
		"cmd-clink": `Save to Clink completions directory:

    germinator completion cmd-clink > ~/AppData/Local/clink/germinator.lua`,
	}

	if instr, ok := instructions[shell]; ok {
		return instr
	}
	return ""
}

// NewCompletionCommand creates the completion command with shell subcommands.
// This replaces Cobra's default completion with Carapace's enhanced completion system.
func NewCompletionCommand(_ *CommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for multiple shells.

Carapace provides enhanced completion support including:
- Dynamic completions for library resources and presets
- Static completions for platform values
- Multi-shell support (10+ shells)

Each subcommand generates a completion script for a specific shell.
The script should be sourced or saved according to your shell's requirements.

Examples:
  # Generate bash completion
  germinator completion bash > /etc/bash_completion.d/germinator

  # Generate zsh completion
  germinator completion zsh > ~/.zfunc/_germinator

  # Generate fish completion
  germinator completion fish > ~/.config/fish/completions/germinator.fish`,
		Run: func(c *cobra.Command, _ []string) {
			_ = c.Help()
		},
	}

	// Add subcommands for each shell
	for _, shell := range completionShells {
		cmd.AddCommand(newCompletionShellCommand(shell))
	}

	return cmd
}

// newCompletionShellCommand creates a completion subcommand for a specific shell.
func newCompletionShellCommand(shell string) *cobra.Command {
	return &cobra.Command{
		Use:   shell,
		Short: fmt.Sprintf("Generate %s completion script", shell),
		Long: fmt.Sprintf(`Generate the completion script for %s.

%s

This command outputs a completion script that can be sourced by your shell
to enable tab-completion for germinator commands, flags, and arguments.

The completion includes:
- Dynamic suggestions for --resources (from library)
- Dynamic suggestions for --preset (from library)
- Dynamic suggestions for 'library show' arguments
- Static suggestions for --platform values`, shell, getShellInstructions(shell)),
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			// Get the root command
			root := cmd.Root()

			// Generate snippet using carapace
			snippet, err := carapace.Gen(root).Snippet(shell)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error generating completion: %v\n", err)
				return
			}

			_, _ = fmt.Fprint(cmd.OutOrStdout(), snippet)
		},
	}
}
