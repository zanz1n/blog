package markdown_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/markdown"
	"github.com/zanz1n/blog/internal/utils"
)

var _ fmt.Stringer = &jsonHelper{}

type jsonHelper struct {
	any
}

// String implements fmt.Stringer.
func (j jsonHelper) String() string {
	buf, err := json.MarshalIndent(j.any, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", j.any)
	}

	return utils.UnsafeString(buf)
}

var testCases = []struct {
	path      string
	headingct int
}{
	{
		path:      "https://raw.githubusercontent.com/markdown-it/markdown-it/refs/heads/master/support/demo_template/sample.md",
		headingct: 24,
	},
	{
		path:      "https://raw.githubusercontent.com/mxstbr/markdown-test-file/refs/heads/master/TEST.md",
		headingct: 13,
	},
}

func TestMarkdownDoc(t *testing.T) {
	for _, tcase := range testCases {
		pathUrl, err := url.Parse(tcase.path)
		require.NoError(t, err)

		_, testFname := path.Split(pathUrl.Path)

		t.Run(fmt.Sprintf("File(%s)", testFname), func(t *testing.T) {
			res, err := http.Get(tcase.path)
			require.NoError(t, err)

			defer res.Body.Close()

			doc, err := markdown.ParseDocument(res.Body)
			require.NoError(t, err)

			t.Run("Index", func(t *testing.T) {
				idx, warnings := doc.Index()
				require.Equal(t, 0, warnings)
				require.Equal(t, tcase.headingct, len(idx), jsonHelper{idx})
			})

			t.Run("Render", func(t *testing.T) {
				_, err := doc.Render()
				require.NoError(t, err)
			})

			t.Run("Read", func(t *testing.T) {
				tempdir := t.TempDir()
				tempfile := path.Join(tempdir, "output.html")

				defer os.Remove(tempfile)

				output, err := doc.Render()
				require.NoError(t, err)

				err = os.WriteFile(tempfile, output, 0666)
				require.NoError(t, err)

				fileBuf, err := os.ReadFile(tempfile)
				require.NoError(t, err)
				require.Equal(t, output, dto.ArticleContent(fileBuf))
			})
		})
	}
}

func BenchmarkMarkdown(b *testing.B) {
	res, err := http.Get(testCases[0].path)
	require.NoError(b, err)

	body, err := io.ReadAll(res.Body)
	require.NoError(b, err)
	defer res.Body.Close()

	b.Run("Parse", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			doc, err := markdown.ParseDocument(bytes.NewReader(body))
			if err != nil {
				panic(err)
			}
			doc.Tree()
		}
	})

	b.Run("Index", func(b *testing.B) {
		doc, err := markdown.ParseDocument(bytes.NewReader(body))
		require.NoError(b, err)

		for i := 0; i < b.N; i++ {
			idx, warnings := doc.Index()
			_, _ = idx, warnings
		}
	})

	b.Run("Render", func(b *testing.B) {
		doc, err := markdown.ParseDocument(bytes.NewReader(body))
		require.NoError(b, err)

		for i := 0; i < b.N; i++ {
			doc.Render()
			doc.ResetRender()
		}
	})
}
