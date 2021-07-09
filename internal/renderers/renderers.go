package renderers

import (
	"errors"
	"net/http"

	"github.com/PatrickRudolph/filebin/internal/filedata"
	"github.com/PatrickRudolph/filebin/internal/renderers/highlight"
	"github.com/PatrickRudolph/filebin/internal/renderers/html"
	"github.com/PatrickRudolph/filebin/internal/renderers/markdown"
	"github.com/PatrickRudolph/filebin/internal/renderers/raw"
)

var (
	reg = []Renderer{
		&markdown.MarkdownRenderer{},
		&html.HtmlRenderer{},
		&highlight.HighlightRenderer{},
		&raw.RawRenderer{},
	}
)

type Renderer interface {
	Supports(mimetype string) bool
	Render(w http.ResponseWriter, r *http.Request, fd *filedata.FileData) error
}

func Lookup(mimetype string) (Renderer, error) {
	for _, r := range reg {
		if r.Supports(mimetype) {
			return r, nil
		}
	}

	return nil, errors.New("renderers: not supported")
}
