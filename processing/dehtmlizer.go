package processing

// Package dehtmlizer proposes tools to convert text in html format to
// something more reasonable like pure text or markdown.
//
// dehtmlier is pretty basic so only a sub-part of html formats are correctly
// interpreted (table are not properly considered, neither order lists). It
// might comes in a future version though, if I get convinced that metadata
// containing html formatted descriptions are relying on such formatting.

import (
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pirmd/cli/style"

	"github.com/pirmd/gostore/store"
)

var (
	reLeadingSpaces   = regexp.MustCompile(`^[\s\p{Zs}]+`)
	reTrailingSpaces  = regexp.MustCompile(`[\s\p{Zs}]+$`)
	reRedundantSpaces = regexp.MustCompile(`[\s\p{Zs}]{2,}`)
)

//DeHTMLizeRecord transforms any frmatting in TML format into markdown
//DeHTMLizeRecord cleans only Record's "Description" field if any.
func DeHTMLizeRecord(r *store.Record) error {
	//TODO(pirmd): Clean Record API to obtain a value
	//TODO(pirmd): Allow configuration of which field to clean (maybe depending
	//on Record's Type like renamer)
	desc := r.GetValue("Description")
	if desc == nil {
		return nil //No Description field, nothing to do
	}

	//TODO(pirmd): we are lead to make a lot of assumption of the type of
	//stored attribute, need to do something beeter than map[string]interface{}
	//I guess
	root, err := html.Parse(strings.NewReader(desc.(string)))
	if err != nil {
		return err
	}

	//TODO(pirmd): change HTML2Markdown to accept string as input?
	mkd := HTML2Markdown(root)

	r.SetValue("Description", mkd)
	return nil
}

//renderNode returns text from all root's descendant text nodes
func renderNode(root *html.Node, st style.Styler) string {
	var txt string

	if root.Type == html.TextNode {
		if t := reLeadingSpaces.ReplaceAllString(root.Data, ""); reTrailingSpaces.ReplaceAllString(t, "") != "" {
			t = reRedundantSpaces.ReplaceAllString(t, " ")
			txt = txt + t
		}
	}

	for node := root.FirstChild; node != nil; node = node.NextSibling {
		switch node.Type {
		case html.ElementNode:
			switch node.DataAtom {
			case atom.A:
				t := renderNode(node, st)
				txt = txt + st.Link(t, getAttr(node, "href"))

			case atom.Img:
				txt = txt + st.Img(getAttr(node, "alt"), getAttr(node, "src"))

			case atom.B, atom.Strong:
				t := renderNode(node, st)
				txt = txt + st.Bold(t)

			case atom.I, atom.Em:
				t := renderNode(node, st)
				txt = txt + st.Italic(t)

			case atom.Del:
				t := renderNode(node, st)
				txt = txt + st.Crossout(t)

			case atom.Br:
				txt = txt + st.Paragraph("")

			case atom.P:
				t := renderNode(node, st)
				txt = txt + st.Paragraph(t)

			case atom.H1:
				t := renderNode(node, st)
				txt = txt + st.Header(1)(t)
			case atom.H2:
				t := renderNode(node, st)
				txt = txt + st.Header(2)(t)
			case atom.H3:
				t := renderNode(node, st)
				txt = txt + st.Header(3)(t)
			case atom.H4:
				t := renderNode(node, st)
				txt = txt + st.Header(4)(t)
			case atom.H5:
				t := renderNode(node, st)
				txt = txt + st.Header(5)(t)
			case atom.H6:
				t := renderNode(node, st)
				txt = txt + st.Header(6)(t)

			case atom.Ul:
				fn := st.BulletedList() //need to get fn first to initialize nested list tracking

				var list []string
				for item := node.FirstChild; item != nil; item = item.NextSibling {
					if item.DataAtom == atom.Li {
						list = append(list, renderNode(item, st))
					}
				}

				if txt == "" || strings.HasSuffix(txt, "\n") {
					txt = txt + fn(list...)
				} else {
					txt = txt + "\n" + fn(list...)
				}

			case atom.Ol:
				fn := st.OrderedList() //need to get fn first to initialize nested list tracking

				var list []string
				for item := node.FirstChild; item != nil; item = item.NextSibling {
					if item.DataAtom == atom.Li {
						list = append(list, renderNode(item, st))
					}
				}

				if txt == "" || strings.HasSuffix(txt, "\n") {
					txt = txt + fn(list...)
				} else {
					txt = txt + "\n" + fn(list...)
				}

			case atom.Table:
				var rows [][]string

				for n := node.FirstChild; n != nil; n = n.NextSibling {
					if n.DataAtom == atom.Tbody {

						for row := n.FirstChild; row != nil; row = row.NextSibling {
							if row.DataAtom == atom.Tr {

								var r []string
								for cell := row.FirstChild; cell != nil; cell = cell.NextSibling {
									if cell.DataAtom == atom.Th || cell.DataAtom == atom.Td {
										r = append(r, renderNode(cell, st))
									}
								}

								rows = append(rows, r)
							}
						}

						break
					}
				}

				txt = txt + st.Table(rows...)

			default:
				txt = txt + renderNode(node, st)
			}

		case html.CommentNode:
			//do nothing

		default:
			txt = txt + renderNode(node, st)
		}

	}

	return txt
}

//HTML2Markdown filters any supplied html into a markdown
func HTML2Markdown(root *html.Node) string {
	return renderNode(root, style.NewMarkdown())
}

//getAttr returns the value of an HTML node attribute.  If no attribute exists
//corresponding to the given name, returns an empty string
func getAttr(node *html.Node, attr string) string {
	for _, a := range node.Attr {
		if a.Key == attr {
			return a.Val
		}
	}
	return ""
}

func init() {
	RecordProcessors["html2md"] = DeHTMLizeRecord
}
