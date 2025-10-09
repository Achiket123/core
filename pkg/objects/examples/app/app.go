package app

import "github.com/urfave/cli/v3"

// NewCommand constructs the root command used to run all consolidated examples.
func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "objects-examples",
		Usage: "Run object storage examples",
		Commands: []*cli.Command{
			simpleCommand(),
			simpleS3Command(),
			multiProviderCommand(),
			multiTenantCommand(),
			openlaneCommand(),
		},
	}
}
