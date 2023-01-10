package generator

import (
	"embed"
	"fmt"
	"strconv"
	"text/template"

	"github.com/vektah/gqlparser/v2/ast"
	"go.uber.org/zap"
	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed *.gotmpl
var tmplfs embed.FS

// Generator generates graphql and resolver code
type Generator struct {
	logs *zap.Logger
	tmpl *template.Template
	opts Options
}

// Options for the generator
type Options struct {
	QueryMessageName        string
	MutationMessageName     string
	SubscriptionMessageName string
}

// New inits the generator
func New(logs *zap.Logger, opts *Options) (g *Generator, err error) {
	g = &Generator{
		logs: logs.Named("generator"),
		tmpl: template.New("root"),
		opts: *opts,
	}

	g.tmpl = g.tmpl.Funcs(template.FuncMap{
		"unquote": strconv.Unquote,
	})

	g.tmpl, err = g.tmpl.ParseFS(tmplfs, "*.gotmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return
}

// NewTarget initializes a new target
func (g *Generator) NewTarget(pf *protogen.File) *Target {
	tg := &Target{
		gen:  g,
		file: pf,
		sch:  &ast.Schema{Types: make(map[string]*ast.Definition)},
	}

	tg.resolvers.mapped = make(map[string]*protogen.Method)
	tg.resolvers.unmapped = make(map[string]*protogen.Method)
	tg.resolvers.services = make(map[*protogen.Service]struct{})
	tg.resolvers.methods = make(map[*protogen.Method]struct{})

	return tg
}
