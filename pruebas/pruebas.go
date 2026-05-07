package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ══════════════════════════════════════════════════════════════════════════════
//  TEMA PERSONALIZADO — escala de grises, estética industrial
// ══════════════════════════════════════════════════════════════════════════════

type shellTheme struct{}

var _ fyne.Theme = (*shellTheme)(nil)

var (
	colBg      = color.NRGBA{R: 0x0e, G: 0x0e, B: 0x0e, A: 0xff}
	colSurface = color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff}
	colBorder  = color.NRGBA{R: 0x3a, G: 0x3a, B: 0x3a, A: 0xff}
	colAccent  = color.NRGBA{R: 0xd0, G: 0xd0, B: 0xd0, A: 0xff}
	colPrimary = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	colMuted   = color.NRGBA{R: 0x70, G: 0x70, B: 0x70, A: 0xff}
	colHover   = color.NRGBA{R: 0x2e, G: 0x2e, B: 0x2e, A: 0xff}
	colDisabled = color.NRGBA{R: 0x44, G: 0x44, B: 0x44, A: 0xff}
	colError   = color.NRGBA{R: 0xc0, G: 0x50, B: 0x50, A: 0xff}
)

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

func (shellTheme) Font(s fyne.TextStyle) fyne.Resource { return theme.DefaultTheme().Font(s) }
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

// ══════════════════════════════════════════════════════════════════════════════
//  HELPERS
// ══════════════════════════════════════════════════════════════════════════════

func labelMono(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	return l
}

func labelMuted(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.TextStyle = fyne.TextStyle{Monospace: true}
	return l
}

func hRule() fyne.CanvasObject {
	r := canvas.NewRectangle(colBorder)
	r.SetMinSize(fyne.NewSize(0, 1))
	return r
}

func spacer() fyne.CanvasObject { return widget.NewLabel("") }

// ══════════════════════════════════════════════════════════════════════════════
//  VALIDACIÓN
// ══════════════════════════════════════════════════════════════════════════════

func validarCredenciales(usuario, contrasena string) (bool, error) {
	archivo, err := os.Open("contraseña.txt")
	if err != nil {
		return false, fmt.Errorf("no se pudo abrir el archivo: %w", err)
	}
	defer archivo.Close()

	hashIngresado := fmt.Sprintf("%x", sha256.Sum256([]byte(contrasena)))
	scanner := bufio.NewScanner(archivo)
	for scanner.Scan() {
		linea := strings.TrimSpace(scanner.Text())
		if linea == "" {
			continue
		}
		partes := strings.SplitN(linea, ":", 2)
		if len(partes) != 2 {
			continue
		}
		if strings.TrimSpace(partes[0]) == usuario &&
			strings.TrimSpace(partes[1]) == hashIngresado {
			return true, nil
		}
	}
	return false, scanner.Err()
}

// ══════════════════════════════════════════════════════════════════════════════
//  SHELL — ejecución de comandos
// ══════════════════════════════════════════════════════════════════════════════

func ejecutarComando(comando string) (salida string, quit bool) {
	comando = strings.TrimSpace(comando)
	if comando == "" {
		return "", false
	}
	if comando == "bye" || comando == "exit" {
		return "[ sesión terminada ]", true
	}
	sl := strings.Fields(comando)
	if sl[0] == "cd" {
		if len(sl) > 1 {
			if err := os.Chdir(sl[1]); err != nil {
				return "  ✗ cd: " + err.Error(), false
			}
			pwd, _ := os.Getwd()
			return "  → " + pwd, false
		}
		return "  ✗ cd: especifica un directorio.", false
	}
	shell := exec.Command("bash", "-c", comando)
	out, err := shell.CombinedOutput()
	resultado := strings.TrimRight(string(out), "\n")
	if err != nil && resultado == "" {
		resultado = "  ✗ error: " + err.Error()
	}
	return resultado, false
}

// ══════════════════════════════════════════════════════════════════════════════
//  VENTANA PRINCIPAL — terminal
// ══════════════════════════════════════════════════════════════════════════════

func mostrarVentanaPrincipal(a fyne.App) {
	w := a.NewWindow("ShellOS")
	w.Resize(fyne.NewSize(940, 620))
	w.CenterOnScreen()

	// ── título de barra ──────────────────────────────────────────────────
	titBar := canvas.NewText("▸ ShellOS  —  terminal interactiva", colAccent)
	titBar.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	titBar.TextSize = 13

	barTitle := container.NewPadded(titBar)

	// ── historial ────────────────────────────────────────────────────────
	historial := widget.NewLabel("")
	historial.TextStyle = fyne.TextStyle{Monospace: true}
	historial.Wrapping = fyne.TextWrapWord

	scroll := container.NewVScroll(historial)

	lineas := []string{}
	agregar := func(linea string) {
		lineas = append(lineas, linea)
		historial.SetText(strings.Join(lineas, "\n"))
		scroll.ScrollToBottom()
	}

	agregar("  ╔═══════════════════════════════════════════════╗")
	agregar("  ║           S H E L L O S   v1.0               ║")
	agregar("  ║      Ubuntu Linux — shell interactiva         ║")
	agregar("  ║   'bye' / 'exit'  para cerrar la sesión       ║")
	agregar("  ╚═══════════════════════════════════════════════╝")
	agregar("")

	// ── barra de entrada ─────────────────────────────────────────────────
	promptLabel := labelMuted("")
	actualizarPrompt := func() {
		pwd, _ := os.Getwd()
		promptLabel.SetText(pwd + "  »")
	}
	actualizarPrompt()

	entrada := widget.NewEntry()
	entrada.SetPlaceHolder("comando…")
	entrada.TextStyle = fyne.TextStyle{Monospace: true}

	procesar := func() {
		texto := strings.TrimSpace(entrada.Text)
		entrada.SetText("")
		if texto == "" {
			return
		}
		pwd, _ := os.Getwd()
		agregar(fmt.Sprintf("  %s  » %s", pwd, texto))

		salida, quit := ejecutarComando(texto)
		if salida != "" {
			for _, l := range strings.Split(salida, "\n") {
				agregar("  " + l)
			}
		}
		agregar("")
		actualizarPrompt()
		if quit {
			a.Quit()
		}
	}

	entrada.OnSubmitted = func(_ string) { procesar() }

	btnEjec := widget.NewButton("↵ Ejecutar", func() { procesar() })
	btnEjec.Importance = widget.HighImportance

	barCmd := container.NewBorder(nil, nil, promptLabel, btnEjec, entrada)
	barCmdPad := container.NewPadded(barCmd)

	// ── layout ───────────────────────────────────────────────────────────
	w.SetContent(container.NewBorder(
		container.NewVBox(barTitle, hRule()),
		container.NewVBox(hRule(), barCmdPad),
		nil, nil,
		container.NewPadded(scroll),
	))
	w.Canvas().Focus(entrada)
	w.Show()
}

// ══════════════════════════════════════════════════════════════════════════════
//  LOGIN
// ══════════════════════════════════════════════════════════════════════════════

func main() {
	intentos := 0

	a := app.New()
	a.Settings().SetTheme(&shellTheme{})

	w := a.NewWindow("ShellOS — Autenticación")
	w.Resize(fyne.NewSize(460, 500))
	w.CenterOnScreen()

	// ── cabecera ─────────────────────────────────────────────────────────
	logo := canvas.NewText("S H E L L O S", colPrimary)
	logo.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	logo.TextSize = 28
	logo.Alignment = fyne.TextAlignCenter

	sub := canvas.NewText("secure authentication layer", colMuted)
	sub.TextStyle = fyne.TextStyle{Monospace: true}
	sub.TextSize = 11
	sub.Alignment = fyne.TextAlignCenter

	deco := canvas.NewRectangle(colAccent)
	deco.SetMinSize(fyne.NewSize(48, 2))

	cabecera := container.NewVBox(
		spacer(),
		container.NewCenter(logo),
		container.NewCenter(sub),
		spacer(),
		container.NewCenter(container.NewHBox(deco)),
		spacer(),
	)

	// ── campos ───────────────────────────────────────────────────────────
	lblUser := canvas.NewText("USUARIO", colMuted)
	lblUser.TextStyle = fyne.TextStyle{Monospace: true}
	lblUser.TextSize = 10

	inpUser := widget.NewEntry()
	inpUser.SetPlaceHolder("ingresa tu usuario")
	inpUser.TextStyle = fyne.TextStyle{Monospace: true}

	lblPass := canvas.NewText("CONTRASEÑA", colMuted)
	lblPass.TextStyle = fyne.TextStyle{Monospace: true}
	lblPass.TextSize = 10

	inpPass := widget.NewPasswordEntry()
	inpPass.SetPlaceHolder("••••••••")
	inpPass.TextStyle = fyne.TextStyle{Monospace: true}

	// ── mensajes de estado ────────────────────────────────────────────────
	msgStatus := canvas.NewText("", colMuted)
	msgStatus.TextStyle = fyne.TextStyle{Monospace: true}
	msgStatus.TextSize = 12
	msgStatus.Alignment = fyne.TextAlignCenter

	msgIntentos := canvas.NewText("", colMuted)
	msgIntentos.TextStyle = fyne.TextStyle{Monospace: true}
	msgIntentos.TextSize = 10
	msgIntentos.Alignment = fyne.TextAlignCenter

	// ── botón ─────────────────────────────────────────────────────────────
	button := widget.NewButton("[ INICIAR SESIÓN ]", nil)
	button.Importance = widget.HighImportance

	accion := func() {
		valido, err := validarCredenciales(inpUser.Text, inpPass.Text)
		if err != nil {
			msgStatus.Text = "✗  error al leer credenciales"
			msgStatus.Color = colError
			msgStatus.Refresh()
			return
		}
		if valido {
			w.Hide()
			mostrarVentanaPrincipal(a)
		} else {
			intentos++
			msgIntentos.Text = fmt.Sprintf("intentos: %d de 3", intentos)
			msgIntentos.Refresh()
			msgStatus.Text = "✗  credenciales incorrectas"
			msgStatus.Color = colError
			msgStatus.Refresh()

			if intentos >= 3 {
				blk1 := canvas.NewText("⛔  ACCESO BLOQUEADO", colError)
				blk1.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
				blk1.TextSize = 18
				blk1.Alignment = fyne.TextAlignCenter

				blk2 := canvas.NewText("usuario sospechoso — plataforma cerrada", colMuted)
				blk2.TextStyle = fyne.TextStyle{Monospace: true}
				blk2.TextSize = 12
				blk2.Alignment = fyne.TextAlignCenter

				blk3 := canvas.NewText(fmt.Sprintf("[ %d intentos fallidos ]", intentos), colMuted)
				blk3.TextStyle = fyne.TextStyle{Monospace: true}
				blk3.TextSize = 11
				blk3.Alignment = fyne.TextAlignCenter

				w.SetContent(container.NewCenter(container.NewVBox(
					spacer(), spacer(),
					container.NewCenter(blk1),
					spacer(),
					container.NewCenter(blk2),
					spacer(),
					container.NewCenter(blk3),
					spacer(), spacer(),
				)))
			}
		}
	}

	button.OnTapped = accion
	inpPass.OnSubmitted = func(_ string) { accion() }

	// ── layout ────────────────────────────────────────────────────────────
	formulario := container.NewVBox(
		lblUser, inpUser,
		spacer(),
		lblPass, inpPass,
	)

	pie := container.NewVBox(
		container.NewCenter(msgStatus),
		container.NewCenter(msgIntentos),
	)

	cuerpo := container.NewVBox(
		cabecera,
		hRule(),
		spacer(),
		formulario,
		spacer(),
		hRule(),
		spacer(),
		container.NewCenter(button),
		spacer(),
		pie,
	)

	w.SetContent(container.NewPadded(container.NewPadded(cuerpo)))
	w.Canvas().Focus(inpUser)
	w.ShowAndRun()
}