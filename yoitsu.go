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

type Yoitsu interface {
	GenerateFile() error
	WriteToDisk(dir string) error
}

type Metadata struct {
	packageName string
}

type Accessors struct {
	Generate         bool
	ById             bool
	GroupByPrimitive bool
}

type yoitsu struct {
	src      Source
	metadata Metadata
	universe Universe

	accessors Accessors

	root interface{}
	file *ast.File
}

func WithUniverse(universe Universe) Option[*yoitsu] {
	return func(y *yoitsu) {
		y.universe = universe
	}
}

func WithPackageName(name string) Option[*yoitsu] {
	return func(y *yoitsu) {
		y.metadata.packageName = name
	}
}

func WithGenerateAccessors(accessorOpt Option[*Accessors]) Option[*yoitsu] {
	return func(y *yoitsu) {
		accessorOpt(&y.accessors)
	}
}

func New(src Source, opts ...Option[*yoitsu]) Yoitsu {
	yt := &yoitsu{
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

	return yt
}

func (y *yoitsu) getRootFromSrc() (interface{}, error) {
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

func (y *yoitsu) GenerateFile() (err error) {
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

	if genType, ok := gType.(*generatedType); ok {
		accessorDecls, accessorImports, err = y.generateMethodAccessors(genType)
		if err != nil {
			return
		}
	}

	var decls []ast.Decl
	var allImportSpecs []ast.Spec

	if len(importSpecs) > 0 {
		allImportSpecs = append(allImportSpecs, importSpecs...)
	}

	if len(accessorImports) > 0 {
		allImportSpecs = append(allImportSpecs, accessorImports...)
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

	y.file = &ast.File{
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Text: fmt.Sprintf("// Generated by yoitsu. DO NOT EDIT!"),
				},
			},
		},
		Name:  ast.NewIdent(y.metadata.packageName),
		Decls: decls,
	}

	return nil
}

func (y *yoitsu) generateJsonTypes() (gType GeneratedType, importSpecs []ast.Spec, structDecls []ast.Decl, err error) {
	gType, err = y.parseRootType()
	if err != nil {
		return
	}

	structDecls = gType.Representation()
	importSpecs = y.imports(gType)
	return
}

func (y *yoitsu) imports(gType GeneratedType) (imports []ast.Spec) {
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

func (y *yoitsu) parseRootType() (gType GeneratedType, err error) {
	if y.root == nil {
		err = ErrNoData
		return
	}

	switch y.root.(type) {
	case JsonMap:
		gType, err = ParseType(y.src.Name(), y.root.(JsonMap), y.universe)
	case []interface{}:
		gType, err = ParseTypes(y.src.Name(), y.root.([]interface{}), y.universe)
	default:
		gType, err = parse(y.src.Name(), y.root, y.universe)
	}

	if err != nil {
		return
	}

	if gType == nil {
		err = ErrNoData
		return
	}

	return
}

func (y *yoitsu) WriteToDisk(dir string) error {
	if y.file == nil {
		return fmt.Errorf("no file generated. Call Yoitsu.GenerateFile first")
	}

	outFile, err := os.Create(filepath.Join(dir, y.src.Name()+".generated.go"))
	if err != nil {
		return err
	}
	defer outFile.Close()

	fset := token.NewFileSet()
	return format.Node(outFile, fset, y.file)
}
