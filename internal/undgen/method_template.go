package undgen

import (
	"go/ast"
	"io"
	"strings"
	"text/template"
)

func (t PlainType) typeNames() (raw, plain string) {
	ts := t.Decl.Specs[0].(*ast.TypeSpec)
	plain = ts.Name.Name
	raw, _ = strings.CutSuffix(plain, "Plain")
	if ts.TypeParams != nil {
		var typeParams strings.Builder
		for _, f := range ts.TypeParams.List {
			if typeParams.Len() > 0 {
				typeParams.WriteByte(',')
			}
			typeParams.WriteString(f.Names[0].Name)
		}

		plain += "[" + typeParams.String() + "]"
		raw += "[" + typeParams.String() + "]"
	}
	return
}

func (t PlainType) PrintToPlain(w io.Writer) error {
	rawTypeName, plainTypeName := t.typeNames()

	param := conversionMethodsParams{
		PrefixComment: DirectivePrefix + DirectiveCommentGenerated,
		SrcTypeName:   rawTypeName,
		DstTypeName:   plainTypeName,
	}

	for _, f := range t.FieldConverters {
		var exp string
		if f.Converter != nil {
			exp = f.Converter.Expr("v." + f.FieldName)
		} else {
			exp = "v." + f.FieldName
		}
		param.Fields = append(param.Fields, field{
			FieldName: f.FieldName,
			FieldExpr: exp,
		})
	}

	return toPlainMethod.Execute(w, param)
}

func (t PlainType) PrintToRaw(w io.Writer) error {
	rawTypeName, plainTypeName := t.typeNames()

	param := conversionMethodsParams{
		PrefixComment: DirectivePrefix + DirectiveCommentGenerated,
		SrcTypeName:   plainTypeName,
		DstTypeName:   rawTypeName,
	}

	for _, f := range t.FieldConverters {
		var exp string
		if f.BackConverter != nil {
			exp = f.BackConverter.Expr("v." + f.FieldName)
		} else {
			exp = "v." + f.FieldName
		}
		param.Fields = append(param.Fields, field{
			FieldName: f.FieldName,
			FieldExpr: exp,
		})
	}

	return toRawMethod.Execute(w, param)
}

type conversionMethodsParams struct {
	PrefixComment string
	SrcTypeName   string
	DstTypeName   string
	Fields        []field
}

type field struct {
	FieldName string
	FieldExpr string
}

var (
	toPlainMethod = template.Must(template.New("").Parse(
		`//{{.PrefixComment}}
func (v {{.SrcTypeName}}) UndPlain() {{.DstTypeName}} {
	return {{.DstTypeName}}{
{{ range .Fields }}                {{.FieldName}}: {{.FieldExpr}},
{{end}}
	}
}`,
	))
	toRawMethod = template.Must(template.New("").Parse(
		`//{{.PrefixComment}}
func (v {{.SrcTypeName}}) UndRaw() {{.DstTypeName}} {
	return {{.DstTypeName}}{
{{ range .Fields }}                {{.FieldName}}: {{.FieldExpr}},
{{end}}
	}
}`,
	))
)
