// Package sanitizer provides functions for sanitizing HTML text.
package sanitizer

import (
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pirmd/gostore/util"
)

// TODO: Filter CSS

// HTML represents an HTML sanitizer with its set of rules based on white-list
// approach.
type HTML struct {
	// SafeElements is the white-list of allowed Tags/Attributes. They are
	// organized by Tags to narrow white-listing of attribute.
	// Attr can be an expression:
	// - *    : all attributes are allowed
	// - a=key: only the given key is accepted
	SafeElements map[atom.Atom][]string

	// SafeSchemes is the white-list of allowed schemes in URL.
	SafeSchemes []string
}

// NewHTML creates a new HTML with given set of allowed tags,
// attributes and schemes.
func NewHTML(safeElmt map[atom.Atom][]string, safeSchemes []string) *HTML {
	return &HTML{
		SafeElements: safeElmt,
		SafeSchemes:  safeSchemes,
	}
}

// Sanitize sanitizes an io.Reader into an io.Writer
func (s *HTML) Sanitize(w io.Writer, r io.Reader) error {
	root, err := html.Parse(r)
	if err != nil {
		return err
	}

	sanitizeErr := &util.MultiErrors{}
	sanitizeErr.Add(s.cleanNode(root))
	sanitizeErr.Add(html.Render(w, root))

	return sanitizeErr.Err()
}

// Scan reports as error any HTML tags or attributes of given io.Reader that is
// seen as unsafe by the sanitizer.
func (s *HTML) Scan(r io.Reader) error {
	root, err := html.Parse(r)
	if err != nil {
		return err
	}

	return s.cleanNode(root)
}

func (s *HTML) cleanNode(n *html.Node) error {
	cleanErr := &util.MultiErrors{}

	switch n.Type {
	case html.DocumentNode:

	case html.TextNode:
		n.Data = html.EscapeString(n.Data)

	case html.ElementNode:
		if !s.isSafeTag(n) {
			cleanErr.Addf("Unsafe tag (%s) in '%s'.", n.DataAtom, pprint(n))
			if n.Parent != nil {
				n.Parent.RemoveChild(n)
			}
			return cleanErr.Err()
		}

		var err error
		n.Attr, err = s.onlySafeAttr(n)
		cleanErr.Add(err)

		cleanErr.Add(s.cleanURL(n))

	case html.CommentNode:
		//cleanErr.Addf("Comment '%s'.", pprint(n))
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
		return cleanErr.Err()

	case html.DoctypeNode:
		//cleanErr.Addf("Doctype '%s'.", pprint(n))
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
		return cleanErr.Err()

	default:
		cleanErr.Addf("Unsafe node of type %d in '%s'.", n.Type, pprint(n))
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
		return cleanErr.Err()

	}

	for c := n.FirstChild; c != nil; {
		// remember next, as s.cleanNode can delete c
		next := c.NextSibling
		cleanErr.Add(s.cleanNode(c))
		c = next
	}

	return cleanErr.Err()
}

func (s *HTML) isSafeTag(n *html.Node) bool {
	for t := range s.SafeElements {
		if n.DataAtom == t {
			return true
		}
	}
	return false
}

func (s *HTML) onlySafeAttr(n *html.Node) ([]html.Attribute, error) {
	sanitizedAttr, unsafeAttrErr := []html.Attribute{}, &util.MultiErrors{}

NextAttr:
	for _, attr := range n.Attr {
		if attr.Val == "" {
			continue
		}

		for _, safe := range s.SafeElements[n.DataAtom] {
			var k, v string
			switch parsed := strings.SplitN(safe, "=", 2); len(parsed) {
			case 1:
				k = parsed[0]
			case 2:
				k, v = parsed[0], parsed[1]
			}

			switch {
			case k == "*":
				sanitizedAttr = append(sanitizedAttr, attr)
				continue NextAttr
			case v == "" && attr.Key == k:
				sanitizedAttr = append(sanitizedAttr, attr)
				continue NextAttr
			case attr.Key == k && attr.Val == v:
				sanitizedAttr = append(sanitizedAttr, attr)
				continue NextAttr
			}
		}

		unsafeAttrErr.Addf("Unsafe attr (%s=%s) in '%s'", attr.Key, attr.Val, pprint(n))
	}

	return sanitizedAttr, unsafeAttrErr.Err()
}

func (s *HTML) isSafeScheme(scheme string) bool {
	for _, safe := range s.SafeSchemes {
		if scheme == safe {
			return true
		}
	}
	return false
}

func (s *HTML) cleanURL(n *html.Node) error {
	var isExternalLink, hasBlankTarget bool

	uncleanURLErr := &util.MultiErrors{}

	for i, a := range n.Attr {
		switch a.Key {
		case "href", "src":
			switch u, err := url.Parse(a.Val); {
			case err != nil:
				uncleanURLErr.Addf("Unparsable url '%s': %v", pprint(n), err)
				n.Attr[i].Val = "about:invalid"

			case u.Scheme != "" && !s.isSafeScheme(u.Scheme):
				uncleanURLErr.Addf("Unsafe URL scheme '%s' in '%s'", u.Scheme, pprint(n))
				n.Attr[i].Val = "about:invalid"

			default:
				n.Attr[i].Val = u.String()
				isExternalLink = (u.Host != "")
			}

		case "target":
			if !hasBlankTarget {
				hasBlankTarget = strings.Contains(a.Val, "_blank")
			}
		}
	}

	if isExternalLink && hasBlankTarget {
		for i, a := range n.Attr {
			if a.Key == "rel" {
				uncleanURLErr.Addf("Unsafe external URL link with target=_blank => rewrite %s", pprint(n))
				n.Attr[i].Val = "noopener noreferrer"
			}
		}
	}

	return uncleanURLErr.Err()
}

func pprint(n *html.Node) string {
	npp := &html.Node{
		Type:      n.Type,
		DataAtom:  n.DataAtom,
		Data:      n.Data,
		Namespace: n.Namespace,
		Attr:      n.Attr,
	}
	npp.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: "...",
	})

	b := &strings.Builder{}
	html.Render(b, npp)
	return b.String()
}
