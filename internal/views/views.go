package views

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PatrickRudolph/filebin/internal/basicauth"
	"github.com/PatrickRudolph/filebin/internal/filedata"
	"github.com/PatrickRudolph/filebin/internal/highlight"
	"github.com/PatrickRudolph/filebin/internal/renderers"
	"github.com/PatrickRudolph/filebin/internal/settings"
	"github.com/PatrickRudolph/filebin/internal/utils"
	"github.com/PatrickRudolph/filebin/internal/version"
	"github.com/gorilla/mux"
)

var (
	logo = `  __ _ _      _     _
 / _(_) | ___| |__ (_)_ __
| |_| | |/ _ \ '_ \| | '_ \
|  _| | |  __/ |_) | | | | |
|_| |_|_|\___|_.__/|_|_| |_|
`
)

func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	fmt.Fprintf(w, "%s\n", logo)
	fmt.Fprintf(w, "Version %s, running at %s\n\n", version.Version, r.Host)
	fmt.Fprintf(w, "Source code: https://github.com/patrickrudolph/filebin\n")
	fmt.Fprintf(w, "Based on: https://github.com/rafaelmartins/filebin\n")
}

func Robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	fmt.Fprintln(w, "User-agent: *")
	fmt.Fprintln(w, "Disallow: /")
}

func Upload(w http.ResponseWriter, r *http.Request) {
	// authentication
	if !basicauth.BasicAuth(w, r) {
		return
	}

	fds, err := filedata.NewFromRequest(r)
	if err != nil {
		if fds == nil {
			utils.Error(w, err)
			return
		}

		log.Printf("error: %s", err)

		// with at least one valid upload we won't return error
		found := false
		for _, fd := range fds {
			if fd != nil {
				found = true
				break
			}
		}
		if !found {
			utils.ErrorBadRequest(w)
			return
		}
	}

	baseUrl := ""
	if s, err := settings.Get(); err == nil {
		baseUrl = s.BaseUrl
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	for _, fd := range fds {
		if fd == nil {
			fmt.Fprintf(w, "failed\n")
			continue
		}
		if baseUrl != "" {
			fmt.Fprintf(w, "%s/%s\n", baseUrl, fd.Id)
		} else {
			fmt.Fprintf(w, "%s\n", fd.Id)
		}
	}
}

func Event(w http.ResponseWriter, r *http.Request) {
	// authentication
	if !basicauth.BasicAuth(w, r) {
		return
	}

	timeout := time.Minute * 5

	done := make(chan struct{})
	go func() {
		filedata.WaitForEvent()
		close(done)
	}()
	select {
	case <-time.After(timeout):
		w.WriteHeader(http.StatusRequestTimeout)
		// timed out
	case <-done:
		// Wait returned
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
}

func List(w http.ResponseWriter, r *http.Request) {
	// authentication
	if !basicauth.BasicAuth(w, r) {
		return
	}

	baseUrl := ""
	if s, err := settings.Get(); err == nil {
		baseUrl = s.BaseUrl
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")

	if r.Header.Get("Content-Type") == "application/json" {
		data, err := filedata.ToJSON()
		if err == nil {
			fmt.Fprintf(w, "%s", string(data))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		} else {
			fmt.Fprintf(w, "Error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}
	} else {
		filedata.ForEach(func(fd *filedata.FileData) {
			if baseUrl != "" {
				fmt.Fprintf(w, "%s: %s (%s) -> %s/%s\n", fd.Timestamp, fd.Filename, fd.Mimetype, baseUrl, fd.Id)
			} else {
				fmt.Fprintf(w, "%s: %s (%s) -> %s\n", fd.Timestamp, fd.Filename, fd.Mimetype, fd.Id)
			}
		})

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	// authentication
	if !basicauth.BasicAuth(w, r) {
		return
	}

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		http.NotFound(w, r)
		return
	}

	if err := filedata.Delete(id); err != nil {
		if err == filedata.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		utils.Error(w, err)
		return
	}
}

func getFile(w http.ResponseWriter, r *http.Request) *filedata.FileData {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		http.NotFound(w, r)
		return nil
	}

	fd, err := filedata.NewFromId(id)
	if err != nil {
		if err == filedata.ErrNotFound {
			http.NotFound(w, r)
			return nil
		}
		utils.Error(w, err)
		return nil
	}
	return fd
}

func File(w http.ResponseWriter, r *http.Request) {
	fd := getFile(w, r)
	if fd == nil {
		return
	}

	renderer, err := renderers.Lookup(fd.Mimetype)
	if err != nil {
		utils.Error(w, err)
	}

	if err := renderer.Render(w, r, fd); err != nil {
		utils.Error(w, err)
	}
}

func FileText(w http.ResponseWriter, r *http.Request) {
	fd := getFile(w, r)
	if fd == nil {
		return
	}

	lexer, err := highlight.GetLexer(fd.Mimetype)
	if err != nil || lexer == nil {
		utils.ErrorBadRequest(w)
		return
	}

	if err := fd.Serve(w, r, fd.GetFilename(), "text/plain; charset=utf-8", false); err != nil {
		utils.Error(w, err)
	}
}

func FileDownload(w http.ResponseWriter, r *http.Request) {
	fd := getFile(w, r)
	if fd == nil {
		return
	}

	if err := fd.Serve(w, r, fd.GetFilename(), fd.Mimetype, true); err != nil {
		utils.Error(w, err)
	}
}

func FileJSON(w http.ResponseWriter, r *http.Request) {
	fd := getFile(w, r)
	if fd == nil {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(fd); err != nil {
		utils.Error(w, err)
	}
}
