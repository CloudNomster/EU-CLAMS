package cli

import (
	"flag"
	"fmt"
	"os"
)

// CLI manages command-line interface operations
type CLI struct {
	configPath string
	verbose    bool
	version    string
}

// NewCLI creates a new CLI instance
func NewCLI(version string) *CLI {
	return &CLI{
		version: version,
	}
}

// Parse parses command-line arguments
func (c *CLI) Parse() {
	flag.StringVar(&c.configPath, "config", "", "path to config file")
	flag.BoolVar(&c.verbose, "verbose", false, "enable verbose logging")

	versionFlag := flag.Bool("version", false, "display version information")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("EU-CLAMS v%s\n", c.version)
		os.Exit(0)
	}
}

// ConfigPath returns the config file path
func (c *CLI) ConfigPath() string {
	return c.configPath
}

// Verbose returns whether verbose logging is enabled
func (c *CLI) Verbose() bool {
	return c.verbose
}

// Version returns the application version
func (c *CLI) Version() string {
	return c.version
}
