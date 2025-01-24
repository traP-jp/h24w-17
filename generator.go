package h24w17

import (
	"os"
	"path"
	"strings"
	"text/template"
)

type Generator struct {
	driverTmpl *template.Template
	stmtTmpl   *template.Template
	data       data
}

type data struct {
	PackageName    string
	CachePlanRaw   string
	TableSchemaRaw string
}

func NewGenerator(cachePlanRaw string, tableSchemaRaw string) *Generator {
	return &Generator{
		driverTmpl: template.Must(template.ParseFiles("template/driver.tmpl")),
		stmtTmpl:   template.Must(template.ParseFiles("template/stmt.tmpl")),
		data:       data{CachePlanRaw: escapeGoString(cachePlanRaw), TableSchemaRaw: escapeGoString(tableSchemaRaw)},
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

	err = g.driverTmpl.Execute(driver, g.data)
	if err != nil {
		panic(err)
	}
	err = g.stmtTmpl.Execute(stmt, g.data)
	if err != nil {
		panic(err)
	}
}

func escapeGoString(s string) string {
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
