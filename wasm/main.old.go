package main

import (
	"syscall/js"

	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"strings"

	"golang.org/x/net/html"
)

type Data map[string]interface{}

type Renderer func(out io.Writer, data Data) error

func render(tmpl string, data Data, id string) string {
	buf := &bytes.Buffer{}
	doc := &html.Node{
		Type: html.ElementNode,
		Data: "doc",
	}
	els, err := html.ParseFragment(strings.NewReader(tmpl), doc)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range els {
		doc.AppendChild(v)
	}
	var f func(*html.Node) Renderer
	f = func(n *html.Node) Renderer {
		if n.Type == html.DocumentNode {
			return func(out io.Writer, data Data) error {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if err := f(c)(out, data); err != nil {
						return err
					}
				}
				return nil
			}
		}
		if n.Type == html.ElementNode {
			return func(out io.Writer, data Data) error {
				var iter []string
				iterName := ""
				attrs := make(map[string]string)
				for _, a := range n.Attr {
					if strings.HasPrefix(a.Key, "v-if") {
						ok, exists := data[a.Val]
						if !exists || !ok.(bool) {
							return nil
						}
					} else if strings.HasPrefix(a.Key, "v-bind:") {
						key := strings.Replace(a.Key, "v-bind:", "", 1)
						attrs[key] = data[a.Val].(string)
					} else if strings.HasPrefix(a.Key, "v-on:") {
						key := strings.Replace(a.Key, "v-on:", "", 1)
						attrs["on"+key] = fmt.Sprintf("%s.%s()", id, a.Val)
					} else if strings.HasPrefix(a.Key, "v-for") {
						parts := strings.Split(a.Val, " in ")
						// TODO: reflect and iterate over idx not a placeholder type
						iter = data[parts[1]].([]string)
						iterName = parts[0]
					} else {
						attrs[a.Key] = a.Val
					}
				}
				fmt.Fprintf(out, "  <%s%s>\n", n.Data, htmlAttrs(attrs))
				indentOut := NewIndentWriter(out, []byte("  "))
				if iter == nil {
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						if err := f(c)(indentOut, data); err != nil {
							return err
						}
					}
				} else {
					for _, el := range iter {
						data[iterName] = el
						for c := n.FirstChild; c != nil; c = c.NextSibling {
							if err := f(c)(indentOut, data); err != nil {
								return err
							}
						}
					}
				}
				fmt.Fprintf(out, "  </%s>\n", n.Data)
				return nil
			}
		}
		if n.Type == html.TextNode {
			t := template.Must(template.New("").Parse(n.Data + "\n"))
			return func(out io.Writer, data Data) error {
				out = NewIndentWriter(out, []byte("  "))
				return t.Execute(out, data)
			}
		}
		return nil
	}
	f(doc)(buf, data)
	html := buf.String()
	return html[7 : len(html)-7]
}

func htmlAttrs(attrs map[string]string) string {
	s := ""
	for k, v := range attrs {
		s = fmt.Sprintf("%s %s=\"%s\"", s, k, v)
	}
	return s
}

// Writer indents each line of its input.
type indentWriter struct {
	w   io.Writer
	bol bool
	pre [][]byte
	sel int
	off int
}

// NewIndentWriter makes a new write filter that indents the input
// lines. Each line is prefixed in order with the corresponding
// element of pre. If there are more lines than elements, the last
// element of pre is repeated for each subsequent line.
func NewIndentWriter(w io.Writer, pre ...[]byte) io.Writer {
	return &indentWriter{
		w:   w,
		pre: pre,
		bol: true,
	}
}

// The only errors returned are from the underlying indentWriter.
func (w *indentWriter) Write(p []byte) (n int, err error) {
	for _, c := range p {
		if w.bol {
			var i int
			i, err = w.w.Write(w.pre[w.sel][w.off:])
			w.off += i
			if err != nil {
				return n, err
			}
		}
		_, err = w.w.Write([]byte{c})
		if err != nil {
			return n, err
		}
		n++
		w.bol = c == '\n'
		if w.bol {
			w.off = 0
			if w.sel < len(w.pre)-1 {
				w.sel++
			}
		}
	}
	return n, nil
}

type Vue struct {
	Element  string
	Template string
	Data     Data
	Methods  Methods
	ID       string
}

type Methods map[string]func(*Vue)

func (v *Vue) Render() string {
	return render(v.Template, v.Data, v.ID)
}

func (v *Vue) Mount() {
	body := js.Global().Get("document").Get("body")
	// win := js.Global().Get("window")
	// el := qs.Get("call").Invoke(win, v.Element)
	body.Set("innerHTML", v.Render())
	methods := make(map[string]interface{})
	for k, m := range v.Methods {
		fn := m
		methods[k] = js.NewCallback(func(args []js.Value) {
			fn(v)
		})
	}
	js.Global().Set(v.ID, methods)
}

type Component struct {
	Element string
	Template string
}

type ButtonCounter struct {
	Counter int
}

func (c *ButtonCounter) Template() string {
	return `<button v-on:click="Greet">Greet</button><p>Here is a number: {{ .Counter }}.</p>`
}

func (c *ButtonCounter) Greet() {
	SetState(c, "Counter", c.Counter+1)
}

func main() {
	v := &Vue{
		ID: "v1",

		Template: `<button v-on:click="greet">Greet</button><p>Here is a number: {{ .Counter }}.</p>`,
		Data: Data{
			"Counter": 4,
		},
		Methods: Methods{
			"greet": func(v *Vue) {
				val := v.Data["Counter"].(int)
				v.Data["Counter"] = val + 1
				v.Mount()
				//js.Global().Get("alert").Invoke("Hello")
			},
		},
	}
	v.Mount()
	select {}
}
