package yoitsu

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
	"slices"
)

type Metadata struct {
	packageName string
}

type Accessors struct {
	// Generate an Accessor struct, loads data from Source
	Generate bool
	// If the jsons is of type JsonArray, will generate methods based on unique fields in the root of the
	// JsonObjects in the array
	ById bool
	// Unused
	GroupByPrimitive bool
}

type Yoitsu struct {
	// May be nil, populated after calling Yoitsu.GenerateFile
	File *ast.File

	src       Source
	metadata  Metadata
	universe  Universe
	accessors Accessors

	root   interface{}
	parser *Parser
}

// WithUniverse includes the passes Universe in Yoitsu.
// See Universe documentation
func WithUniverse(universe Universe) Option[*Yoitsu] {
	return func(y *Yoitsu) {
		y.universe = universe
	}
}

// WithPackageName sets the package name, defaults to "generated"
func WithPackageName(name string) Option[*Yoitsu] {
	return func(y *Yoitsu) {
		y.metadata.packageName = name
	}
}

// WithGenerateAccessors sets the accessor options, defaults to false on all
func WithGenerateAccessors(accessorOpt Option[*Accessors]) Option[*Yoitsu] {
	return func(y *Yoitsu) {
		accessorOpt(&y.accessors)
	}
}

// WithMetadata to further customize Metadata, currently unused
func WithMetadata(metaOpt Option[*Metadata]) Option[*Yoitsu] {
	return func(y *Yoitsu) {
		metaOpt(&y.metadata)
	}
}

// New create a new Yoitsu instance, it is recommended to create a new one per generation.
// See documentation for options on how to customize the output
func New(src Source, opts ...Option[*Yoitsu]) *Yoitsu {
	yt := &Yoitsu{
		src:       src,
		metadata:  Metadata{},
		accessors: Accessors{},
	}

	for _, opt := range opts {
		opt(yt)
	}

	if yt.metadata.packageName == "" {
		yt.metadata.packageName = "generated"
	}

	if yt.universe == nil {
		yt.universe = EmptyUniverse()
	}

	yt.parser = NewParser(yt)
	return yt
}

func (y *Yoitsu) getRootFromSrc() (interface{}, error) {
	b, err := y.src.Json()
	if err != nil {
		return nil, err
	}

	var root interface{}
	err = json.Unmarshal(b, &root)
	if err != nil {
		return nil, err
	}

	return root, nil
}

// GenerateFile parses the json from Source, and sets the Yoitsu.File field
func (y *Yoitsu) GenerateFile() (err error) {
	y.root, err = y.getRootFromSrc()
	if err != nil {
		return
	}

	var (
		gType       GeneratedType
		importSpecs []ast.Spec
		structDecls []ast.Decl

		accessorImports []ast.Spec
		accessorDecls   []ast.Decl
	)

	gType, importSpecs, structDecls, err = y.generateJsonTypes()
	if err != nil {
		return
	}

	accessorDecls, accessorImports, err = y.generateMethodAccessors(gType)
	if err != nil {
		return
	}

	var decls []ast.Decl
	var allImportSpecs []ast.Spec

	for _, spec := range append(importSpecs, accessorImports...) {
		alreadyAdded := slices.ContainsFunc(allImportSpecs, func(s ast.Spec) bool {
			return spec.(*ast.ImportSpec).Path.Value == s.(*ast.ImportSpec).Path.Value
		})
		if !alreadyAdded {
			allImportSpecs = append(allImportSpecs, spec)
		}
	}

	if len(allImportSpecs) > 0 {
		decls = append(decls, &ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: allImportSpecs,
		})
	}

	if len(structDecls) > 0 {
		decls = append(decls, structDecls...)
	}

	if len(accessorDecls) > 0 {
		decls = append(decls, accessorDecls...)
	}

	y.File = &ast.File{
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Text: fmt.Sprintf("// Generated by Yoitsu. DO NOT EDIT!"),
				},
			},
		},
		Name:  ast.NewIdent(y.metadata.packageName),
		Decls: decls,
	}

	return nil
}

func (y *Yoitsu) generateJsonTypes() (gType GeneratedType, importSpecs []ast.Spec, structDecls []ast.Decl, err error) {
	gType, err = y.parser.ParseRoot(y.src.Name(), y.root)
	if err != nil {
		return
	}

	structDecls = gType.Representation()
	importSpecs = y.imports(gType)
	return
}

func (y *Yoitsu) imports(gType GeneratedType) (imports []ast.Spec) {
	var addedImports []string

	for _, s := range gType.Imports() {
		if slices.Contains(addedImports, s) {
			continue
		}

		imports = append(imports, &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf("\"%s\"", s),
			},
		})

		addedImports = append(addedImports, s)
	}

	return
}

// WriteToDisk cannot be called while Yoitsu.File is nil, call Yoitsu.GenerateFile first. Writes file to disk
func (y *Yoitsu) WriteToDisk(dir string) error {
	if y.File == nil {
		return fmt.Errorf("no File generated. Call Yoitsu.GenerateFile first")
	}

	outFile, err := os.Create(filepath.Join(dir, y.src.Name()+".generated.go"))
	if err != nil {
		return err
	}
	defer outFile.Close()

	fset := token.NewFileSet()
	return format.Node(outFile, fset, y.File)
}
