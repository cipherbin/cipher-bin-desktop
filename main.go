package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"fyne.io/fyne/app"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/cipherbin/cipher-bin-cli/pkg/api"
	"github.com/cipherbin/cipher-bin-desktop/internal/desktop"
)

func newDesktopClient(httpClient *http.Client) (*desktop.Client, error) {
	dc := new(desktop.Client)
	dc.App = app.NewWithID("cipherb.in.desktop")
	dc.App.SetIcon(theme.FyneLogo())
	*dc.Window = dc.App.NewWindow("cipherb.in")

	ac, err := api.NewClient(desktop.WebBaseURL, desktop.APIBaseURL, httpClient)
	if err != nil {
		fmt.Printf("Error creating API client. Err: %v", err)
		os.Exit(1)
		return nil, err
	}
	dc.APIClient = ac

	dc.WriteInput = widget.NewMultiLineEntry()
	dc.ReadInput = widget.NewMultiLineEntry()

	return dc, nil
}

func main() {
	client, err := newDesktopClient(&http.Client{Timeout: 15 * time.Second})
	if err != nil {
		fmt.Printf("Error creating desktop client, err: %s", err.Error())
		os.Exit(1)
	}
	client.InitializeForms()
	client.InitializeContainers()
	client.InitializeTabs()
	client.InitializeWindow()
	win := *client.Window
	win.ShowAndRun()
}
