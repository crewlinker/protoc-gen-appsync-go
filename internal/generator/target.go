package generator

import (
	"fmt"
	"io"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/formatter"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Target provides methods for generating from a single proto file
type Target struct {
	gen *Generator
	sch *ast.Schema

	file *protogen.File

	resolvers struct {
		unmapped map[string]*protogen.Method
		mapped   map[string]*protogen.Method
		methods  map[*protogen.Method]struct{}
		services map[*protogen.Service]struct{}
	}
}

// TargetData is exposed to our templates
type TargetData struct {
	*protogen.File
	Resolvers        map[string]*protogen.Method
	ResolverMethods  map[*protogen.Method]struct{}
	ResolverServices map[*protogen.Service]struct{}
}

// Generate the target and write an graph schema and resolver code.
func (tg *Target) Generate(graphw, resolvew io.Writer) error {

	// index service methods and if they are marked to resolve a field
	for _, svc := range tg.file.Services {
		for _, met := range svc.Methods {
			mopts := MethodOptions(met)
			if mopts != nil && len(mopts.Resolves) > 0 {
				for _, res := range mopts.Resolves {
					// unmapped resolvers will be mapped during schema generation
					tg.resolvers.unmapped[res] = met

					// map unique services that resolve
					tg.resolvers.methods[met] = struct{}{}
					tg.resolvers.services[met.Parent] = struct{}{}
				}
			}
		}
	}

	// generate the graphql schema, also populating the field index
	if err := tg.generateSchema(); err != nil {
		return fmt.Errorf("failed to generate schema: %w", err)
	}

	// fail if resolving was configured but no field hooked it up after generating the schema
	if len(tg.resolvers.unmapped) > 0 {
		for q, met := range tg.resolvers.unmapped {
			return fmt.Errorf("%s.%s resolves field '%s' but it was not found under the root messages (%s, %s or %s)",
				met.Parent.Desc.Name(), met.Desc.Name(),
				q,
				tg.gen.opts.QueryMessageName,
				tg.gen.opts.MutationMessageName,
				tg.gen.opts.SubscriptionMessageName)
		}
	}

	// output the graphql schema text
	formatter.NewFormatter(graphw).FormatSchema(tg.sch)

	// generate and output the resolving code
	if err := tg.gen.tmpl.ExecuteTemplate(resolvew, "resolve.gotmpl", TargetData{
		File:             tg.file,
		Resolvers:        tg.resolvers.mapped,
		ResolverMethods:  tg.resolvers.methods,
		ResolverServices: tg.resolvers.services,
	}); err != nil {
		return fmt.Errorf("failed to generate resolving code: %w", err)
	}

	return nil
}

// generateSchema populate the target's graphql schema definition
func (tg *Target) generateSchema() error {

	// find the messages that make up the root graphql types: Query, Mutation and Subscription
	for _, msg := range tg.file.Messages {
		switch string(msg.Desc.Name()) {
		case tg.gen.opts.QueryMessageName, tg.gen.opts.MutationMessageName, tg.gen.opts.SubscriptionMessageName:

			// one of the roots of the graphql tree, start recursing down to generate schema definitions
			_, err := tg.generateMessage(false, msg)
			if err != nil {
				return fmt.Errorf("failed to generate message '%s': %w", msg.Desc.Name(), err)
			}

		default:
			continue
		}
	}

	return nil
}

// generateMessage generates graphql object/input type definitions from protobuf messages
func (tg *Target) generateMessage(isInput bool, msg *protogen.Message) (def *ast.Definition, err error) {
	def = &ast.Definition{Name: msg.GoIdent.GoName, Kind: ast.Object, Fields: ast.FieldList{}}

	// if we're traversing the input side of the graph, create input defs instead
	if isInput {
		def.Kind = ast.InputObject
		def.Name = def.Name + "Input" // prevent name collisions if inputs are the same message
	}

	// if it's already defined we don't do it again, else it causes infinite loops in case of recursion
	if _, ok := tg.sch.Types[def.Name]; ok {
		return tg.sch.Types[def.Name], nil
	}

	// add the type in the graphql schema, return the name
	tg.sch.Types[def.Name] = def

	// generate graphql field definitions for each field in the message
	for _, fld := range msg.Fields {
		if fopts := FieldOptions(fld); fopts != nil && fopts.Ignore != nil && *fopts.Ignore {
			continue // skip ignored field
		}

		fdef, err := tg.generateField(isInput, fld)
		if err != nil {
			return nil, fmt.Errorf("failed to generate field '%s': %w", fld.Desc.Name(), err)
		}

		def.Fields = append(def.Fields, fdef)
	}

	return def, nil
}

// generateField generates graphql field definitions from the protobuf message field
func (tg *Target) generateField(isInput bool, fld *protogen.Field) (def *ast.FieldDefinition, err error) {
	def = &ast.FieldDefinition{Name: fld.Desc.JSONName(), Type: &ast.Type{NonNull: true}}

	// if a rpc method was configured to be resolving this field, add any arguments.
	// if we're building input the fields never have arguments
	protoQualifier := fmt.Sprintf("%s.%s", fld.Parent.Desc.Name(), fld.Desc.Name())
	graphQualifier := fmt.Sprintf("%s.%s", fld.Parent.Desc.Name(), fld.Desc.JSONName())
	if resolver, ok := tg.resolvers.unmapped[protoQualifier]; ok && !isInput {
		if def.Arguments, err = tg.generateArguments(fld, resolver); err != nil {
			return nil, fmt.Errorf("failed to generate arguments: %w", err)
		}

		delete(tg.resolvers.unmapped, protoQualifier)  // remove from map so we can error if some resolves failed
		tg.resolvers.mapped[graphQualifier] = resolver // add to resolver map for generating go resolver code
	}

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
		mdef, err := tg.generateMessage(isInput, fld.Message)
		if err != nil {
			return nil, fmt.Errorf("failed to generate nested message definition: %w", err)
		}

		def.Type.NamedType = mdef.Name

	// enums are also supported in graphql
	case fld.Desc.Kind() == protoreflect.EnumKind:

		// generate num type
		edef, err := tg.generateEnum(isInput, fld.Enum)
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

// generateEnum generates graphql enum type from protobuf enum field
func (tg *Target) generateEnum(isInput bool, enum *protogen.Enum) (def *ast.Definition, err error) {
	def = &ast.Definition{Kind: ast.Enum, Name: enum.GoIdent.GoName, EnumValues: ast.EnumValueList{}}
	for _, val := range enum.Values {

		// we cut off the first prefix to make the protojson decoding work. The Go ident always includes the
		// enum name in the first part before the "_" so we turn "Status_STATUS_OK" into "STATUS_OK"
		_, name, _ := strings.Cut(val.GoIdent.GoName, "_")
		def.EnumValues = append(def.EnumValues, &ast.EnumValueDefinition{Name: name})
	}

	tg.sch.Types[def.Name] = def
	return def, nil
}

// generateArguments generates graphql arguments from the service method in the options
func (tg *Target) generateArguments(fld *protogen.Field, res *protogen.Method) (def ast.ArgumentDefinitionList, err error) {
	for _, infld := range res.Input.Fields {
		if fopts := FieldOptions(infld); fopts != nil && fopts.Ignore != nil && *fopts.Ignore {
			continue // skip argument if field is ignored
		}

		fdef, err := tg.generateField(true, infld)
		if err != nil {
			return nil, fmt.Errorf("failed to generate argument field: %w", err)
		}

		def = append(def, &ast.ArgumentDefinition{Name: fdef.Name, Type: fdef.Type})
	}

	return
}
