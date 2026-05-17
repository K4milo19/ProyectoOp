package main

// theme.go
// Paleta de colores, implementación de fyne.Theme y helpers de UI reutilizables.
// No contiene lógica de negocio ni llamadas al sistema.

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ── Paleta ────────────────────────────────────────────────────────────────────

var (
	colBg       = color.NRGBA{R: 0x0e, G: 0x0e, B: 0x0e, A: 0xff}
	colSurface  = color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff}
	colBorder   = color.NRGBA{R: 0x3a, G: 0x3a, B: 0x3a, A: 0xff}
	colAccent   = color.NRGBA{R: 0xd0, G: 0xd0, B: 0xd0, A: 0xff}
	colPrimary  = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	colMuted    = color.NRGBA{R: 0x70, G: 0x70, B: 0x70, A: 0xff}
	colHover    = color.NRGBA{R: 0x2e, G: 0x2e, B: 0x2e, A: 0xff}
	colDisabled = color.NRGBA{R: 0x44, G: 0x44, B: 0x44, A: 0xff}
	colError    = color.NRGBA{R: 0xc0, G: 0x50, B: 0x50, A: 0xff}
	colGreen    = color.NRGBA{R: 0x50, G: 0xc0, B: 0x70, A: 0xff}
)

// ── shellTheme ────────────────────────────────────────────────────────────────

type shellTheme struct{}

var _ fyne.Theme = (*shellTheme)(nil)

func (shellTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return colBg
	case theme.ColorNameButton:
		return colSurface
	case theme.ColorNameDisabledButton:
		return colDisabled
	case theme.ColorNameDisabled:
		return colMuted
	case theme.ColorNameError:
		return colError
	case theme.ColorNameFocus:
		return colAccent
	case theme.ColorNameForeground:
		return colPrimary
	case theme.ColorNameHover:
		return colHover
	case theme.ColorNameInputBackground:
		return colSurface
	case theme.ColorNameInputBorder:
		return colBorder
	case theme.ColorNameMenuBackground:
		return colSurface
	case theme.ColorNameOverlayBackground:
		return colSurface
	case theme.ColorNamePlaceHolder:
		return colMuted
	case theme.ColorNamePressed:
		return colAccent
	case theme.ColorNamePrimary:
		return colAccent
	case theme.ColorNameScrollBar:
		return colBorder
	case theme.ColorNameSeparator:
		return colBorder
	case theme.ColorNameShadow:
		return color.NRGBA{A: 0xcc}
	case theme.ColorNameSuccess:
		return colAccent
	case theme.ColorNameWarning:
		return color.NRGBA{R: 0xaa, G: 0x88, B: 0x44, A: 0xff}
	}
	return theme.DefaultTheme().Color(n, v)
}

func (shellTheme) Font(s fyne.TextStyle) fyne.Resource    { return theme.DefaultTheme().Font(s) }
func (shellTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }

func (shellTheme) Size(n fyne.ThemeSizeName) float32 {
	switch n {
	case theme.SizeNamePadding:
		return 10
	case theme.SizeNameInnerPadding:
		return 8
	case theme.SizeNameText:
		return 13
	case theme.SizeNameHeadingText:
		return 20
	case theme.SizeNameSubHeadingText:
		return 15
	case theme.SizeNameInputBorder:
		return 2
	case theme.SizeNameScrollBar:
		return 6
	case theme.SizeNameScrollBarSmall:
		return 3
	}
	return theme.DefaultTheme().Size(n)
}

// ── Helpers de UI ─────────────────────────────────────────────────────────────

// labelMono devuelve un Label monoespaciado en negrita.
func labelMono(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	return l
}

// labelMuted devuelve un Label monoespaciado normal (sin negrita).
func labelMuted(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.TextStyle = fyne.TextStyle{Monospace: true}
	return l
}

// hRule devuelve una línea horizontal de 1 px con el color de borde.
func hRule() fyne.CanvasObject {
	r := canvas.NewRectangle(colBorder)
	r.SetMinSize(fyne.NewSize(0, 1))
	return r
}

// spacer devuelve un widget vacío que actúa como separador vertical.
func spacer() fyne.CanvasObject { return widget.NewLabel("") }