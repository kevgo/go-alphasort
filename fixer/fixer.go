// Package fixer reorders the free-standing functions in Go source code alphabetically.
package fixer

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strings"
)

// SortFuncs returns the given Go source code with all free-standing functions
// (functions without a receiver) sorted alphabetically by name.
//
// Only the positions occupied by free-standing functions get rearranged.
// Methods, types, variables, constants, imports, floating comments,
// and the whitespace between declarations stay exactly where they are.
// Each function moves together with its doc comment
// and any comment that trails it on the same line.
//
// Names compare by plain byte order, so capitalized (exported) functions
// come before lowercase ones.
// The sort is stable, so functions with identical names (like multiple init functions)
// keep their relative order.
func SortFuncs(filename string, src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	funcs := findFuncs(fset, file)
	sorted := slices.Clone(funcs)
	slices.SortStableFunc(sorted, func(a, b funcChunk) int {
		return strings.Compare(a.name, b.name)
	})
	var result bytes.Buffer
	result.Grow(len(src))
	cursor := 0
	for i, slot := range funcs {
		result.Write(src[cursor:slot.start])
		replacement := sorted[i]
		result.Write(src[replacement.start:replacement.end])
		cursor = slot.end
	}
	result.Write(src[cursor:])
	output := result.Bytes()
	// safety net: never hand out code that no longer parses
	if _, err := parser.ParseFile(token.NewFileSet(), filename, output, parser.ParseComments|parser.SkipObjectResolution); err != nil {
		return nil, fmt.Errorf("internal error: sorting functions produced invalid Go code: %w", err)
	}
	return output, nil
}

// findFuncs provides the source ranges of all free-standing functions in the given file,
// in the order in which they appear.
func findFuncs(fset *token.FileSet, file *ast.File) []funcChunk {
	tokFile := fset.File(file.Pos())
	result := []funcChunk{}
	for i, decl := range file.Decls {
		funcDecl, isFunc := decl.(*ast.FuncDecl)
		if !isFunc || funcDecl.Recv != nil {
			continue
		}
		start := funcDecl.Pos()
		if funcDecl.Doc != nil {
			start = funcDecl.Doc.Pos()
		}
		limit := token.NoPos
		if i+1 < len(file.Decls) {
			limit = file.Decls[i+1].Pos()
		}
		end := trailingCommentEnd(tokFile, file.Comments, funcDecl.End(), limit)
		result = append(result, funcChunk{
			name:  funcDecl.Name.Name,
			start: tokFile.Offset(start),
			end:   tokFile.Offset(end),
		})
	}
	return result
}

// trailingCommentEnd provides the end position of the comments
// that trail a declaration ending at declEnd on the same line.
// It never extends past limit, the start of the next declaration (unless NoPos).
func trailingCommentEnd(tokFile *token.File, groups []*ast.CommentGroup, declEnd, limit token.Pos) token.Pos {
	result := declEnd
	line := tokFile.Line(declEnd)
	for _, group := range groups {
		for _, comment := range group.List {
			if comment.Pos() < declEnd {
				continue
			}
			if tokFile.Line(comment.Pos()) != line {
				return result
			}
			if limit.IsValid() && comment.End() > limit {
				return result
			}
			result = comment.End()
		}
	}
	return result
}

// funcChunk describes the source code range of a free-standing function,
// including its doc comment and trailing same-line comments.
type funcChunk struct {
	name  string
	start int // byte offset of the first character
	end   int // byte offset after the last character
}
