package main

import (
	"fmt"
	"sort"
	"html"
	"log/slog"
	"net/http"
	"github.com/go-chi/chi/v5"
)

type MyCustomReader interface {
	FileNames() []string 
	ReadFile(name string) (int64, []byte, error) 
	RawFilter() string 
	Name() string 
	SizeOfFile(name string) string
}

var (
	reader MyCustomReader
	logger *slog.Logger
)


func createRouter(reader_ MyCustomReader, logger_ *slog.Logger) http.Handler {
	reader = reader_
	logger = logger_
	router := chi.NewRouter()

	router.Get("/", mainGetRoute)
	
	router.Get("/main.css", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("main.css")
		http.ServeFile(w, r, "./main.css")
	})

	router.Get("/test/{id}", fileGetRoute)

	return router
}


func mainGetRoute(w http.ResponseWriter, r *http.Request) {
	retHtml := `<html>
	<head>
		<link rel="stylesheet" type="text/css" href="/main.css"/>
		<title>Zip viewer</title>
	</head>
	<body>
	<div class="content" style>
	<span class="name">Archive "%s"</span>
	<span class="ext">ext: "%s"</span>
	`
	w.Write([]byte(fmt.Sprintf(retHtml, html.EscapeString(reader.Name()), html.EscapeString(reader.RawFilter()))))
	names := reader.FileNames()
	sort.Strings(names)
	if len(names)>0{
		for _,v := range names {
			w.Write([]byte(fmt.Sprintf("<a href=\"/test/%s\"%s>%s</a>", html.EscapeString(v), titleBySize(reader.SizeOfFile(v)), html.EscapeString(v))))
		}
	} else {
		w.Write([]byte("<span>No data for presents</span>"))
	}
	w.Write([]byte("</div></body></html>"))
	logger.Info("main page request")
}

func fileGetRoute(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "id")
	if name != "" {
		logger.Info("file in zip read", "Filename", name)
		size, data, err := reader.ReadFile(name)
		if err != nil {
			http.NotFound(w, r)
			logger.Warn("File reading error", "Error", fmt.Sprintf("%v", err))
			return
		}
		if size==0 {
			data = []byte("{no data}")
		}
		w.Write([]byte(fmt.Sprintf(`<html>
			<head>
				<title>%s/%s</title>
				<style>
					body {
						background-color: #333; 
						color: #ccc; 
						font-family: Noto Mono, Ubuntu Mono, monospace;
						white-space: pre;
					}
				</style>
			</head>
			<body><p>`, html.EscapeString(reader.Name()), html.EscapeString(name))))
		w.Write([]byte(html.EscapeString(string(data))))
		w.Write([]byte("</p></body></html>"))
		logger.Debug("Access to existing file", "File", name)
	} else {
		logger.Warn("File reading with empty name")
		http.NotFound(w, r)
	}
}


func titleBySize(fileSize string) string {
	if fileSize=="" {
		return ""
	}
	return fmt.Sprintf(` title="%s"`, fileSize)
}
