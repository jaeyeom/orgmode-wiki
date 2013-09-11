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
	p.ParseDocument(r)
	p.WriteXML(os.Stdout, true)
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
	p.ParseDocument(r)
	p.WriteXML(os.Stdout, true)
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
	p.ParseDocument(r)
	p.WriteXML(os.Stdout, true)
	// Output:
	// <Document>
	//   <Header level="1">Test</Header>
	//   <Paragraph><Link link="hello"></Link></Paragraph>
	// </Document>
}
