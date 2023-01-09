package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/sh"
)

// init performs some sanity checks before running anything
func init() {
	mustBeInRoot()
}

// Test perform the whole project's unit tests
func Test() error {
	return sh.Run(
		"go", "run", "-mod=readonly", "github.com/onsi/ginkgo/v2/ginkgo",
		"-p", "-randomize-all", "-repeat=5", "--fail-on-pending", "--race",
		"--trace", "--junit-report=test-report.xml", "./...")
}

// profileFromEnv determines the AWS credentials profile from the env argument
func profileFromEnv(env string) string {
	if env != "prod" && env != "stag" && env != "dev" {
		panic("invalid env: '" + env + "', supports: 'prod', 'stag' or 'dev'")
	}
	return fmt.Sprintf("cl-%s", env)
}

// mustBeInRoot checks that the command is run in the project root
func mustBeInRoot() {
	if _, err := os.Stat("go.mod"); err != nil {
		panic("must be in root, couldn't stat go.mod file: " + err.Error())
	}
}
