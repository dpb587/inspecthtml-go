package inspecthtml

import (
	"io"
	"strings"

	"github.com/dpb587/cursorio-go/cursorio"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type ParserOption interface {
	apply(*ParserConfig)
}

type Parser struct {
	r       *parserReader
	rActual io.Reader

	parseRoot *html.Node
	parseErr  error
	offsets   *ParseMetadata
}

func NewParser(r io.Reader, opts ...ParserOption) *Parser {
	p := &Parser{
		r: &parserReader{
			tokenizer:     html.NewTokenizer(r),
			nodeTagByKey:  map[string]*NodeMetadata{},
			nodeSwapByKey: map[string]parserNodeSwap{},
		},
	}

	cfg := &ParserConfig{
		initialOffset: &cursorio.TextOffset{},
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	p.r.doc = cursorio.NewTextWriter(*cfg.initialOffset)

	if cfg.tokenizerInterceptor != nil {
		p.r.tokenizer = cfg.tokenizerInterceptor(p.r.tokenizer)
	}

	if cfg.readerInterceptor != nil {
		p.rActual = cfg.readerInterceptor(p.r)
	} else {
		p.rActual = p.r
	}

	return p
}

func (p *Parser) Parse() (*html.Node, *ParseMetadata, error) {
	if p.parseRoot == nil && p.parseErr == nil {
		p.parseRoot, p.parseErr = html.Parse(p.rActual)
		if p.parseErr == nil {
			p.rebuild(p.parseRoot)
		}
	}

	return p.parseRoot, p.offsets, p.parseErr
}

func (p *Parser) ParseWithOptions(opts ...html.ParseOption) (*html.Node, *ParseMetadata, error) {
	if p.parseRoot == nil && p.parseErr == nil {
		p.parseRoot, p.parseErr = html.ParseWithOptions(p.rActual, opts...)
		if p.parseErr == nil {
			p.rebuild(p.parseRoot)
		}
	}

	return p.parseRoot, p.offsets, p.parseErr
}

func (p *Parser) rebuild(root *html.Node) {
	p.offsets = &ParseMetadata{
		metadataByNode: map[*html.Node]*NodeMetadata{},
	}

	p.rebuildNode(root, rebuildState{})
}

func (p *Parser) rebuildNode(n *html.Node, state rebuildState) {
	switch n.Type {
	case html.TextNode:
		split := strings.SplitN(n.Data[1:], "t", 2)
		swap := p.r.nodeSwapByKey[split[0]]
		n.Data = swap.original

		// Mirror textIM: strip a leading \r/\n from the first text child of
		// <textarea>. The standard parser does this at parse time, but since
		// we encode text as t{key}, the outer tokenizer never sees the raw
		// newline to strip it. Restore the same behavior here.
		if n.Parent != nil && n.Parent.DataAtom == atom.Textarea && n.Parent.FirstChild == n {
			if len(n.Data) > 0 && n.Data[0] == '\r' {
				n.Data = n.Data[1:]
			}
			if len(n.Data) > 0 && n.Data[0] == '\n' {
				n.Data = n.Data[1:]
			}
			if n.Data == "" {
				n.Parent.RemoveChild(n)
				return
			}
		}

		p.offsets.metadataByNode[n] = &NodeMetadata{
			TokenOffsets: swap.offsetRange,
		}

		if len(split) > 1 {
			inject := &html.Node{
				Parent:      n.Parent,
				PrevSibling: n,
				NextSibling: n.NextSibling,
				Type:        html.TextNode,
				Data:        "t" + split[1],
			}

			n.NextSibling = inject
			n = inject

			if n.NextSibling != nil {
				n.NextSibling.PrevSibling = n
			} else if n.Parent != nil {
				n.Parent.LastChild = n
			}
		}

		return
	case html.CommentNode:
		switch n.Data[0] {
		case 'c':
			pnt := p.r.nodeSwapByKey[n.Data[1:]]
			n.Data = pnt.original

			p.offsets.metadataByNode[n] = &NodeMetadata{
				TokenOffsets: pnt.offsetRange,
			}
		case 't':
			pnt := p.r.nodeSwapByKey[n.Data[1:]]

			if state.wsDrop {
				n.Parent.RemoveChild(n)

				return
			}

			n.Type = html.TextNode
			n.Data = pnt.original

			p.offsets.metadataByNode[n] = &NodeMetadata{
				TokenOffsets: pnt.offsetRange,
			}

			if state.wsReparent != nil && n.Parent != state.wsReparent {
				n.Parent.RemoveChild(n)
				state.wsReparent.AppendChild(n)
			}
		default:
			if p.offsets.metadataByNode[n.PrevSibling] != nil {
				// if it was already set, html parser must have reordered nodes
				// first encountered offset should be most accurate
				if p.offsets.metadataByNode[n.PrevSibling].EndTagTokenOffsets == nil {
					// this should never error given the assumed deterministic tree
					v, _ := cursorio.ParseTextOffsetRange(n.Data)

					p.offsets.metadataByNode[n.PrevSibling].EndTagTokenOffsets = &v
				}
			} else {
				// missing meta; html parser must have injected/restarted a previously open tag
				// rather than fake TokenOffsets + TagNameOffsets, drop the metadata
			}

			if n.NextSibling != nil {
				n.NextSibling.PrevSibling = n.PrevSibling
			} else if n.Parent != nil {
				n.Parent.LastChild = n.PrevSibling
			}

			if n.PrevSibling != nil {
				n.PrevSibling.NextSibling = n.NextSibling
			} else if n.Parent != nil {
				n.Parent.FirstChild = n.NextSibling
			}
		}

		return
	case html.ElementNode:
		if len(n.Attr) == 0 {
			break
		}

		if lastAttr := n.Attr[len(n.Attr)-1]; lastAttr.Key == "o" {
			if metadata := p.r.nodeTagByKey[lastAttr.Val]; metadata != nil {
				p.offsets.metadataByNode[n] = metadata
				n.Attr = n.Attr[:len(n.Attr)-1]
			}
		}
	}

	var childState rebuildState
	if n.Type == html.DocumentNode || n.DataAtom == atom.Html {
		childState.wsDrop = true
	}

	for c := n.FirstChild; c != nil; {
		next := c.NextSibling
		p.rebuildNode(c, childState)

		//   - Document child before <html>: drop (initialIM / beforeHTMLIM)
		//   - Document child after <html>: reparent to <body> (afterAfterBodyIM)
		//   - <html> child before <head>: drop (beforeHeadIM)
		//   - <html> child between <head> and <body>: keep (afterHeadIM)
		//   - <html> child after <body>: reparent to <body> (afterBodyIM)
		if c.Type == html.ElementNode {
			if n.Type == html.DocumentNode && c.DataAtom == atom.Html {
				childState.wsDrop = false
				for body := c.FirstChild; body != nil; body = body.NextSibling {
					if body.Type == html.ElementNode && body.DataAtom == atom.Body {
						childState.wsReparent = body
						break
					}
				}
			} else if n.DataAtom == atom.Html {
				switch c.DataAtom {
				case atom.Head:
					childState.wsDrop = false
				case atom.Body:
					childState.wsReparent = c
				}
			}
		}

		if c.NextSibling != nil && c.NextSibling != next {
			c = c.NextSibling
		} else {
			c = next
		}
	}
}

type rebuildState struct {
	wsDrop     bool
	wsReparent *html.Node
}
