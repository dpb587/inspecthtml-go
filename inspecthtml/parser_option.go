package inspecthtml

import (
	"github.com/dpb587/cursorio-go/cursorio"
	"golang.org/x/net/html"
)

type ParserOption interface {
	apply(*Parser)
}

//

type parserOptionFunc func(*Parser)

func (f parserOptionFunc) apply(p *Parser) {
	f(p)
}

//

func InitialOffsetParserOption(b cursorio.TextOffset) ParserOption {
	return parserOptionFunc(func(p *Parser) {
		p.r.doc = cursorio.NewTextWriter(b)
	})
}

func TokenizerParserOption(f func(t *html.Tokenizer)) ParserOption {
	return parserOptionFunc(func(p *Parser) {
		f(p.r.tokenizer)
	})
}
