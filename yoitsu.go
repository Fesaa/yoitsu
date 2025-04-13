package yoitsu

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
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
	}

	if err != nil {
		return err
	}

	structDecl := gType.Representation()
	decls := make([]ast.Decl, 0)

	for _, sd := range structDecl {
		decls = append(decls, sd)
	}

	y.file = &ast.File{
		Name:  ast.NewIdent(y.packageName),
		Decls: decls,
	}

	return nil
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
