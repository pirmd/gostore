package sanitizer

import (
	"golang.org/x/net/html/atom"
)

var (
	// EPUBElmt is the default list of accepted EPUB tags/attributes
	EPUBElmt = map[atom.Atom][]string{
		atom.A:          append([]string{"href"}, globalHTMLAttr...),
		atom.B:          globalHTMLAttr,
		atom.Blockquote: globalHTMLAttr,
		atom.Body:       globalHTMLAttr,
		atom.Br:         globalHTMLAttr,
		atom.Caption:    globalHTMLAttr,
		atom.Cite:       globalHTMLAttr,
		atom.Col:        append([]string{"span"}, globalHTMLAttr...),
		atom.Colgroup:   append([]string{"span"}, globalHTMLAttr...),
		atom.Dd:         globalHTMLAttr,
		atom.Del:        globalHTMLAttr,
		atom.Dfn:        globalHTMLAttr,
		atom.Div:        globalHTMLAttr,
		atom.Em:         globalHTMLAttr,
		atom.H1:         globalHTMLAttr,
		atom.H2:         globalHTMLAttr,
		atom.H3:         globalHTMLAttr,
		atom.H4:         globalHTMLAttr,
		atom.H5:         globalHTMLAttr,
		atom.H6:         globalHTMLAttr,
		atom.Head:       {},
		atom.Hr:         globalHTMLAttr,
		atom.Html:       {"lang", "xmlns", "xml:lang"},
		atom.I:          globalHTMLAttr,
		atom.Img:        append([]string{"height", "src", "width"}, globalHTMLAttr...),
		atom.Li:         globalHTMLAttr,
		atom.Link:       {"href", "rel=stylesheet", "type=text/css"},
		atom.Ol:         globalHTMLAttr,
		atom.P:          globalHTMLAttr,
		atom.S:          globalHTMLAttr,
		atom.Small:      globalHTMLAttr,
		atom.Span:       globalHTMLAttr,
		atom.Strong:     globalHTMLAttr,
		atom.Sub:        globalHTMLAttr,
		atom.Sup:        globalHTMLAttr,
		atom.Table:      globalHTMLAttr,
		atom.Tbody:      globalHTMLAttr,
		atom.Td:         append([]string{"colspan", "rowspan"}, globalHTMLAttr...),
		atom.Tfoot:      globalHTMLAttr,
		atom.Th:         append([]string{"abbr", "colspan", "rowspan"}, globalHTMLAttr...),
		atom.Thead:      globalHTMLAttr,
		atom.Title:      globalHTMLAttr,
		atom.Tr:         globalHTMLAttr,
		atom.U:          globalHTMLAttr,
		atom.Ul:         globalHTMLAttr,
	}

	// EPUBSchemes lists the accepted schemes to be found in an epub's link.
	EPUBSchemes = []string{"http", "https", "mailto"}

	// EPUB is the default EPUB sanitizer.
	EPUB = NewHTML(EPUBElmt, EPUBSchemes)

	globalHTMLAttr = []string{"class", "id", "lang", "style", "title"}
)
