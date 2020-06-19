package main

import (
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
	encryptForm.AppendItem(&widget.FormItem{
		Text:   "message",
		Widget: widget.NewTextGridFromString("TODO: text input"),
	})

	decryptForm := &widget.Form{BaseWidget: widget.BaseWidget{}, Items: []*widget.FormItem{}}
	decryptForm.ExtendBaseWidget(decryptForm)
	decryptForm.AppendItem(&widget.FormItem{
		Text:   "message url",
		Widget: widget.NewTextGridFromString("TODO: text input"),
	})

	encryptContainer := fyne.NewContainerWithLayout(layout.NewGridLayout(1), encryptForm)
	decryptContainer := fyne.NewContainerWithLayout(layout.NewGridLayout(1), decryptForm)

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
