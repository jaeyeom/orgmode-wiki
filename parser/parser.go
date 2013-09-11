package parser

import (
	"fmt"
	"io"
	"log"
)

type Element struct {
	Attr     map[string]string
	Parent   *Element
	Children []*Element
	Type     string
	Text     string
}

type Parser struct {
	root    *Element
	current *Element
	pos     int
	line    int
	column  int
}

func AddElement(parent *Element, typeName string) *Element {
	newElement := &Element{
		Attr:     map[string]string{},
		Parent:   parent,
		Children: []*Element{},
		Type:     typeName,
		Text:     "",
	}
	if parent != nil {
		parent.Children = append(parent.Children, newElement)
	}
	return newElement
}

func (p *Parser) StartElement(typeName string) {
	p.current = AddElement(p.current, typeName)
}

func (p *Parser) EndElement() {
	p.current = p.current.Parent
}

func (p *Parser) NextColumn() {
	p.pos += 1
	p.column += 1
}

func (p *Parser) NextLine() {
	p.pos += 1
	p.line += 1
	p.column = 0
}

func (p *Parser) AddError(from, msg string) {
	log.Printf("%s: %s at Line %d, Column %d\n", from, msg, p.line, p.column)
	fmt.Printf("%s: %s at Line %d, Column %d\n", from, msg, p.line, p.column)
}

// ParseDocument parses a document.
func (p *Parser) ParseDocument(r io.ByteScanner) {
	_, err := r.ReadByte()
	if err != nil {
		return
	}
	r.UnreadByte()
	p.root = AddElement(nil, "Document")
	p.current = p.root
	for {
		_, err := r.ReadByte()
		if err != nil {
			break
		}
		r.UnreadByte()
		p.ParseLine(r)
	}
}

// ParseLine parses an unknown line.
func (p *Parser) ParseLine(r io.ByteScanner) {
	level := 0
	for {
		c, err := r.ReadByte()
		if err != nil {
			return
		}
		if c == '*' && level == 0 {
			r.UnreadByte()
			p.ParseHeader(r)
		} else if c == ' ' {
			level += 1
			p.NextColumn()
		} else if c == '\n' {
			if p.current.Type == "Paragraph" {
				p.EndElement()
			}
			p.NextLine()
			break
		} else {
			if p.current.Type != "Paragraph" {
				p.StartElement("Paragraph")
			}
			r.UnreadByte()
			p.StartElement("Text")
			p.current.Attr["level"] = fmt.Sprint(level)
			p.ParseTextLine(r)
			p.EndElement()
		}
	}
}

// ParseTextLine parses a text line. This does not create an element.
func (p *Parser) ParseTextLine(r io.ByteScanner) {
	for {
		c, err := r.ReadByte()
		if err != nil {
			return
		}
		if c == '\n' {
			p.NextLine()
			break
		}
		if c == ']' {
			if p.current.Type == "Link" || (p.current.Type == "Text" && p.current.Parent.Type == "Link") {
				r.UnreadByte()
				return
			} else {
				p.AddError("ParseTextLink", "unexpected ]")
				p.NextColumn()
				return
			}
		}
		if c == '[' {
			if p.current.Type == "Text" {
				p.NextColumn()
				p.EndElement()
				p.ParseLink(r)
				p.StartElement("Text")
			} else {
				p.AddError("ParseTextLink", "type is "+p.current.Type)
				p.NextColumn()
			}
			continue
		}
		p.NextColumn()
		p.current.Text += string(c)
	}
}

// ParseHeader parses a header. Format is "* Header1" or "** Header2", etc.
func (p *Parser) ParseHeader(r io.ByteScanner) {
	p.StartElement("Header")
	defer p.EndElement()
	p.ParseHeaderBullet(r)
	p.StartElement("Text")
	p.ParseTextLine(r)
	p.EndElement()
}

// ParseHeaderBullet parses a bullet of the header.
func (p *Parser) ParseHeaderBullet(r io.ByteScanner) {
	level := 0
	for {
		c, err := r.ReadByte()
		if err != nil {
			p.AddError("ParseHeaderBullet", "unexpected EOF")
			return
		}
		if c == '*' {
			level += 1
			p.NextColumn()
		} else if c == ' ' {
			p.NextColumn()
			break
		} else {
			p.NextColumn()
			p.AddError("ParseHeaderBullet", "* or space expected")
			return
		}
	}
	p.current.Attr["level"] = fmt.Sprint(level)
}

// ParseLink parses a link. Format is [[link]] or [[link][text]].
func (p *Parser) ParseLink(r io.ByteScanner) {
	// Start of link is already consumed.
	state := "start"
	p.StartElement("Link")
	defer p.EndElement()
	for {
		c, err := r.ReadByte()
		if err != nil {
			p.AddError("ParseLink", "unexpected EOF")
			return
		}
		if c == '[' {
			p.NextColumn()
			if state == "start" {
				state = "link"
			} else if state == "middle" {
				state = "text"
			} else {
				p.AddError("ParseLink", "unexpected [")
				return
			}
		} else if c == ']' {
			p.NextColumn()
			if state == "link" {
				state = "middle"
			} else if state == "text" {
				state = "end"
			} else if state == "middle" {
				break
			} else if state == "end" {
				return
			} else {
				p.AddError("ParseLink", "unexpected ]")
				return
			}
		} else if c == ' ' {
			p.NextColumn()
			continue
		} else {
			if state == "link" {
				r.UnreadByte()
				p.ParseTextLine(r)
				p.current.Attr["link"] += p.current.Text
				p.current.Text = ""
			} else if state == "text" {
				r.UnreadByte()
				p.StartElement("Text")
				p.ParseTextLine(r)
				p.EndElement()
			} else {
				p.NextColumn()
				p.AddError("ParseLink", "unexpected character")
				return
			}
		}
	}
	if len(p.current.Children) == 0 {
		p.StartElement("Text")
		p.EndElement()
	}
}

// WriteXML writes the parse tree to w in XML. If pretty is true, it
// prints out as indented XML.
func (p *Parser) WriteXML(w io.Writer, pretty bool) {
	if p.root == nil {
		return
	}
	current := p.root
	writer := NewXMLWriter(w)
	var writeFunc func(node *Element, level int)
	writeFunc = func(node *Element, level int) {
		if node.Type == "Text" {
			writer.Text(node)
		}
		for i := 0; i < level; i++ {
			fmt.Fprint(w, "  ")
		}
		writer.StartElement(node)
		hasText := false
		for _, child := range node.Children {
			if child.Type == "Text" {
				hasText = true
				break
			}
		}

		prevChildType := ""
		for _, child := range node.Children {
			if prevChildType == "Text" && child.Type == "Text" && child.Text != "" {
				fmt.Fprint(w, " ")
			}
			if pretty && !hasText {
				fmt.Fprintln(w)
				writeFunc(child, level+1)
				for i := 0; i < level; i++ {
					fmt.Fprint(w, "  ")
				}
			} else {
				writeFunc(child, 0)
			}
			prevChildType = child.Type
		}
		if pretty && !hasText && node.Type != "Text" {
			fmt.Fprintln(w)
		}
		writer.EndElement(node)
	}
	writeFunc(current, 0)
}
