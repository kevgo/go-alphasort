package fix

import (
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"path/filepath"
)

// ScanDir provides the Go package in the given directory.
// Files that cannot be read or parsed are ignored.
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
// Sorting struct instantiations needs to know the types declared in all files of the package,
// for example to tell instantiations of struct types apart from literals of map types.
// The zero value is an empty package.
type Package struct {
	files map[string]*ast.File // the parsed files of the package, by file name
}

// analyze determines the types declared in the files of this package.
func (p Package) analyze() packageInfo {
	result := packageInfo{
		types: map[string][]ast.Expr{},
	}
	for _, file := range p.files {
		ast.Inspect(file, func(node ast.Node) bool {
			if spec, ok := node.(*ast.TypeSpec); ok {
				result.types[spec.Name.Name] = append(result.types[spec.Name.Name], spec.Type)
			}
			return true
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

// packageInfo describes the types declared in the files of a package.
type packageInfo struct {
	types map[string][]ast.Expr // the definitions of the types declared in the package, by type name
}

// isNonStruct indicates whether the given type expression is known to not describe a struct.
func (info packageInfo) isNonStruct(typ ast.Expr) bool {
	underlying := info.underlying(typ)
	return underlying != nil && !isStructType(underlying)
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
