package fix

import (
	"go/ast"
	"go/token"
)

// declaredTypes provides the types declared in the given file,
// together with the scopes in which they are visible and the names that the file declares as something other than a type.
func declaredTypes(file *ast.File) typeInfo {
	result := typeInfo{
		nonTypes:   map[string]bool{},
		types:      map[string][]ast.Expr{},
		typeScopes: map[string][]scope{},
	}
	declare := func(names ...*ast.Ident) {
		for _, name := range names {
			result.nonTypes[name.Name] = true
		}
	}
	declareFields := func(fields *ast.FieldList) {
		if fields != nil {
			for _, field := range fields.List {
				declare(field.Names...)
			}
		}
	}
	declareExprs := func(exprs ...ast.Expr) {
		for _, expr := range exprs {
			if ident, ok := expr.(*ast.Ident); ok {
				declare(ident)
			}
		}
	}
	var ancestors []ast.Node
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			ancestors = ancestors[:len(ancestors)-1]
			return true
		}
		switch n := node.(type) {
		case *ast.TypeSpec:
			result.types[n.Name.Name] = append(result.types[n.Name.Name], n.Type)
			result.typeScopes[n.Name.Name] = append(result.typeScopes[n.Name.Name], typeScope(file, n, ancestors))
		case *ast.FuncDecl:
			if n.Recv == nil {
				declare(n.Name)
			}
			if n.Body != nil {
				// parameter and result names enter a scope only in functions that have a body
				declareFields(n.Recv)
				declareFields(n.Type.Params)
				declareFields(n.Type.Results)
			}
		case *ast.FuncLit:
			declareFields(n.Type.Params)
			declareFields(n.Type.Results)
		case *ast.ValueSpec:
			declare(n.Names...)
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE {
				declareExprs(n.Lhs...)
			}
		case *ast.RangeStmt:
			if n.Tok == token.DEFINE {
				declareExprs(n.Key, n.Value)
			}
		}
		ancestors = append(ancestors, node)
		return true
	})
	return result
}

// registerElided remembers the type of the given expression if it is a composite literal with an elided type.
func registerElided(elided map[*ast.CompositeLit]ast.Expr, expr, typ ast.Expr) {
	if star, ok := typ.(*ast.StarExpr); ok {
		// in a []*T, the element {...} stands for &T{...}
		typ = star.X
	}
	if lit, ok := expr.(*ast.CompositeLit); ok && lit.Type == nil {
		elided[lit] = typ
	}
}

// scope describes the range of source code in which a declared name is visible.
type scope struct {
	end   token.Pos
	start token.Pos
}

// typeInfo describes the types declared in a file.
type typeInfo struct {
	nonTypes   map[string]bool       // names that the file declares as functions, variables, constants, parameters, or results in any scope
	types      map[string][]ast.Expr // the definitions of the types declared in the file, by type name
	typeScopes map[string][]scope    // where in the file the types declared in it are visible, by type name
}

// isArray indicates whether the given type expression is known to describe an array or slice.
func (info typeInfo) isArray(typ ast.Expr) bool {
	_, ok := info.underlying(typ).(*ast.ArrayType)
	return ok
}

// isType indicates whether the given identifier is known to refer to a type declared in this file.
// It is conservative: names that the file also declares as something other than a type never count as types.
func (info typeInfo) isType(ident *ast.Ident) bool {
	if info.nonTypes[ident.Name] {
		return false
	}
	for _, s := range info.typeScopes[ident.Name] {
		if ident.Pos() >= s.start && ident.Pos() < s.end {
			return true
		}
	}
	return false
}

// underlying provides the type literal that the given type expression describes,
// or nil if it is unknown.
func (info typeInfo) underlying(typ ast.Expr) ast.Expr {
	for range 10 { // limit the depth to survive cyclic type definitions
		switch t := typ.(type) {
		case *ast.ParenExpr:
			typ = t.X
		case *ast.IndexExpr:
			typ = t.X
		case *ast.IndexListExpr:
			typ = t.X
		case *ast.Ident:
			definitions := info.types[t.Name]
			if len(definitions) != 1 {
				return nil
			}
			typ = definitions[0]
		case *ast.SelectorExpr, nil:
			return nil
		default:
			return typ
		}
	}
	return nil
}

// typeScope provides the part of the given file in which the given type declaration is visible.
// Types declared inside a function are visible from their name to the end of the innermost enclosing block.
func typeScope(file *ast.File, spec *ast.TypeSpec, ancestors []ast.Node) scope {
	for i := len(ancestors) - 1; i >= 0; i-- {
		switch ancestors[i].(type) {
		case *ast.BlockStmt, *ast.CaseClause, *ast.CommClause:
			return scope{end: ancestors[i].End(), start: spec.Name.Pos()}
		}
	}
	return scope{end: file.FileEnd, start: file.FileStart}
}

// walkLiterals calls the given function for each composite literal in the given file,
// together with the type of that literal, which might be nil if it is unknown.
// For literals whose type is elided, it determines the type from the enclosing literal.
func walkLiterals(file *ast.File, info typeInfo, visit func(lit *ast.CompositeLit, typ ast.Expr)) {
	elided := map[*ast.CompositeLit]ast.Expr{}
	ast.Inspect(file, func(node ast.Node) bool {
		lit, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}
		typ := lit.Type
		if typ == nil {
			typ = elided[lit]
		}
		visit(lit, typ)
		var keyType, elemType ast.Expr
		switch t := info.underlying(typ).(type) {
		case *ast.ArrayType:
			elemType = t.Elt
		case *ast.MapType:
			keyType, elemType = t.Key, t.Value
		default:
			return true
		}
		for _, elt := range lit.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				registerElided(elided, kv.Key, keyType)
				elt = kv.Value
			}
			registerElided(elided, elt, elemType)
		}
		return true
	})
}
