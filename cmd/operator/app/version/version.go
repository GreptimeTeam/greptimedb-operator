package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/greptime/greptimedb-operator/pkg/version"
)

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of greptimedb-operator and exit",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GreptimeDB Operator Version\n---------------------------\n")
			fmt.Printf("%s", version.Get())
		},
	}
}
