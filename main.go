package main

import (
	"fmt"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
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

	encryptForm.OnSubmit = func() {
		fmt.Printf("encryptInput.Text: %s\n", encryptInput.Text)
		encryptInput.Text = ""
		encryptContainer.Refresh()
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
