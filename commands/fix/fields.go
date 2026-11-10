package fix

import (
	"go/ast"
	"go/token"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// compareEdits orders edits by their start, enclosing edits first.
func compareEdits(a, b edit) int {
	if a.start != b.start {
		return a.start - b.start
	}
	return b.end - a.end
}

// edit describes a part of the source code that gets replaced.
type edit struct {
	end    int
	render func() string // provides the replacement text
	start  int
}

// editor applies nested edits to source code.
type editor struct {
	comments []*ast.Comment // all comments in the file, in source order
	edits    []edit         // sorted by start, enclosing edits before the edits they contain
	info     typeInfo
	src      string
	tokFile  *token.File
}

// addList registers an edit that brings the given list of items into the given order.
func (e *editor) addList(items []listItem, order []int) {
	if slices.IsSorted(order) {
		return
	}
	e.edits = append(e.edits, edit{
		end: items[len(items)-1].trailEnd,
		render: func() string {
			var result strings.Builder
			for slot := range items {
				if slot > 0 {
					result.WriteString(e.src[items[slot-1].trailEnd:items[slot].leadStart])
				}
				item := items[order[slot]]
				result.WriteString(e.text(item.leadStart, item.coreEnd))
				result.WriteString(e.src[items[slot].coreEnd:items[slot].sepEnd])
				result.WriteString(e.text(item.sepEnd, item.trailEnd))
			}
			return result.String()
		},
		start: items[0].leadStart,
	})
}

// addLiteral registers the edits that sort the keyed elements of the given struct or map literal.
func (e *editor) addLiteral(lit *ast.CompositeLit) {
	if len(lit.Elts) < 2 {
		return
	}
	keys := make([]string, len(lit.Elts))
	sideEffects := 0
	for i, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			return
		}
		key, ok := literalKey(kv.Key)
		if !ok {
			return
		}
		keys[i] = key
		if e.info.mayHaveSideEffects(kv.Value) {
			sideEffects++
		}
	}
	if sideEffects > 1 {
		// sorting would change the order in which Go evaluates these side effects
		return
	}
	nodes := make([]ast.Node, len(lit.Elts))
	for i, elt := range lit.Elts {
		nodes[i] = elt
	}
	e.addList(e.listItems(nodes, lit.Lbrace, lit.Rbrace, ','), sortOrder(keys))
}

// addNames registers an edit that sorts the names of the given struct field.
func (e *editor) addNames(field *ast.Field) {
	names := make([]string, len(field.Names))
	for i, name := range field.Names {
		names[i] = name.Name
	}
	if slices.IsSorted(names) {
		return
	}
	start, end := field.Names[0].Pos(), field.Names[len(field.Names)-1].End()
	for _, comment := range e.comments {
		if comment.Pos() >= start && comment.End() <= end {
			return
		}
	}
	slices.Sort(names)
	e.edits = append(e.edits, edit{
		start:  e.tokFile.Offset(start),
		end:    e.tokFile.Offset(end),
		render: func() string { return strings.Join(names, ", ") },
	})
}

// addStruct registers the edits that sort the fields of the given struct type.
func (e *editor) addStruct(structType *ast.StructType) {
	fields := structType.Fields.List
	for _, field := range fields {
		for _, name := range field.Names {
			if name.Name == "_" {
				// blank fields usually serve as padding or markers whose position matters
				return
			}
		}
	}
	keys := make([]string, len(fields))
	nodes := make([]ast.Node, len(fields))
	for i, field := range fields {
		if len(field.Names) > 1 {
			e.addNames(field)
		}
		keys[i] = fieldName(field)
		nodes[i] = field
	}
	if len(fields) < 2 {
		return
	}
	e.addList(e.listItems(nodes, structType.Fields.Opening, structType.Fields.Closing, ';'), sortOrder(keys))
}

// firstCommentAt provides the index of the first comment that starts at or after the given offset.
func (e *editor) firstCommentAt(offset int) int {
	pos := e.tokFile.Pos(offset)
	return sort.Search(len(e.comments), func(c int) bool { return e.comments[c].Pos() >= pos })
}

// lineOf provides the line number of the given offset.
func (e *editor) lineOf(offset int) int {
	return e.tokFile.Line(e.tokFile.Pos(offset))
}

// listItems determines the source code ranges of the given list elements,
// which are enclosed by the given opening and closing token and separated by the given separator.
func (e *editor) listItems(nodes []ast.Node, opening, closing token.Pos, separator byte) []listItem {
	items := make([]listItem, len(nodes))
	for i, node := range nodes {
		items[i].coreStart = e.tokFile.Offset(node.Pos())
		items[i].coreEnd = e.tokFile.Offset(node.End())
	}
	closingOffset := e.tokFile.Offset(closing)
	boundaryLine := e.tokFile.Line(opening)
	for i := range items {
		item := &items[i]
		// separator
		item.sepEnd = item.coreEnd
		rest := strings.TrimLeft(e.src[item.coreEnd:], " \t")
		if strings.HasPrefix(rest, string(separator)) {
			item.sepEnd = len(e.src) - len(rest) + 1
		}
		// trailing comments on the same line
		limit := closingOffset
		if i+1 < len(items) {
			limit = items[i+1].coreStart
		}
		item.trailEnd = item.sepEnd
		coreEndLine := e.lineOf(item.coreEnd)
		for _, comment := range e.comments[e.firstCommentAt(item.sepEnd):] {
			start, end := e.tokFile.Offset(comment.Pos()), e.tokFile.Offset(comment.End())
			if end > limit || e.lineOf(start) != coreEndLine {
				break
			}
			item.trailEnd = end
		}
		// leading comments on the lines directly above, which belong to this item
		item.leadStart = item.coreStart
		for c := e.firstCommentAt(item.coreStart) - 1; c >= 0; c-- {
			comment := e.comments[c]
			start, end := e.tokFile.Offset(comment.Pos()), e.tokFile.Offset(comment.End())
			if e.lineOf(start) <= boundaryLine || e.lineOf(end) < e.lineOf(item.leadStart)-1 {
				break
			}
			item.leadStart = start
		}
		boundaryLine = e.lineOf(item.trailEnd)
	}
	return items
}

// text provides the source code between the given offsets, with all edits inside of it applied.
func (e *editor) text(start, end int) string {
	var result strings.Builder
	cursor := start
	first := sort.Search(len(e.edits), func(i int) bool { return e.edits[i].start >= start })
	for _, ed := range e.edits[first:] {
		if ed.start >= end {
			break
		}
		if ed.start < cursor {
			continue // nested inside an edit that was already applied
		}
		if ed.end > end {
			continue // encloses the requested range, i.e. it is the edit that requests this text
		}
		result.WriteString(e.src[cursor:ed.start])
		result.WriteString(ed.render())
		cursor = ed.end
	}
	result.WriteString(e.src[cursor:end])
	return result.String()
}

// listItem describes the source code of an element in a list of struct fields or composite literal elements.
// Its parts are, in this order:
// the leading comments, the element itself, the separator after it, and the trailing comments on the same line.
// When sorting, the separator stays in place and the other parts move.
type listItem struct {
	coreEnd   int // end of the element
	coreStart int // start of the element
	leadStart int // start of the comments above the element
	sepEnd    int // end of the separator after the element
	trailEnd  int // end of the comments after the element on the same line
}

// fieldName provides the name of the given struct field, which for embedded fields is the name of their type.
func fieldName(field *ast.Field) string {
	if len(field.Names) > 0 {
		names := make([]string, len(field.Names))
		for i, name := range field.Names {
			names[i] = name.Name
		}
		return slices.Min(names)
	}
	expr := field.Type
	for {
		switch t := expr.(type) {
		case *ast.StarExpr:
			expr = t.X
		case *ast.ParenExpr:
			expr = t.X
		case *ast.IndexExpr:
			expr = t.X
		case *ast.IndexListExpr:
			expr = t.X
		case *ast.SelectorExpr:
			return t.Sel.Name
		case *ast.Ident:
			return t.Name
		default:
			return ""
		}
	}
}

// literalKey provides the text by which to sort the given key of a struct or map literal.
// It supports identifiers, qualified identifiers, and string literals.
func literalKey(key ast.Expr) (string, bool) {
	switch k := key.(type) {
	case *ast.Ident:
		return k.Name, true
	case *ast.SelectorExpr:
		if pkg, ok := k.X.(*ast.Ident); ok {
			return pkg.Name + "." + k.Sel.Name, true
		}
	case *ast.BasicLit:
		if k.Kind == token.STRING {
			text, err := strconv.Unquote(k.Value)
			return text, err == nil
		}
	}
	return "", false
}

// isConversion indicates whether a call of the given expression is a type conversion or a call of a side-effect free builtin function.
// Identifiers count only if the file doesn't declare them as anything that could be called.
func (info typeInfo) isConversion(fun ast.Expr) bool {
	switch f := fun.(type) {
	case *ast.Ident:
		if info.isType(f) {
			return true
		}
		return pureBuiltins[f.Name] && !info.nonTypes[f.Name]
	case *ast.ParenExpr:
		return info.isConversion(f.X)
	case *ast.IndexExpr:
		return info.isConversion(f.X)
	case *ast.IndexListExpr:
		return info.isConversion(f.X)
	case *ast.StarExpr, *ast.ArrayType, *ast.MapType, *ast.ChanType, *ast.FuncType, *ast.InterfaceType, *ast.StructType:
		return true
	}
	return false
}

// sortFields provides the given source code with the fields of its struct types
// and the keyed elements of its struct instantiations sorted alphabetically.
func sortFields(fset *token.FileSet, file *ast.File, src []byte) string {
	info := declaredTypes(file)
	tokFile := fset.File(file.Pos())
	e := editor{
		comments: []*ast.Comment{},
		info:     info,
		src:      string(src),
		tokFile:  tokFile,
	}
	for _, group := range file.Comments {
		e.comments = append(e.comments, group.List...)
	}
	ast.Inspect(file, func(node ast.Node) bool {
		if structType, ok := node.(*ast.StructType); ok {
			e.addStruct(structType)
		}
		return true
	})
	walkLiterals(file, info, func(lit *ast.CompositeLit, typ ast.Expr) {
		if !info.isArray(typ) {
			e.addLiteral(lit)
		}
	})
	slices.SortStableFunc(e.edits, compareEdits)
	return e.text(0, len(src))
}

// mayHaveSideEffects indicates whether evaluating the given expression might call functions or communicate over channels.
func (info typeInfo) mayHaveSideEffects(expr ast.Expr) bool {
	result := false
	ast.Inspect(expr, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncLit:
			return false // defining a function doesn't run it
		case *ast.CallExpr:
			if !info.isConversion(n.Fun) {
				result = true
			}
		case *ast.UnaryExpr:
			if n.Op == token.ARROW {
				result = true
			}
		}
		return !result
	})
	return result
}

// sortOrder provides the indexes of the given keys in alphabetical order of these keys.
func sortOrder(keys []string) []int {
	order := make([]int, len(keys))
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int { return strings.Compare(keys[a], keys[b]) })
	return order
}

// pureBuiltins contains the predeclared functions and types whose calls have no side effects.
var pureBuiltins = map[string]bool{
	"any": true, "bool": true, "byte": true, "cap": true, "complex": true, "complex128": true, "complex64": true,
	"error": true, "float32": true, "float64": true, "imag": true, "int": true, "int16": true, "int32": true,
	"int64": true, "int8": true, "len": true, "make": true, "max": true, "min": true, "new": true, "real": true,
	"rune": true, "string": true, "uint": true, "uint16": true, "uint32": true, "uint64": true, "uint8": true,
	"uintptr": true,
}
