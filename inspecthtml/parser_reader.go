package inspecthtml

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"unicode"

	"github.com/dpb587/cursorio-go/cursorio"
	"golang.org/x/net/html"
)

var emptyQuotes = []byte(`""`)
var equalEmptyQuotes = []byte(`=""`)

var reTagName = regexp.MustCompile(`^<([^\s/<>]+)`)
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

	nodeIdx       int64
	nodeTagByKey  map[string]*NodeMetadata
	nodeSwapByKey map[string]parserNodeSwap
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

		var lastAttrSuffix []byte

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

						lastAttrSuffix = nil
					} else if rawCutset[0] == '\'' {
						closeMatcher := reAttrValueSingleQuote.FindSubmatchIndex(rawCutset[1:])

						if closeMatcher == nil {
							// weird
						} else {
							consumeLen = closeMatcher[1] + 1
						}

						lastAttrSuffix = nil
					} else if !unicode.IsSpace(rune(rawCutset[0])) && rawCutset[0] != '>' {
						closeMatcher := reAttrValueUnquoted.FindSubmatchIndex(rawCutset)

						if closeMatcher == nil {
							// weird
						} else {
							consumeLen = closeMatcher[1]
						}

						lastAttrSuffix = nil
					}

					if consumeLen > 0 {
						valueOffsetRange := r.doc.WriteForOffsetRange(rawCutset[:consumeLen])
						tagAttrProfile.ValueOffsets = &valueOffsetRange

						rawCutset = rawCutset[consumeLen:]
					} else {
						lastAttrSuffix = emptyQuotes
					}
				} else if len(attrValue) > 0 {
					// an edge case worth fixing; almost panic-worthy; subsequent attributes may no longer be correct
					fmt.Fprintf(os.Stderr, "inspecthtml: regex attr failed (raw=%q, key=%q, val=%q)\n", string(rawCutset), string(attrKey), string(attrValue))
				} else {
					rawCutset = rawCutset[rawAttrMatcher[3]:]
					lastAttrSuffix = equalEmptyQuotes
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

		var wSuffix []byte

		if len(lastAttrSuffix) > 0 {
			wSuffix = append(lastAttrSuffix[:], wSuffix...)
		}

		wSuffix = fmt.Appendf(wSuffix, " o=%q", nodeKey)

		if bytes.HasSuffix(raw, []byte("/>")) {
			wSuffix = append(wSuffix, '/', '>')
			raw = raw[:len(raw)-2]
		} else if bytes.HasSuffix(raw, []byte(">")) {
			wSuffix = append(wSuffix, '>')
			raw = raw[:len(raw)-1]
		}

		r.buf = append(raw, wSuffix...)
	case html.EndTagToken:
		docOffsetRange := r.doc.WriteForOffsetRange(raw)

		r.buf = append(raw, []byte("<!--"+docOffsetRange.OffsetRangeString()+"-->")...)
	case html.CommentToken:
		r.nodeIdx++
		nodeKey := strconv.FormatInt(r.nodeIdx, 10)

		var commentContent string

		if len(raw) >= 7 {
			// <!--content-->
			commentContent = string(raw)[4 : len(raw)-3]
		} else if len(raw) > 4 {
			// malformed; but tokenizer recovered; e.g. <!--content
			commentContent = string(raw)[4:]
		}

		r.nodeSwapByKey[nodeKey] = parserNodeSwap{
			original:    commentContent,
			offsetRange: r.doc.WriteForOffsetRange(raw),
		}

		r.buf = []byte("<!--c" + nodeKey + "-->")
	case html.TextToken:
		r.nodeIdx++
		nodeKey := strconv.FormatInt(r.nodeIdx, 10)
		r.nodeSwapByKey[nodeKey] = parserNodeSwap{
			original:    r.tokenizer.Token().Data,
			offsetRange: r.doc.WriteForOffsetRange(raw),
		}

		// WS vs non-WS is significant to tokenizer; maybe only relevant outside of body?
		// for now, avoid special-casing (e.g. <head> </head>; <script>text</script>)
		if bytes.ContainsFunc(raw, func(r rune) bool {
			if uint32(r) <= unicode.MaxLatin1 {
				switch r {
				case '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0:
					return false
				}
				return true
			}

			return !unicode.Is(unicode.White_Space, r)
		}) {
			r.buf = []byte("t" + nodeKey)
		} else {
			r.buf = []byte("<!--t" + nodeKey + "-->")
		}
	default:
		r.doc.Write(raw)
		r.buf = raw
	}

	return nil
}
