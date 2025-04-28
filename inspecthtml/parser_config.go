package inspecthtml

import (
	"github.com/dpb587/cursorio-go/cursorio"
	"golang.org/x/net/html"
)

type ParserConfig struct {
	initialOffset        *cursorio.TextOffset
	tokenizerInterceptor func(t *html.Tokenizer) *html.Tokenizer
}

var _ ParserOption = ParserConfig{}

func (c ParserConfig) apply(o *ParserConfig) {
	if c.initialOffset != nil {
		o.initialOffset = c.initialOffset
	}

	if c.tokenizerInterceptor != nil {
		o.tokenizerInterceptor = c.tokenizerInterceptor
	}
}

func (c ParserConfig) SetInitialOffset(v cursorio.TextOffset) ParserConfig {
	c.initialOffset = &v

	return c
}

func (c ParserConfig) SetTokenizerInterceptor(f func(t *html.Tokenizer) *html.Tokenizer) ParserConfig {
	c.tokenizerInterceptor = f

	return c
}
