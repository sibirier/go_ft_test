package main

import (
	"flag"
	"fmt"

	"github.com/sibirier/go_ft_test/zipreader"

	"log/slog"
	"context"
	"net/http"
	"os"
	"os/signal"
)

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
	reader, err := zipreader.CreateReader(*filename, *ext_filter)
	if err!=nil {
		logger.Error(err.Error())
		return
	}
	defer reader.Close()

	// start
	logger.Info("ZipViewer started", "Archive", *filename, "Filter",*ext_filter)

	// setting up server with router
	srv := http.Server{Addr: fmt.Sprintf(":%d", *port), Handler: createRouter(reader, logger)}

	// make supply graceful shutdown
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
	
	// run
	if err := srv.ListenAndServe(); err!=nil && err!=http.ErrServerClosed {
		logger.Error(err.Error())
		return
	}
	// wait shutdown
	<-waitEndCh
	logger.Info("HTTP server shutdown graceful")
}

func fitFlag(f string) string {
	if f=="" {
		return f
	}
	if b_l := f[len(f)-1]; b_l=='"' || b_l=='\'' {
		f = f[0:len(f)-1]
	}
	if f=="" {
		return f
	}
	if b_f := f[0]; b_f=='"' || b_f=='\'' {
		f = f[1:len(f)-1]
	}
	return f
}
