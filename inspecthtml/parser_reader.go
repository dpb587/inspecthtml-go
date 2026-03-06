package inspecthtml

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/dpb587/cursorio-go/cursorio"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var reTagName = regexp.MustCompile(`^<([^\s/>]+)`)
var reAttrKeyValue = regexp.MustCompile(`.*?[\s<>]*([^=\s/<>]+)((\s*=\s*)(.))?`)
var reAttrValueDoubleQuote = regexp.MustCompile(`.*?"`)
var reAttrValueSingleQuote = regexp.MustCompile(`.*?'`)
var reAttrValueUnquoted = regexp.MustCompile(`[^\s>]+`)

type parserNodeSwap struct {
	original    string
	offsetRange cursorio.TextOffsetRange
}

type parserReader struct {
	tokenizer *html.Tokenizer
	doc       *cursorio.TextWriter

	err  error
	buf  []byte
	bufi int

	nodeIdx                int64
	nodeRawTextMode        bool
	nodeTagByKey           map[string]*NodeMetadata
	nodeSwapByKey          map[string]parserNodeSwap
	endTagOffsetRangeByKey map[string]cursorio.TextOffsetRange
	wsOffsetRangeByKey     map[string]cursorio.TextOffsetRange
}

func (r *parserReader) Read(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	} else if r.bufi >= len(r.buf) {
		r.err = r.next()
		if r.err != nil {
			return 0, r.err
		}
	}

	var l = copy(p, r.buf[r.bufi:])

	r.bufi += l

	return l, nil
}

func (r *parserReader) next() error {
	r.bufi = 0

	tt := r.tokenizer.Next()
	if tt == html.ErrorToken {
		return r.tokenizer.Err()
	}

	// copy and avoid append reusing tokenizer's byte slice
	raw := []byte(string(r.tokenizer.Raw()))

	switch tt {
	case html.SelfClosingTagToken, html.StartTagToken:
		rawCutset := raw

		docOffset := r.doc.GetTextOffset()
		tagProfile := &NodeMetadata{
			TokenOffsets: cursorio.TextOffsetRange{
				From: docOffset,
			},
			TagSelfClosing: tt == html.SelfClosingTagToken,
		}

		tagNameMatcher := reTagName.FindSubmatchIndex(rawCutset)
		if tagNameMatcher != nil {
			r.doc.Write(rawCutset[:tagNameMatcher[2]])

			tagNameOffsets := r.doc.WriteForOffsetRange(rawCutset[tagNameMatcher[2]:tagNameMatcher[3]])
			tagProfile.TagNameOffsets = &tagNameOffsets

			rawCutset = rawCutset[tagNameMatcher[3]:]
		}

		_, hasAttr := r.tokenizer.TagName()
		for hasAttr {
			attrKey, attrValue, more := r.tokenizer.TagAttr()

			rawAttrMatcher := reAttrKeyValue.FindSubmatchIndex(rawCutset)

			if rawAttrMatcher == nil {
				// <script async src="https://example.com/asset?shop="quoteful.example.com"></script>

				// ignore
				// risky to not advance cursor; possible early regex match for next attribute?
				tagProfile.TagAttr = append(tagProfile.TagAttr, nil)
			} else {
				r.doc.Write(rawCutset[:rawAttrMatcher[2]])

				tagAttrProfile := NodeAttributeMetadata{
					KeyOffsets: r.doc.WriteForOffsetRange(rawCutset[rawAttrMatcher[2]:rawAttrMatcher[3]]),
				}

				if rawAttrMatcher[4] > -1 {
					r.doc.Write(rawCutset[rawAttrMatcher[6]:rawAttrMatcher[7]])
					rawCutset = rawCutset[rawAttrMatcher[8]:]

					var consumeLen int

					if rawCutset[0] == '"' {
						closeMatcher := reAttrValueDoubleQuote.FindSubmatchIndex(rawCutset[1:])

						if closeMatcher == nil {
							// weird
						} else {
							consumeLen = closeMatcher[1] + 1
						}
					} else if rawCutset[0] == '\'' {
						closeMatcher := reAttrValueSingleQuote.FindSubmatchIndex(rawCutset[1:])

						if closeMatcher == nil {
							// weird
						} else {
							consumeLen = closeMatcher[1] + 1
						}
					} else if !unicode.IsSpace(rune(rawCutset[0])) && rawCutset[0] != '>' {
						closeMatcher := reAttrValueUnquoted.FindSubmatchIndex(rawCutset)

						if closeMatcher == nil {
							// weird
						} else {
							consumeLen = closeMatcher[1]
						}
					}

					if consumeLen > 0 {
						valueOffsetRange := r.doc.WriteForOffsetRange(rawCutset[:consumeLen])
						tagAttrProfile.ValueOffsets = &valueOffsetRange

						rawCutset = rawCutset[consumeLen:]
					}
				} else if len(attrValue) > 0 {
					// an edge case worth fixing; almost panic-worthy; subsequent attributes may no longer be correct
					fmt.Fprintf(os.Stderr, "inspecthtml: regex attr failed (raw=%q, key=%q, val=%q)\n", string(rawCutset), string(attrKey), string(attrValue))
				} else {
					rawCutset = rawCutset[rawAttrMatcher[3]:]
				}

				tagProfile.TagAttr = append(tagProfile.TagAttr, &tagAttrProfile)
			}

			hasAttr = more
		}

		r.doc.Write(rawCutset)

		tagProfile.TokenOffsets.Until = r.doc.GetTextOffset()

		r.nodeIdx++
		nodeKey := strconv.FormatInt(r.nodeIdx, 10)

		r.nodeTagByKey[nodeKey] = tagProfile

		if tagNameMatcher != nil {
			switch atom.Lookup(bytes.ToLower(raw[tagNameMatcher[2]:tagNameMatcher[3]])) {
			// https://html.spec.whatwg.org/multipage/parsing.html#parsing-html-fragments
			case atom.Script, atom.Style, atom.Textarea, atom.Title, atom.Plaintext, atom.Iframe, atom.Xmp, atom.Noembed, atom.Noframes, atom.Noscript:
				r.nodeRawTextMode = true
			}
		}

		// always first attribute to avoid mangling that may happen upstream for malformed user input
		// including trailing space to avoid any accidental overlap with malformed tags (e.g., `<body</div>`)
		oAttr := fmt.Appendf(nil, " o=%q ", nodeKey)
		if tagNameMatcher != nil {
			insertAt := tagNameMatcher[3]
			r.buf = append(raw[:insertAt:insertAt], oAttr...)
			r.buf = append(r.buf, raw[insertAt:]...)
		} else {
			r.buf = raw
		}
	case html.EndTagToken:
		r.nodeIdx++
		nodeKey := strconv.FormatInt(r.nodeIdx, 10)

		r.endTagOffsetRangeByKey[nodeKey] = r.doc.WriteForOffsetRange(raw)
		r.buf = append(raw, []byte("<!--e"+nodeKey+"-->")...)

		r.nodeRawTextMode = false
	case html.CommentToken:
		r.nodeIdx++
		nodeKey := strconv.FormatInt(r.nodeIdx, 10)

		var commentContent string

		if !bytes.HasPrefix(raw, []byte("<!--")) {
			// Bogus comment token (e.g. <?php...?> or <!tag...>): the tokenizer
			// already computes the correct comment data, so use it directly.
			commentContent = r.tokenizer.Token().Data
		} else {
			if len(raw) >= 7 {
				// <!--content--> or <!--content--!>
				// HTML5 allows --!> as a comment terminator (comment end bang state).
				// Strip 4 chars for --!> endings vs the usual 3 for -->.
				if bytes.HasSuffix(raw, []byte("--!>")) {
					commentContent = string(raw)[4 : len(raw)-4]
				} else {
					commentContent = string(raw)[4 : len(raw)-3]
				}
			} else if len(raw) > 4 {
				// malformed; but tokenizer recovered; e.g. <!--content or <!-->
				commentContent = string(raw)[4:]
			}

			// Normalize line endings per the HTML spec input stream preprocessing
			// (\r\n → \n, \r → \n). Token().Data does this but for non-bogus comments
			// we use raw slicing (since Token().Data returns incorrect data for malformed
			// short comments like <!--> yielding "" instead of ">").
			commentContent = strings.ReplaceAll(commentContent, "\r\n", "\n")
			commentContent = strings.ReplaceAll(commentContent, "\r", "\n")
			commentContent = html.UnescapeString(commentContent)
		}

		r.nodeSwapByKey[nodeKey] = parserNodeSwap{
			original:    commentContent,
			offsetRange: r.doc.WriteForOffsetRange(raw),
		}

		r.buf = []byte("<!--c" + nodeKey + "-->")
	case html.TextToken:
		original := r.tokenizer.Token().Data

		if !r.nodeRawTextMode {
			// The upstream html.Parse has complex logic for WS (dropping before <head>, preserving in <head>, reparenting
			// after </body>, active formatting etc.). Rather than duplicating and maintaining the logic, propagate it and
			// rely on a trailing comment for backfilling the whitespace text node if it remains.
			//
			// It is possible this whitespace gets squashed with other text nodes, in which case the offsets range currently
			// will be lost.
			if !strings.ContainsFunc(original, func(r rune) bool {
				if uint32(r) <= unicode.MaxLatin1 {
					switch r {
					case '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0:
						return false
					}

					return true
				}

				return !unicode.Is(unicode.White_Space, r)
			}) {
				r.nodeIdx++
				nodeKey := strconv.FormatInt(r.nodeIdx, 10)

				r.wsOffsetRangeByKey[nodeKey] = r.doc.WriteForOffsetRange(raw)
				r.buf = append(raw, []byte("<!--w"+nodeKey+"-->")...)

				return nil
			}
		}

		// approximate behavior of upstream parser; it does not seem to care about other control characters?
		// see https://www.w3.org/International/questions/qa-controls.en.html#support
		original = strings.ReplaceAll(original, "\x00", "")
		if len(original) == 0 {
			// quietly drop the token
			return r.next()
		}

		r.nodeIdx++
		nodeKey := strconv.FormatInt(r.nodeIdx, 10)
		r.nodeSwapByKey[nodeKey] = parserNodeSwap{
			original:    original,
			offsetRange: r.doc.WriteForOffsetRange(raw),
		}

		r.buf = []byte("t" + nodeKey)
	default:
		r.doc.Write(raw)
		r.buf = raw
	}

	return nil
}
