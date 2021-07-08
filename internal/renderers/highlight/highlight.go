package highlight

import (
	"net/http"
	"strings"

	"github.com/PatrickRudolph/filebin/internal/filedata"
	"github.com/PatrickRudolph/filebin/internal/highlight"
)

type HighlightRenderer struct{}

func (h *HighlightRenderer) Supports(mimetype string) bool {
	lexer, err := highlight.GetLexer(mimetype)
	if err != nil {
		return false
	}

	if lexer.Config().Name != "" {
		return true
	}

	return strings.HasPrefix(mimetype, "text/")
}

func (h *HighlightRenderer) Render(w http.ResponseWriter, r *http.Request, fd *filedata.FileData) error {
	return highlightFile(w, fd)
}
