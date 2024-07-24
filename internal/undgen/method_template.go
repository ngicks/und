package undgen

import (
	"go/ast"
	"io"
	"strings"
	"text/template"
)

func (t PlainType) PrintToPlain(w io.Writer) error {
	ts := t.Decl.Specs[0].(*ast.TypeSpec)
	dstTypeName := ts.Name.Name
	srcTypeName, _ := strings.CutSuffix(dstTypeName, "Plain")

	if ts.TypeParams != nil {
		var typeParams strings.Builder
		for _, f := range ts.TypeParams.List {
			if typeParams.Len() > 0 {
				typeParams.WriteByte(',')
			}
			typeParams.WriteString(f.Names[0].Name)
		}

		dstTypeName += "[" + typeParams.String() + "]"
		srcTypeName += "[" + typeParams.String() + "]"
	}

	param := ToPlainMethodsParams{
		PrefixComment: DirectivePrefix + DirectiveCommentGenerated,
		SrcTypeName:   srcTypeName,
		DstTypeName:   dstTypeName,
	}

	for _, f := range t.ToPlain {
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

	return toPlainMethods.Execute(w, param)
}

type ToPlainMethodsParams struct {
	PrefixComment string
	SrcTypeName   string
	DstTypeName   string
	Fields        []field
}

type field struct {
	FieldName string
	FieldExpr string
}

var toPlainMethods = template.Must(template.New("").Parse(
	`//{{.PrefixComment}}
func (v {{.SrcTypeName}}) ToPlain() {{.DstTypeName}} {
	return {{.DstTypeName}}{
{{ range .Fields }}                {{.FieldName}}: {{.FieldExpr}},
{{end}}
	}
}`,
))
