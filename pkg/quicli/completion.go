package quicli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateCompletion returns a shell completion script for the given Cli and shell.
func generateCompletion(c *Cli, shell string) (string, error) {
	switch shell {
	case "bash":
		return generateBashCompletion(c), nil
	case "zsh":
		return generateZshCompletion(c), nil
	case "fish":
		return generateFishCompletion(c), nil
	default:
		return "", fmt.Errorf("unsupported shell %q — use bash, zsh, or fish", shell)
	}
}

func cliProgName() string {
	return filepath.Base(os.Args[0])
}

// allSubcommandNames returns all subcommand names and aliases as a flat slice.
func allSubcommandNames(subs Subcommands) []string {
	var names []string
	for _, s := range subs {
		names = append(names, s.Name)
		names = append(names, s.Aliases...)
	}
	return names
}

// allFlagNames returns --long and -short flag names for the given flags.
func allFlagNames(flags []Flag) []string {
	var names []string
	seen := map[string]bool{}
	for _, f := range flags {
		names = append(names, "--"+f.Name)
		short := f.ShortName
		if short == "" {
			short = f.Name[0:1]
		}
		if !f.NoShortName && !seen[short] {
			names = append(names, "-"+short)
			seen[short] = true
		}
	}
	return names
}

func generateBashCompletion(c *Cli) string {
	prog := cliProgName()
	fnName := "_" + strings.ReplaceAll(prog, "-", "_") + "_completion"
	cmds := strings.Join(allSubcommandNames(c.Subcommands), " ")
	flags := strings.Join(allFlagNames(c.Flags), " ")

	var b strings.Builder
	fmt.Fprintf(&b, "%s() {\n", fnName)
	fmt.Fprintf(&b, "    local cur=\"${COMP_WORDS[COMP_CWORD]}\"\n")
	if len(c.Subcommands) > 0 {
		fmt.Fprintf(&b, "    local commands=\"%s\"\n", cmds)
	}
	fmt.Fprintf(&b, "    local flags=\"%s\"\n", flags)
	fmt.Fprintf(&b, "\n")
	if len(c.Subcommands) > 0 {
		fmt.Fprintf(&b, "    if [[ $COMP_CWORD -eq 1 ]]; then\n")
		fmt.Fprintf(&b, "        COMPREPLY=( $(compgen -W \"$commands $flags\" -- \"$cur\") )\n")
		fmt.Fprintf(&b, "    else\n")
		fmt.Fprintf(&b, "        COMPREPLY=( $(compgen -W \"$flags\" -- \"$cur\") )\n")
		fmt.Fprintf(&b, "    fi\n")
	} else {
		fmt.Fprintf(&b, "    COMPREPLY=( $(compgen -W \"$flags\" -- \"$cur\") )\n")
	}
	fmt.Fprintf(&b, "}\n")
	fmt.Fprintf(&b, "complete -F %s %s\n", fnName, prog)
	return b.String()
}

func generateZshCompletion(c *Cli) string {
	prog := cliProgName()
	fnName := "_" + strings.ReplaceAll(prog, "-", "_")

	var b strings.Builder
	fmt.Fprintf(&b, "#compdef %s\n", prog)
	fmt.Fprintf(&b, "%s() {\n", fnName)

	if len(c.Subcommands) > 0 {
		fmt.Fprintf(&b, "    local -a commands\n")
		fmt.Fprintf(&b, "    commands=(\n")
		for _, s := range c.Subcommands {
			fmt.Fprintf(&b, "        '%s:%s'\n", s.Name, s.Description)
		}
		fmt.Fprintf(&b, "    )\n")
		fmt.Fprintf(&b, "    local -a flags\n")
		fmt.Fprintf(&b, "    flags=(\n")
		for _, f := range c.Flags {
			fmt.Fprintf(&b, "        '--%s[%s]'\n", f.Name, f.Description)
		}
		fmt.Fprintf(&b, "    )\n")
		fmt.Fprintf(&b, "    if (( CURRENT == 2 )); then\n")
		fmt.Fprintf(&b, "        _describe 'commands' commands\n")
		fmt.Fprintf(&b, "    else\n")
		fmt.Fprintf(&b, "        _arguments $flags\n")
		fmt.Fprintf(&b, "    fi\n")
	} else {
		fmt.Fprintf(&b, "    local -a flags\n")
		fmt.Fprintf(&b, "    flags=(\n")
		for _, f := range c.Flags {
			fmt.Fprintf(&b, "        '--%s[%s]'\n", f.Name, f.Description)
		}
		fmt.Fprintf(&b, "    )\n")
		fmt.Fprintf(&b, "    _arguments $flags\n")
	}

	fmt.Fprintf(&b, "}\n")
	fmt.Fprintf(&b, "%s \"$@\"\n", fnName)
	return b.String()
}

func generateFishCompletion(c *Cli) string {
	prog := cliProgName()
	var b strings.Builder

	for _, s := range c.Subcommands {
		fmt.Fprintf(&b, "complete -c %s -n \"__fish_use_subcommand\" -f -a %s -d '%s'\n",
			prog, s.Name, s.Description)
		if s.Aliases != nil {
			for _, alias := range s.Aliases {
				fmt.Fprintf(&b, "complete -c %s -n \"__fish_use_subcommand\" -f -a %s -d '%s (alias)'\n",
					prog, alias, s.Description)
			}
		}
	}
	for _, f := range c.Flags {
		short := f.ShortName
		if short == "" {
			short = f.Name[0:1]
		}
		if f.NoShortName {
			fmt.Fprintf(&b, "complete -c %s -l %s -d '%s'\n", prog, f.Name, f.Description)
		} else {
			fmt.Fprintf(&b, "complete -c %s -s %s -l %s -d '%s'\n", prog, short, f.Name, f.Description)
		}
	}
	return b.String()
}
