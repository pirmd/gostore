package dehtmlizer

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/net/html"

	"github.com/pirmd/cli/style"
	"github.com/pirmd/verify"
)

const (
	testdataPath = "./testdata"
)

func TestRender(t *testing.T) {
	testCases := []struct {
		in   string
		want string
	}{
		{
			in:   "<p>Hello Gophers!</p>",
			want: "Hello Gophers!\n",
		},
		{
			in: `
			<div>
				<p>Hello Gophers!</p>
				<p>Golang is nice</p>
			</div>`,
			want: "Hello Gophers!\n\nGolang is nice\n",
		},
		{
			in:   "<p>Hello <span>Gophers</span>!</p>",
			want: "Hello Gophers!\n",
		},
		{
			in:   `Hello <b>Gophers</b>!`,
			want: `Hello __Gophers__!`,
		},
		{
			in:   `Hello <i>Gophers</i>!`,
			want: `Hello *Gophers*!`,
		},
		{
			in:   `Hello <del>Gophers</del>!`,
			want: `Hello ~~Gophers~~!`,
		},
		{
			in:   `<h1>Hello <i>Gophers</i>!</h1>`,
			want: "# HELLO *GOPHERS*!\n",
		},
		{
			in:   `Hello <b>Gophers</b>!`,
			want: `Hello __Gophers__!`,
		},
		{
			in:   `Hello  <b> Gophers</b>!`,
			want: `Hello __Gophers__!`,
		},
		{
			in:   `  Hello  <b> Gophers</b>!`,
			want: `Hello __Gophers__!`,
		},
		{
			in:   `    <b>Hello Gophers</b>!`,
			want: `__Hello Gophers__!`,
		},
		{
			in:   "  Hello\n  <b> Gophers</b>!",
			want: `Hello __Gophers__!`,
		},
		{
			in:   `<a href="http://interesting.com/">Link</a>`,
			want: `[Link](http://interesting.com/)`,
		},
		{
			in:   `<a onclick="alert(42)">Link</a>`,
			want: `[Link]()`,
		},
		{
			in: `
            <ul>
                <li>todo</li>
                <li>really need todo</li>
            </ul>`,
			want: `* todo

* really need todo
`,
		},
		{
			in: `
            <ol>
                <li>first thing</li>
                <li>second thing</li>
            </ol>`,
			want: `1. first thing

2. second thing
`,
		},
		{
			in: `
            <ul>
            <li>item1
                <ol>
                    <li>item1.1</li>
                    <li>item1.2</li>
                    <li>item1.3</li>
                    <li>item1.4</li>
                 </ol>
            </li>
            <li>item2
                 <ul>
                    <li>item2.1</li>
                    <li>item2.2</li>
                 </ul>
            </li>
            </ul>`,
			want: `* item1 

  1. item1.1

  2. item1.2

  3. item1.3

  4. item1.4

* item2 

  + item2.1

  + item2.2
`,
		},
		{
			in: `
            <table>
            <tr>
                <th>Col1</th>
                <th>Col2</th>
            </tr>
            <tr>
                <td>Col1.1</td>
                <td>Col2.1</td>
            </tr>
            <tr>
                <td>Col1.2</td>
                <td>Col2.2</td>
            </tr>
            </table>`,
			want: `------ ------
Col1   Col2  
------ ------
Col1.1 Col2.1
             
Col1.2 Col2.2
------ ------
`,
		},
	}

	for _, tc := range testCases {
		got, err := html2txt(tc.in, style.NewMarkdown())
		if err != nil {
			t.Errorf("Fail to render %s to Markdown: %v", tc.in, err)
		}

		verify.EqualString(t, got, tc.want, "Fail to render %s to Markdown.", tc.in)
	}
}

func TestRenderFullHtml(t *testing.T) {
	testCases, err := filepath.Glob(filepath.Join(testdataPath, "*.html"))
	if err != nil {
		t.Fatalf("cannot read test data in %s:%v", testdataPath, err)
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			r, err := os.Open(tc)
			if err != nil {
				t.Errorf("cannot open test file '%s': %s", tc, err)
			}

			root, err := html.Parse(r)
			if err != nil {
				t.Errorf("Fail to parse input %s: %v", tc, err)
			}

			got := renderNode(root, style.NewMarkdown())
			verify.MatchGolden(t, got, "Generated Markdown is not as expected.")
		})
	}
}
