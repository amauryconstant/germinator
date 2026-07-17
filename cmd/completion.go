package cmd

import (
	"context"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"gitlab.com/amoconst/germinator/internal/cmdutil"
	"gitlab.com/amoconst/germinator/internal/iostreams"
)

// shells returns the list of supported completion shell names. A
// fresh slice is returned on each call so callers cannot mutate the
// underlying storage, satisfying the cmd/AGENTS.md "no package-level
// mutable state" rule (Phase 6 migration; replacing the previous
// `var completionShells = []string{...}` declaration that the widened
// forbidigo pattern in .golangci.yml now flags as forbidden).
func shells() []string {
	return []string{
		"bash", "zsh", "fish", "powershell", "elvish",
		"nushell", "oil", "tcsh", "xonsh", "cmd-clink",
	}
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

    germinator completion fish > ~/.config/germinator/completions/germinator.fish`,
		"powershell": `Add to your PowerShell profile:

    germinator completion powershell | Out-String | Invoke-Expression

Also enable menu completion:

    Set-PSReadLineKeyHandler -Key Tab -Function MenuComplete`,
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

// completionOptions holds the runtime state for a `completion <shell>`
// invocation. IO and Ctx come from the Factory; Shell is set by the
// selected subcommand.
type completionOptions struct {
	IO    *iostreams.IOStreams
	Ctx   context.Context
	Shell string
}

// NewCmdCompletion creates the completion command tree with shell
// subcommands, wired through the canonical NewCmdXxx(f, runF)
// pattern. Carapace generates the actual completion scripts; this
// command only dispatches to the per-shell snippet generator.
func NewCmdCompletion(f *cmdutil.Factory, runF func(*completionOptions) error) *cobra.Command {
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
		Run: func(c *cobra.Command, _ []string) { _ = c.Help() },
	}

	for _, shell := range shells() {
		shell := shell // capture for closure
		cmd.AddCommand(newCompletionShellCommand(f, runF, shell))
	}

	return cmd
}

// newCompletionShellCommand creates a completion subcommand for a
// specific shell. The Factory is passed through so the per-shell
// command reads IO/Ctx from it.
func newCompletionShellCommand(f *cmdutil.Factory, runF func(*completionOptions) error, shell string) *cobra.Command {
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
		RunE: func(c *cobra.Command, _ []string) error {
			opts := &completionOptions{
				IO:    f.IOStreams,
				Ctx:   c.Context(),
				Shell: shell,
			}
			if runF != nil {
				return runF(opts)
			}
			return runCompletion(c, opts)
		},
	}
}

// runCompletion generates the carapace snippet for the requested shell
// and writes it to opts.IO.Out. Errors from carapace are wrapped so
// output.FormatError in main.go renders them once.
func runCompletion(cmd *cobra.Command, opts *completionOptions) error {
	snippet, err := carapace.Gen(cmd.Root()).Snippet(opts.Shell)
	if err != nil {
		return fmt.Errorf("generating %s completion snippet: %w", opts.Shell, err)
	}
	if _, err := fmt.Fprint(opts.IO.Out, snippet); err != nil {
		return fmt.Errorf("writing completion snippet: %w", err)
	}
	return nil
}
