package markdown

import (
	"sync"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
)

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.Linkify,
		extension.Table,
		extension.Strikethrough,
		extension.TaskList,
		extension.Typographer,
		extension.DefinitionList,
		highlighting.NewHighlighting(
			highlighting.WithFormatOptions(
				chromahtml.WithLineNumbers(true),
				chromahtml.WithClasses(true),
			),
		),
	),
)

var documentPool = sync.Pool{New: newDocumentInpool}

func newDocumentInpool() any {
	return NewDocument()
}

func GetDocument() *Document {
	return documentPool.Get().(*Document)
}

func PutDocument(doc *Document) {
	doc.Reset()
	documentPool.Put(doc)
}
