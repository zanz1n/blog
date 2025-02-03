package markdown

import (
	"bytes"
	"fmt"
	"io"
	"iter"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/utils"
)

var (
	_ io.Writer = &Document{}
	_ io.Reader = &Document{}
)

type Document struct {
	source bytes.Buffer

	output   bytes.Buffer
	rendered bool

	tree ast.Node
}

func NewDocument() *Document {
	return &Document{}
}

// Read implements io.Reader.
func (d *Document) Read(p []byte) (n int, err error) {
	if !d.rendered {
		return 0, io.EOF
	}

	return d.output.Read(p)
}

// Write implements io.Writer.
func (d *Document) Write(p []byte) (n int, err error) {
	if d.tree != nil {
		return 0, io.ErrClosedPipe
	}

	return d.source.Write(p)
}

func (d *Document) Tree() ast.Node {
	return d.tree
}

func (d *Document) Source() []byte {
	return d.source.Bytes()
}

func (d *Document) Content() dto.ArticleContent {
	return dto.ArticleContent(d.output.Bytes())
}

func (d *Document) Parse() {
	rd := text.NewReader(d.Source())
	d.tree = md.Parser().Parse(rd)
	d.sanitizeAst()
}

func (d *Document) Render() error {
	err := md.Renderer().Render(&d.output, d.Source(), d.tree)
	if err == nil {
		d.rendered = true
	}
	return err
}

// This function must only be called after parse
func (d *Document) Index() (idx dto.ArticleIndexing, warnings int) {
	if d.tree == nil {
		panic("(*Document).Index() must not be called after (*Document).Parse()")
	}

	idx = dto.ArticleIndexing{}
	warnings = 0
	headingC := 0

	for _, node := range iterChild(d.tree) {
		if node.Kind() != ast.KindHeading {
			continue
		}
		headingC++

		if node.Type() != ast.TypeBlock {
			// TODO: Warn about not block heading component
			warnings++
			continue
		}

		nodeh, ok := node.(*ast.Heading)
		if !ok {
			// TODO: Warn about it
			warnings++
			continue
		}

		headingType := dto.HeadingType(nodeh.Level)

		lines := node.Lines()
		name := utils.UnsafeString(lines.Value(d.Source()))

		idaatr := fmt.Sprintf("idx-%d-%d", headingType, headingC)
		node.SetAttributeString("id", idaatr)

		idx = append(idx, dto.ArticleIndexingUnit{
			Head: headingType,
			Name: name,
			ID:   idaatr,
		})
	}

	return
}

func (d *Document) Reset() {
	d.source.Reset()
	d.output.Reset()
	d.rendered = false
	d.tree = nil
}

func (d *Document) sanitizeAst() {
	if d.tree == nil {
		panic("(*Document).sanitizeAst() must not be called after (*Document).Parse()")
	}
}

func iterChild(tree ast.Node) iter.Seq2[int, ast.Node] {
	return func(yield func(int, ast.Node) bool) {
		i := 0
		node := tree

		for {
			if i == 0 {
				if node = node.FirstChild(); node == nil {
					return
				}
			} else {
				if node = node.NextSibling(); node == nil {
					return
				}
			}

			if !yield(i, node) {
				return
			}
			i++
		}
	}
}
