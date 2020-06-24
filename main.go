package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cipherbin/cipher-bin-desktop/internal/desktop"
)

func main() {
	client, err := desktop.NewClient(&http.Client{Timeout: 15 * time.Second})
	if err != nil {
		fmt.Printf("Error creating desktop client, err: %s", err.Error())
		os.Exit(1)
	}
	client.Run()
}
