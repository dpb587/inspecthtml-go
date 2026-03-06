package inspecthtml

import (
	"io"

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

	p.rebuildNode(root)
}

func (p *Parser) rebuildNode(n *html.Node) {
	switch n.Type {
	case html.TextNode:
		var from int
		var isTextRef bool
		var expanded []*html.Node

		appendTextRef := func(i int) {
			swap := p.r.nodeSwapByKey[n.Data[from+1:i]]

			inject := &html.Node{
				Type: html.TextNode,
				Data: swap.original,
			}

			expanded = append(expanded, inject)
			p.offsets.metadataByNode[inject] = &NodeMetadata{
				TokenOffsets: swap.offsetRange,
			}
		}

		for i, c := range n.Data {
			if isTextRef {
				if c < '0' || c > '9' {
					appendTextRef(i)

					from = i
					isTextRef = false
				} else {
					continue
				}
			}

			if c == 't' {
				if from < i {
					inject := &html.Node{
						Type: html.TextNode,
						Data: n.Data[from:i],
					}

					expanded = append(expanded, inject)
				}

				from = i
				isTextRef = true
			}
		}

		if isTextRef {
			appendTextRef(len(n.Data))
		} else if from > 0 {
			inject := &html.Node{
				Type: html.TextNode,
				Data: n.Data[from:],
			}

			expanded = append(expanded, inject)
		}

		if len(expanded) > 0 {
			for i, c := range expanded {
				c.Parent = n.Parent

				if i == 0 {
					c.PrevSibling = n.PrevSibling
					if c.PrevSibling != nil {
						c.PrevSibling.NextSibling = c
					} else if c.Parent != nil {
						c.Parent.FirstChild = c
					}
				} else {
					c.PrevSibling = expanded[i-1]
					c.PrevSibling.NextSibling = c
				}

				if i == len(expanded)-1 {
					c.NextSibling = n.NextSibling
					if c.NextSibling != nil {
						c.NextSibling.PrevSibling = c
					} else if c.Parent != nil {
						c.Parent.LastChild = c
					}
				}
			}

			// assumes only the first text node may contain the newlines that must be trimmed next; unlikely bug?
			n = expanded[0]
		}

		if n.Parent != nil && n.Parent.FirstChild == n &&
			(n.Parent.DataAtom == atom.Textarea ||
				n.Parent.DataAtom == atom.Pre ||
				n.Parent.DataAtom == atom.Listing) {
			if len(n.Data) > 0 && n.Data[0] == '\r' {
				n.Data = n.Data[1:]
			}

			if len(n.Data) > 0 && n.Data[0] == '\n' {
				n.Data = n.Data[1:]
			}

			if len(n.Data) == 0 {
				n.Parent.RemoveChild(n)
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

	for c := n.FirstChild; c != nil; {
		next := c.NextSibling
		p.rebuildNode(c)
		if c.NextSibling != nil && c.NextSibling != next {
			c = c.NextSibling
		} else {
			c = next
		}
	}
}
