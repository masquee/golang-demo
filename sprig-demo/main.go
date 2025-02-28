package main

import (
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func main() {
	tpl := `{{ .Variable | default "default value" }}`

	t, err := template.New("example").Funcs(sprig.FuncMap()).Parse(tpl)
	if err != nil {
		panic(err)
	}

	data := map[string]interface{}{
		"Variable": "some value", // Uncomment to see the difference
	}

	err = t.Execute(os.Stdout, data)
	if err != nil {
		panic(err)
	}
}
