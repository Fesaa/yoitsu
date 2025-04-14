package yoitsu

import (
	"go/ast"
	"go/token"
)

func ifErrNotNilStmt() ast.Stmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  ast.NewIdent("err"),
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{ast.NewIdent("err")},
				},
			},
		},
	}
}

func deferStmt(x string, sel string) ast.Stmt {
	return &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(x),
				Sel: ast.NewIdent(sel),
			},
			Args: []ast.Expr{},
		},
	}
}

func unmarshallStmt(x string, sel string) ast.Stmt {
	return &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent("json"),
					Sel: ast.NewIdent("Unmarshal"),
				},
				Args: []ast.Expr{
					ast.NewIdent("data"),
					&ast.UnaryExpr{
						Op: token.AND,
						X: &ast.SelectorExpr{
							X:   ast.NewIdent(x),
							Sel: ast.NewIdent(sel),
						},
					},
				},
			},
		},
	}
}
