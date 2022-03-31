package quill

import (
	"bytes"
	"io/ioutil"
	"strconv"
	"testing"
)

func TestSimple(t *testing.T) {

	cases := map[string]struct {
		ops  string
		want string
	}{
		"empty": {
			ops:  `[{"insert": "\n"}]`,
			want: "<p><br/></p>",
		},
		"two paragraphs (single op)": {
			ops:  `[{"insert": "line1\nline2\n"}]`,
			want: "<p>line1</p><p>line2</p>",
		},
		"blank line": {
			ops:  `[{"insert": "line1\n\nline3\n"}]`,
			want: "<p>line1</p><p><br/></p><p>line3</p>",
		},
		"blockquote": {
			ops:  `[{"insert": "bkqt"}, {"attributes": {"blockquote": true}, "insert": "\n"}]`,
			want: "<blockquote>bkqt</blockquote>",
		},
		"color": {
			ops:  `[{"attributes": {"color": "#a10000"}, "insert": "colored"}, {"insert": "\n"}]`,
			want: `<p><span style="color:#a10000;">colored</span></p>`,
		},
		"strikethrough": {
			ops:  `[{"attributes":{"strike":true},"insert":"striked"},{"insert":"\n"}]`,
			want: "<p><s>striked</s></p>",
		},
		"list": {
			ops:  `[{"insert":"abc "},{"attributes":{"bold":true},"insert":"bld"},{"attributes":{"list":"bullet"},"insert":"\n"}]`,
			want: "<ul><li>abc <strong>bld</strong></li></ul>",
		},
		"image": {
			ops:  `[{"insert":{"image":"source-url"}},{"insert":"\n"}]`,
			want: `<p><img src="source-url"></p>`,
		},
		"image wrapped": {
			ops:  `[{"insert":"text "},{"insert":{"image":"source-url"}},{"insert":" more text\n"}]`,
			want: `<p>text <img src="source-url"> more text</p>`,
		},
		"background": {
			ops:  `[{"insert":"abc "},{"attributes":{"background":"#66a3e0"},"insert":"bkg colored"},{"insert":" plain\n"}]`,
			want: `<p>abc <span style="background-color:#66a3e0;">bkg colored</span> plain</p>`,
		},
		"underlined": {
			ops:  `[{"attributes":{"underline":true},"insert":"underlined"},{"insert":"\n"}]`,
			want: "<p><u>underlined</u></p>",
		},
		"size": {
			ops: `[{"insert":"stuff "},
				{"insert":"large","attributes":{"size":"large"}},{"insert":" other "},
				{"insert":"small","attributes":{"size":"small"}},{"insert":"\n"}]`,
			want: `<p>stuff <span class="ql-size-large">large</span> other <span class="ql-size-small">small</span></p>`,
		},
		"superscript": {
			ops:  `[{"insert":"plain"},{"attributes":{"script":"super"},"insert":"super"},{"insert":"\n"}]`,
			want: "<p>plain<sup>super</sup></p>",
		},
		"subscript": {
			ops:  `[{"insert":"plain"},{"attributes":{"script":"sub"},"insert":"sub"},{"insert":"\n"}]`,
			want: "<p>plain<sub>sub</sub></p>",
		},
	}

	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			got, err := Render([]byte(tc.ops))
			if err != nil {
				t.Fatalf("%s", err)
			}
			if string(got) != tc.want {
				t.Errorf("bad rendering; got: %s", got)
			}
		})
	}

}

func TestRender(t *testing.T) {

	pairNames := []string{"ops1", "nested", "ordering", "list1", "list2", "list3", "list4", "indent", "code1", "code2", "code3"}

	for _, n := range pairNames {
		t.Run(n, func(t *testing.T) {

			ops, err := ioutil.ReadFile("./testdata/" + n + ".json")
			if err != nil {
				t.Fatalf("could not read %s.json; %s", n, err)
			}

			html, err := ioutil.ReadFile("./testdata/" + n + ".html")
			if err != nil {
				t.Fatalf("could not read %s.html; %s", n, err)
			}

			got, err := Render(ops)
			if err != nil {
				t.Errorf("error rendering; %v", err)
			}

			if !bytes.Equal(html, got) {
				t.Errorf("bad rendering:\nwanted: \n%s\ngot: \n%s", html, got)
			}

		})
	}

}

func TestClassesList(t *testing.T) {
	cases := []struct {
		classes []string
		expect  string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"abc"}, ` class="abc"`},
		{[]string{"abc", "ee-abcd"}, ` class="abc ee-abcd"`},
	}
	for i, tc := range cases {
		t.Run("case_"+strconv.Itoa(i), func(t *testing.T) {
			got := classesList(tc.classes)
			if got != tc.expect {
				t.Errorf("expected %q but got %q", tc.expect, got)
			}
		})
	}
}

func BenchmarkRender_ops1(b *testing.B) {
	bts, err := ioutil.ReadFile("./testdata/ops1.json")
	if err != nil {
		b.Fatalf("could not read ops file: %s", err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		bts, err := Render(bts)
		if err != nil {
			b.Errorf("error rendering: %s", err)
		}
		_ = bts
	}
}
