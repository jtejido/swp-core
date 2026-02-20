package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os/signal"
	"syscall"

	"swp-spec-kit/poc/internal/server"
)

func main() {
	listen := flag.String("listen", ":7777", "TCP listen address")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	log.Printf("swp-server listening on %s", *listen)
	s := server.New(log.Default())
	if err := s.Serve(ctx, ln); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
