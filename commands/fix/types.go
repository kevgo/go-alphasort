package fix

import (
	"go/ast"
)

// declaredTypes provides the types declared in the given file.
func declaredTypes(file *ast.File) typeInfo {
	result := typeInfo{
		types: map[string][]ast.Expr{},
	}
	ast.Inspect(file, func(node ast.Node) bool {
		if spec, ok := node.(*ast.TypeSpec); ok {
			result.types[spec.Name.Name] = append(result.types[spec.Name.Name], spec.Type)
		}
		return true
	})
	return result
}

// typeInfo describes the types declared in a file.
type typeInfo struct {
	types map[string][]ast.Expr // the definitions of the types declared in the file, by type name
}

// isArray indicates whether the given type expression is known to describe an array or slice.
func (info typeInfo) isArray(typ ast.Expr) bool {
	_, ok := info.underlying(typ).(*ast.ArrayType)
	return ok
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
