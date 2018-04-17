package quill

import (
	"bytes"
	"sort"
	"strconv"
)

// A formatState holds the current state of open tag, class, or style formats.
type formatState []*Format // the list of currently open attribute tags

// hasSet says if the given format is already opened.
func (fs *formatState) hasSet(fm *Format) bool {
	for i := range *fs {
		if (*fs)[i].Place == fm.Place && (*fs)[i].Val == fm.Val {
			return true
		}
	}
	return false
}

// closePrevious checks if the previous ops opened any formats that are not set on the current Op and closes those formats
// in the opposite order in which they were opened.
func (fs *formatState) closePrevious(buf *bytes.Buffer, o *Op, doingBlock bool) {

	closedTemp := make(formatState, 0, 1)

	for i := len(*fs) - 1; i >= 0; i-- { // Start with the last format opened.

		f := (*fs)[i]

		// If this format is not set on the current Op, close it.
		if (!f.wrap && !f.fm.HasFormat(o)) || (f.wrap && f.fm.(FormatWrapper).Close(*fs, o, doingBlock)) {

			// If we need to close a tag after which there are tags that should stay open, close the following tags for now.
			if i < len(*fs)-1 {
				for ij := len(*fs) - 1; ij > i; ij-- {
					closedTemp.add((*fs)[ij])
					fs.pop(buf)
				}
			}

			fs.pop(buf)

		}

	}

	// Re-open the temporarily closed formats.
	closedTemp.writeFormats(buf)
	*fs = append(*fs, closedTemp...) // Copy after the sorting.

}

// pop removes the last format from the state of currently open formats.
func (fs *formatState) pop(buf *bytes.Buffer) {
	indx := len(*fs) - 1
	if (*fs)[indx].wrap {
		buf.WriteString((*fs)[indx].wrapPost)
	} else if (*fs)[indx].Place == Tag {
		closeTag(buf, (*fs)[indx].Val)
	} else {
		closeTag(buf, "span")
	}
	*fs = (*fs)[:indx]
}

// add adds a format that the string that will be written to buf right after this will have.
// Before calling add, check if the Format is already opened up earlier.
// Do not use add to write block-level styles (those are written by o.writeBlock after being merged).
func (fs *formatState) add(f *Format) {
	if f.Place < 3 { // Check if the Place is valid.
		*fs = append(*fs, f)
	}
}

// writeFormats sorts the formats in the current formatState and writes them all out to buf. If a format implements
// the FormatWrapper interface, that format's opening wrap is printed.
func (fs *formatState) writeFormats(buf *bytes.Buffer) {

	sort.Sort(fs) // Ensure that the serialization is consistent even if attribute ordering in a map changes.

	for _, f := range *fs {

		if f.wrap {
			buf.WriteString(f.Val) // The complete opening or closing wrap is given.
			continue
		}

		buf.WriteByte('<')

		switch f.Place {
		case Tag:
			buf.WriteString(f.Val)
		case Class:
			buf.WriteString("span class=")
			buf.WriteString(strconv.Quote(f.Val))
		case Style:
			buf.WriteString("span style=")
			buf.WriteString(strconv.Quote(f.Val))
		}

		buf.WriteByte('>')

	}

}

// Implement the sort.Interface interface.

func (fs *formatState) Len() int { return len(*fs) }

func (fs *formatState) Less(i, j int) bool {

	fsi, fsj := (*fs)[i], (*fs)[j]

	// Formats that implement the FormatWrapper interface are written first.
	if _, ok := fsi.fm.(FormatWrapper); ok {
		return true
	} else if _, ok := fsj.fm.(FormatWrapper); ok {
		return false
	}

	// Tags are written first, then classes, and then style attributes.
	if fsi.Place != fsj.Place {
		return fsi.Place < fsj.Place
	}

	// Simply check values.
	return fsi.Val < fsj.Val

}

func (fs *formatState) Swap(i, j int) {
	(*fs)[i], (*fs)[j] = (*fs)[j], (*fs)[i]
}
