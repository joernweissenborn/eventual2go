package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/codegangsta/cli"
)

//flag vars

var pkgName string
var typeNames []string
var names []string
var export bool
var verbose bool
var print bool
var outputFile string

func main() {

	app := cli.NewApp()

	app.Name = "event_generator"

	app.Usage = "Generate typed events for eventual2go"

	app.Version = "0.1"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "export, e",
			Usage:       "exports the events by rasing the first letter of the typename, use for unexported types or internals, like string",
			Destination: &export,
		},
		cli.BoolFlag{
			Name:        "verbose",
			Usage:       "verbose output",
			Destination: &verbose,
		},
		cli.StringFlag{
			Name:        "package, p",
			Usage:       "name of package",
			Destination: &pkgName,
		},
		cli.BoolFlag{
			Name:        "print, pr",
			Usage:       "outputs to the commandline instead writing to file",
			Destination: &print,
		},
		cli.StringFlag{
			Name:        "output, o",
			Usage:       "file to output, leave empty for setting output to %CURRENT_DIR%/TYPE_NAME_events.go. If multiple types are specified, the first is used for the filename",
			Destination: &outputFile,
		},
		cli.StringSliceFlag{
			Name:  "type, t",
			Usage: "name of the type(s), set multiple for multiple types",
		},
		cli.StringSliceFlag{
			Name:  "name, n",
			Usage: "name of the type(s) to be used by methods and type declartion, set multiple for multiple types. Usful if generate events for slices.",
		},
	}

	app.Action = run

	app.Run(os.Args)

}

func run(c *cli.Context) {

	logger := log.New(os.Stdout, "event_generator: ", 0)

	if pkgName == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("error retrieving pkgName from working dir", err)
		}
		pkgName = filepath.Base(cwd)
	}

	typeNames = c.StringSlice("type")
	names = c.StringSlice("name")

	if len(typeNames) == 0 {
		cli.ShowAppHelp(c)
		logger.Fatal("No types defined")
	}
	if verbose {
		logger.Println("Generating for Package", pkgName)
	}
	out := &bytes.Buffer{}

	generateHeader(out)

	for i, typeName := range typeNames {
		if typeName == "" {
			logger.Fatal("Empty type name")
		}

		name := typeName

		if i < len(names) {
			name = names[i]
		} else if export {
			name = strings.ToUpper(string(name[0])) + name[1:]
		}
		generateType(out, typeName, name)
	}
	if print {

		fmt.Print(out)
	} else {
		if outputFile == "" {
			var name string
			if len(names) != 0 {
				name = names[0]
			} else {
				name = typeNames[0]
			}
			outputFile = fmt.Sprintf("%s_events.go", name)
		}
		outputFile = strings.ToLower(outputFile)
		path, err := filepath.Abs(outputFile)
		if err != nil {
			logger.Fatalf("error creating file", err)
		}

		err = ioutil.WriteFile(path, out.Bytes(), os.ModePerm)
		if err != nil {
			logger.Fatalf("error creating file", err)
		}

	}

}

func generateHeader(out io.Writer) {
	t, _ := template.New("").Parse(tmplHeader)

	header := struct{ PkgName string }{pkgName}

	t.Execute(out, header)
}

func generateType(out io.Writer, typeName, name string) {
	t, _ := template.New("").Parse(tmplType)

	tname := struct {
		TypeName string
		Name     string
	}{typeName, name}

	t.Execute(out, tname)
}

