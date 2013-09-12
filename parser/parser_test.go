package parser

import (
	"bytes"
	"os"
)

func ExampleHeaderAndParagraph() {
	wiki := "* Header1\n"
	wiki += "** Header2\n"
	wiki += "How are you\n"
	wiki += "doing?\n"
	wiki += "\n"
	wiki += "Next paragraph.\n"
	p := Parser{}
	r := bytes.NewBufferString(wiki)
	p.Parse(r)
	p.Write(NewXMLWriter(os.Stdout), true)
	// Output:
	// <Document>
	//   <Header level="1">Header1</Header>
	//   <Header level="2">Header2</Header>
	//   <Paragraph>How are you doing?</Paragraph>
	//   <Paragraph>Next paragraph.</Paragraph>
	// </Document>
}

func ExampleLink() {
	wiki := "Link to [[hello]] or [[http://www][www]]."
	p := Parser{}
	r := bytes.NewBufferString(wiki)
	p.Parse(r)
	p.Write(NewXMLWriter(os.Stdout), true)
	// Output:
	// <Document>
	//   <Paragraph>Link to <Link link="hello"></Link> or <Link link="http://www">www</Link>.</Paragraph>
	// </Document>
}

func ExampleComplex() {
	wiki := "* Test\n"
	wiki += "  [[hello]]\n"
	p := Parser{}
	r := bytes.NewBufferString(wiki)
	p.Parse(r)
	p.Write(NewXMLWriter(os.Stdout), true)
	// Output:
	// <Document>
	//   <Header level="1">Test</Header>
	//   <Paragraph><Link link="hello"></Link></Paragraph>
	// </Document>
}

func ExampleComplexHTML() {
	wiki := "* Test\n"
	wiki += "** Test\n"
	wiki += "   [[hello]], [[http://world][world]]\n"
	p := Parser{}
	r := bytes.NewBufferString(wiki)
	p.Parse(r)
	p.Write(NewHTMLWriter(os.Stdout), false)
	// Output:
	// <h1>Test</h1><h2>Test</h2><p><a href="/view/hello">hello</a>, <a href="http://world">world</a></p>
}
