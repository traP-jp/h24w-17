package template

import (
	"embed"
	"os"
	"path"
	"strings"
	"text/template"
)

type Generator struct {
	driverTmpl *template.Template
	stmtTmpl   *template.Template
	cacheTmpl  *template.Template
	data       data
}

type data struct {
	PackageName    string
	CachePlanRaw   string
	TableSchemaRaw string
}

//go:embed *.tmpl
var templates embed.FS

func NewGenerator(cachePlanRaw string, tableSchemaRaw string) *Generator {
	return &Generator{
		driverTmpl: template.Must(template.ParseFS(templates, "driver.tmpl")),
		stmtTmpl:   template.Must(template.ParseFS(templates, "stmt.tmpl")),
		cacheTmpl:  template.Must(template.ParseFS(templates, "cache.tmpl")),
		data:       data{CachePlanRaw: toEscapedGoStringLiteral(cachePlanRaw), TableSchemaRaw: toEscapedGoStringLiteral(tableSchemaRaw)},
	}
}

func (g *Generator) Generate(destDir string) {
	g.data.PackageName = path.Base(destDir)
	driver, err := os.Create(path.Join(destDir, "driver.go"))
	if err != nil {
		panic(err)
	}
	defer driver.Close()
	stmt, err := os.Create(path.Join(destDir, "stmt.go"))
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	cache, err := os.Create(path.Join(destDir, "cache.go"))
	if err != nil {
		panic(err)
	}
	defer cache.Close()

	err = g.driverTmpl.Execute(driver, g.data)
	if err != nil {
		panic(err)
	}
	err = g.stmtTmpl.Execute(stmt, g.data)
	if err != nil {
		panic(err)
	}
	err = g.cacheTmpl.Execute(cache, g.data)
	if err != nil {
		panic(err)
	}
}

func toEscapedGoStringLiteral(s string) string {
	split := strings.Split(s, "`")
	var b strings.Builder
	b.WriteByte('`')
	for i, v := range split {
		b.WriteString(v)
		if i != len(split)-1 {
			b.WriteString("` + \"`\" + `")
		}
	}
	b.WriteByte('`')
	return b.String()
}
