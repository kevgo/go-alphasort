package fix

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"maps"
	"os"
	"path/filepath"
)

// ScanDir provides the Go package in the given directory.
// Files that cannot be read or parsed are ignored,
// since they cannot contain code that sorting other files could break.
func ScanDir(dir string) Package {
	result := Package{files: map[string]*ast.File{}}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		if !IsGoFile(entry) {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, entry.Name()), nil, parser.SkipObjectResolution)
		if err == nil {
			result.files[entry.Name()] = file
		}
	}
	return result
}

// Package contains the files of a Go package.
// Sorting the fields of a struct type needs to know how all files in the package use that type.
// The zero value is an empty package.
type Package struct {
	files map[string]*ast.File // the parsed files of the package, by file name
}

// analyze determines how the files in this package use their types.
func (p Package) analyze() packageInfo {
	result := packageInfo{
		positionalNames:   map[string]bool{},
		positionalStructs: map[string]bool{},
		types:             map[string][]ast.Expr{},
	}
	for _, file := range p.files {
		ast.Inspect(file, func(node ast.Node) bool {
			if spec, ok := node.(*ast.TypeSpec); ok {
				result.types[spec.Name.Name] = append(result.types[spec.Name.Name], spec.Type)
			}
			return true
		})
	}
	for _, file := range p.files {
		walkLiterals(file, result, func(lit *ast.CompositeLit, typ ast.Expr) {
			if !isPositional(lit) {
				return
			}
			names, structType := result.resolve(typ)
			for _, name := range names {
				result.positionalNames[name] = true
			}
			if structType != nil && len(names) == 0 {
				result.positionalStructs[types.ExprString(structType)] = true
			}
		})
	}
	return result
}

// with provides a copy of this package that contains the given file under the given name,
// replacing an existing file with that name.
func (p Package) with(filename string, file *ast.File) Package {
	files := make(map[string]*ast.File, len(p.files)+1)
	maps.Copy(files, p.files)
	files[filepath.Base(filename)] = file
	return Package{files: files}
}

// packageInfo describes how the files of a package use their types.
type packageInfo struct {
	positionalNames   map[string]bool       // names of the struct types that are instantiated with positional fields
	positionalStructs map[string]bool       // anonymous struct types that are instantiated with positional fields
	types             map[string][]ast.Expr // the definitions of the types declared in the package, by type name
}

// isNonStruct indicates whether the given type expression is known to not describe a struct.
func (info packageInfo) isNonStruct(typ ast.Expr) bool {
	underlying := info.underlying(typ)
	return underlying != nil && !isStructType(underlying)
}

// keepsFieldOrder indicates whether the fields of the given struct type must keep their order
// because code instantiates this struct with positional fields.
func (info packageInfo) keepsFieldOrder(structType *ast.StructType, typeName string) bool {
	if typeName != "" {
		return info.positionalNames[typeName]
	}
	return info.positionalStructs[types.ExprString(structType)]
}

// resolve provides the names of the types that the given type expression refers to,
// and the struct type it ultimately describes, or nil if it doesn't describe a struct known to this package.
func (info packageInfo) resolve(typ ast.Expr) ([]string, *ast.StructType) {
	names := []string{}
	for range 10 { // limit the depth to survive cyclic type definitions
		switch t := typ.(type) {
		case *ast.StructType:
			return names, t
		case *ast.ParenExpr:
			typ = t.X
		case *ast.IndexExpr:
			typ = t.X
		case *ast.IndexListExpr:
			typ = t.X
		case *ast.SelectorExpr:
			// a type from another package: remember its name to be on the safe side
			return append(names, t.Sel.Name), nil
		case *ast.Ident:
			names = append(names, t.Name)
			definitions := info.types[t.Name]
			if len(definitions) != 1 {
				return names, nil
			}
			typ = definitions[0]
		default:
			return names, nil
		}
	}
	return names, nil
}

// underlying provides the type literal that the given type expression describes,
// or nil if it is unknown.
func (info packageInfo) underlying(typ ast.Expr) ast.Expr {
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

// isPositional indicates whether the given composite literal lists its elements without keys.
func isPositional(lit *ast.CompositeLit) bool {
	if len(lit.Elts) == 0 {
		return false
	}
	_, keyed := lit.Elts[0].(*ast.KeyValueExpr)
	return !keyed
}

// isStructType indicates whether the given expression is a struct type literal.
func isStructType(expr ast.Expr) bool {
	_, ok := expr.(*ast.StructType)
	return ok
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
func walkLiterals(file *ast.File, info packageInfo, visit func(lit *ast.CompositeLit, typ ast.Expr)) {
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
