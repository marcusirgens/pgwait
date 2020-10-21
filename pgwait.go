package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"os"
	"time"
)

var host string
var port uint
var timeoutSeconds uint

func main() {
	if flag.NArg() != 3 {
		fmt.Fprint(os.Stderr, "expected exactly 3 arguments: username password dbname")
		os.Exit(1)
	}

	url := fmt.Sprintf(
		"host=%s user=%s password=%s port=%d dbname=%s sslmode=disable",
		host,
		flag.Arg(0),
		flag.Arg(1),
		port,
		flag.Arg(2),
	)

	db, err := sql.Open("pgx", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db: %v\n", err)
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutSeconds))
	done := make(chan struct{})

	go func() {
		defer close(done)
		t := time.NewTicker(time.Second)
		for {
			if err := db.Ping(); err == nil {
				cancel()
				done <- struct{}{}
				return
			}
			select {
			case <-ctx.Done():
				done <- struct{}{}
				return
			case <-t.C:
			}
		}

	}()

	<-done
	<-ctx.Done()
	
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Fprintln(os.Stderr, "timeout")
			os.Exit(3)
		}

		if errors.Is(err, context.Canceled) {
			os.Exit(0)
		}
	}
}

func init() {
	flag.StringVar(&host, "host", "localhost", "Host")
	flag.UintVar(&timeoutSeconds, "timeout", 60, "Timeout in seconds")
	flag.UintVar(&port, "port", 5432, "Port")
	flag.Parse()
}
