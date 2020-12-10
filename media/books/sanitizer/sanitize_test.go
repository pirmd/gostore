package sanitizer

import (
	"bytes"
	"strings"

	"testing"
)

func TestSanitize(t *testing.T) {
	testCases := []struct {
		in, out string
	}{
		// Scripts
		{
			`<script src="evil.js"/>`,
			``,
		},
		{
			`a<br>b`,
			`a<br/>b`,
		},
		// Lists
		{
			`<ul foo="bar"> <li x="y">a</li> <li>a</li> </ul>`,
			`<ul> <li>a</li> <li>a</li> </ul>`,
		},
		// Links
		{
			`<a href="https://openbsd.org">link to puffy</a>`,
			`<a href="https://openbsd.org">link to puffy</a>`,
		},
		{
			`<a href="javascript:evil.js">hello</a>`,
			`<a href="about:invalid">hello</a>`,
		},
		{
			`<a href="about:blank">hello</a>`,
			`<a href="about:invalid">hello</a>`,
		},
		{
			`<a href="%">hello</a>`,
			`<a href="about:invalid">hello</a>`,
		},
		// Other
		{
			`<div><strong>hello</strong></div>`,
			`<div><strong>hello</strong></div>`,
		},
		{
			`&lt;`,
			`&amp;lt;`,
		},
		{
			`<div><p>foo</p>`,
			`<div><p>foo</p></div>`,
		},
		{
			`<p></a alt="blah"></p>`,
			`<p></p>`,
		},
	}

	for _, tc := range testCases {
		got := &bytes.Buffer{}
		_ = EPUB.Sanitize(got, strings.NewReader(tc.in))

		// TODO: manage Errors at this point to check that the error content is
		// consistent with the filtering cause
		// t.Logf("Sanitize output %v", err)

		want := "<html><head></head><body>" + tc.out + "</body></html>"
		if got.String() != want {
			t.Errorf("Sanitize failed of '%s'.\n Got : %s\n Want: %s\n", tc.in, got, want)
		}
	}
}
