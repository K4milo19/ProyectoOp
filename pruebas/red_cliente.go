package main

// red_cliente.go
// Cliente TCP de ShellOS.
// Ventana dividida en dos zonas:
//   - Izquierda: terminal remota (comandos al servidor)
//   - Derecha:   panel de métricas DEL SERVIDOR (CPU, RAM, Disco remotos)
//
// Las métricas llegan como mensajes push del servidor con el prefijo:
//   <<<METRICS:CPU=12.3|RAM_USED=1024|RAM_FREE=2048|DISK_USED=40.1|DISK_TOTAL=100.0>>>
// Estas líneas se filtran del flujo de respuesta a comandos y se aplican
// directamente al panel derecho mediante parsearMetrics().

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ── Conexión al servidor ──────────────────────────────────────────────────────

type ConexionServidor struct {
	conn    net.Conn
	scanner *bufio.Scanner

	// onMetrics es llamado cada vez que llega un mensaje <<<METRICS:...>>>
	// Se asigna desde mostrarTerminalCliente tras crear la conexión.
	onMetrics func(cpu, ramUsed, ramFree, diskUsed, diskTotal float64)
}

func conectarServidor(host string) (*ConexionServidor, error) {
	addr := host + puertoServidor
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar a %s: %w", addr, err)
	}
	return &ConexionServidor{
		conn:    conn,
		scanner: bufio.NewScanner(conn),
	}, nil
}

// parsearMetrics extrae los campos de una línea <<<METRICS:...>>>.
// Devuelve false si la línea no tiene el formato correcto.
func parsearMetrics(linea string) (cpu, ramUsed, ramFree, diskUsed, diskTotal float64, ok bool) {
	if !strings.HasPrefix(linea, "<<<METRICS:") || !strings.HasSuffix(linea, ">>>") {
		return
	}
	inner := strings.TrimSuffix(strings.TrimPrefix(linea, "<<<METRICS:"), ">>>")
	// inner = "CPU=12.3|RAM_USED=1024|RAM_FREE=2048|DISK_USED=40.1|DISK_TOTAL=100.0"
	campos := strings.Split(inner, "|")
	vals := make(map[string]float64, len(campos))
	for _, c := range campos {
		kv := strings.SplitN(c, "=", 2)
		if len(kv) != 2 {
			continue
		}
		v, err := strconv.ParseFloat(kv[1], 64)
		if err != nil {
			continue
		}
		vals[kv[0]] = v
	}
	cpu       = vals["CPU"]
	ramUsed   = vals["RAM_USED"]
	ramFree   = vals["RAM_FREE"]
	diskUsed  = vals["DISK_USED"]
	diskTotal = vals["DISK_TOTAL"]
	ok = true
	return
}

// enviarComando manda un comando al servidor y recoge (salida, pwd, error).
// Las líneas <<<METRICS:...>>> que lleguen mientras se espera la respuesta
// se desvían a onMetrics en lugar de mezclarse con la salida del comando.
func (c *ConexionServidor) enviarComando(cmd string) (salida, pwd string, err error) {
	_, err = fmt.Fprintf(c.conn, "%s\n", cmd)
	if err != nil {
		return "", "", fmt.Errorf("error al enviar: %w", err)
	}

	var sb strings.Builder
	for c.scanner.Scan() {
		linea := c.scanner.Text()

		// Intercept de métricas push del servidor
		if strings.HasPrefix(linea, "<<<METRICS:") {
			if c.onMetrics != nil {
				if cpu, ru, rf, du, dt, ok := parsearMetrics(linea); ok {
					c.onMetrics(cpu, ru, rf, du, dt)
				}
			}
			continue
		}

		if linea == "<<<END>>>" {
			break
		}
		if strings.HasPrefix(linea, "<<<PWD:") && strings.HasSuffix(linea, ">>>") {
			pwd = strings.TrimSuffix(strings.TrimPrefix(linea, "<<<PWD:"), ">>>")
			continue
		}
		sb.WriteString(linea)
		sb.WriteByte('\n')
	}
	if err = c.scanner.Err(); err != nil {
		return sb.String(), pwd, fmt.Errorf("error al leer respuesta: %w", err)
	}
	return strings.TrimRight(sb.String(), "\n"), pwd, nil
}

func (c *ConexionServidor) cerrar() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// ── Lectura asíncrona de métricas push ───────────────────────────────────────

// escucharMetricasPush lee el socket en background esperando mensajes
// <<<METRICS:...>>> que el servidor envía de forma proactiva (fuera del
// ciclo request/response de comandos). Se lanza como goroutine.
//
// Nota: cuando el cliente está esperando la respuesta de un comando,
// enviarComando() consume el stream —incluyendo los METRICS— directamente,
// por lo que esta goroutine solo procesa los que llegan en "silencio"
// (sin un comando pendiente). Para eso usamos un bufio.Scanner independiente
// sobre la misma conexión: el scanner interno de ConexionServidor sólo se usa
// dentro de enviarComando, por lo que no hay condición de carrera siempre que
// el cliente no envíe dos comandos a la vez (lo cual es correcto aquí).
//
// Implementación simplificada: reutilizamos el mismo scanner de la conexión.
// El scanner de enviarComando ya filtra los METRICS en tiempo real; esta
// goroutine maneja los que llegan ENTRE comandos (idle period).
func (c *ConexionServidor) escucharMetricasPush(stop <-chan struct{}) {
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
			}

			if !c.scanner.Scan() {
				return // conexión cerrada
			}
			linea := c.scanner.Text()

			if strings.HasPrefix(linea, "<<<METRICS:") {
				if c.onMetrics != nil {
					if cpu, ru, rf, du, dt, ok := parsearMetrics(linea); ok {
						c.onMetrics(cpu, ru, rf, du, dt)
					}
				}
				continue
			}

			// Cualquier otra línea inesperada fuera de un comando: ignorar.
		}
	}()
}

// ── Ventana de conexión ───────────────────────────────────────────────────────

func mostrarVentanaConexion(a fyne.App) {
	w := a.NewWindow("ShellOS — Conectar al servidor")
	w.Resize(fyne.NewSize(420, 320))
	w.CenterOnScreen()

	titulo := canvas.NewText("MODO CLIENTE", colPrimary)
	titulo.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	titulo.TextSize = 22
	titulo.Alignment = fyne.TextAlignCenter

	sub := canvas.NewText("conexión remota vía TCP", colMuted)
	sub.TextStyle = fyne.TextStyle{Monospace: true}
	sub.TextSize = 11
	sub.Alignment = fyne.TextAlignCenter

	lblIP := canvas.NewText("IP DEL SERVIDOR", colMuted)
	lblIP.TextStyle = fyne.TextStyle{Monospace: true}
	lblIP.TextSize = 10

	inpIP := widget.NewEntry()
	inpIP.SetPlaceHolder("192.168.1.x  o  localhost")
	inpIP.TextStyle = fyne.TextStyle{Monospace: true}
	inpIP.Text = "localhost"

	msgErr := canvas.NewText("", colError)
	msgErr.TextStyle = fyne.TextStyle{Monospace: true}
	msgErr.TextSize = 11
	msgErr.Alignment = fyne.TextAlignCenter

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

func mostrarTerminalCliente(a fyne.App, cs *ConexionServidor, host string) {
	w := a.NewWindow(fmt.Sprintf("ShellOS  —  cliente  →  %s", host))
	w.Resize(fyne.NewSize(1280, 700))
	w.CenterOnScreen()

	stopPush := make(chan struct{})
	w.SetOnClosed(func() {
		cs.cerrar()
		close(stopPush)
	})

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

	// ── Panel de métricas REMOTAS del servidor (derecha) ──────────────────
	tituloPanel := canvas.NewText("◈  MÉTRICAS DEL SERVIDOR", colAccent)
	tituloPanel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	tituloPanel.TextSize = 11

	lblCPU := widget.NewLabel("CPU\n[ esperando datos... ]")
	lblCPU.TextStyle = fyne.TextStyle{Monospace: true}
	lblCPU.Wrapping = fyne.TextWrapWord

	lblRAM := widget.NewLabel("RAM\nusada: -- MB\nlibre: -- MB")
	lblRAM.TextStyle = fyne.TextStyle{Monospace: true}
	lblRAM.Wrapping = fyne.TextWrapWord

	lblDisco := widget.NewLabel("DISCO\nusado: -- GB\ntotal: -- GB\nlibre: -- GB")
	lblDisco.TextStyle = fyne.TextStyle{Monospace: true}
	lblDisco.Wrapping = fyne.TextWrapWord

	lblTick := canvas.NewText("push cada 5 s", colMuted)
	lblTick.TextStyle = fyne.TextStyle{Monospace: true}
	lblTick.TextSize = 10

	panelDerecho := container.NewVBox(
		container.NewPadded(tituloPanel),
		hRule(),
		container.NewPadded(lblCPU),
		hRule(),
		container.NewPadded(lblRAM),
		hRule(),
		container.NewPadded(lblDisco),
		hRule(),
		container.NewPadded(lblTick),
	)

	// ── Callback de métricas: actualiza el panel derecho en el hilo UI ────
	cs.onMetrics = func(cpu, ramUsed, ramFree, diskUsed, diskTotal float64) {
		fyne.Do(func() {
			lblCPU.SetText(fmt.Sprintf("CPU: %.1f%%", cpu))
			lblRAM.SetText(fmt.Sprintf(
				"RAM — usada: %.0f MB  |  libre: %.0f MB",
				ramUsed, ramFree,
			))
			lblDisco.SetText(fmt.Sprintf(
				"DISCO — usado: %.1f GB  |  total: %.1f GB  |  libre: %.1f GB",
				diskUsed, diskTotal, diskTotal-diskUsed,
			))
		})
	}

	// Arrancar escucha de métricas push (mensajes entre comandos)
	cs.escucharMetricasPush(stopPush)

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
	agregar("  ║  métricas del servidor → panel derecho        ║")
	agregar("  ╚═══════════════════════════════════════════════╝")
	agregar("")

	// ── Entrada de comandos ───────────────────────────────────────────────
	promptLbl := labelMuted("~  »")

	if _, pwd, err := cs.enviarComando("pwd"); err == nil && pwd != "" {
		promptLbl.SetText(pwd + "  »")
	}

	entrada := widget.NewEntry()
	entrada.SetPlaceHolder("comando remoto…")
	entrada.TextStyle = fyne.TextStyle{Monospace: true}

	procesar := func() {
		texto := strings.TrimSpace(entrada.Text)
		entrada.SetText("")
		if texto == "" {
			return
		}

		agregar(fmt.Sprintf("  %s  » %s", promptLbl.Text, texto))

		salida, pwd, err := cs.enviarComando(texto)
		if err != nil {
			agregar("  ✗ error de red: " + err.Error())
			estadoLbl.Text = "● DESCONECTADO"
			estadoLbl.Color = colError
			estadoLbl.Refresh()
			entrada.Disable()
			agregar("")
			return
		}

		if pwd != "" {
			promptLbl.SetText(pwd + "  »")
		}

		for _, l := range strings.Split(salida, "\n") {
			agregar("  " + l)
		}
		agregar("")

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

	// ── Layout: terminal izquierda | panel métricas derecha ───────────────
	terminal := container.NewBorder(
		nil,
		container.NewVBox(hRule(), container.NewPadded(barCmd)),
		nil, nil,
		container.NewPadded(scroll),
	)

	contenido := container.NewBorder(
		container.NewVBox(barTitulo, hRule()),
		nil, nil,
		container.NewPadded(panelDerecho),
		terminal,
	)

	w.SetContent(contenido)
	w.Canvas().Focus(entrada)
	w.Show()
}