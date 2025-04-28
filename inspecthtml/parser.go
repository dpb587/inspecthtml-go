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

	parseRoot       *html.Node
	parseErr        error
	offsets         *ParseMetadata
	deferredRebuild []func()
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

	for _, fn := range p.deferredRebuild {
		fn()
	}

	p.deferredRebuild = nil
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
		if n.Data[0] == 'c' {
			pnt := p.r.nodeSwapByKey[n.Data[1:]]
			n.Data = pnt.original

			p.offsets.metadataByNode[n] = &NodeMetadata{
				TokenOffsets: pnt.offsetRange,
			}
		} else if n.Data[0] == 't' {
			pnt := p.r.nodeSwapByKey[n.Data[1:]]
			n.Type = html.TextNode
			n.Data = pnt.original

			p.offsets.metadataByNode[n] = &NodeMetadata{
				TokenOffsets: pnt.offsetRange,
			}
		} else {
			// this should never error given the assumed deterministic tree
			v, _ := cursorio.ParseTextOffsetRange(n.Data)
			p.offsets.metadataByNode[n.PrevSibling].EndTagTokenOffsets = &v

			if n.NextSibling != nil {
				n.NextSibling.PrevSibling = n.PrevSibling
			} else {
				n.Parent.LastChild = n.PrevSibling
			}

			n.PrevSibling.NextSibling = n.NextSibling
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

	if n.Type == html.ElementNode && p.offsets.metadataByNode[n] != nil {
		if p.offsets.metadataByNode[n].TagSelfClosing {
			// nothing to do
		} else if p.offsets.metadataByNode[n].EndTagTokenOffsets == nil {
			if n.LastChild != nil {
				if p.offsets.metadataByNode[n.LastChild] != nil {
					p.offsets.metadataByNode[n].EndTagTokenOffsets = &cursorio.TextOffsetRange{
						From:  p.offsets.metadataByNode[n.LastChild].TokenOffsets.Until,
						Until: p.offsets.metadataByNode[n.LastChild].TokenOffsets.Until,
					}
				}
			} else {
				p.deferredRebuild = append(p.deferredRebuild, func() {
					if n.NextSibling != nil {
						if p.offsets.metadataByNode[n.NextSibling] != nil {
							p.offsets.metadataByNode[n].EndTagTokenOffsets = &cursorio.TextOffsetRange{
								From:  p.offsets.metadataByNode[n.NextSibling].TokenOffsets.From,
								Until: p.offsets.metadataByNode[n.NextSibling].TokenOffsets.From,
							}
						}
					} else if n.Parent != nil {
						if p.offsets.metadataByNode[n.Parent] != nil {
							p.offsets.metadataByNode[n].EndTagTokenOffsets = &cursorio.TextOffsetRange{
								From:  p.offsets.metadataByNode[n.Parent].EndTagTokenOffsets.From,
								Until: p.offsets.metadataByNode[n.Parent].EndTagTokenOffsets.From,
							}
						}
					}
				})
			}
		}
	}
}
