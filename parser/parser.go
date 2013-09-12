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
	Name     string
	Text     string
}

type Parser struct {
	root    *Element
	current *Element
	pos     int
	line    int
	column  int
}

func addElement(parent *Element, name string) *Element {
	newElement := &Element{
		Attr:     map[string]string{},
		Parent:   parent,
		Children: []*Element{},
		Name:     name,
		Text:     "",
	}
	if parent != nil {
		parent.Children = append(parent.Children, newElement)
	}
	return newElement
}

func (p *Parser) startElement(name string) {
	p.current = addElement(p.current, name)
}

func (p *Parser) endElement() {
	p.current = p.current.Parent
}

func (p Parser) isInElement(name string) bool {
	current := p.current
	for current != nil {
		if current.Name == name {
			return true
		}
		current = current.Parent
	}
	return false
}

func (p *Parser) openElement(name string) {
	if !p.isInElement(name) {
		p.startElement(name)
	}
}

func (p *Parser) closeElement(name string) {
	if p.isInElement(name) {
		for p.current.Name != name {
			p.endElement()
		}
		p.endElement()
	}
}

func (p *Parser) nextColumn() {
	p.pos += 1
	p.column += 1
}

func (p *Parser) nextLine() {
	p.pos += 1
	p.line += 1
	p.column = 0
}

func (p Parser) addError(from, msg string) {
	log.Printf("%s: %s at Line %d, Column %d\n", from, msg, p.line, p.column)
	fmt.Printf("%s: %s at Line %d, Column %d\n", from, msg, p.line, p.column)
}

// Parse parses a document.
func (p *Parser) Parse(r io.ByteScanner) {
	p.parseDocument(r)
}

// parseDocument parses a document.
func (p *Parser) parseDocument(r io.ByteScanner) {
	_, err := r.ReadByte()
	if err != nil {
		return
	}
	r.UnreadByte()
	p.root = addElement(nil, "Document")
	p.current = p.root
	for {
		_, err := r.ReadByte()
		if err != nil {
			break
		}
		r.UnreadByte()
		p.parseLine(r)
	}
}

// parseLine parses an unknown line.
func (p *Parser) parseLine(r io.ByteScanner) {
	level := 0
	for {
		c, err := r.ReadByte()
		if err != nil {
			return
		}
		if c == '*' && level == 0 {
			p.closeElement("Paragraph")
			r.UnreadByte()
			p.parseHeader(r)
		} else if c == ' ' {
			level += 1
			p.nextColumn()
		} else if c == '\n' {
			p.closeElement("Paragraph")
			p.nextLine()
			break
		} else if c == '\r' {
			// Ignore CR character.
			p.nextColumn()
		} else {
			p.openElement("Paragraph")
			r.UnreadByte()
			p.startElement("Text")
			p.current.Attr["level"] = fmt.Sprint(level)
			p.parseTextLine(r)
			p.endElement()
		}
	}
}

// parseTextLine parses a text line. This does not create an element.
func (p *Parser) parseTextLine(r io.ByteScanner) {
	for {
		c, err := r.ReadByte()
		if err != nil {
			return
		}
		if c == '\n' {
			p.nextLine()
			break
		}
		if c == '\r' {
			// Ignore CR character.
			p.nextColumn()
			continue
		}
		if c == ']' {
			if p.current.Name == "Link" || (p.current.Name == "Text" && p.current.Parent.Name == "Link") {
				r.UnreadByte()
				return
			} else {
				p.addError("ParseTextLink", "unexpected ]")
				p.nextColumn()
				return
			}
		}
		if c == '[' {
			if p.current.Name == "Text" {
				p.nextColumn()
				p.endElement()
				p.parseLink(r)
				p.startElement("Text")
			} else {
				p.addError("ParseTextLink", "type is "+p.current.Name)
				p.nextColumn()
			}
			continue
		}
		p.nextColumn()
		p.current.Text += string(c)
	}
}

// parseHeader parses a header. Format is "* Header1" or "** Header2", etc.
func (p *Parser) parseHeader(r io.ByteScanner) {
	p.startElement("Header")
	defer p.endElement()
	p.parseHeaderBullet(r)
	p.startElement("Text")
	p.parseTextLine(r)
	p.endElement()
}

// parseHeaderBullet parses a bullet of the header.
func (p *Parser) parseHeaderBullet(r io.ByteScanner) {
	level := 0
	for {
		c, err := r.ReadByte()
		if err != nil {
			p.addError("ParseHeaderBullet", "unexpected EOF")
			return
		}
		if c == '*' {
			level += 1
			p.nextColumn()
		} else if c == ' ' {
			p.nextColumn()
			break
		} else {
			p.nextColumn()
			p.addError("ParseHeaderBullet", "* or space expected")
			return
		}
	}
	p.current.Attr["level"] = fmt.Sprint(level)
}

// parseLink parses a link. Format is [[link]] or [[link][text]].
func (p *Parser) parseLink(r io.ByteScanner) {
	// Start of link is already consumed.
	const (
		start = iota
		link
		middle
		text
		end
	)
	state := start
	p.startElement("Link")
	defer p.endElement()
	for {
		c, err := r.ReadByte()
		if err != nil {
			p.addError("ParseLink", "unexpected EOF")
			return
		}
		if c == '[' {
			p.nextColumn()
			if state == start {
				state = link
			} else if state == middle {
				state = text
			} else {
				p.addError("ParseLink", "unexpected [")
				return
			}
		} else if c == ']' {
			p.nextColumn()
			if state == link {
				state = middle
			} else if state == text {
				state = end
			} else if state == middle {
				break
			} else if state == end {
				return
			} else {
				p.addError("ParseLink", "unexpected ]")
				return
			}
		} else if c == ' ' {
			p.nextColumn()
			continue
		} else {
			if state == link {
				r.UnreadByte()
				p.parseTextLine(r)
				p.current.Attr["link"] += p.current.Text
				p.current.Text = ""
			} else if state == text {
				r.UnreadByte()
				p.startElement("Text")
				p.parseTextLine(r)
				p.endElement()
			} else {
				p.nextColumn()
				p.addError("ParseLink", "unexpected character")
				return
			}
		}
	}
	if len(p.current.Children) == 0 {
		p.startElement("Text")
		p.endElement()
	}
}

// WriteXML writes the parse tree to w in XML. If pretty is true, it
// prints out as indented XML.
func (p Parser) Write(w Writer, pretty bool) {
	if p.root == nil {
		return
	}
	current := p.root
	var writeFunc func(node *Element, level int)
	writeFunc = func(node *Element, level int) {
		if node.Name == "Text" {
			w.Text(node)
		}
		for i := 0; i < level; i++ {
			fmt.Fprint(w, "  ")
		}
		w.StartElement(node)
		hasText := false
		for _, child := range node.Children {
			if child.Name == "Text" {
				hasText = true
				break
			}
		}

		seenText := false
		for _, child := range node.Children {
			if seenText && child.Text != "" {
				fmt.Fprint(w, " ")
			}
			if child.Text != "" {
				seenText = true
			} else if child.Name != "Text" {
				seenText = false
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
		}
		if pretty && !hasText && node.Name != "Text" {
			fmt.Fprintln(w)
		}
		w.EndElement(node)
	}
	writeFunc(current, 0)
}
