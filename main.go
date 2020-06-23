package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/cipherbin/cipher-bin-cli/pkg/aes256"
	"github.com/cipherbin/cipher-bin-cli/pkg/api"
	"github.com/cipherbin/cipher-bin-cli/pkg/colors"
	"github.com/cipherbin/cipher-bin-cli/pkg/randstring"
	"github.com/cipherbin/cipher-bin-desktop/example/fyne_demo/data"
	"github.com/cipherbin/cipher-bin-server/db"
	uuid "github.com/satori/go.uuid"
)

const (
	prefCurrentTab = "currentTab"
	webBaseURL     = "https://cipherb.in"
	apiBaseURL     = "https://api.cipherb.in"
)

type desktopClient struct {
	app        fyne.App
	window     *fyne.Window
	binClient  *api.Client
	writeInput *widget.Entry
	readInput  *widget.Entry
	writeForm  *widget.Form
	readForm   *widget.Form
}

type appLayout struct{}

func main() {
	client := new(desktopClient)
	client.app = app.NewWithID("cipherb.in.desktop")
	client.app.SetIcon(theme.FyneLogo())
	w := client.app.NewWindow("cipherb.in")
	client.window = &w
	httpclient := http.Client{Timeout: 15 * time.Second}

	binClient, err := api.NewClient(webBaseURL, apiBaseURL, &httpclient)
	if err != nil {
		fmt.Printf("Error creating API client. Err: %v", err)
		os.Exit(1)
		return
	}
	client.binClient = binClient

	client.writeInput = widget.NewMultiLineEntry()
	client.writeForm = &widget.Form{
		Items:    []*widget.FormItem{{Text: "Message", Widget: client.writeInput}},
		OnCancel: func() { client.writeInput.Text = "" },
		OnSubmit: func() {
			defer func() { client.writeInput.Text = "" }()
			fmt.Println(client.writeInput)
			uuidv4 := uuid.NewV4().String()
			key := randstring.New(32)

			// Encrypt the message using AES-256
			encryptedMsg, err := aes256.Encrypt([]byte(client.writeInput.Text), key)
			if err != nil {
				colors.Println(err.Error(), colors.Red)
				os.Exit(1)
			}

			// Create one time use URL with format {host}?bin={uuidv4};{ecryption_key}
			oneTimeURL := fmt.Sprintf("%s/msg?bin=%s;%s", webBaseURL, uuidv4, key)
			msg := db.Message{UUID: uuidv4, Message: encryptedMsg}

			if err := binClient.PostMessage(&msg); err != nil {
				os.Exit(1)
			}

			fmt.Printf("One time URL: %s\n", oneTimeURL)
		},
	}

	client.readInput = widget.NewMultiLineEntry()
	client.readForm = &widget.Form{
		Items:    []*widget.FormItem{{Text: "Message", Widget: client.readInput}},
		OnCancel: func() { client.readInput.Text = "" },
		OnSubmit: func() {
			defer func() { client.readInput.Text = "" }()

			url := client.readInput.Text
			if !validURL(url, webBaseURL) {
				fmt.Println("sorry, this message has either already been viewed and destroyed or it never existed at all")
				os.Exit(1)
				return
			}

			// If we've gotten here, the open in browser flag was not provided, so we
			// replace the browser url with the api url to fetch the message here
			url = strings.Replace(url, webBaseURL, apiBaseURL, -1)

			encryptedMsg, err := binClient.GetMessage(url)
			if err != nil {
				fmt.Printf("error fetching message: %+v\n", err)
				os.Exit(1)
				return
			}

			var key string

			// Ensure we have what looks like an AES key and set the key var if so
			urlParts := strings.Split(url, ";")
			if len(urlParts) == 2 {
				key = urlParts[1]
			}

			// Length of urlParts != 2. In other words, if it's an invalid link.
			if key == "" {
				fmt.Printf("sorry, it seems you have an invalid link: %+v", err)
				os.Exit(1)
				return
			}

			// Decrypt the message returned from GetMessage
			plainTextMsg, err := aes256.Decrypt(encryptedMsg.Message, key)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Println(plainTextMsg)
		},
	}

	wmc := fyne.NewContainerWithLayout(
		layout.NewBorderLayout(widget.NewToolbar(), nil, nil, nil),
		widget.NewTabContainer(widget.NewTabItem("Message", client.writeForm)),
	)

	rmc := fyne.NewContainerWithLayout(
		layout.NewBorderLayout(widget.NewToolbar(), nil, nil, nil),
		widget.NewTabContainer(widget.NewTabItem("Message", client.readForm)),
	)

	tabs := widget.NewTabContainer(
		widget.NewTabItemWithIcon("Welcome", theme.HomeIcon(), homeWindow(client.app)),
		widget.NewTabItemWithIcon("Write Message", theme.DocumentCreateIcon(), wmc),
		widget.NewTabItemWithIcon("Read Message", theme.FolderOpenIcon(), rmc),
	)
	tabs.SetTabLocation(widget.TabLocationLeading)
	tabs.SelectTabIndex(client.app.Preferences().Int(prefCurrentTab))
	tabs.OnChanged = func(tab *widget.TabItem) {
		client.writeInput.Text = ""
		client.readInput.Text = ""
		client.writeInput.Refresh()
		client.readInput.Refresh()
	}

	win := *client.window
	win.SetContent(tabs)
	win.ShowAndRun()
	client.app.Preferences().SetInt(prefCurrentTab, tabs.CurrentTabIndex())

	win.SetContent(fyne.NewContainerWithLayout(layout.NewBorderLayout(widget.NewToolbar(), nil, nil, nil), tabs))
	win.ShowAndRun()
}

// validURL takes a string url and checks whether it's a valid cipherb.in link
func validURL(url, apiBaseURL string) bool {
	return strings.HasPrefix(url, fmt.Sprintf("%s/msg?bin=", apiBaseURL))
}

func homeWindow(a fyne.App) fyne.CanvasObject {
	logo := canvas.NewImageFromResource(data.FyneScene)
	logo.SetMinSize(fyne.NewSize(228, 167))

	return widget.NewVBox(
		layout.NewSpacer(),
		widget.NewLabelWithStyle(
			"welcome to cipherb.in",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		),
		widget.NewHBox(layout.NewSpacer(), logo, layout.NewSpacer()),
		widget.NewHBox(
			layout.NewSpacer(),
			widget.NewHyperlink("cipherb.in", parseURL("https://cipherb.in/")),
			widget.NewLabel("-"),
			widget.NewHyperlink("github", parseURL("https://github.com/cipherbin/cipher-bin-desktop")),
			layout.NewSpacer(),
		),
		layout.NewSpacer(),
		widget.NewGroup(
			"Theme",
			fyne.NewContainerWithLayout(
				layout.NewGridLayout(2),
				widget.NewButton("Dark", func() {
					a.Settings().SetTheme(theme.DarkTheme())
				}),
				widget.NewButton("Light", func() {
					a.Settings().SetTheme(theme.LightTheme())
				}),
			),
		),
	)
}

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}
	return link
}

func (a *appLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := 400, 400
	for _, o := range objects {
		childSize := o.MinSize()
		w += childSize.Width
		h += childSize.Height
	}
	return fyne.NewSize(w, h)
}

func (a *appLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	pos := fyne.NewPos(0, containerSize.Height-a.MinSize(objects).Height)
	for _, o := range objects {
		size := o.MinSize()
		o.Resize(size)
		o.Move(pos)
		pos = pos.Add(fyne.NewPos(size.Width, size.Height))
	}
}
