// Copyright (c) Autovia GmbH
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/autovia/dyn/handlers"
	S "github.com/autovia/dyn/structs"

	"github.com/containerd/containerd"
)

func main() {
	app := &S.App{}
	app.Addr = flag.String("addr", ":3000", "TCP address for the server to listen on, in the form host:port")
	app.AccessKey = flag.String("access-key", "user", "aws_access_key_id")
	app.SecretKey = flag.String("secret-key", "password", "aws_secret_access_key")
	app.Mount = flag.String("mount", "./mount", "root directory for container volumes")
	app.Metadata = flag.String("metadata", ".dyn", "root directory for metadata")
	flag.Parse()

	// Router
	app.Router = http.NewServeMux()
	app.Router.Handle("/", S.Auth{App: app, R: map[string]any{
		"POST": handlers.Post,
	}})

	// Check fs folders
	_, err := os.Stat(*app.Mount)
	if os.IsNotExist(err) {
		if err := os.Mkdir(*app.Mount, os.ModePerm); err != nil {
			log.Fatalf("Can not create storage directoy at %s", *app.Mount)
		}
		log.Printf("Storage directory created at %s", *app.Mount)
	}

	metadata := filepath.Join(*app.Mount, *app.Metadata)
	if _, err := os.Stat(metadata); os.IsNotExist(err) {
		if err := os.Mkdir(metadata, os.ModePerm); err != nil {
			log.Fatalf("Can not create metadata directoy at %s", *app.Mount)
		}
		log.Printf("Metadata directory created at %s", metadata)
	}

	// Containerd
	app.C, err = containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		log.Fatalf("Containerd error %s", err)
	}
	defer app.C.Close()

	// Server
	srv := &http.Server{
		Addr:    *app.Addr,
		Handler: app.Router,
		//TLSConfig:    cfg,
		//TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	log.Printf("Listen on %s", *app.Addr)
	log.Fatal(srv.ListenAndServe())
}
