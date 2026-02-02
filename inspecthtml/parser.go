package inspecthtml

import (
	"io"
	"strings"

	"github.com/dpb587/cursorio-go/cursorio"
	"golang.org/x/net/html"
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
		split := strings.SplitN(n.Data[1:], "t", 2)
		swap := p.r.nodeSwapByKey[split[0]]
		n.Data = swap.original

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
			n.Type = html.TextNode
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

		p.offsets.metadataByNode[n] = p.r.nodeTagByKey[n.Attr[len(n.Attr)-1].Val]
		n.Attr = n.Attr[:len(n.Attr)-1]
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.rebuildNode(c)
	}
}
