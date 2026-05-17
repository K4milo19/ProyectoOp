package main

// reporte.go
// Lectura de métricas del sistema operativo (CPU, RAM, Red) y goroutine
// que actualiza los labels de la GUI cada 5 segundos.
// La lógica de lectura no depende de GUI; solo los parámetros de reporte() sí.

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2/canvas"
)

// ── Estructura de datos ───────────────────────────────────────────────────────

// MetricasSistema agrupa los valores leídos en un ciclo de monitoreo.
type MetricasSistema struct {
	CPUPorcentaje float64
	RAMUsadaMB    float64
	RAMTotalMB    float64
	RAMPorcentaje float64
	NetRxBytes    uint64
	NetTxBytes    uint64
}

// ── Estado interno de CPU (delta entre lecturas) ──────────────────────────────

var (
	prevIdle  uint64
	prevTotal uint64
)

// ── Lecturas de /proc ─────────────────────────────────────────────────────────

// leerCPU calcula el porcentaje de uso de CPU desde la última llamada
// comparando los contadores acumulados de /proc/stat.
func leerCPU() float64 {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		linea := scanner.Text()
		if !strings.HasPrefix(linea, "cpu ") {
			continue
		}
		campos := strings.Fields(linea)
		if len(campos) < 8 {
			break
		}
		parse := func(s string) uint64 {
			v, _ := strconv.ParseUint(s, 10, 64)
			return v
		}
		user    := parse(campos[1])
		nice    := parse(campos[2])
		system  := parse(campos[3])
		idle    := parse(campos[4])
		iowait  := parse(campos[5])
		irq     := parse(campos[6])
		softirq := parse(campos[7])

		idleNow  := idle + iowait
		totalNow := user + nice + system + idle + iowait + irq + softirq

		deltaIdle  := idleNow - prevIdle
		deltaTotal := totalNow - prevTotal

		prevIdle  = idleNow
		prevTotal = totalNow

		if deltaTotal == 0 {
			return 0
		}
		return (1.0 - float64(deltaIdle)/float64(deltaTotal)) * 100.0
	}
	return 0
}

// leerRAM obtiene la memoria usada, total (en MB) y el porcentaje desde /proc/meminfo.
func leerRAM() (usadaMB, totalMB, porcentaje float64) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer f.Close()

	vals := map[string]uint64{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		partes := strings.Fields(scanner.Text())
		if len(partes) < 2 {
			continue
		}
		clave := strings.TrimSuffix(partes[0], ":")
		v, _ := strconv.ParseUint(partes[1], 10, 64)
		vals[clave] = v
	}

	total        := vals["MemTotal"]
	libre         := vals["MemFree"]
	buffers      := vals["Buffers"]
	cached        := vals["Cached"]
	sreclaimable := vals["SReclaimable"]

	disponible := libre + buffers + cached + sreclaimable
	usada       := total - disponible

	totalMB = float64(total) / 1024.0
	usadaMB = float64(usada) / 1024.0
	if total > 0 {
		porcentaje = float64(usada) / float64(total) * 100.0
	}
	return
}

// leerRed suma los bytes recibidos y transmitidos de todas las interfaces
// de red (excepto loopback) desde /proc/net/dev.
func leerRed() (rxBytes, txBytes uint64) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan() // cabecera 1
	scanner.Scan() // cabecera 2
	for scanner.Scan() {
		linea := strings.TrimSpace(scanner.Text())
		idx := strings.Index(linea, ":")
		if idx < 0 {
			continue
		}
		iface := strings.TrimSpace(linea[:idx])
		if iface == "lo" {
			continue
		}
		campos := strings.Fields(linea[idx+1:])
		if len(campos) < 9 {
			continue
		}
		rx, _ := strconv.ParseUint(campos[0], 10, 64)
		tx, _ := strconv.ParseUint(campos[8], 10, 64)
		rxBytes += rx
		txBytes += tx
	}
	return
}

// ── Combinación ───────────────────────────────────────────────────────────────

// obtenerMetricas ejecuta las tres lecturas y las empaqueta en MetricasSistema.
func obtenerMetricas() MetricasSistema {
	cpu := leerCPU()
	usada, total, ramPct := leerRAM()
	rx, tx := leerRed()
	return MetricasSistema{
		CPUPorcentaje: cpu,
		RAMUsadaMB:    usada,
		RAMTotalMB:    total,
		RAMPorcentaje: ramPct,
		NetRxBytes:    rx,
		NetTxBytes:    tx,
	}
}

// ── Helpers de formato ────────────────────────────────────────────────────────

// barraTexto devuelve una barra ASCII proporcional al porcentaje dado.
//
//	Ejemplo: barraTexto(65, 20) → "[█████████████░░░░░░░]"
func barraTexto(pct float64, ancho int) string {
	llenos := int(pct / 100.0 * float64(ancho))
	if llenos > ancho {
		llenos = ancho
	}
	return "[" + strings.Repeat("█", llenos) + strings.Repeat("░", ancho-llenos) + "]"
}

// formatBytes convierte bytes crudos a una cadena legible (B, KB, MB, GB).
func formatBytes(b uint64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.2f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.2f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.2f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// ── Goroutine de reporte ──────────────────────────────────────────────────────

// reporte lanza un goroutine que actualiza los tres canvas.Text con las métricas
// del sistema cada 5 segundos. Se detiene cuando se cierra el canal stop.
//
// Parámetros:
//   - lblCPU, lblRAM, lblRed: labels de la ventana principal a actualizar.
//   - stop: canal que se cierra cuando el usuario cierra la ventana.
func reporte(lblCPU, lblRAM, lblRed *canvas.Text, stop <-chan struct{}) {
	// Primera lectura para inicializar prevIdle/prevTotal (el primer delta sería 0).
	obtenerMetricas()

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				m := obtenerMetricas()

				lblCPU.Text = fmt.Sprintf(
					"CPU  %s  %5.1f%%",
					barraTexto(m.CPUPorcentaje, 20),
					m.CPUPorcentaje,
				)
				lblCPU.Refresh()

				lblRAM.Text = fmt.Sprintf(
					"RAM  %s  %5.1f%%   (%s / %s)",
					barraTexto(m.RAMPorcentaje, 20),
					m.RAMPorcentaje,
					fmt.Sprintf("%.0f MB", m.RAMUsadaMB),
					fmt.Sprintf("%.0f MB", m.RAMTotalMB),
				)
				lblRAM.Refresh()

				lblRed.Text = fmt.Sprintf(
					"NET  ↓ %-12s   ↑ %-12s",
					formatBytes(m.NetRxBytes),
					formatBytes(m.NetTxBytes),
				)
				lblRed.Refresh()
			}
		}
	}()
}