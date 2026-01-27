package inspecthtml

import (
	"github.com/dpb587/cursorio-go/cursorio"
)

type NodeMetadata struct {
	TokenOffsets cursorio.TextOffsetRange

	TagNameOffsets *cursorio.TextOffsetRange
	TagAttr        []*NodeAttributeMetadata
	TagSelfClosing bool

	EndTagTokenOffsets *cursorio.TextOffsetRange
}

func (n NodeMetadata) GetOuterOffsets() cursorio.TextOffsetRange {
	if n.EndTagTokenOffsets == nil {
		return n.TokenOffsets
	}

	return cursorio.TextOffsetRange{
		From:  n.TokenOffsets.From,
		Until: n.EndTagTokenOffsets.Until,
	}
}

func (n NodeMetadata) HasInner() bool {
	return n.EndTagTokenOffsets != nil
}

func (n NodeMetadata) GetInnerOffsets() *cursorio.TextOffsetRange {
	if n.EndTagTokenOffsets == nil {
		return nil
	}

	return &cursorio.TextOffsetRange{
		From:  n.TokenOffsets.Until,
		Until: n.EndTagTokenOffsets.From,
	}
}

//

type NodeAttributeMetadata struct {
	KeyOffsets   cursorio.TextOffsetRange
	ValueOffsets *cursorio.TextOffsetRange
}
