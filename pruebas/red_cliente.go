package main

// red_cliente.go
// Cliente TCP de ShellOS.
// Ventana dividida en dos zonas:
//   - Izquierda: terminal remota (comandos al servidor)
//   - Derecha:   panel de métricas DEL SERVIDOR (push cada 5 s)
//
// Arquitectura de lectura:
//   Un único goroutine (lectorSocket) lee todas las líneas del socket.
//   Clasifica cada línea:
//     · <<<METRICS:...>>>  → canal metricas (no bloqueante)
//     · cualquier otra    → canal respuestas (leído por enviarComando)
//   Así no hay race condition entre lecturas concurrentes.

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
	conn      net.Conn
	// respuestas recibe todas las líneas del socket que NO son métricas.
	// enviarComando lee de este canal.
	respuestas chan string
	// metricas recibe las líneas <<<METRICS:...>>> del servidor.
	metricas   chan string
	// onMetrics se asigna desde mostrarTerminalCliente para actualizar la UI.
	onMetrics  func(cpu, ramUsed, ramFree, diskUsed, diskTotal float64)
}

func conectarServidor(host string) (*ConexionServidor, error) {
	addr := host + puertoServidor
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar a %s: %w", addr, err)
	}
	cs := &ConexionServidor{
		conn:       conn,
		respuestas: make(chan string, 256),
		metricas:   make(chan string, 16),
	}
	go cs.lectorSocket()
	return cs, nil
}

// lectorSocket es el ÚNICO goroutine que lee del socket.
// Clasifica cada línea y la envía al canal correspondiente.
// Se detiene cuando la conexión se cierra.
func (c *ConexionServidor) lectorSocket() {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		linea := scanner.Text()
		if strings.HasPrefix(linea, "<<<METRICS:") {
			// Envío no bloqueante: si nadie lee métricas, se descarta
			select {
			case c.metricas <- linea:
			default:
			}
		} else {
			c.respuestas <- linea
		}
	}
	// Conexión cerrada: cerrar canales para que enviarComando no quede colgado
	close(c.respuestas)
	close(c.metricas)
}

// despachadorMetricas lee el canal de métricas y llama a onMetrics.
// Se lanza como goroutine separada desde mostrarTerminalCliente.
func (c *ConexionServidor) despachadorMetricas(stop <-chan struct{}) {
	go func() {
		for {
			select {
			case <-stop:
				return
			case linea, ok := <-c.metricas:
				if !ok {
					return // canal cerrado (conexión caída)
				}
				if c.onMetrics != nil {
					if cpu, ru, rf, du, dt, ok2 := parsearMetrics(linea); ok2 {
						c.onMetrics(cpu, ru, rf, du, dt)
					}
				}
			}
		}
	}()
}

// parsearMetrics extrae los 5 valores de una línea <<<METRICS:...>>>.
func parsearMetrics(linea string) (cpu, ramUsed, ramFree, diskUsed, diskTotal float64, ok bool) {
	if !strings.HasPrefix(linea, "<<<METRICS:") || !strings.HasSuffix(linea, ">>>") {
		return
	}
	inner := strings.TrimSuffix(strings.TrimPrefix(linea, "<<<METRICS:"), ">>>")
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
// Lee exclusivamente del canal c.respuestas (lectorSocket ya separó las métricas).
func (c *ConexionServidor) enviarComando(cmd string) (salida, pwd string, err error) {
	_, err = fmt.Fprintf(c.conn, "%s\n", cmd)
	if err != nil {
		return "", "", fmt.Errorf("error al enviar: %w", err)
	}

	var sb strings.Builder
	for linea := range c.respuestas {
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
	return strings.TrimRight(sb.String(), "\n"), pwd, nil
}

func (c *ConexionServidor) cerrar() {
	if c.conn != nil {
		c.conn.Close()
	}
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

	stopMetrics := make(chan struct{})
	w.SetOnClosed(func() {
		cs.cerrar()
		close(stopMetrics)
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

	lblRAM := widget.NewLabel("RAM\nusada: -- MB  |  libre: -- MB")
	lblRAM.TextStyle = fyne.TextStyle{Monospace: true}
	lblRAM.Wrapping = fyne.TextWrapWord

	lblDisco := widget.NewLabel("DISCO\nusado: -- GB  |  total: -- GB  |  libre: -- GB")
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

	// Callback que actualiza el panel derecho desde cualquier goroutine
	cs.onMetrics = func(cpu, ramUsed, ramFree, diskUsed, diskTotal float64) {
		fyne.Do(func() {
			lblCPU.SetText(fmt.Sprintf("CPU: %.1f%%", cpu))
			lblRAM.SetText(fmt.Sprintf(
				"RAM — usada: %.0f MB  |  libre: %.0f MB", ramUsed, ramFree,
			))
			lblDisco.SetText(fmt.Sprintf(
				"DISCO — usado: %.1f GB  |  total: %.1f GB  |  libre: %.1f GB",
				diskUsed, diskTotal, diskTotal-diskUsed,
			))
		})
	}

	// Arrancar despachador de métricas (lee canal, llama onMetrics)
	cs.despachadorMetricas(stopMetrics)

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

	// pwd inicial: se ejecuta en goroutine para no bloquear el hilo de UI
	go func() {
		if _, pwd, err := cs.enviarComando("pwd"); err == nil && pwd != "" {
			fyne.Do(func() { promptLbl.SetText(pwd + "  »") })
		}
	}()

	entrada := widget.NewEntry()
	entrada.SetPlaceHolder("comando remoto…")
	entrada.TextStyle = fyne.TextStyle{Monospace: true}

	// cmdMu evita enviar dos comandos simultáneos (la UI es secuencial,
	// pero el botón podría pulsarse dos veces rápido)
	var cmdEnCurso bool

	procesar := func() {
		if cmdEnCurso {
			return
		}
		texto := strings.TrimSpace(entrada.Text)
		entrada.SetText("")
		if texto == "" {
			return
		}

		cmdEnCurso = true
		agregar(fmt.Sprintf("  %s  » %s", promptLbl.Text, texto))

		// Ejecutar en goroutine para no congelar la UI mientras espera respuesta
		go func() {
			salida, pwd, err := cs.enviarComando(texto)
			fyne.Do(func() {
				cmdEnCurso = false
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
			})
		}()
	}

	entrada.OnSubmitted = func(_ string) { procesar() }

	btnEjec := widget.NewButton("↵ Ejecutar", func() { procesar() })
	btnEjec.Importance = widget.HighImportance

	barCmd := container.NewBorder(nil, nil, promptLbl, btnEjec, entrada)

	// ── Layout ────────────────────────────────────────────────────────────
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