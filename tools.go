//go:build tools

package tools

import (
	_ "github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go"
	_ "github.com/envoyproxy/protoc-gen-validate"
	_ "github.com/magefile/mage"
	_ "github.com/onsi/ginkgo/v2/ginkgo"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
