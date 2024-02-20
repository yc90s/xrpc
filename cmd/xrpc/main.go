package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/yc90s/xrpc"
)

const tmplService = `// Code generated by xrpc. DO NOT EDIT.
{{$root := . -}}
package {{$root.Name}}

import "github.com/yc90s/xrpc"
{{- range $_, $im := $root.Imports}}
import "{{$im}}"
{{- end}}

{{range $_, $m := $root.Services}}
type I{{$m.Name}} interface {
    {{- range $_, $method := $m.Methods}}
    {{$method.Name -}}(
	{{- range $index, $arg := $method.Args}}
		{{- $arg -}}
		{{if ne $index (sub (len $method.Args) 1)}}, {{end}}
	{{- end -}})
    {{- if len $method.Returns }} (
		{{- range $index, $ret := $method.Returns}}
		{{- $ret -}}
		{{if ne $index (sub (len $method.Returns) 1)}}, {{end}}
		{{- end}})
    {{- end}}
    {{- end}}
}

func Register{{$m.Name}}Server(rpc *xrpc.RPCServer, s I{{$m.Name}}) {
	{{- range $_, $method := $m.Methods }}
    rpc.Register("{{$method.Name}}", s.{{$method.Name}})
	{{- end}}
}

type {{$m.Name}}Client struct {
    *xrpc.RPCClient
}

func New{{$m.Name}}Client(c *xrpc.RPCClient) *{{$m.Name}}Client {
    return &{{$m.Name}}Client{c}
}
{{range $_, $method := $m.Methods}}
func (c *{{$m.Name}}Client) {{$method.Name}}(subj string
	{{- range $index, $arg := $method.Args -}}
		, arg{{$index}} {{$arg -}}
	{{- end -}}) 
	{{- if len $method.Returns }} (
		{{- range $index, $ret := $method.Returns}}
		{{- $ret -}}
		{{if ne $index (sub (len $method.Returns) 1)}}, {{end}}
		{{- end}})
	{{- else}} error
	{{- end}} {

	{{- if len $method.Returns }}
    var reply {{decayReply (getReply $method.Returns)}}
    err := c.Call(subj, "{{$method.Name}}", &reply
    {{- range $index, $arg := $method.Args -}}
		, arg{{$index}}
	{{- end -}})
	{{- if isPointerReply (getReply $method.Returns)}}
    return &reply, err
	{{- else}}
    return reply, err
	{{- end}}

	{{- else}}
    err := c.Cast(subj, "{{$method.Name}}"
	{{- range $index, $arg := $method.Args -}}
		, arg{{$index}}
	{{- end -}})
    return err
	{{- end}}
}
{{- end -}}
{{- end -}}
`

func generate(ast *PackageAST, file *os.File) error {
	funcs := template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
		"getReply": func(returns []string) string {
			if len(returns) == 0 {
				panic("no return")
			}
			return returns[0]
		},
		"decayReply": func(reply string) string {
			if len(reply) < 1 {
				return reply
			}
			if reply[0] == '*' {
				return reply[1:]
			}
			return reply
		},
		"isPointerReply": func(reply string) bool {
			if len(reply) < 1 {
				return false
			}
			return reply[0] == '*'
		},
	}
	t := template.Must(template.New("").Funcs(funcs).Parse(tmplService))

	return t.Execute(file, ast)
}

func main() {
	var version bool
	var outDir string
	flag.BoolVar(&version, "version", false, "print version")
	flag.StringVar(&outDir, "out", ".", "output directory")
	flag.Parse()

	if version {
		fmt.Println(xrpc.Version)
		return
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s [flags] IDL:\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, "  %s -out out_dir {{file_pattern}}\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fmt.Fprintf(os.Stderr, "  -version\n    Show the version of xrpc\n")
		fmt.Fprintf(os.Stderr, "  -out string\n    Specify output directory (default \".\")\n")
		fmt.Fprintf(os.Stderr, "  file_pattern\n    Specify service file pattern\n")
	}

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		return
	}

	for _, inFilePattern := range args {
		if inFilePattern == "" {
			continue
		}
		files, err := filepath.Glob(inFilePattern)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, inFile := range files {
			content, err := os.ReadFile(inFile)
			if err != nil {
				fmt.Println(err)
				continue
			}

			lexer := NewLexer(content)
			parser := NewParser(lexer)
			ast, err := parser.parse()
			if err != nil {
				panic(err)
			}

			outFile := fmt.Sprintf("%s/%s.go", outDir, filepath.Base(inFile))

			file, err := os.Create(outFile)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			err = generate(ast, file)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}
}