package main

// main.go
// Punto de entrada de ShellOS.
// Flujo: selección de modo (Servidor / Cliente) → login → ventana principal.
//
// Servidor: abre un socket TCP en :9000, muestra log en vivo de conexiones
//           y comandos ejecutados, y registra todo en logs.txt.
// Cliente:  pide la IP del servidor, se conecta y abre una terminal remota.

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// modoApp indica si el usuario eligió operar como servidor o cliente.
type modoApp int

const (
	modoServidor modoApp = iota
	modoCliente
)

// ══════════════════════════════════════════════════════════════════════════════
//  VENTANA DE SELECCIÓN DE MODO
// ══════════════════════════════════════════════════════════════════════════════

// mostrarSeleccionModo muestra la primera ventana donde el usuario elige
// si este nodo actuará como Servidor o como Cliente.
// Al elegir llama a mostrarLogin con el modo seleccionado.
func mostrarSeleccionModo(a fyne.App) {
	w := a.NewWindow("ShellOS — Selección de modo")
	w.Resize(fyne.NewSize(480, 420))
	w.CenterOnScreen()

	// ── Cabecera ──────────────────────────────────────────────────────────
	logo := canvas.NewText("S H E L L O S", colPrimary)
	logo.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	logo.TextSize = 28
	logo.Alignment = fyne.TextAlignCenter

	sub := canvas.NewText("selecciona el modo de operación", colMuted)
	sub.TextStyle = fyne.TextStyle{Monospace: true}
	sub.TextSize = 11
	sub.Alignment = fyne.TextAlignCenter

	deco := canvas.NewRectangle(colAccent)
	deco.SetMinSize(fyne.NewSize(48, 2))

	// ── Tarjeta SERVIDOR ──────────────────────────────────────────────────
	icnSrv := canvas.NewText("▣", colAccent)
	icnSrv.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	icnSrv.TextSize = 28

	tituloSrv := canvas.NewText("SERVIDOR", colPrimary)
	tituloSrv.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	tituloSrv.TextSize = 15

	descSrv := canvas.NewText("Acepta conexiones remotas.\nEjecuta comandos y guarda logs.txt.", colMuted)
	descSrv.TextStyle = fyne.TextStyle{Monospace: true}
	descSrv.TextSize = 11

	btnSrv := widget.NewButton("[ INICIAR COMO SERVIDOR ]", func() {
		w.Hide()
		mostrarLogin(a, modoServidor)
	})
	btnSrv.Importance = widget.HighImportance

	cardSrv := container.NewVBox(
		container.NewCenter(icnSrv),
		container.NewCenter(tituloSrv),
		container.NewCenter(descSrv),
		spacer(),
		container.NewCenter(btnSrv),
	)

	// ── Tarjeta CLIENTE ───────────────────────────────────────────────────
	icnCli := canvas.NewText("▷", colAccent)
	icnCli.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	icnCli.TextSize = 28

	tituloCli := canvas.NewText("CLIENTE", colPrimary)
	tituloCli.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	tituloCli.TextSize = 15

	descCli := canvas.NewText("Conecta a un servidor remoto.\nEnvía comandos y ve la salida.", colMuted)
	descCli.TextStyle = fyne.TextStyle{Monospace: true}
	descCli.TextSize = 11

	btnCli := widget.NewButton("[ INICIAR COMO CLIENTE ]", func() {
		w.Hide()
		mostrarLogin(a, modoCliente)
	})
	btnCli.Importance = widget.MediumImportance

	cardCli := container.NewVBox(
		container.NewCenter(icnCli),
		container.NewCenter(tituloCli),
		container.NewCenter(descCli),
		spacer(),
		container.NewCenter(btnCli),
	)

	// ── Layout ────────────────────────────────────────────────────────────
	cabecera := container.NewVBox(
		spacer(),
		container.NewCenter(logo),
		container.NewCenter(sub),
		spacer(),
		container.NewCenter(container.NewHBox(deco)),
		spacer(),
	)

	tarjetas := container.NewGridWithColumns(2, cardSrv, cardCli)

	cuerpo := container.NewVBox(
		cabecera,
		hRule(),
		spacer(),
		tarjetas,
		spacer(),
	)

	w.SetContent(container.NewPadded(container.NewPadded(cuerpo)))
	w.ShowAndRun()
}

// ══════════════════════════════════════════════════════════════════════════════
//  LOGIN
// ══════════════════════════════════════════════════════════════════════════════

// mostrarLogin muestra la pantalla de autenticación.
// Si el login es exitoso abre la ventana correspondiente al modo elegido.
func mostrarLogin(a fyne.App, modo modoApp) {
	intentos := 0

	w := a.NewWindow("ShellOS — Autenticación")
	w.Resize(fyne.NewSize(460, 500))
	w.CenterOnScreen()

	// ── Cabecera ──────────────────────────────────────────────────────────
	logo := canvas.NewText("S H E L L O S", colPrimary)
	logo.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	logo.TextSize = 28
	logo.Alignment = fyne.TextAlignCenter

	modotxt := "[ modo: SERVIDOR ]"
	if modo == modoCliente {
		modotxt = "[ modo: CLIENTE ]"
	}
	subModo := canvas.NewText(modotxt, colAccent)
	subModo.TextStyle = fyne.TextStyle{Monospace: true}
	subModo.TextSize = 11
	subModo.Alignment = fyne.TextAlignCenter

	sub := canvas.NewText("secure authentication layer", colMuted)
	sub.TextStyle = fyne.TextStyle{Monospace: true}
	sub.TextSize = 11
	sub.Alignment = fyne.TextAlignCenter

	deco := canvas.NewRectangle(colAccent)
	deco.SetMinSize(fyne.NewSize(48, 2))

	cabecera := container.NewVBox(
		spacer(),
		container.NewCenter(logo),
		container.NewCenter(subModo),
		container.NewCenter(sub),
		spacer(),
		container.NewCenter(container.NewHBox(deco)),
		spacer(),
	)

	// ── Campos ───────────────────────────────────────────────────────────
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

	// ── Estado ────────────────────────────────────────────────────────────
	msgStatus := canvas.NewText("", colMuted)
	msgStatus.TextStyle = fyne.TextStyle{Monospace: true}
	msgStatus.TextSize = 12
	msgStatus.Alignment = fyne.TextAlignCenter

	msgIntentos := canvas.NewText("", colMuted)
	msgIntentos.TextStyle = fyne.TextStyle{Monospace: true}
	msgIntentos.TextSize = 10
	msgIntentos.Alignment = fyne.TextAlignCenter

	// ── Botón ─────────────────────────────────────────────────────────────
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
			switch modo {
			case modoServidor:
				mostrarVentanaServidor(a)
			case modoCliente:
				mostrarVentanaConexion(a) // definido en red_cliente.go
			}
			return
		}

		intentos++
		msgIntentos.Text = fmt.Sprintf("intentos: %d de 3", intentos)
		msgIntentos.Refresh()
		msgStatus.Text = "✗  credenciales incorrectas"
		msgStatus.Color = colError
		msgStatus.Refresh()

		if intentos >= 3 {
			mostrarPantallaBloqueo(w, intentos)
		}
	}

	button.OnTapped = accion
	inpPass.OnSubmitted = func(_ string) { accion() }

	// ── Layout ────────────────────────────────────────────────────────────
	formulario := container.NewVBox(lblUser, inpUser, spacer(), lblPass, inpPass)
	pie := container.NewVBox(
		container.NewCenter(msgStatus),
		container.NewCenter(msgIntentos),
	)
	cuerpo := container.NewVBox(
		cabecera, hRule(), spacer(),
		formulario, spacer(), hRule(), spacer(),
		container.NewCenter(button), spacer(), pie,
	)

	w.SetContent(container.NewPadded(container.NewPadded(cuerpo)))
	w.Canvas().Focus(inpUser)
	w.Show()
}

// ══════════════════════════════════════════════════════════════════════════════
//  VENTANA SERVIDOR
// ══════════════════════════════════════════════════════════════════════════════

// mostrarVentanaServidor abre el panel del servidor: inicia el socket TCP
// y muestra un log en tiempo real de todas las conexiones y comandos.
func mostrarVentanaServidor(a fyne.App) {
	w := a.NewWindow("ShellOS — Servidor")
	w.Resize(fyne.NewSize(860, 620))
	w.CenterOnScreen()

	stop := make(chan struct{})
	w.SetOnClosed(func() { close(stop) })

	// ── Título ────────────────────────────────────────────────────────────
	titBar := canvas.NewText("▣  ShellOS  —  modo SERVIDOR", colAccent)
	titBar.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	titBar.TextSize = 13

	estadoLbl := canvas.NewText("● iniciando...", colMuted)
	estadoLbl.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	estadoLbl.TextSize = 11

	barTitulo := container.NewBorder(nil, nil, nil,
		container.NewPadded(estadoLbl),
		container.NewPadded(titBar),
	)

	// ── IPs locales (obtenidas de red_servidor.go) ───────────────────────
	ips := obtenerIPs() // puede devolver varias interfaces

	// Línea resumen para la barra de info
	ipResumen := strings.Join(ips, "  |  ")

	infoLbl := canvas.NewText(
		fmt.Sprintf("  OS: %s    Puerto: %s    Log: %s", sistemaOperativo(), puertoServidor, archivoLog),
		colMuted,
	)
	infoLbl.TextStyle = fyne.TextStyle{Monospace: true}
	infoLbl.TextSize = 11

	// Label destacado con las IPs para que el operador las comparta fácilmente
	ipTitulo := canvas.NewText("  IP(s) para conectar clientes:", colAccent)
	ipTitulo.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	ipTitulo.TextSize = 12

	ipValor := canvas.NewText("  "+ipResumen, colPrimary)
	ipValor.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	ipValor.TextSize = 14

	panelIP := container.NewVBox(
		container.NewPadded(infoLbl),
		container.NewPadded(ipTitulo),
		container.NewPadded(ipValor),
	)

	// ── Log en vivo ───────────────────────────────────────────────────────
	logLabel := widget.NewLabel("")
	logLabel.TextStyle = fyne.TextStyle{Monospace: true}
	logLabel.Wrapping = fyne.TextWrapWord
	scroll := container.NewVScroll(logLabel)

	var lineas []string
	// onLog es llamado desde goroutines; canal para serializar al hilo GUI
	logCh := make(chan string, 64)
	go func() {
		for msg := range logCh {
			lineas = append(lineas, "  "+msg)
			logLabel.SetText(joinLines(lineas))
			scroll.ScrollToBottom()
		}
	}()
	onLog := func(msg string) { logCh <- msg }

	onLog("═══════════════════════════════════════════════════════════")
	onLog("  ShellOS  —  Servidor TCP iniciando")
	onLog("═══════════════════════════════════════════════════════════")
	for _, ip := range ips {
		onLog(fmt.Sprintf("  ◈  IP disponible:  %s%s", ip, puertoServidor))
	}
	onLog("═══════════════════════════════════════════════════════════")

	// ── Iniciar servidor ──────────────────────────────────────────────────
	err := iniciarServidor(onLog, stop)
	if err != nil {
		estadoLbl.Text = "● ERROR"
		estadoLbl.Color = colError
		estadoLbl.Refresh()
		onLog("✗  " + err.Error())
	} else {
		estadoLbl.Text = fmt.Sprintf("● ESCUCHANDO  %s", puertoServidor)
		estadoLbl.Color = colGreen
		estadoLbl.Refresh()
	}

	// ── Layout ────────────────────────────────────────────────────────────
	top := container.NewVBox(barTitulo, hRule(), panelIP, hRule())

	w.SetContent(container.NewBorder(
		top, nil, nil, nil,
		container.NewPadded(scroll),
	))
	w.Show()
}

// ══════════════════════════════════════════════════════════════════════════════
//  HELPERS
// ══════════════════════════════════════════════════════════════════════════════

func joinLines(ls []string) string {
	sb := ""
	for i, l := range ls {
		if i > 0 {
			sb += "\n"
		}
		sb += l
	}
	return sb
}

// mostrarPantallaBloqueo reemplaza el contenido de la ventana con un aviso
// de acceso bloqueado tras agotar los intentos permitidos.
func mostrarPantallaBloqueo(w fyne.Window, intentosFallidos int) {
	blk1 := canvas.NewText("⛔  ACCESO BLOQUEADO", colError)
	blk1.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	blk1.TextSize = 18
	blk1.Alignment = fyne.TextAlignCenter

	blk2 := canvas.NewText("usuario sospechoso — plataforma cerrada", colMuted)
	blk2.TextStyle = fyne.TextStyle{Monospace: true}
	blk2.TextSize = 12
	blk2.Alignment = fyne.TextAlignCenter

	blk3 := canvas.NewText(fmt.Sprintf("[ %d intentos fallidos ]", intentosFallidos), colMuted)
	blk3.TextStyle = fyne.TextStyle{Monospace: true}
	blk3.TextSize = 11
	blk3.Alignment = fyne.TextAlignCenter

	w.SetContent(container.NewCenter(container.NewVBox(
		spacer(), spacer(),
		container.NewCenter(blk1), spacer(),
		container.NewCenter(blk2), spacer(),
		container.NewCenter(blk3),
		spacer(), spacer(),
	)))
}

// ══════════════════════════════════════════════════════════════════════════════
//  MAIN
// ══════════════════════════════════════════════════════════════════════════════

func main() {
	a := app.New()
	a.Settings().SetTheme(&shellTheme{})
	mostrarSeleccionModo(a)
}