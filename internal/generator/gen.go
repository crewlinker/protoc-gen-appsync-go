package generator

import (
	"io"

	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/formatter"
	"go.uber.org/zap"
)

// Generator generates graphql and resolver code
type Generator struct{ logs *zap.Logger }

// New inits the generator
func New(logs *zap.Logger) *Generator {
	return &Generator{logs: logs.Named("generator")}
}

// GenerateSchema generates the graphql schema
func (g Generator) GenerateSchema(w io.Writer) error {
	sch := &ast.Schema{Types: make(map[string]*ast.Definition)}

	formatter.NewFormatter(w).FormatSchema(sch)
	return nil
}
