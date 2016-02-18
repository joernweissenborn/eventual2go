package main

import (
	"os"
	"text/template"
)

type EventType struct {
	PkgName  string
	TypeName string
}

func main() {
	t := EventType{"myp", "Myt"}
	tpl := template.Must(template.ParseFiles("typed_events.tmpl"))
	tpl.Execute(os.Stderr, t)
}
