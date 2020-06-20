package main

import (
	"log"

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

	encryptForm.AppendItem(&widget.FormItem{
		Text:   "message",
		Widget: encryptInput,
	})

	decryptForm := &widget.Form{BaseWidget: widget.BaseWidget{}, Items: []*widget.FormItem{}}
	decryptForm.ExtendBaseWidget(decryptForm)

	decryptInput := widget.NewMultiLineEntry()

	decryptForm.AppendItem(&widget.FormItem{
		Text:   "message url",
		Widget: decryptInput,
	})

	encryptButton := widget.NewButton("encrypt", func() {
		log.Printf("encrypt clicked: %s", encryptInput.Text)
	})

	decryptButton := widget.NewButton("decrypt", func() {
		log.Printf("decrypt clicked: %s", decryptInput.Text)
	})

	encryptContainer := fyne.NewContainerWithLayout(layout.NewGridLayout(1), encryptForm, encryptButton)
	decryptContainer := fyne.NewContainerWithLayout(layout.NewGridLayout(1), decryptForm, decryptButton)

	tabs := widget.NewTabContainer(
		&widget.TabItem{
			Text:    "encrypt",
			Icon:    nil,
			Content: encryptContainer,
		},
		&widget.TabItem{
			Text:    "decrypt",
			Icon:    nil,
			Content: decryptContainer,
		},
	)

	w.SetContent(fyne.NewContainerWithLayout(&cipherbin{}, tabs))

	w.ShowAndRun()
}

func (c *cipherbin) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := 300, 300
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
