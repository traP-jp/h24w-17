package h24w17

import (
	"os"
	"path"
	"text/template"
)

type Generator struct {
	driverTmpl *template.Template
	stmtTmpl   *template.Template
	data       data
}

type data struct {
	PackageName  string
	CachePlanRaw string
}

func NewGenerator(cachePlanRaw string) *Generator {
	return &Generator{
		driverTmpl: template.Must(template.ParseFiles("template/driver.tmpl")),
		stmtTmpl:   template.Must(template.ParseFiles("template/stmt.tmpl")),
		data:       data{CachePlanRaw: cachePlanRaw},
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
