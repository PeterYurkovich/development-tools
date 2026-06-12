package mode

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func HasAllRequiredFlags(cmd *cobra.Command, requiredFlags []string) bool {
	for _, flagName := range requiredFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			return false
		}

		if !flag.Changed {
			return false
		}
	}
	return true
}

func GetMissingFlags(cmd *cobra.Command, requiredFlags []string) []string {
	missing := []string{}
	for _, flagName := range requiredFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil || !flag.Changed {
			missing = append(missing, flagName)
		}
	}
	return missing
}

func DetermineMode(cmd *cobra.Command, requiredFlags []string) (useTUI bool, err error) {
	inTerminal := IsTerminal()
	hasAllFlags := HasAllRequiredFlags(cmd, requiredFlags)

	if !inTerminal {
		if !hasAllFlags {
			missing := GetMissingFlags(cmd, requiredFlags)
			flagNames := make([]string, len(missing))
			for i, name := range missing {
				flagNames[i] = "--" + name
			}
			return false, fmt.Errorf("not running in terminal and missing required flags: %s", strings.Join(flagNames, ", "))
		}
		return false, nil
	}

	if hasAllFlags {
		return false, nil
	}

	return true, nil
}
