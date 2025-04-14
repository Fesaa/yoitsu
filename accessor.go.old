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

	tokenMap string = "map[%s]%s"

	tokenReceiver string = "a"
	tokenPointer  string = "*"

	tokenMethodLoadName  string = "LoadData"
	tokenMethodGroupData string = "GroupData"
	tokenError           string = "error"

	tokenAllMethod string = "Raw"
)

func (y *Yoitsu) generateMethodAccessors(gType GeneratedType) (decls []ast.Decl, importSpecs []ast.Spec, err error) {
	if !y.accessors.Generate {
		return
	}

	fieldList := ast.FieldList{}

	accessorsStruct := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(gType.Name() + tokenAccessor),
				Type: &ast.StructType{
					Fields: &fieldList,
				},
			},
		},
	}

	fieldList.List = append(fieldList.List, &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(tokenData)},
		Type:  ast.NewIdent(gType.JsonType().TypeName()),
	})

	decls = append(decls, accessorsStruct)

	var (
		decl       ast.Decl
		importSpec []ast.Spec
	)

	decl, importSpec = y.loadMethod(gType.Name())
	decls = append(decls, decl)
	importSpecs = append(importSpecs, importSpec...)

	decls = append(decls, y.allMethod(gType))

	if y.accessors.ById {
		gat, ok := gType.(*generatedArrayType)
		if ok {
			var uniqueDecls []ast.Decl
			uniqueDecls, importSpec = y.uniqueGetters(&gat.generatedType, &fieldList)

			if len(uniqueDecls) > 0 {
				decls = append(decls, uniqueDecls...)
			}
			if len(importSpec) > 0 {
				importSpecs = append(importSpecs, importSpec...)
			}
		}

		gmt, ok := gType.(*generatedMapType)
		if ok {
			decls = append(decls, y.getByIdMethod(gmt.generatedType))
		}

	}

	return
}

func (y *Yoitsu) getByIdMethod(gType generatedType) ast.Decl {
	funcName := "ByID"

	receiver := &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{ast.NewIdent(tokenReceiver)},
				Type:  ast.NewIdent(tokenPointer + gType.Name() + tokenAccessor),
			},
		},
	}

	funcType := &ast.FuncType{
		Params: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent(tokenIdentifier)},
					Type:  ast.NewIdent("string"),
				},
			},
		},
		Results: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: ast.NewIdent(gType.JsonType().TypeName()),
				},
			},
		},
	}

	return &ast.FuncDecl{
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Text: fmt.Sprintf("\n// ByID returns the %s identified by the passed id", gType.JsonType().TypeName()),
				},
			},
		},
		Recv: receiver,
		Name: ast.NewIdent(funcName),
		Type: funcType,
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.IndexExpr{
							X: &ast.SelectorExpr{
								X:   ast.NewIdent(tokenReceiver),
								Sel: ast.NewIdent(tokenData),
							},
							Index: ast.NewIdent(tokenIdentifier),
						},
					},
				},
			},
		},
	}
}

func (y *Yoitsu) loadMethod(structName string) (ast.Decl, []ast.Spec) {
	receiver := &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{ast.NewIdent(tokenReceiver)},
				Type:  ast.NewIdent(tokenPointer + structName + tokenAccessor),
			},
		},
	}

	funcName := ast.NewIdent(tokenMethodLoadName)
	funcType := &ast.FuncType{
		Params: &ast.FieldList{},
		Results: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: ast.NewIdent(tokenError),
				},
			},
		},
	}

	funcBody, importSpec := y.src.LoadMethod()

	doc := fmt.Sprintf("\n// %s retrieves the data.", tokenMethodLoadName)
	if y.accessors.ById {
		doc += fmt.Sprintf(" Must be called before %s.%s", structName, tokenMethodGroupData)
	}

	return &ast.FuncDecl{
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Text: doc,
				},
			},
		},
		Recv: receiver,
		Name: funcName,
		Type: funcType,
		Body: funcBody,
	}, importSpec
}

func (y *Yoitsu) allMethod(gType GeneratedType) ast.Decl {
	receiver := &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{ast.NewIdent(tokenReceiver)},
				Type:  ast.NewIdent(tokenPointer + gType.Name() + tokenAccessor),
			},
		},
	}

	funcName := ast.NewIdent(tokenAllMethod)
	funcType := &ast.FuncType{
		Params: &ast.FieldList{},
		Results: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: ast.NewIdent(gType.JsonType().TypeName()),
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
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Text: fmt.Sprintf("\n// %s returns the raw data.", tokenAllMethod),
				},
			},
		},
		Recv: receiver,
		Name: funcName,
		Type: funcType,
		Body: funcBody,
	}
}

func (y *Yoitsu) uniqueGetters(gType *generatedType, fieldList *ast.FieldList) (decls []ast.Decl, importSpec []ast.Spec) {
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

func (y *Yoitsu) groupByMethod(gType *generatedType, ujps []GeneratedType) ast.Decl {
	receiver := &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{ast.NewIdent(tokenReceiver)},
				Type:  ast.NewIdent(tokenPointer + gType.Name() + tokenAccessor),
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
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Text: fmt.Sprintf("\n// %s groups the data by their unique ids.\n// Can be called manually in conjunction with %s.%s to preload everything", tokenMethodGroupData, gType.Name()+tokenAccessor, tokenMethodLoadName),
				},
			},
		},
		Recv: receiver,
		Name: ast.NewIdent(tokenMethodGroupData),
		Type: &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{List: []*ast.Field{}},
		},
		Body: funcBody,
	}
}

func (y *Yoitsu) uniqueJsonPrimitivesAccessor(gType *generatedType, ujp GeneratedType) ast.Decl {
	structField := ast.NewIdent(tokenData + ujp.Name())
	funcName := fmt.Sprintf("By%s", ujp.Name())

	receiver := &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{ast.NewIdent(tokenReceiver)},
				Type:  ast.NewIdent(tokenPointer + gType.Name() + tokenAccessor),
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

func (y *Yoitsu) uniqueJsonPrimitives(gType *generatedType) (found []GeneratedType) {
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
