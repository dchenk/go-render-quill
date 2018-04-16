package quill

import (
	"bytes"
	"sort"
	"testing"
)

func TestFormatState_add(t *testing.T) {

	cases := []struct {
		current []*Format
		keyword string
		o       *Op
		want    []*Format
	}{
		{
			current: []*Format{}, // no formats
			keyword: "italic",
			o: &Op{
				Data:  "stuff",
				Type:  "text",
				Attrs: map[string]string{"italic": "y"},
			},
			want: []*Format{
				{
					Val:   "em",
					Place: Tag,
				},
			},
		},
		{
			current: []*Format{
				{ // One format already set.
					Val:   "em",
					Place: Tag,
				},
			},
			keyword: "italic",
			o: &Op{
				Data:  "stuff",
				Type:  "text",
				Attrs: map[string]string{"italic": "y"},
			},
			want: []*Format{ // The way add works, it does not check if the format is already added.
				{
					Val:   "em",
					Place: Tag,
				},
				{
					Val:   "em",
					Place: Tag,
				},
			},
		},
	}

	fs := make(formatState, 0, 2) // reuse

	for i, ca := range cases {

		fs = ca.current

		fmTer := ca.o.getFormatter(ca.keyword, nil)
		fm := fmTer.Fmt()
		fm.fm = fmTer

		fs.add(fm)

		if len(ca.want) != len(fs) {
			t.Fatalf("(index %d) unequal count of formats", i)
		}

		for j := range fs {
			if ca.want[j].Val != fs[j].Val {
				t.Errorf("did not add format Val correctly (index %d); got %q", i, fs[j].Val)
			}
			if ca.want[j].Place != fs[j].Place {
				t.Errorf("did not add format Place correctly (index %d); got %v", i, fs[j].Place)
			}
			if ca.want[j].Block != fs[j].Block {
				t.Errorf("did not add format Block correctly (index %d); got %v", i, fs[j].Block)
			}
		}

	}

}

func TestFormatState_closePrevious(t *testing.T) {

	o1 := blankOp()
	o1.Attrs["italic"] = "y"
	o1.Attrs["bold"] = "y"

	o2 := blankOp()
	o2.Attrs["background"] = "#e0e0e0"
	o2.Attrs["italic"] = "y"

	cases := []formatState{
		{
			{"em", Tag, false, false, "", "", o1.getFormatter("italic", nil)},
			{"strong", Tag, false, false, "", "", o1.getFormatter("bold", nil)},
		},
		{
			{"background-color:#e0e0e0;", Style, false, false, "", "", o2.getFormatter("background", nil)},
			{"em", Tag, false, false, "", "", o2.getFormatter("italic", nil)},
		},
	}

	want := []string{"</strong></em>", "</em></span>"}

	var buf bytes.Buffer

	for i := range cases {

		o := blankOp()

		cases[i].closePrevious(&buf, o, false)
		got := buf.String()
		if got != want[i] || len(cases[i]) != 0 {
			t.Errorf("closed formats wrong (index %d); wanted %q; got %q", i, want[i], got)
		}

		buf.Reset()

	}

}

func TestFormatState_Sort(t *testing.T) {

	o := &Op{
		Data: "stuff",
		Type: "text",
		// no attributes set
	}

	cases := [][]struct {
		Val     string
		Place   FormatPlace
		keyword string
	}{
		{
			{"strong", Tag, "bold"},
			{"em", Tag, "italic"},
		},
		{
			{"u", Tag, "underline"},
			{"align-center", Class, "align"},
			{"strong", Tag, "bold"},
		},
		{
			{"color:#e0e0e0;", Style, "color"},
			{"em", Tag, "italic"},
		},
		{
			{"em", Tag, "italic"},
			{`<a href="https://widerwebs.com" target="_blank">`, Tag, "link"}, // link wrapper
		},
	}

	want := [][]struct {
		Val     string
		Place   FormatPlace
		keyword string
	}{
		{
			{"em", Tag, "italic"},
			{"strong", Tag, "bold"},
		},
		{
			{"strong", Tag, "bold"},
			{"u", Tag, "underline"},
			{"align-center", Class, "align"},
		},
		{
			{"em", Tag, "italic"},
			{"color:#e0e0e0;", Style, "color"},
		},
		{
			{`<a href="https://widerwebs.com" target="_blank">`, Tag, "link"}, // link wrapper
			{"em", Tag, "italic"},
		},
	}

	for i := range cases {

		fsCase := make(formatState, 0, 1)
		for _, s := range cases[i] {
			fsCase = append(fsCase, &Format{
				Val:   s.Val,
				Place: s.Place,
				fm:    o.getFormatter(s.keyword, nil),
			})
		}

		sort.Sort(&fsCase)

		fsWant := make(formatState, 0, 1)
		for _, s := range want[i] {
			fsWant = append(fsWant, &Format{
				Val:   s.Val,
				Place: s.Place,
				fm:    o.getFormatter(s.keyword, nil),
			})
		}

		ok := true
		for j := range fsCase {
			if fsCase[j].Val != fsWant[j].Val {
				ok = false
			} else if fsCase[j].Place != fsWant[j].Place {
				ok = false
			}
		}
		if !ok {
			t.Errorf("bad sorting (index %d); got:\n", i)
			for k := range fsCase {
				t.Errorf("  (%d) %+v\n", k, fsCase[k])
			}
		}

	}

}
