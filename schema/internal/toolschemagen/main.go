// Command toolschemagen maintains JSON Schema extension carriers after generation.
package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

func main() {
	data, err := os.ReadFile("types.go")
	if err == nil {
		data, err = rewrite(data)
	}
	if err == nil {
		err = os.WriteFile("types.go", data, 0644)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rewrite(data []byte) ([]byte, error) {
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "types.go", data, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	found := 0
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if !ok || (spec.Name.Name != "ToolInputSchema" && spec.Name.Name != "ToolOutputSchema") {
			return true
		}
		body, ok := spec.Type.(*ast.StructType)
		if !ok {
			return true
		}
		for _, field := range body.Fields.List {
			if len(field.Names) == 1 && field.Names[0].Name == "AdditionalProperties" {
				field.Tag = &ast.BasicLit{Kind: token.STRING, Value: "`json:\"-\" yaml:\"-\" mapstructure:\",remain\"`"}
				found++
			}
		}
		return true
	})
	if found != 2 {
		return nil, fmt.Errorf("expected two tool schema extension carriers, found %d", found)
	}
	// Generated tool schema and call decoding share the exact JSON number adapter;
	// official versioned schema snapshots remain pristine.
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "UnmarshalJSON" || fn.Recv == nil || len(fn.Recv.List) != 1 {
			continue
		}
		pointer, ok := fn.Recv.List[0].Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		owner, ok := pointer.X.(*ast.Ident)
		if !ok {
			continue
		}
		switch owner.Name {
		case "ToolInputSchema", "CallToolRequestParams", "CallToolRequest", "CallToolResult", "CallToolResultResponse", "CallToolResultResponseResult":
		default:
			continue
		}
		if owner.Name == "CallToolRequestParams" || owner.Name == "CallToolResult" || owner.Name == "CallToolResultResponseResult" {
			normalized := false
			for _, statement := range fn.Body.List {
				expression, ok := statement.(*ast.ExprStmt)
				if !ok {
					continue
				}
				call, ok := expression.X.(*ast.CallExpr)
				if !ok {
					continue
				}
				name, ok := call.Fun.(*ast.Ident)
				if ok && name.Name == "normalizeToolCallNumbers" {
					normalized = true
				}
			}
			if !normalized {
				for index, statement := range fn.Body.List {
					assignment, ok := statement.(*ast.AssignStmt)
					if !ok || len(assignment.Lhs) != 1 {
						continue
					}
					if _, ok := assignment.Lhs[0].(*ast.StarExpr); !ok {
						continue
					}
					normalization := &ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent("normalizeToolCallNumbers"), Args: []ast.Expr{ast.NewIdent("j")}}}
					fn.Body.List = append(fn.Body.List[:index+1], append([]ast.Stmt{normalization}, fn.Body.List[index+1:]...)...)
					break
				}
			}
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			packageName, ok := selector.X.(*ast.Ident)
			if ok && packageName.Name == "json" && selector.Sel.Name == "Unmarshal" {
				call.Fun = ast.NewIdent("unmarshalToolSchemaJSON")
			}
			return true
		})
	}
	var output strings.Builder
	if err := format.Node(&output, set, file); err != nil {
		return nil, err
	}
	return []byte(output.String()), nil
}
