package backends

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/PatrickRudolph/filebin/internal/filedata/backends/local"
)

type Backend interface {
	Name() string
	List() ([]string, error)
	Read(id string) (io.ReadCloser, error)
	ReadMetadata(id string) (string, string, int64, time.Time, error)
	Write(id string, r io.ReadSeeker, filename string, mimetype string) (int64, error)
	Delete(id string) error
	Serve(w http.ResponseWriter, r *http.Request, id string, filename string, mimetype string, attachment bool) error
}

func Lookup(dir string) (Backend, error) {

	if dir != "" {
		return local.NewLocal(dir)
	}

	return nil, errors.New("filedata: no data backend configured")
}
