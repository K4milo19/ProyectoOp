package main

// reporte.go
// Lectura de métricas del sistema — solo Linux (/proc).
// Las tres lecturas (CPU, RAM, Red) se ejecutan en goroutines paralelas
// y sus resultados se combinan con un WaitGroup antes de actualizar la GUI.

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2/canvas"
)

// ── Estructura de datos ───────────────────────────────────────────────────────

type MetricasSistema struct {
	CPUPorcentaje float64
	RAMUsadaMB    float64
	RAMTotalMB    float64
	RAMPorcentaje float64
	NetRxBytes    uint64
	NetTxBytes    uint64
}

// estado interno para el cálculo delta de CPU entre lecturas
var (
	prevIdle  uint64
	prevTotal uint64
	cpuMu     sync.Mutex // protege prevIdle y prevTotal
)

// ── Lecturas /proc ────────────────────────────────────────────────────────────

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

		cpuMu.Lock()
		deltaIdle  := idleNow - prevIdle
		deltaTotal := totalNow - prevTotal
		prevIdle   = idleNow
		prevTotal  = totalNow
		cpuMu.Unlock()

		if deltaTotal == 0 {
			return 0
		}
		return (1.0 - float64(deltaIdle)/float64(deltaTotal)) * 100.0
	}
	return 0
}

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
	libre        := vals["MemFree"]
	buffers      := vals["Buffers"]
	cached       := vals["Cached"]
	sreclaimable := vals["SReclaimable"]

	disponible := libre + buffers + cached + sreclaimable
	usada      := total - disponible

	totalMB = float64(total) / 1024.0
	usadaMB = float64(usada) / 1024.0
	if total > 0 {
		porcentaje = float64(usada) / float64(total) * 100.0
	}
	return
}

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

// ── obtenerMetricas — lecturas paralelas con goroutines ───────────────────────

// obtenerMetricas lanza las tres lecturas de /proc en goroutines separadas
// y espera a que todas terminen antes de devolver el resultado combinado.
func obtenerMetricas() MetricasSistema {
	var (
		m   MetricasSistema
		wg  sync.WaitGroup
		mu  sync.Mutex
	)

	wg.Add(3)

	// Goroutine 1: CPU
	go func() {
		defer wg.Done()
		v := leerCPU()
		mu.Lock()
		m.CPUPorcentaje = v
		mu.Unlock()
	}()

	// Goroutine 2: RAM
	go func() {
		defer wg.Done()
		usada, total, pct := leerRAM()
		mu.Lock()
		m.RAMUsadaMB    = usada
		m.RAMTotalMB    = total
		m.RAMPorcentaje = pct
		mu.Unlock()
	}()

	// Goroutine 3: Red
	go func() {
		defer wg.Done()
		rx, tx := leerRed()
		mu.Lock()
		m.NetRxBytes = rx
		m.NetTxBytes = tx
		mu.Unlock()
	}()

	wg.Wait()
	return m
}

// ── Helpers de formato ────────────────────────────────────────────────────────

func barraTexto(pct float64, ancho int) string {
	llenos := int(pct / 100.0 * float64(ancho))
	if llenos > ancho {
		llenos = ancho
	}
	return "[" + strings.Repeat("█", llenos) + strings.Repeat("░", ancho-llenos) + "]"
}

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

// ── Goroutine del reporte ─────────────────────────────────────────────────────

// reporte lanza un goroutine que cada 5 segundos llama a obtenerMetricas
// (la cual internamente usa 3 goroutines paralelas) y actualiza los labels.
// Se detiene limpiamente al cerrarse el canal stop.
func reporte(lblCPU, lblRAM, lblRed *canvas.Text, stop <-chan struct{}) {
	obtenerMetricas() // primera lectura para inicializar deltas de CPU

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