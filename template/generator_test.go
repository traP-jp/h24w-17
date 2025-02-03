package template

import (
	"embed"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

//go:embed *.go
var expectedFiles embed.FS

func TestGenerator(t *testing.T) {
	g := NewGenerator("", "")
	g.data.PackageName = "template"

	writer := strings.Builder{}

	tests := []struct {
		name     string
		template *template.Template
	}{
		{"driver", g.driverTmpl},
		{"stmt", g.stmtTmpl},
		{"cache", g.cacheTmpl},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected, err := expectedFiles.ReadFile(test.name + ".go")
			assert.NoError(t, err)
			writer.Reset()
			err = test.template.Execute(&writer, g.data)
			assert.NoError(t, err)
			assert.Equal(t, string(expected), writer.String())
		})
	}
}
