// Package `quill` takes a Quill-based Delta (https://github.com/quilljs/delta) as a JSON array of `insert` operations
// and renders the defined HTML document.
//
// This library is designed to be easily extendable. Simply call RenderExtended with a function that may provide its
// own formats for certain kinds of ops and attributes.
package quill

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Render takes a Delta array of insert operations and returns the rendered HTML using the built-in settings.
// If an error occurs while rendering, any HTML already rendered is returned.
func Render(ops []byte) ([]byte, error) {
	return RenderExtended(ops, nil)
}

// RenderExtended takes a Delta array of insert operations and, optionally, a function that may provide a Formatter to
// customize the way certain kinds of inserts are rendered. If the given Formatter is nil, then the default one that is
// built in is used. If an error occurs while rendering, any HTML already rendered is returned.
func RenderExtended(ops []byte, customFormats func(string, *Op) Formatter) (html []byte, err error) {

	var raw []rawOp
	if err = json.Unmarshal(ops, &raw); err != nil {
		return nil, err
	}

	var (
		finalBuf = new(bytes.Buffer)     // the final output
		tempBuf  = new(bytes.Buffer)     // temporary buffer reused for each block element
		fs       = new(formatState)      // the tags currently open in the order in which they were opened
		o        = new(Op)               // allocate memory for an Op to reuse for all iterations
		fms      = make([]*Format, 0, 4) // the Formatter types defined for each Op
	)
	o.Attrs = make(map[string]string, 3) // initialize once here only

	for i := range raw {

		if err = raw[i].makeOp(o); err != nil {
			return finalBuf.Bytes(), err
		}

		fms = fms[:0] // Reset the slice for the current Op iteration.

		// To set up fms, first check the Op insert type.
		typeFmTer := o.getFormatter(o.Type, customFormats)
		if typeFmTer == nil {
			return finalBuf.Bytes(), fmt.Errorf("an op does not have a format defined for its type: %v", raw[i])
		}
		fms = o.addFmTer(fms, typeFmTer, fs, tempBuf)

		// Get a Formatter out of each of the attributes.
		for attr := range o.Attrs {
			fms = o.addFmTer(fms, o.getFormatter(attr, customFormats), fs, tempBuf)
		}

		// Open the a block element, write its body, and close it to move on only when the ending "\n" of the block is reached.
		if strings.IndexByte(o.Data, '\n') != -1 {

			if o.Data == "\n" { // Write a block element and flush the temporary buffer.

				// Avoid empty paragraphs and "\n" in the output.
				if tempBuf.Len() == 0 {
					o.Data = "<br>"
				} else {
					o.Data = ""
				}

				o.writeBlock(fs, tempBuf, finalBuf, fms)

			} else { // Extract the block-terminating line feeds and write each part as its own Op.

				split := strings.Split(o.Data, "\n")

				for i := range split {

					o.Data = split[i]

					// If the current o.Data still has an "\n" following (its not the last in split), then it ends a block.
					if i < len(split)-1 {

						// Avoid having empty paragraphs.
						if tempBuf.Len() == 0 && o.Data == "" {
							o.Data = "<br>"
						}

						o.writeBlock(fs, tempBuf, finalBuf, fms)

					} else if o.Data != "" { // If the last element in split is just "" then the last character in the rawOp is "\n".
						o.writeInline(fs, tempBuf, fms)
					}

				}

			}

		} else { // We are just adding stuff inline.
			o.writeInline(fs, tempBuf, fms)
		}

	}

	// Before writing out the final buffer, close the last remaining tags set by a FormatWrapper.
	// The FormatWrapper should see that all styling is now done.
	fs.closePrevious(finalBuf, blankOp(), true)

	html = finalBuf.Bytes()
	return

}

// addFmTer adds the format from fmTer to fms if the format is not already set on fs. All FormatWrapper formats are added regardless
// of whether they are already set on fs.
func (o *Op) addFmTer(fms []*Format, fmTer Formatter, fs *formatState, buf *bytes.Buffer) []*Format {
	if fmTer == nil {
		return fms
	}
	fm := fmTer.Fmt()
	if fm == nil {
		// Check if the format is a FormatWriter. If it is, just write it out and continue.
		if wr, ok := fmTer.(FormatWriter); ok {
			wr.Write(buf)
			o.Data = ""
		}
		return fms
	}
	fm.fm = fmTer
	if fw, ok := fmTer.(FormatWrapper); ok {
		fm.wrap = true
		fm.wrapPre, fm.wrapPost = fw.Wrap()
		return append(fms, fm)
	} else if !fs.hasSet(fmTer.Fmt()) {
		return append(fms, fm)
	}
	return fms
}

// An Op is a Delta insert operations (https://github.com/quilljs/delta#insert) that has been converted into this format for
// usability with the type safety in Go.
type Op struct {
	Data  string            // the text to insert or the value of the embed object (http://quilljs.com/docs/delta/#embeds)
	Type  string            // the type of the op (typically "text", but any other type can be registered)
	Attrs map[string]string // key is attribute name; value is either the attribute value or "y" (meaning true)
}

// writeBlock writes a block element (which may be nested inside another block element if it is a FormatWrapper).
// The opening HTML tag of a block element is written to the main buffer only after the "\n" character terminating the
// block is reached (the Op with the "\n" character holds the information about the block element).
func (o *Op) writeBlock(fs *formatState, tempBuf *bytes.Buffer, finalBuf *bytes.Buffer, newFms []*Format) {

	// Close the inline formats opened within the block to the tempBuf and block formats of wrappers to finalBuf.
	//wrapFs := &formatState{make([]*Format, 0, 2)}
	//inlineFs := &formatState{make([]*Format, 0, 2)}
	//for i := range fs.open {
	//	if fs.open[i].wrap {
	//		wrapFs.add(fs.open[i])
	//	} else {
	//		inlineFs.add(fs.open[i])
	//	}
	//}
	//wrapFs.closePrevious(finalBuf, o, true)
	//inlineFs.closePrevious(tempBuf, o, true)
	//fs.open = append(wrapFs.open, inlineFs.open...) // There may be something left open.
	//fs.closePrevious(tempBuf, o, true)
	closedTemp := &formatState{}

	for i := len(fs.open) - 1; i >= 0; i-- { // Start with the last format opened.

		f := fs.open[i]

		// If this format is not set on the current Op, close it.
		if (!f.wrap && !f.fm.HasFormat(o)) || (f.wrap && f.fm.(FormatWrapper).Close(fs.open, o, true)) {

			// If we need to close a tag after which there are tags that should stay open, close the following tags for now.
			if i < len(fs.open)-1 {
				for ij := len(fs.open) - 1; ij > i; ij-- {
					closedTemp.add(fs.open[ij])
					if f.wrap && f.Block {
						fs.pop(finalBuf)
					} else {
						fs.pop(tempBuf)
					}
				}
			}

			if f.wrap && f.Block {
				fs.pop(finalBuf)
			} else {
				fs.pop(tempBuf)
			}

		}

	}

	// Re-open the temporarily closed formats.
	closedTemp.writeFormats(tempBuf)
	fs.open = append(fs.open, closedTemp.open...) // Copy after the sorting.

	var block struct {
		tagName string
		classes []string
		style   string
	}

	// At least a format from the Op.Type should be set.
	if len(newFms) == 0 {
		return
	}

	// Merge all formats into a single tag.
	for i := range newFms {
		fm := newFms[i]
		// Apply only block-level formats.
		if fm.Block {
			v := fm.Val
			switch fm.Place {
			case Tag:
				// If an opening tag is not specified by the Op insert type, it may be specified by an attribute.
				if v != "" {
					block.tagName = v // Override whatever value is set.
				}
			case Class:
				block.classes = append(block.classes, v)
			case Style:
				block.style += v
			}
		}
		// Write out all of FormatWrapper opening text (if there is any).
		if fm.wrap && fm.fm.(FormatWrapper).Open(fs.open, o) {
			fm.Val = fm.wrapPre
			fs.add(fm)
			finalBuf.WriteString(fm.Val)
		}
	}

	if block.tagName != "" {
		finalBuf.WriteByte('<')
		finalBuf.WriteString(block.tagName)
		finalBuf.WriteString(classesList(block.classes))
		if block.style != "" {
			finalBuf.WriteString(" style=")
			finalBuf.WriteString(strconv.Quote(block.style))
		}
		finalBuf.WriteByte('>')
	}

	finalBuf.Write(tempBuf.Bytes()) // Copy the temporary buffer to the final output.

	finalBuf.WriteString(o.Data) // Copy the data of the current Op (usually just "<br>" or blank).

	if block.tagName != "" {
		closeTag(finalBuf, block.tagName)
	}

	// Write out the closes by FormatWrapper formats, starting from the last written.
	//fs.closePrevious(finalBuf, o, true)

	tempBuf.Reset()

}

func (o *Op) writeInline(fs *formatState, buf *bytes.Buffer, newFms []*Format) {

	fs.closePrevious(buf, o, false)

	// Save the formats being written now separately from fs.
	addNow := &formatState{make([]*Format, 0, len(newFms))}

	for _, f := range newFms {
		// Apply only inline formats.
		if !f.Block {
			if f.wrap {
				// Add FormatWrapper formats only if they need to be written now.
				if f.fm.(FormatWrapper).Open(fs.open, o) {
					f.Val = f.wrapPre
					addNow.add(f)
				}
			} else {
				addNow.add(f)
			}
		}
	}

	addNow.writeFormats(buf)
	fs.open = append(fs.open, addNow.open...) // Copy after the sorting.

	buf.WriteString(o.Data)

}

// HasAttr says if the Op is not nil and has the attribute set to a non-blank value.
func (o *Op) HasAttr(attr string) bool {
	return o != nil && o.Attrs[attr] != ""
}

// getFormatter returns a formatter based on the keyword (either "text" or "" or an attribute name) and the Op settings.
// For every Op, first its Type is passed through here as the keyword, and then its attributes.
func (o *Op) getFormatter(keyword string, customFormats func(string, *Op) Formatter) Formatter {

	if customFormats != nil {
		if custom := customFormats(keyword, o); custom != nil {
			return custom
		}
	}

	switch keyword { // This is the list of currently recognized "keywords".
	case "text":
		return new(textFormat)
	case "header":
		return &headerFormat{
			level: o.Attrs["header"],
		}
	case "list":
		lf := &listFormat{
			indent: indentDepths[o.Attrs["indent"]],
		}
		if o.Attrs["list"] == "bullet" {
			lf.lType = "ul"
		} else {
			lf.lType = "ol"
		}
		return lf
	case "blockquote":
		return new(blockQuoteFormat)
	case "align":
		return &alignFormat{
			val: o.Attrs["align"],
		}
	case "image":
		return &imageFormat{
			src: o.Data,
		}
	case "link":
		return &linkFormat{
			href: o.Attrs["link"],
		}
	case "bold":
		return new(boldFormat)
	case "italic":
		return new(italicFormat)
	case "underline":
		return new(underlineFormat)
	case "color":
		return &colorFormat{
			c: o.Attrs["color"],
		}
	case "indent":
		return &indentFormat{
			in: o.Attrs["indent"],
		}
	}

	return nil

}

// A FormatPlace is either an HTML tag name, a CSS class, or a style attribute value.
type FormatPlace uint8

const (
	Tag FormatPlace = iota
	Class
	Style
)

// A Formatter is able to give a Format and say whether a given Op should have that Format applied.
type Formatter interface {
	Fmt() *Format       // Format gives the string to write and where to place it.
	HasFormat(*Op) bool // Say if the Op has the Format that Fmt returns.
}

// A Formatter may also be a FormatWriter if it wishes to write the body of the Op in some custom way (useful for embeds).
type FormatWriter interface {
	Formatter
	Write(io.Writer) // Write the entire body of the element.
}

// A FormatWrapper wraps text with additional text of any kind (such as "<ul>" for lists).
type FormatWrapper interface {
	Formatter
	Wrap() (pre, post string)        // Say what opening and closing wraps will be written.
	Open([]*Format, *Op) bool        // Given the open formats and current Op, say if to write the pre string.
	Close([]*Format, *Op, bool) bool // Given the open formats, current Op, and if the Op closes a block, say if to write the post string.
}

// A Format specifies how styling to text is applied. The Val string is what is printed in the place given by Place. Block indicates
// if this is a block-level format.
type Format struct {
	Val               string      // the value to print
	Place             FormatPlace // where this format is placed in the text
	Block             bool        // indicate whether this is a block-level format (not printed until a "\n" is reached)
	wrap              bool        // indicates whether this format was written as a FormatWrapper
	wrapPre, wrapPost string      // If this Format is a wrap, then Val holds the open and wrapPost holds the close.
	fm                Formatter   // where this instance of a Format came from
}

// A blankOp can be used to signal any FormatWrapper formats to write the final closing wrap.
func blankOp() *Op {
	return &Op{"", "text", make(map[string]string)}
}

// If cl has something, then classesList returns the class attribute to add to an HTML element with a space before the
// "class" attribute and spaces between each class name.
func classesList(cl []string) (classAttr string) {
	if len(cl) > 0 {
		classAttr = " class=" + strconv.Quote(strings.Join(cl, " "))
	}
	return
}

// closeTag writes a complete closing tag to buf.
func closeTag(buf *bytes.Buffer, tagName string) {
	buf.WriteString("</")
	buf.WriteString(tagName)
	buf.WriteByte('>')
}
