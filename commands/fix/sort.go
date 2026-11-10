package fix

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strings"
)

// Sort provides the given Go source code with its elements sorted alphabetically:
// the fields of struct types, the keyed elements of struct and map literals, and the top-level functions and methods.
// It only reorders code and leaves formatting, such as the alignment of moved struct fields, to gofmt.
func Sort(filename string, src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parse(fset, filename, src)
	if err != nil {
		return nil, err
	}
	output := []byte(sortFields(fset, file, src))
	if !bytes.Equal(output, src) {
		fset = token.NewFileSet()
		file, err = parse(fset, filename, output)
		if err != nil {
			return nil, fmt.Errorf("internal error: sorting struct fields produced invalid Go code: %w", err)
		}
	}
	output = []byte(sortFuncs(fset, file, output))
	// safety net: never hand out code that no longer parses
	if _, err := parse(token.NewFileSet(), filename, output); err != nil {
		return nil, fmt.Errorf("internal error: sorting produced invalid Go code: %w", err)
	}
	return output, nil
}

// attachToTypes moves the constructors and methods of the types declared in the given pieces
// directly behind the respective type declaration.
func attachToTypes(pieces []piece, nl string) []piece {
	localTypes := map[string]bool{}
	for _, p := range pieces {
		for _, name := range p.typeNames {
			localTypes[name] = true
		}
	}
	constructors := map[string]piece{}
	methods := map[string][]piece{}
	moved := make([]bool, len(pieces))
	for i, p := range pieces {
		switch p.kind {
		case kindMethod:
			if localTypes[p.receiver] {
				methods[p.receiver] = append(methods[p.receiver], p)
				moved[i] = true
			}
		case kindFunc:
			typeName, isNew := strings.CutPrefix(p.name, "New")
			if _, exists := constructors[typeName]; isNew && localTypes[typeName] && !exists {
				p.attached = true
				constructors[typeName] = p
				moved[i] = true
			}
		case kindOther, kindType:
		}
	}
	result := make([]piece, 0, len(pieces))
	// When pieces at the beginning move away, the first remaining piece
	// takes over the whitespace that separated the moved pieces from the package clause.
	leadingGap, hasLeadingGap := "", false
	for i, p := range pieces {
		if moved[i] {
			if len(result) == 0 && !hasLeadingGap {
				leadingGap, hasLeadingGap = p.gapBefore, true
			}
			continue
		}
		if hasLeadingGap {
			p.gapBefore = leadingGap
			hasLeadingGap = false
		}
		result = append(result, p)
		for _, typeName := range p.typeNames {
			if constructor, has := constructors[typeName]; has {
				result = append(result, withLineGap(constructor, nl))
			}
			typeMethods := methods[typeName]
			slices.SortStableFunc(typeMethods, comparePieces)
			for _, method := range typeMethods {
				result = append(result, withLineGap(method, nl))
			}
		}
	}
	return result
}

// comparePieces orders pieces by name, using plain byte order.
func comparePieces(a, b piece) int {
	return strings.Compare(a.name, b.name)
}

// newline provides the line ending used in the given source code.
func newline(src []byte) string {
	if strings.Contains(string(src), "\r\n") {
		return "\r\n"
	}
	return "\n"
}

// parse parses the given Go source code.
func parse(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
	return parser.ParseFile(fset, filename, src, parser.ParseComments|parser.SkipObjectResolution)
}

// receiverTypeName provides the name of the type that the given method receiver refers to.
func receiverTypeName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	expr := recv.List[0].Type
	for {
		switch e := expr.(type) {
		case *ast.StarExpr:
			expr = e.X
		case *ast.ParenExpr:
			expr = e.X
		case *ast.IndexExpr:
			expr = e.X
		case *ast.IndexListExpr:
			expr = e.X
		case *ast.Ident:
			return e.Name
		default:
			return ""
		}
	}
}

// sortFreeFuncs sorts the free-standing functions that aren't attached to a type
// alphabetically within the positions they occupy.
// The whitespace before these positions stays in place.
func sortFreeFuncs(pieces []piece) {
	slots := []int{}
	for i, p := range pieces {
		if p.kind == kindFunc && !p.attached {
			slots = append(slots, i)
		}
	}
	funcs := make([]piece, len(slots))
	for i, slot := range slots {
		funcs[i] = pieces[slot]
	}
	slices.SortStableFunc(funcs, comparePieces)
	for i, slot := range slots {
		funcs[i].gapBefore = pieces[slot].gapBefore
		pieces[slot] = funcs[i]
	}
}

// sortFuncs provides the given source code with its functions and methods reordered:
// constructors and methods move behind their type, and the remaining free-standing functions get sorted alphabetically.
func sortFuncs(fset *token.FileSet, file *ast.File, src []byte) string {
	head, pieces, tail := splitPieces(fset, file, src)
	pieces = attachToTypes(pieces, newline(src))
	sortFreeFuncs(pieces)
	var result strings.Builder
	result.Grow(len(src))
	result.WriteString(head)
	for _, piece := range pieces {
		result.WriteString(piece.gapBefore)
		result.WriteString(piece.text)
	}
	result.WriteString(tail)
	return result.String()
}

// splitPieces splits the given source code into the part up to the package name,
// the top-level declarations and floating comments after it,
// and the remaining text after the last piece.
func splitPieces(fset *token.FileSet, file *ast.File, src []byte) (string, []piece, string) {
	tokFile := fset.File(file.Pos())
	headEnd := tokFile.Offset(file.Name.End())
	type span struct {
		begin, end int
		piece      piece
	}
	spans := []span{}
	for i, decl := range file.Decls {
		begin := decl.Pos()
		p := piece{kind: kindOther}
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Doc != nil {
				begin = d.Doc.Pos()
			}
			p.name = d.Name.Name
			p.kind = kindFunc
			if d.Recv != nil {
				p.kind = kindMethod
				p.receiver = receiverTypeName(d.Recv)
			}
		case *ast.GenDecl:
			if d.Doc != nil {
				begin = d.Doc.Pos()
			}
			if d.Tok == token.TYPE {
				p.kind = kindType
				for _, spec := range d.Specs {
					p.typeNames = append(p.typeNames, spec.(*ast.TypeSpec).Name.Name)
				}
			}
		}
		limit := token.NoPos
		if i+1 < len(file.Decls) {
			limit = file.Decls[i+1].Pos()
		}
		end := trailingCommentEnd(tokFile, file.Comments, decl.End(), limit)
		spans = append(spans, span{begin: tokFile.Offset(begin), end: tokFile.Offset(end), piece: p})
	}
	// floating comments: comments after the package name that aren't part of a declaration
	declCount := len(spans)
	declIndex := 0
	for _, group := range file.Comments {
		for _, comment := range group.List {
			begin, end := tokFile.Offset(comment.Pos()), tokFile.Offset(comment.End())
			if begin < headEnd {
				continue
			}
			for declIndex < declCount && spans[declIndex].end <= begin {
				declIndex++
			}
			if declIndex < declCount && spans[declIndex].begin <= begin {
				continue // inside a declaration
			}
			spans = append(spans, span{begin: begin, end: end, piece: piece{kind: kindOther}})
		}
	}
	slices.SortFunc(spans, func(a, b span) int { return a.begin - b.begin })
	pieces := make([]piece, len(spans))
	cursor := headEnd
	for i, s := range spans {
		s.piece.gapBefore = string(src[cursor:s.begin])
		s.piece.text = string(src[s.begin:s.end])
		pieces[i] = s.piece
		cursor = s.end
	}
	return string(src[:headEnd]), pieces, string(src[cursor:])
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

// withLineGap provides the given piece, ensuring that it starts on a new line.
func withLineGap(p piece, nl string) piece {
	if strings.TrimSpace(p.gapBefore) != "" || !strings.Contains(p.gapBefore, "\n") {
		p.gapBefore = nl + nl
	}
	return p
}

// piece is a top-level declaration or a floating comment,
// together with the text that precedes it.
type piece struct {
	attached  bool   // whether this is a constructor that moved behind its type
	gapBefore string // the whitespace between the previous piece and this one
	kind      pieceKind
	name      string   // the name of the function or method
	receiver  string   // for methods: the name of the receiver type
	text      string   // the source code of this piece, including doc and trailing comments
	typeNames []string // for type declarations: the declared type names
}

// pieceKind describes what a piece contains.
type pieceKind int

const (
	kindOther  pieceKind = iota // something that stays in place
	kindFunc                    // a free-standing function
	kindMethod                  // a method
	kindType                    // a type declaration
)
