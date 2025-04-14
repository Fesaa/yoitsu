package yoitsu

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// Source tells the Parser and Yoitsu where to find the Json data.
type Source interface {
	Json() ([]byte, error)
	// Name is used for the root type and file name (Source.Name + ".generated.go")
	Name() string
}

// LoadAbleSource returns the body, and required imports to load the data when using the generated file.
// Implement this when you wish to generate Accessors. The provided sources both implement this interface
type LoadAbleSource interface {
	LoadMethod() (stmt *ast.BlockStmt, importSpec []ast.Spec)
}

func NewFileSource(name string, f string) Source {
	return &fileSource{
		f:    f,
		name: name,
	}
}

func NewUrlSource(name string, u string, opts ...Option[urlSource]) Source {
	us := urlSource{
		url:  u,
		name: name,
	}

	for _, opt := range opts {
		opt(us)
	}

	if us.httpClient == nil {
		us.httpClient = http.DefaultClient
	}

	return &us
}

type fileSource struct {
	f    string
	b    []byte
	name string
}

func (src *fileSource) Json() ([]byte, error) {
	if src.b != nil {
		return src.b, nil
	}

	var file *os.File
	var err error

	file, err = os.Open(src.f)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	src.b, err = io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return src.b, nil
}

func (src *fileSource) Name() string {
	return src.name
}

func (src *fileSource) LoadMethod() (*ast.BlockStmt, []ast.Spec) {
	return &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						ast.NewIdent("f"),
						ast.NewIdent("err"),
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("os"),
								Sel: ast.NewIdent("Open"),
							},
							Args: []ast.Expr{
								&ast.BasicLit{
									Kind:  token.STRING,
									Value: fmt.Sprintf(`"%s"`, src.f),
								},
							},
						},
					},
				},
				ifErrNotNilStmt(),
				deferStmt("f", "Close"),
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						ast.NewIdent("data"),
						ast.NewIdent("err"),
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("io"),
								Sel: ast.NewIdent("ReadAll"),
							},
							Args: []ast.Expr{
								ast.NewIdent("f"),
							},
						},
					},
				},
				ifErrNotNilStmt(),
				unmarshallStmt(tokenReceiver, tokenData),
			},
		}, []ast.Spec{
			&ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote("os"),
				},
			},
			&ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote("io"),
				},
			},
			&ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote("encoding/json"),
				},
			},
		}
}

type urlSource struct {
	httpClient *http.Client
	url        string
	b          []byte
	name       string
}

func (src *urlSource) Json() ([]byte, error) {
	if src.b != nil {
		return src.b, nil
	}

	parsedUrl, err := url.Parse(src.url)
	if err != nil {
		return nil, err
	}

	src.url = parsedUrl.String()

	resp, err := src.httpClient.Get(src.url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	src.b = b
	return b, nil
}

func (src *urlSource) Name() string {
	return src.name
}

func (src *urlSource) LoadMethod() (*ast.BlockStmt, []ast.Spec) {
	return &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						ast.NewIdent("res"),
						ast.NewIdent("err"),
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("http"),
								Sel: ast.NewIdent("Get"),
							},
							Args: []ast.Expr{
								&ast.BasicLit{
									Kind:  token.STRING,
									Value: fmt.Sprintf(`"%s"`, src.url),
								},
							},
						},
					},
				},
				ifErrNotNilStmt(),
				deferStmt("res.Body", "Close"),
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						ast.NewIdent("data"),
						ast.NewIdent("err"),
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("io"),
								Sel: ast.NewIdent("ReadAll"),
							},
							Args: []ast.Expr{
								ast.NewIdent("res.Body"),
							},
						},
					},
				},
				ifErrNotNilStmt(),
				unmarshallStmt(tokenReceiver, tokenData),
			},
		}, []ast.Spec{
			&ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote("net/http"),
				},
			},
			&ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote("io"),
				},
			},
			&ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote("encoding/json"),
				},
			},
		}
}

func UrlSourceWithHttpClient(c *http.Client) Option[urlSource] {
	return func(source urlSource) {
		source.httpClient = c
	}
}
