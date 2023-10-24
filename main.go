package main

import (
	"flag"
	"fmt"
	// "strings"
	"html"
	// "encoding/json"

	"github.com/sibirier/go_ft_test/zipreader"

	"log/slog"
	"context"
	"net/http"
	"os"
	"os/signal"
	"github.com/go-chi/chi/v5"
	// "github.com/go-chi/chi/v5/middleware"
)

func fitFlag(f string) string {
	if f=="" {
		return f
	}
	if b_l := f[len(f)-1]; b_l=='"' || b_l=='\'' {
		f = f[1:len(f)-1]
	}
	if f=="" {
		return f
	}
	if b_f := f[0]; b_f=='"' || b_f=='\'' {
		f = f[1:len(f)-1]
	}
	return f
}

func main() {
	// flags
	port := flag.Int("p", 8080, "port for server")
	ext_filter := flag.String("ext", "", "filter for extensions to display files")
	filename := flag.String("file", "", "filter for extensions to display files")
	flag.Parse()

	//logger
	th := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(th)

	if filename==nil || *filename=="" {
		logger.Error("file must be specified")
		return
	}

	// prepare input data
	*filename = fitFlag(*filename)
	*ext_filter = fitFlag(*ext_filter)

	// zip
	r, err := zipreader.CreateReader(*filename, *ext_filter)
	if err!=nil {
		logger.Error(err.Error())
		return
	}
	defer r.Close()

	// fire
	logger.Info("ZipViewer started", "Archive", *filename, "Filter",*ext_filter)

	srv := http.Server{Addr: fmt.Sprintf(":%d", *port), Handler: createRouter(r, logger)}

	waitEndCh := make(chan int, 1)
	go func(){
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		<-c

		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error(err.Error())
		}
		close(waitEndCh)
	}()
	
	if err := srv.ListenAndServe(); err!=nil && err!=http.ErrServerClosed {
		logger.Error(err.Error())
		return
	}
	<-waitEndCh
	logger.Info("HTTP server shutdown graceful")
}


func titleBySize(fileSize string) string {
	if fileSize=="" {
		return ""
	}
	return fmt.Sprintf(` title="%s"`, fileSize)
}

// might (and pleasure) place into other file, where will has own interface for reader, compatible with zipreader.MyReader
func createRouter(reader *zipreader.MyReader, logger *slog.Logger) http.Handler {
	router := chi.NewRouter()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
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
		w.Write([]byte(fmt.Sprintf(retHtml, reader.Name(), reader.RawFilter())))
		names := reader.FileNames()
		if len(names)>0{
			for _,v := range names {
				w.Write([]byte(fmt.Sprintf("<a href=\"/test/%s\"%s>%s</a>", v, titleBySize(reader.SizeOfFile(v)),v)))
			}
		} else {
			w.Write([]byte("<span>No data for presents</span>"))
		}
		w.Write([]byte("</div></body></html>"))
		logger.Info("main page request")
	})
	
	router.Get("/main.css", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("main.css")
		http.ServeFile(w, r, "./main.css")
	})

	router.Get("/test/{id}", func(w http.ResponseWriter, r *http.Request) {
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
				<body><p>`, reader.Name(), name)))
			w.Write([]byte(html.EscapeString(string(data))))
			w.Write([]byte("</p></body></html>"))
			logger.Debug("Access to existing file", "File", name)
		} else {
			logger.Warn("File reading with empty name")
			http.NotFound(w, r)
		}
	})

	return router
}