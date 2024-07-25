package undgen

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"strconv"
	"strings"
	"text/template"
	"unicode"
)

type fieldConverter interface {
	Expr(field string) string
}

type genericConverter struct {
	Selector string
	// AdditionalImports map[string]string
	Method string
	Nil    bool
	// added before field.
	Args     []string
	TypePram []string
}

func nullishConverter(imports UndImports) *genericConverter {
	return &genericConverter{
		Selector: imports.conversion,
		Method:   "UndNullish",
	}
}

func (m *genericConverter) Expr(
	field string,
) string {
	if m.Nil {
		return "nil"
	}
	if m.Selector == "" {
		return field + "." + m.Method + "()"
	}

	var instantiation string
	if len(m.TypePram) > 0 {
		instantiation = "[" + strings.Join(m.TypePram, ",") + "]"
	}
	ident := m.Selector
	if ident == "." {
		ident = ""
	} else {
		ident += "."
	}
	var args []string
	if len(m.Args) > 0 {
		args = append(args, m.Args...)
	}
	args = append(args, field)
	return ident + m.Method + instantiation + "(" + strings.Join(args, ",") + ")"
}

type nestedConverter struct {
	g        *genericConverter
	wrappers []fieldConverter
}

func (c *nestedConverter) Expr(
	field string,
) string {
	expr := c.g.Expr(field)
	for _, wrapper := range c.wrappers {
		expr = wrapper.Expr(expr)
		var ok bool
		expr, ok = strings.CutSuffix(expr, ")")
		if ok {
			s := strings.TrimLeftFunc(expr, unicode.IsSpace)
			switch s[len(s)-1] {
			case '(', '\n':
				expr += ")"
			default:
				expr += ",\n)"
			}
		}
	}
	return expr
}

type templateConverter struct {
	t *template.Template
	p templateParams
}

func (c *templateConverter) Expr(
	field string,
) string {
	param := c.p
	param.Arg = field
	var buf bytes.Buffer
	err := c.t.Execute(&buf, param)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

type templateParams struct {
	OptionPkg  string
	UndPkg     string
	ElasticPkg string
	Arg        string
	TypeParam  string
	Size       string
}

func newTemplateParams(
	imports UndImports,
	isSlice bool,
	arg string,
	typeParam ast.Node,
	size int,
) templateParams {
	var buf bytes.Buffer
	fset := token.NewFileSet()
	err := printer.Fprint(&buf, fset, typeParam)
	if err != nil {
		panic(err)
	}
	var sizeStr string
	if size > 0 {
		sizeStr = strconv.FormatInt(int64(size), 10)
	}
	return templateParams{
		OptionPkg:  imports.option,
		UndPkg:     imports.Und(isSlice),
		ElasticPkg: imports.Elastic(isSlice),
		Arg:        arg,
		TypeParam:  buf.String(),
		Size:       sizeStr,
	}
}

func suffixSlice(s string, suffix bool) string {
	if suffix {
		return s + "Slice"
	}
	return s
}

var (
	undFixedSize = template.Must(template.New("").Parse(
		`{{.UndPkg}}.Map(
	{{.Arg}},
	func(s []{{.OptionPkg}}.Option[{{.TypeParam}}]) (r [{{.Size}}]{{.OptionPkg}}.Option[{{.TypeParam}}]) {
		copy(r[:], s)
		return
	},
)`))
	mapUndNonNullFixedSize = template.Must(template.New("").Parse(
		`{{.UndPkg}}.Map(
	{{.Arg}},
	func(s [{{.Size}}]{{.OptionPkg}}.Option[{{.TypeParam}}]) (r [{{.Size}}]{{.TypeParam}}) {
		for i := 0; i < {{.Size}}; i++ {
			r[i] = s[i].Value()
		}
		return
	},
)`,
	))
)
