package main

// red_cliente.go
// Cliente TCP de ShellOS.
// Muestra una ventana de terminal donde el usuario escribe comandos que se
// envían al servidor y cuya salida se recibe y muestra en pantalla.
// No ejecuta nada localmente: todo ocurre en el servidor.

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ── Conexión al servidor ──────────────────────────────────────────────────────

// ConexionServidor encapsula el socket y el reader/writer hacia el servidor.
type ConexionServidor struct {
	conn    net.Conn
	scanner *bufio.Scanner
}

// conectarServidor intenta conectarse al servidor TCP.
// Devuelve error si no lo logra en 5 segundos.
func conectarServidor(host string) (*ConexionServidor, error) {
	addr := host + puertoServidor // puertoServidor definido en red_servidor.go
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar a %s: %w", addr, err)
	}
	return &ConexionServidor{
		conn:    conn,
		scanner: bufio.NewScanner(conn),
	}, nil
}

// enviarComando manda una línea al servidor y recoge la respuesta completa
// (el servidor termina cada respuesta con "<<<END>>>").
func (c *ConexionServidor) enviarComando(cmd string) (string, error) {
	_, err := fmt.Fprintf(c.conn, "%s\n", cmd)
	if err != nil {
		return "", fmt.Errorf("error al enviar: %w", err)
	}

	var sb strings.Builder
	for c.scanner.Scan() {
		linea := c.scanner.Text()
		if linea == "<<<END>>>" {
			break
		}
		sb.WriteString(linea)
		sb.WriteByte('\n')
	}
	if err := c.scanner.Err(); err != nil {
		return sb.String(), fmt.Errorf("error al leer respuesta: %w", err)
	}
	return strings.TrimRight(sb.String(), "\n"), nil
}

// cerrar cierra la conexión TCP.
func (c *ConexionServidor) cerrar() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// ── Ventana de conexión (pide IP del servidor) ────────────────────────────────

// mostrarVentanaConexion muestra un diálogo para ingresar la IP del servidor.
// Al conectar exitosamente abre la terminal remota.
func mostrarVentanaConexion(a fyne.App) {
	w := a.NewWindow("ShellOS — Conectar al servidor")
	w.Resize(fyne.NewSize(420, 320))
	w.CenterOnScreen()

	// ── Cabecera ──────────────────────────────────────────────────────────
	titulo := canvas.NewText("MODO CLIENTE", colPrimary)
	titulo.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	titulo.TextSize = 22
	titulo.Alignment = fyne.TextAlignCenter

	sub := canvas.NewText("conexión remota vía TCP", colMuted)
	sub.TextStyle = fyne.TextStyle{Monospace: true}
	sub.TextSize = 11
	sub.Alignment = fyne.TextAlignCenter

	// ── Campo IP ──────────────────────────────────────────────────────────
	lblIP := canvas.NewText("IP DEL SERVIDOR", colMuted)
	lblIP.TextStyle = fyne.TextStyle{Monospace: true}
	lblIP.TextSize = 10

	inpIP := widget.NewEntry()
	inpIP.SetPlaceHolder("192.168.1.x  o  localhost")
	inpIP.TextStyle = fyne.TextStyle{Monospace: true}
	inpIP.Text = "localhost"

	// ── Estado ────────────────────────────────────────────────────────────
	msgErr := canvas.NewText("", colError)
	msgErr.TextStyle = fyne.TextStyle{Monospace: true}
	msgErr.TextSize = 11
	msgErr.Alignment = fyne.TextAlignCenter

	// ── Botón conectar ────────────────────────────────────────────────────
	btnConectar := widget.NewButton("[ CONECTAR ]", nil)
	btnConectar.Importance = widget.HighImportance

	accion := func() {
		host := strings.TrimSpace(inpIP.Text)
		if host == "" {
			msgErr.Text = "✗  ingresa una IP o hostname"
			msgErr.Refresh()
			return
		}
		msgErr.Text = "  conectando..."
		msgErr.Color = colMuted
		msgErr.Refresh()

		cs, err := conectarServidor(host)
		if err != nil {
			msgErr.Text = "✗  " + err.Error()
			msgErr.Color = colError
			msgErr.Refresh()
			return
		}

		w.Hide()
		mostrarTerminalCliente(a, cs, host)
	}

	btnConectar.OnTapped = accion
	inpIP.OnSubmitted = func(_ string) { accion() }

	// ── Layout ────────────────────────────────────────────────────────────
	cuerpo := container.NewVBox(
		spacer(),
		container.NewCenter(titulo),
		container.NewCenter(sub),
		spacer(),
		hRule(),
		spacer(),
		lblIP, inpIP,
		spacer(),
		container.NewCenter(btnConectar),
		spacer(),
		container.NewCenter(msgErr),
	)
	w.SetContent(container.NewPadded(container.NewPadded(cuerpo)))
	w.Canvas().Focus(inpIP)
	w.Show()
}

// ── Terminal remota ───────────────────────────────────────────────────────────

// mostrarTerminalCliente abre la ventana de terminal del cliente.
// Los comandos se envían al servidor y la respuesta se muestra en el historial.
func mostrarTerminalCliente(a fyne.App, cs *ConexionServidor, host string) {
	w := a.NewWindow(fmt.Sprintf("ShellOS  —  cliente  →  %s", host))
	w.Resize(fyne.NewSize(980, 680))
	w.CenterOnScreen()
	w.SetOnClosed(func() { cs.cerrar() })

	// ── Barra de título ───────────────────────────────────────────────────
	titBar := canvas.NewText(
		fmt.Sprintf("▸ ShellOS  —  terminal remota  [ servidor: %s%s ]", host, puertoServidor),
		colAccent,
	)
	titBar.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	titBar.TextSize = 12

	estadoLbl := canvas.NewText("● CONECTADO", colGreen)
	estadoLbl.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	estadoLbl.TextSize = 11

	barTitulo := container.NewBorder(nil, nil, nil,
		container.NewPadded(estadoLbl),
		container.NewPadded(titBar),
	)

	// ── Historial ─────────────────────────────────────────────────────────
	historial := widget.NewLabel("")
	historial.TextStyle = fyne.TextStyle{Monospace: true}
	historial.Wrapping = fyne.TextWrapWord
	scroll := container.NewVScroll(historial)

	var lineas []string
	agregar := func(linea string) {
		lineas = append(lineas, linea)
		historial.SetText(strings.Join(lineas, "\n"))
		scroll.ScrollToBottom()
	}

	agregar("  ╔═══════════════════════════════════════════════╗")
	agregar("  ║        S H E L L O S  —  C L I E N T E       ║")
	agregar(fmt.Sprintf("  ║  servidor: %-36s║", host+puertoServidor))
	agregar("  ║  los comandos se ejecutan en el servidor      ║")
	agregar("  ╚═══════════════════════════════════════════════╝")
	agregar("")

	// ── Entrada de comandos ───────────────────────────────────────────────
	promptLbl := labelMuted(host + "  »")

	entrada := widget.NewEntry()
	entrada.SetPlaceHolder("comando remoto…")
	entrada.TextStyle = fyne.TextStyle{Monospace: true}

	procesar := func() {
		texto := strings.TrimSpace(entrada.Text)
		entrada.SetText("")
		if texto == "" {
			return
		}

		agregar(fmt.Sprintf("  %s  » %s", host, texto))

		salida, err := cs.enviarComando(texto)
		if err != nil {
			agregar("  ✗ error de red: " + err.Error())
			estadoLbl.Text = "● DESCONECTADO"
			estadoLbl.Color = colError
			estadoLbl.Refresh()
			entrada.Disable()
			agregar("")
			return
		}

		for _, l := range strings.Split(salida, "\n") {
			agregar("  " + l)
		}
		agregar("")

		// Si el servidor cerró la sesión
		if texto == "bye" || texto == "exit" {
			estadoLbl.Text = "● DESCONECTADO"
			estadoLbl.Color = colError
			estadoLbl.Refresh()
			entrada.Disable()
		}
	}

	entrada.OnSubmitted = func(_ string) { procesar() }

	btnEjec := widget.NewButton("↵ Ejecutar", func() { procesar() })
	btnEjec.Importance = widget.HighImportance

	barCmd := container.NewBorder(nil, nil, promptLbl, btnEjec, entrada)

	// ── Layout ────────────────────────────────────────────────────────────
	top := container.NewVBox(
		barTitulo,
		hRule(),
	)
	bottom := container.NewVBox(
		hRule(),
		container.NewPadded(barCmd),
	)

	w.SetContent(container.NewBorder(
		top, bottom, nil, nil,
		container.NewPadded(scroll),
	))
	w.Canvas().Focus(entrada)
	w.Show()
}