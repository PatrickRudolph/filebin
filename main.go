package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/PatrickRudolph/filebin/internal/filedata"
	"github.com/PatrickRudolph/filebin/internal/settings"
	"github.com/PatrickRudolph/filebin/internal/views"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/nanmu42/gzip"
)

func usage(err error) {
	fmt.Fprintln(os.Stderr, "usage: filebin")
	if err != nil {
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "error:", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func logHandler(w io.Writer, params handlers.LogFormatterParams) {
	uri := params.Request.RequestURI
	if params.Request.ProtoMajor == 2 && params.Request.Method == "CONNECT" {
		uri = params.Request.Host
	}
	if uri == "" {
		uri = params.URL.RequestURI()
	}

	fmt.Fprintf(w, "[%s] %s %q %d %d\n",
		params.TimeStamp.UTC().Format("2006-01-02 15:04:05 MST"),
		params.Request.Method,
		uri,
		params.StatusCode,
		params.Size,
	)
}

func removeOldFilesCheck(s *settings.Settings) {

	filedata.ForEach(func(fd *filedata.FileData) {
		if !fd.Timestamp.Add(s.MaxAge).After(time.Now()) {
			go filedata.Delete(fd.Id)
		}
	})

	go func() {
		time.Sleep(time.Hour)
		removeOldFilesCheck(s)
	}()
}

func main() {
	s, err := settings.Get()
	if err != nil {
		usage(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", views.Upload).Methods("POST")
	r.HandleFunc("/", views.Index)
	r.HandleFunc("/robots.txt", views.Robots)
	r.HandleFunc("/list", views.List)
	r.HandleFunc("/event", views.Event)
	r.HandleFunc("/{id}.json", views.FileJSON)
	r.HandleFunc("/{id}.txt", views.FileText)
	r.HandleFunc("/{id}/download", views.FileDownload)
	r.HandleFunc("/{id}", views.Delete).Methods("DELETE")
	r.HandleFunc("/{id}", views.File)

	h := handlers.CustomLoggingHandler(os.Stderr, r, logHandler)
	h = handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(h)

	if err := filedata.Init(); err != nil {
		usage(err)
	}

	go removeOldFilesCheck(s)

	fmt.Fprintf(os.Stderr, " * Listening on %s (backend: %s)\n", s.ListenAddr, s.Backend.Name())
	if err := http.ListenAndServe(s.ListenAddr, gzip.DefaultHandler().WrapHandler(h)); err != nil {
		usage(err)
	}
}
