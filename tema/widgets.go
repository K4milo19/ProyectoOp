package tema

// widgets.go
// Helpers de UI reutilizables en toda la aplicación.
// Devuelven widgets o CanvasObjects con el estilo de ShellOS ya aplicado.

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// LabelMono devuelve un Label monoespaciado en negrita.
func LabelMono(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	return l
}

// LabelMuted devuelve un Label monoespaciado sin negrita.
func LabelMuted(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.TextStyle = fyne.TextStyle{Monospace: true}
	return l
}

// HRule devuelve una línea horizontal de 1 px con el color de borde.
func HRule() fyne.CanvasObject {
	r := canvas.NewRectangle(ColBorder)
	r.SetMinSize(fyne.NewSize(0, 1))
	return r
}

// Spacer devuelve un widget vacío que actúa como separador vertical.
func Spacer() fyne.CanvasObject { return widget.NewLabel("") }
