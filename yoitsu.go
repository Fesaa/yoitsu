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
	Generate() error
	WriteToDisk(dir string) error
}

type yoitsu struct {
	src           Source
	existingTypes []GeneratedType
	file          *ast.File
	packageName   string
}

func WithExistingType(gt GeneratedType) Option[*yoitsu] {
	return func(y *yoitsu) {
		y.existingTypes = append(y.existingTypes, gt)
	}
}

func WithPackageName(name string) Option[*yoitsu] {
	return func(y *yoitsu) {
		y.packageName = name
	}
}

func New(src Source, opts ...Option[*yoitsu]) Yoitsu {
	yt := &yoitsu{src: src}

	for _, opt := range opts {
		opt(yt)
	}

	if yt.packageName == "" {
		yt.packageName = "generated"
	}

	return yt
}

func (y *yoitsu) root() (interface{}, error) {
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

func (y *yoitsu) Generate() error {
	uni := NewUniverse()
	for _, t := range y.existingTypes {
		uni.AddType(t)
	}

	root, err := y.root()
	if err != nil {
		return err
	}

	var gType GeneratedType
	switch root.(type) {
	case JsonMap:
		gType, err = ParseType(y.src.Name(), root.(JsonMap), uni)
	case []interface{}:
		gType, err = ParseTypes(y.src.Name(), root.([]interface{}), uni)
	default:
		gType, err = parse(y.src.Name(), root, uni)
	}

	if err != nil {
		return err
	}

	if gType == nil {
		return ErrorNoData
	}

	structDecl := gType.Representation()
	decls := make([]ast.Decl, 0)

	for _, sd := range structDecl {
		decls = append(decls, sd)
	}

	imports := y.imports(gType)
	if imports != nil {
		decls = append([]ast.Decl{imports}, decls...)
	}

	y.file = &ast.File{
		Name:  ast.NewIdent(y.packageName),
		Decls: decls,
	}

	return nil
}

func (y *yoitsu) imports(gType GeneratedType) ast.Decl {
	var imports []ast.Spec
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

	if len(imports) == 0 {
		return nil
	}

	return &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: imports,
	}
}

func (y *yoitsu) WriteToDisk(dir string) error {
	if y.file == nil {
		return fmt.Errorf("no file generated. Call Yoitsu.Generate first")
	}

	outFile, err := os.Create(filepath.Join(dir, y.src.Name()+".generated.go"))
	if err != nil {
		return err
	}
	defer outFile.Close()

	fset := token.NewFileSet()
	return format.Node(outFile, fset, y.file)
}
