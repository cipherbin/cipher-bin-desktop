package main

import (
	"fmt"
	"net/http"
	"os"
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

		// Create a v4 uuid for message identification and to eliminate
		// almost any chance of stumbling upon the url
		uuidv4 := uuid.NewV4().String()

		// Generate a random 32 byte string
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
		fmt.Printf("decryptInput.Text: %s\n", decryptInput.Text)
		decryptInput.Text = ""
		decryptContainer.Refresh()
	}

	encryptForm.Refresh()
	decryptForm.Refresh()

	w.SetContent(fyne.NewContainerWithLayout(&cipherbin{}, tabs))
	w.ShowAndRun()
}

// TODO:
// - Actually submit data to the api
// - Size things better

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
