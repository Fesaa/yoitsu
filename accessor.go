package yoitsu

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"
)

const (
	tokenAccessor   string = "Accessor"
	tokenData       string = "_data"
	tokenIdentifier string = "identifier"

	tokenArray string = "[]"
	tokenMap   string = "map[%s]%s"

	tokenReceiver string = "a"
	tokenPointer  string = "*"

	tokenMethodLoadName  string = "LoadData"
	tokenMethodGroupData string = "GroupData"
	tokenError           string = "error"

	tokenAllMethod string = "All"
)

func (y *yoitsu) generateMethodAccessors(gType *generatedType) (decls []ast.Decl, importSpecs []ast.Spec, err error) {
	if !y.accessors.Generate {
		return
	}

	_, isArray := y.root.([]interface{})
	if !isArray {
		err = ErrAccessorsAreNotGeneratedForArrays
		return
	}

	fieldList := ast.FieldList{}
	safeName := toSafeGoName(gType.JsonType().TypeName())
	structName := safeName + tokenAccessor

	accessorsStruct := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(structName),
				Type: &ast.StructType{
					Fields: &fieldList,
				},
			},
		},
	}

	fieldList.List = append(fieldList.List, &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(tokenData)},
		Type:  ast.NewIdent(tokenArray + safeName),
	})

	decls = append(decls, accessorsStruct)

	var (
		decl       ast.Decl
		importSpec []ast.Spec
	)

	decl, importSpec = y.src.LoadMethod(structName)
	decls = append(decls, decl)
	importSpecs = append(importSpecs, importSpec...)

	decls = append(decls, y.allMethod(gType))

	if y.accessors.ById {
		var uniqueDecls []ast.Decl
		uniqueDecls, importSpec = y.uniqueGetters(gType, &fieldList)

		if len(uniqueDecls) > 0 {
			decls = append(decls, uniqueDecls...)
		}
		if len(importSpec) > 0 {
			importSpecs = append(importSpecs, importSpec...)
		}
	}

	return
}

func (y *yoitsu) allMethod(gType GeneratedType) ast.Decl {
	safeName := toSafeGoName(gType.JsonType().TypeName())
	structName := safeName + tokenAccessor

	receiver := &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{ast.NewIdent(tokenReceiver)},
				Type:  ast.NewIdent(tokenPointer + structName),
			},
		},
	}

	funcName := ast.NewIdent(tokenAllMethod)
	funcType := &ast.FuncType{
		Params: &ast.FieldList{},
		Results: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: ast.NewIdent(tokenArray + safeName),
				},
			},
		},
	}

	funcBody := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.SelectorExpr{
						X:   ast.NewIdent(tokenReceiver),
						Sel: ast.NewIdent(tokenData),
					},
				},
			},
		},
	}

	return &ast.FuncDecl{
		Recv: receiver,
		Name: funcName,
		Type: funcType,
		Body: funcBody,
	}
}

func (y *yoitsu) uniqueGetters(gType *generatedType, fieldList *ast.FieldList) (decls []ast.Decl, importSpec []ast.Spec) {
	unqiueJsonPrimitives := y.uniqueJsonPrimitives(gType)

	if len(unqiueJsonPrimitives) == 0 {
		return
	}

	decls = append(decls, y.groupByMethod(gType, unqiueJsonPrimitives))

	for _, ujp := range unqiueJsonPrimitives {

		decl := y.uniqueJsonPrimitivesAccessor(gType, ujp)
		if decl == nil {
			continue
		}

		fieldList.List = append(fieldList.List, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(tokenData + toSafeGoName(ujp.Name()))},
			Type:  ast.NewIdent(fmt.Sprintf(tokenMap, ujp.JsonType().TypeName(), gType.Name())),
		})

		decls = append(decls, decl)
	}

	return
}

func (y *yoitsu) groupByMethod(gType *generatedType, ujps []GeneratedType) ast.Decl {
	safeName := toSafeGoName(gType.JsonType().TypeName())
	structName := safeName + tokenAccessor

	receiver := &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{ast.NewIdent(tokenReceiver)},
				Type:  ast.NewIdent(tokenPointer + structName),
			},
		},
	}

	l := make([]ast.Stmt, len(ujps))
	for i, ujp := range ujps {
		l[i] = &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.IndexExpr{
					X: &ast.SelectorExpr{
						X:   ast.NewIdent(tokenReceiver),
						Sel: ast.NewIdent(tokenData + toSafeGoName(ujp.Name())),
					},
					Index: &ast.SelectorExpr{
						X:   ast.NewIdent("d"),
						Sel: ast.NewIdent(toSafeGoName(ujp.Name())),
					},
				},
			},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				ast.NewIdent("d"),
			},
		}
	}

	funcBody := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.RangeStmt{
				Key:   ast.NewIdent("_"),
				Value: ast.NewIdent("d"),
				Tok:   token.DEFINE,
				X: &ast.SelectorExpr{
					X:   ast.NewIdent(tokenReceiver),
					Sel: ast.NewIdent(tokenData),
				},
				Body: &ast.BlockStmt{
					List: l,
				},
			},
		},
	}

	return &ast.FuncDecl{
		Recv: receiver,
		Name: ast.NewIdent(tokenMethodGroupData),
		Type: &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{List: []*ast.Field{}},
		},
		Body: funcBody,
	}
}

func (y *yoitsu) uniqueJsonPrimitivesAccessor(gType *generatedType, ujp GeneratedType) ast.Decl {
	safeName := toSafeGoName(gType.JsonType().TypeName())
	structName := safeName + tokenAccessor
	structField := ast.NewIdent(tokenData + toSafeGoName(ujp.Name()))
	funcName := fmt.Sprintf("Get%sBy%s", safeName, toSafeGoName(ujp.Name()))

	receiver := &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{ast.NewIdent(tokenReceiver)},
				Type:  ast.NewIdent(tokenPointer + structName),
			},
		},
	}

	funcType := &ast.FuncType{
		Params: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent(tokenIdentifier)},
					Type:  ast.NewIdent(ujp.JsonType().TypeName()),
				},
			},
		},
		Results: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: ast.NewIdent(gType.JsonType().TypeName()),
				},
				{
					Type: ast.NewIdent(tokenError),
				},
			},
		},
	}

	funcBody := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X: &ast.SelectorExpr{
						X:   ast.NewIdent(tokenReceiver),
						Sel: structField,
					},
					Op: token.EQL,
					Y:  ast.NewIdent("nil"),
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.AssignStmt{
							Lhs: []ast.Expr{
								ast.NewIdent("err"),
							},
							Tok: token.DEFINE,
							Rhs: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   ast.NewIdent(tokenReceiver),
										Sel: ast.NewIdent(tokenMethodLoadName),
									},
								},
							},
						},
						&ast.IfStmt{
							Cond: &ast.BinaryExpr{
								X:  ast.NewIdent("err"),
								Op: token.NEQ,
								Y:  ast.NewIdent("nil"),
							},
							Body: &ast.BlockStmt{
								List: []ast.Stmt{
									&ast.ReturnStmt{
										Results: []ast.Expr{
											&ast.CompositeLit{
												Type: ast.NewIdent(gType.JsonType().TypeName()),
											},
											ast.NewIdent("err"),
										},
									},
								},
							},
						},
						&ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   ast.NewIdent(tokenReceiver),
									Sel: ast.NewIdent(tokenMethodGroupData),
								},
								Args: []ast.Expr{},
							},
						},
					},
				},
			},
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.IndexExpr{
						X: &ast.SelectorExpr{
							X:   ast.NewIdent(tokenReceiver),
							Sel: structField,
						},
						Index: ast.NewIdent(tokenIdentifier),
					},
					ast.NewIdent("nil"),
				},
			},
		},
	}

	return &ast.FuncDecl{
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Text: fmt.Sprintf("\n// %s returns the %s uniquely identified by %s\n//\n// Error is only non-nil if the source errors out", funcName, gType.JsonType().TypeName(), ujp.Name()),
				},
			},
		},
		Recv: receiver,
		Name: ast.NewIdent(funcName),
		Type: funcType,
		Body: funcBody,
	}
}

func (y *yoitsu) uniqueJsonPrimitives(gType *generatedType) (found []GeneratedType) {
	data := y.root.([]interface{})

	for name, field := range gType.types {
		prim, ok := field.JsonType().(JsonPrimitive)
		if !ok {
			continue
		}

		switch prim.TypeName() {
		case JsonFloat64.TypeName():
			unique, success := extractUniqueValues[float64](data, name)
			if success && len(unique) == len(data) {
				found = append(found, field)
			}
		case JsonString.TypeName():
			unique, success := extractUniqueValues[string](data, name)
			if success && len(unique) == len(data) {
				found = append(found, field)
			}
		}
	}

	return
}

func extractUniqueValues[T comparable](data []interface{}, field string) ([]T, bool) {
	values := make([]T, 0)

	for _, v := range data {
		d, ok := v.(map[string]any)
		if !ok {
			return nil, false
		}

		value, ok := d[field]
		if !ok {
			continue
		}

		typedValue, ok := value.(T)
		if !ok {
			return nil, false
		}

		if !slices.Contains(values, typedValue) {
			values = append(values, typedValue)
		}
	}

	return values, true
}
