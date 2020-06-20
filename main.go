package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"

	"github.com/cipherbin/cipher-bin-cli/pkg/aes256"
	"github.com/cipherbin/cipher-bin-cli/pkg/api"
	"github.com/cipherbin/cipher-bin-cli/pkg/colors"
	"github.com/cipherbin/cipher-bin-cli/pkg/randstring"
	"github.com/cipherbin/cipher-bin-server/db"
	uuid "github.com/satori/go.uuid"
)

type cipherbin struct{}

func main() {
	w := app.New().NewWindow("cipherbin")

	encryptForm := &widget.Form{BaseWidget: widget.BaseWidget{}, Items: []*widget.FormItem{}}
	encryptForm.ExtendBaseWidget(encryptForm)
	encryptInput := widget.NewMultiLineEntry()
	encryptForm.AppendItem(&widget.FormItem{Text: "message", Widget: encryptInput})

	decryptForm := &widget.Form{BaseWidget: widget.BaseWidget{}, Items: []*widget.FormItem{}}
	decryptForm.ExtendBaseWidget(decryptForm)
	decryptInput := widget.NewMultiLineEntry()
	decryptForm.AppendItem(&widget.FormItem{Text: "message url", Widget: decryptInput})

	encryptContainer := fyne.NewContainerWithLayout(layout.NewGridLayout(1), encryptForm)
	decryptContainer := fyne.NewContainerWithLayout(layout.NewGridLayout(1), decryptForm)

	tabs := widget.NewTabContainer(&widget.TabItem{
		Text:    "encrypt",
		Icon:    nil,
		Content: encryptContainer,
	}, &widget.TabItem{
		Text:    "decrypt",
		Icon:    nil,
		Content: decryptContainer,
	})

	tabs.OnChanged = func(tab *widget.TabItem) {
		encryptInput.Text = ""
		decryptInput.Text = ""
		encryptContainer.Refresh()
		decryptContainer.Refresh()
	}

	httpclient := http.Client{Timeout: 15 * time.Second}
	webBaseURL := "https://cipherb.in"
	apiBaseURL := "https://api.cipherb.in"

	binClient, err := api.NewClient(webBaseURL, apiBaseURL, &httpclient)
	if err != nil {
		fmt.Printf("Error creating API client. Err: %v", err)
		os.Exit(1)
		return
	}

	encryptForm.OnSubmit = func() {
		defer func() {
			encryptInput.Text = ""
			encryptContainer.Refresh()
		}()
		fmt.Printf("encryptInput.Text: %s\n", encryptInput.Text)

		uuidv4 := uuid.NewV4().String()
		key := randstring.New(32)

		// Encrypt the message using AES-256
		encryptedMsg, err := aes256.Encrypt([]byte(encryptInput.Text), key)
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
	}

	decryptForm.OnSubmit = func() {
		defer func() {
			decryptInput.Text = ""
			decryptContainer.Refresh()
		}()
		fmt.Printf("decryptInput.Text: %s\n", decryptInput.Text)

		url := decryptInput.Text
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

		// Decrypt the message returned from APIClient.GetMessage
		plainTextMsg, err := aes256.Decrypt(encryptedMsg.Message, key)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Print the decrypted message to the terminal
		fmt.Println(plainTextMsg)
	}

	encryptForm.Refresh()
	decryptForm.Refresh()

	w.SetContent(fyne.NewContainerWithLayout(&cipherbin{}, tabs))
	w.ShowAndRun()
}

// validURL takes a string url and checks whether it's a valid cipherb.in link
func validURL(url, apiBaseURL string) bool {
	return strings.HasPrefix(url, fmt.Sprintf("%s/msg?bin=", apiBaseURL))
}

func (c *cipherbin) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := 400, 400
	for _, o := range objects {
		childSize := o.MinSize()
		w += childSize.Width
		h += childSize.Height
	}
	return fyne.NewSize(w, h)
}

func (c *cipherbin) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	pos := fyne.NewPos(0, containerSize.Height-c.MinSize(objects).Height)
	for _, o := range objects {
		size := o.MinSize()
		o.Resize(size)
		o.Move(pos)
		pos = pos.Add(fyne.NewPos(size.Width, size.Height))
	}
}
