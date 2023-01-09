package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/crewlinker/protoc-gen-appsync-go/infra/cdknaming"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Infra namespace holds infrastructure management
type Infra mg.Namespace

// Bootstarp bootstrap the CDK toolkit
func (Infra) Bootstrap(env string) error {
	profile := profileFromEnv(env)
	return infraRun("cdk", "bootstrap",
		"--profile", profile,
		"--cloudformation-execution-policies", strings.Join([]string{
			"arn:aws:iam::aws:policy/AmazonSSMFullAccess",
			"arn:aws:iam::aws:policy/AWSLambda_FullAccess",
			"arn:aws:iam::aws:policy/CloudWatchFullAccess",
			"arn:aws:iam::aws:policy/AmazonS3FullAccess",
			"arn:aws:iam::aws:policy/IAMFullAccess",
			"arn:aws:iam::aws:policy/AWSAppSyncAdministrator",
		}, ","),
	)
}

// Build builds artifacts required for deploying to AWS
func (Infra) Build() (err error) {
	for _, p := range [][]string{
		{"lambda", "examplesimple"},
		{"lambda", "examplenested"},
	} {
		if err = buildLambda(p...); err != nil {
			return err
		}
	}

	return
}

// Diff calculates the diff for our infrastructure deploy
func (Infra) Diff(env, instance string) error {
	profile := profileFromEnv(env)
	return infraRun("cdk", "diff",
		"--profile", profile,
		"--context", "instance="+instance,
	)
}

// Deploy deploy our infrastructure
func (Infra) Deploy(env, instance string) error {
	profile := profileFromEnv(env)
	return infraRun("cdk", "deploy",
		"--require-approval", "never",
		"--profile", profile,
		"--context", "instance="+instance,
		"--outputs-file", filepath.Join("cdk.outputs.json"),
	)
}

// buildLambda will generate and build a lambda executable for the lambda code in directory 'p'.
func buildLambda(p ...string) (err error) {
	dstdir := filepath.Join(p...)
	tmpdir, _ := os.MkdirTemp("", "cwrs_build_*")
	err = runIfNoErr(err, nil, "rm", "-f", filepath.Join(dstdir, "pkg.zip"))
	err = runIfNoErr(err, map[string]string{"GOOS": "linux", "GOARCH": "amd64"},
		"go", "build", "-trimpath", "-tags", "lambda.norpc",
		"-o", filepath.Join(tmpdir, "bootstrap"), "."+string(filepath.Separator)+dstdir)
	err = runIfNoErr(err, nil, "touch", "-t", "200906122350", filepath.Join(tmpdir, "bootstrap"))
	err = runIfNoErr(err, nil, "zip", "-r", "-j", "--latest-time", "-X", filepath.Join(dstdir, "pkg.zip"), tmpdir)
	return
}

// runIfNoErr will only run cmd with args if 'err' is nil, else it will return err. This allows us to
// make somewhat readable automation around scripts
func runIfNoErr(err error, env map[string]string, cmd string, args ...string) error {
	if err != nil {
		return err
	}
	return sh.RunWith(env, cmd, args...)
}

// infraRun will call sh.Run but just after calling chdir into the infra dir
func infraRun(cmd string, args ...string) error {
	if err := os.Chdir("infra"); err != nil {
		return fmt.Errorf("failed to chdir: %w", err)
	}

	// setup qualifier settings so we isolate our bootstrap between projects
	args = append(args, []string{
		"--toolkit-stack-name", cdknaming.Qualifier + "Bootstrap",
		"--qualifier", strings.ToLower(cdknaming.Qualifier),
	}...)

	return sh.Run(cmd, args...)
}
