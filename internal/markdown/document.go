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

func ParseDocument(r io.Reader) (*Document, error) {
	src := bytes.NewBuffer([]byte{})

	_, err := io.Copy(src, r)
	if err != nil {
		return nil, err
	}

	rd := text.NewReader(src.Bytes())
	tree := md.Parser().Parse(rd)

	return &Document{
		src:  src,
		tree: tree,
	}, nil
}

type Document struct {
	src  *bytes.Buffer
	dst  *bytes.Buffer
	tree ast.Node
}

func (d *Document) Tree() ast.Node {
	return d.tree
}

func (d *Document) Source() []byte {
	return d.src.Bytes()
}

func (d *Document) Index() (idx dto.ArticleIndexing, warnings int) {
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

func (d *Document) Render() (dto.ArticleContent, error) {
	if d.dst != nil {
		return d.dst.Bytes(), nil
	}

	d.dst = bytes.NewBuffer([]byte{})
	err := md.Renderer().Render(d.dst, d.Source(), d.tree)
	if err != nil {
		d.dst = nil
		return nil, err
	}

	return d.dst.Bytes(), nil
}

func (d *Document) ResetRender() {
	d.dst = nil
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
