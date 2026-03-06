package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yoshiya0503/mahjongsoul-overlay/pkg/server"
)

//go:embed web/*
var webFS embed.FS

func main() {
	addr := flag.String("addr", ":8787", "listen address")
	certFile := flag.String("cert", "server.crt", "TLS certificate file")
	keyFile := flag.String("key", "server.key", "TLS key file")
	flag.Parse()

	webContent, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal(err)
	}

	srv := server.New(webContent)

	go func() {
		log.Printf("mahjongsoul-overlay starting on https://localhost%s", *addr)
		if err := srv.ListenAndServe(*addr, *certFile, *keyFile); err != nil {
			log.Fatal(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down")
	srv.Shutdown()
}
