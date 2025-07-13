package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/TecharoHQ/yeet/internal/gpgtest"
)

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("wrong args")
	}

	fname := flag.Arg(0)
	keyID, err := gpgtest.MakeKey(context.Background(), fname)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("export GPG_KEY_FILE=" + fname)
	fmt.Println("export GPG_KEY_ID=" + keyID)
}
