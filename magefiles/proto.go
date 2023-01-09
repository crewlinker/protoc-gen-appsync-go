package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Infra namespace holds protobuf management
type Proto mg.Namespace

// Lint lint the protobuf
func (Proto) Lint() error {
	return sh.Run("buf", "lint")
}

// Gen generate the protobuf message code
func (Proto) Gen() error {
	return sh.Run("buf", "generate")
}
