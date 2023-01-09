package generator

import (
	"embed"
	"fmt"
	"io"
	"strings"
	"text/template"

	appsyncv1 "github.com/crewlinker/protoc-gen-appsync-go/proto/appsync/v1"
	"github.com/iancoleman/strcase"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/formatter"
	"go.uber.org/zap"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
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
		"method_options": g.MethodOptions,
	})

	g.tmpl, err = g.tmpl.ParseFS(tmplfs, "*.gotmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return
}

// methodOptions returns our plugin specific options for a method. Returns nil if no options are set
// for a method
func (g Generator) MethodOptions(m *protogen.Method) *appsyncv1.MethodOptions {
	mopts, ok := m.Desc.Options().(*descriptorpb.MethodOptions)
	if !ok {
		return nil
	}
	mext, ok := proto.GetExtension(mopts, appsyncv1.E_Method).(*appsyncv1.MethodOptions)
	if !ok {
		return nil
	}
	if mext == nil {
		return nil
	}

	// default the field name to lowerCamelCase of the method name
	if mext.ResolveField == nil {
		mname := strcase.ToLowerCamel(m.GoName)
		mext.ResolveField = &mname
	}

	return mext
}

// GenerateResolve generates resolver code
func (g Generator) GenerateResolve(w io.Writer, pf *protogen.File) error {
	return g.tmpl.ExecuteTemplate(w, "resolve.gotmpl", pf)
}

// generateEnum generates graphql enum type from protobuf enum field
func (g Generator) generateEnum(sch *ast.Schema, pf *protogen.File, isInput bool, depth int, enum *protogen.Enum) (def *ast.Definition, err error) {
	def = &ast.Definition{Kind: ast.Enum, Name: enum.GoIdent.GoName, EnumValues: ast.EnumValueList{}}
	for _, val := range enum.Values {
		// we cut off the first prefix to make the protojson decoding work. The Go ident always includes the
		// enum name in the first part before the "_" so we turn "Status_STATUS_OK" into "STATUS_OK"
		_, name, _ := strings.Cut(val.GoIdent.GoName, "_")
		def.EnumValues = append(def.EnumValues, &ast.EnumValueDefinition{Name: name})
	}

	sch.Types[def.Name] = def
	return def, nil
}

// generateField generates graphql field definitions from the protobuf message field
func (g Generator) generateField(sch *ast.Schema, pf *protogen.File, isInput bool, depth int, fld *protogen.Field) (def *ast.FieldDefinition, err error) {
	def = &ast.FieldDefinition{Name: fld.Desc.JSONName(), Type: &ast.Type{NonNull: true}}

	// the explicit optional keyword is different from the "optional" cardinality
	if fld.Desc.HasOptionalKeyword() {
		def.Type.NonNull = false
	}

	switch {

	// basic scalar types
	case fld.Desc.Kind() == protoreflect.StringKind:
		def.Type.NamedType = "String"

	// messages are an object and recurse
	case fld.Desc.Kind() == protoreflect.MessageKind:

		// recurse, so we can include nested messages
		mdef, err := g.generateMessage(sch, pf, isInput, depth+1, fld.Message)
		if err != nil {
			return nil, fmt.Errorf("failed to generate nested message definition: %w", err)
		}

		def.Type.NamedType = mdef.Name

	// enums are also supported in graphql
	case fld.Desc.Kind() == protoreflect.EnumKind:

		// generate num type
		edef, err := g.generateEnum(sch, pf, isInput, depth, fld.Enum)
		if err != nil {
			return nil, fmt.Errorf("failed to generate nested enum definition: %w", err)
		}

		def.Type.NamedType = edef.Name

	default:
		return nil, fmt.Errorf("unsupported field: Kind=%v Desc=%v", fld.Desc.Kind(), fld.Desc)
	}

	// for repeated fields we turn the field type into the element instead
	if fld.Desc.Cardinality() == protoreflect.Repeated {
		switch {
		case fld.Desc.IsList():
			// we never allow "null" to be passed as list value
			def.Type.Elem = &ast.Type{NamedType: def.Type.NamedType, NonNull: true}
			def.Type.NamedType = "" // reset to non-named, to allow elem
		case fld.Desc.IsMap():
			return nil, fmt.Errorf("maps using the 'shortcut' notation are not supported, use the 'legacy' structure instead (https://github.com/golang/protobuf/issues/1511)")
		default:
			return nil, fmt.Errorf("unsupported repeated cardinality, not List or Map")
		}
	}

	return
}

// generateMessage generates graphql object/input type definitions from protobuf messages
func (g Generator) generateMessage(sch *ast.Schema, pf *protogen.File, isInput bool, depth int, msg *protogen.Message) (def *ast.Definition, err error) {
	def = &ast.Definition{Name: msg.GoIdent.GoName, Kind: ast.Object, Fields: ast.FieldList{}}

	// generate graphql field definitions for each field in the message
	for _, fld := range msg.Fields {
		fdef, err := g.generateField(sch, pf, isInput, depth, fld)
		if err != nil {
			return nil, fmt.Errorf("failed to generate field '%s': %w", fld.Desc.Name(), err)
		}

		def.Fields = append(def.Fields, fdef)
	}

	// add the type in the graphql schema, return the name
	sch.Types[def.Name] = def
	return def, nil
}

// GenerateSchema generates the graphql schema
func (g Generator) GenerateSchema(w io.Writer, pf *protogen.File) error {
	sch := &ast.Schema{Types: make(map[string]*ast.Definition)}

	for _, msg := range pf.Messages {
		switch string(msg.Desc.Name()) {
		case g.opts.QueryMessageName, g.opts.MutationMessageName, g.opts.SubscriptionMessageName:

			// one of the roots of the graphql tree, start recursing down to generate schema definitions
			_, err := g.generateMessage(sch, pf, false, 0, msg)
			if err != nil {
				return fmt.Errorf("failed to generate message '%s': %w", msg.Desc.Name(), err)
			}

		default:
			continue
		}
	}

	// @TODO instead, add a "resolver" field option that selects the service.method name
	// @TODO allow an option to resolve a member of the "Response" message. For example if a field
	// needs to return a scalar.

	// @TODO 1. first walk messages to create the "output" side of the graphql schema. That is
	// start at top level Query, Mutation, etc. And create all the object (non) input types
	// @TODO 2. then walk the rpc methods and "hook" them up to the the tree created in step one (i.e)
	// check if the required methods exist
	// @TODO 3. then for each method, walk the Input side by walking the method.Input field.

	// for _, svc := range pf.Services {
	// 	for _, met := range svc.Methods {
	// 		opts := g.MethodOptions(met)
	// 		if opts == nil || opts.ResolveOn == nil {
	// 			continue // no method options, or not configured as resolver, skip further generation
	// 		}

	// 		// for methods that are declared on Query or Mutation types we
	// 		// don't expect protobuf to declare them so we need to add them
	// 		// themselves
	// 		switch *opts.ResolveOn {
	// 		case "Query":
	// 			def := &ast.Definition{Kind: ast.Object, Name: "Query", Fields: ast.FieldList{}}
	// 			sch.Types[def.Name] = def
	// 		case "Mutation":
	// 			def := &ast.Definition{Kind: ast.Object, Name: "Mutation", Fields: ast.FieldList{}}
	// 			sch.Types[def.Name] = def
	// 		}

	// 		// @TODO allow configuration of
	// 		fieldName := strcase.ToLowerCamel(met.GoName)
	// 		typName := opts.ResolveOn

	// 		g.printf("AAAAA: %s %s %v %v", svc.GoName, met.GoName, *typName, fieldName)
	// 	}
	// }

	// output by formatting the schema
	formatter.NewFormatter(w).FormatSchema(sch)
	return nil
}

// convenient print method for debugging
func (g Generator) printf(s string, v ...any) {
	g.logs.Info("print", zap.String("line:", fmt.Sprintf(s, v...)))
}
