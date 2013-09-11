package parser

import (
	"io"
	"fmt"
	"strings"
)

type XMLWriter struct {
	writer io.Writer
}

func NewXMLWriter(writer io.Writer) *XMLWriter {
	return &XMLWriter{writer}
}

func (w XMLWriter) StartElement(node *Element) {
	if node.Type == "Text" {
		return
	}
	attrs := []string{}
	for k, v := range node.Attr {
		attrs = append(attrs, fmt.Sprintf("%s=\"%s\"", k, v))
	}
	if len(attrs) == 0 {
		fmt.Fprintf(w.writer, "<%s>", node.Type)
	} else {
		fmt.Fprintf(w.writer, "<%s %s>", node.Type, strings.Join(attrs, " "))
	}
}

func (w XMLWriter) Text(node *Element) {
	fmt.Fprintf(w.writer, node.Text)
}

func (w XMLWriter) EndElement(node *Element) {
	if node.Type == "Text" {
		return
	}
	fmt.Fprintf(w.writer, "</%s>", node.Type)
}
