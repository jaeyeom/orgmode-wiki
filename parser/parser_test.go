package parser

import (
	"bytes"
	"os"
)

func ExampleHeaderAndParagraph() {
	wiki := "* Header1\n"
	wiki += "** Header2\n"
	wiki += "How *are* you\n"
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
	//   <Paragraph>How <Bold>are</Bold> you doing?</Paragraph>
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
	wiki += "** Test\n"
	wiki += "   _*[[hello]]*_, [[http://world][world]]\n"
	p := Parser{}
	r := bytes.NewBufferString(wiki)
	p.Parse(r)
	p.Write(NewXMLWriter(os.Stdout), true)
	// Output:
	// <Document>
	//   <Header level="1">Test</Header>
	//   <Header level="2">Test</Header>
	//   <Paragraph><Underline><Bold><Link link="hello"></Link></Bold></Underline>, <Link link="http://world">world</Link></Paragraph>
	// </Document>
}

func ExampleExampleLine() {
	wiki := "* Test\n"
	wiki += "paragraph.\n"
	wiki += ": example 1.\n"
	wiki += "  : example 2.\n"
	wiki += "  :normal text.\n"
	p := Parser{}
	r := bytes.NewBufferString(wiki)
	p.Parse(r)
	p.Write(NewXMLWriter(os.Stdout), true)
	// Output:
	// <Document>
	//   <Header level="1">Test</Header>
	//   <Paragraph>paragraph.</Paragraph>
	//   <Example>example 1.</Example>
	//   <Example>example 2.</Example>
	//   <Paragraph>:normal text.</Paragraph>
	// </Document>
}

func ExampleComplexHTML() {
	wiki := "* Test\n"
	wiki += "** Test\n"
	wiki += "   _*[[hello]]*_, [[http://world][world]]\n"
	wiki += "   : example"
	p := Parser{}
	r := bytes.NewBufferString(wiki)
	p.Parse(r)
	p.Write(NewHTMLWriter(os.Stdout), false)
	// Output:
	// <h1>Test</h1><h2>Test</h2><p><u><b><a href="/view/hello">hello</a></b></u>, <a href="http://world">world</a></p><pre>example</pre>
}
