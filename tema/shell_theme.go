package tema

// shell_theme.go
// Implementación de fyne.Theme para ShellOS.
// Mapea cada nombre de color/tamaño de Fyne a la paleta de ShellOS.

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ShellTheme implementa fyne.Theme con la paleta oscura de ShellOS.
type ShellTheme struct{}

var _ fyne.Theme = (*ShellTheme)(nil)

func (ShellTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return ColBg
	case theme.ColorNameButton:
		return ColSurface
	case theme.ColorNameDisabledButton:
		return ColDisabled
	case theme.ColorNameDisabled:
		return ColMuted
	case theme.ColorNameError:
		return ColError
	case theme.ColorNameFocus:
		return ColAccent
	case theme.ColorNameForeground:
		return ColPrimary
	case theme.ColorNameHover:
		return ColHover
	case theme.ColorNameInputBackground:
		return ColSurface
	case theme.ColorNameInputBorder:
		return ColBorder
	case theme.ColorNameMenuBackground:
		return ColSurface
	case theme.ColorNameOverlayBackground:
		return ColSurface
	case theme.ColorNamePlaceHolder:
		return ColMuted
	case theme.ColorNamePressed:
		return ColAccent
	case theme.ColorNamePrimary:
		return ColAccent
	case theme.ColorNameScrollBar:
		return ColBorder
	case theme.ColorNameSeparator:
		return ColBorder
	case theme.ColorNameShadow:
		return color.NRGBA{A: 0xcc}
	case theme.ColorNameSuccess:
		return ColAccent
	case theme.ColorNameWarning:
		return color.NRGBA{R: 0xaa, G: 0x88, B: 0x44, A: 0xff}
	}
	return theme.DefaultTheme().Color(n, v)
}

func (ShellTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DefaultTheme().Font(s) }
func (ShellTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }

func (ShellTheme) Size(n fyne.ThemeSizeName) float32 {
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
