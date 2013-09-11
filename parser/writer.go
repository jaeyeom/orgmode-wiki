package parser

import (
	"io"
	"fmt"
	"strings"
)

type Writer interface {
	io.Writer
	StartElement(node *Element)
	Text(node *Element)
	EndElement(node *Element)
}

type XMLWriter struct {
	io.Writer
}

func NewXMLWriter(writer io.Writer) *XMLWriter {
	return &XMLWriter{writer}
}

func (w XMLWriter) StartElement(node *Element) {
	if node.Name == "Text" {
		return
	}
	attrs := []string{}
	for k, v := range node.Attr {
		attrs = append(attrs, fmt.Sprintf("%s=\"%s\"", k, v))
	}
	if len(attrs) == 0 {
		fmt.Fprintf(w, "<%s>", node.Name)
	} else {
		fmt.Fprintf(w, "<%s %s>", node.Name, strings.Join(attrs, " "))
	}
}

func (w XMLWriter) Text(node *Element) {
	fmt.Fprintf(w, node.Text)
}

func (w XMLWriter) EndElement(node *Element) {
	if node.Name == "Text" {
		return
	}
	fmt.Fprintf(w, "</%s>", node.Name)
}

type HTMLWriter struct {
	io.Writer
}

func NewHTMLWriter(writer io.Writer) *HTMLWriter {
	return &HTMLWriter{writer}
}

func (w HTMLWriter) StartElement(node *Element) {
	name := ""
	attrs := map[string]string{}
	switch node.Name {
	case "Link":
		name = "a"
		attrs["href"] = node.Attr["link"]
	case "Paragraph":
		name = "p"
	case "Header":
		name = "h" + node.Attr["level"]
	}
	if name == "" {
		return
	}
	attrsStr := []string{}
	for k, v := range attrs {
		if v == "" {
			continue
		}
		attrsStr = append(attrsStr, fmt.Sprintf("%s=\"%s\"", k, v))
	}
	if len(attrs) == 0 {
		fmt.Fprintf(w, "<%s>", name)
	} else {
		fmt.Fprintf(w, "<%s %s>", name, strings.Join(attrsStr, " "))
	}
}

func (w HTMLWriter) Text(node *Element) {
	fmt.Fprintf(w, node.Text)
}

func (w HTMLWriter) EndElement(node *Element) {
	name := ""
	switch node.Name {
	case "Link":
		name = "a"
	case "Paragraph":
		name = "p"
	case "Header":
		name = "h" + node.Attr["level"]
	}
	if name == "" {
		return
	}
	fmt.Fprintf(w, "</%s>", name)
}
