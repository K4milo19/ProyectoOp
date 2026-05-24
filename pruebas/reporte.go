package main

// reporte.go
// Monitor del sistema — solo Linux.
// Todo el estado vive dentro de la goroutine, sin variables globales.

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2/widget"
)

// ── Lectura de CPU ────────────────────────────────────────────────────────────
// Devuelve (idle, total) crudos para que la goroutine calcule el delta.

func leerStatCPU() (idle, total uint64) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return
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
			return
		}
		p := func(s string) uint64 {
			v, _ := strconv.ParseUint(s, 10, 64)
			return v
		}
		user    := p(campos[1])
		nice    := p(campos[2])
		system  := p(campos[3])
		idleT   := p(campos[4])
		iowait  := p(campos[5])
		irq     := p(campos[6])
		softirq := p(campos[7])

		idle  = idleT + iowait
		total = user + nice + system + idleT + iowait + irq + softirq
		return
	}
	return
}

// ── Lectura de RAM ────────────────────────────────────────────────────────────

func leerRAM() (usadoMB, libreMB float64, ok bool) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer f.Close()

	v := map[string]uint64{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		partes := strings.Fields(scanner.Text())
		if len(partes) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(partes[1], 10, 64)
		v[strings.TrimSuffix(partes[0], ":")] = val
	}

	total, existeTotal := v["MemTotal"]
	if !existeTotal || total == 0 {
		return
	}

	libre := v["MemFree"] + v["Buffers"] + v["Cached"] + v["SReclaimable"]
	usado := total - libre

	usadoMB = float64(usado) / 1024.0
	libreMB = float64(libre) / 1024.0
	ok = true
	return
}

// ── Lectura de Disco ──────────────────────────────────────────────────────────

func leerDisco() (usadoGB, totalGB float64, ok bool) {
	out, err := exec.Command("df", "-B1", "--output=size,used", "/").Output()
	if err != nil {
		return
	}
	lineas := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lineas) < 2 {
		return
	}
	campos := strings.Fields(lineas[1])
	if len(campos) < 2 {
		return
	}
	total, err1 := strconv.ParseUint(campos[0], 10, 64)
	usado, err2 := strconv.ParseUint(campos[1], 10, 64)
	if err1 != nil || err2 != nil || total == 0 {
		return
	}
	totalGB = float64(total) / (1 << 30)
	usadoGB = float64(usado) / (1 << 30)
	ok = true
	return
}

// ── Goroutine principal ───────────────────────────────────────────────────────

// reporte lanza una goroutine que cada 5 segundos actualiza los tres labels.
// El estado del delta de CPU vive dentro de la goroutine — sin variables globales.
func reporte(lblCPU, lblRAM, lblDisco *widget.Label, stop <-chan struct{}) {
	go func() {
		// Estado del delta de CPU encapsulado aquí dentro
		var prevIdle, prevTotal uint64

		// Primera lectura para inicializar el delta
		prevIdle, prevTotal = leerStatCPU()

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				// ── CPU ──────────────────────────────────────────
				idle, total := leerStatCPU()
				dIdle  := idle - prevIdle
				dTotal := total - prevTotal
				prevIdle  = idle
				prevTotal = total

				var cpu float64
				if dTotal > 0 {
					cpu = (1.0 - float64(dIdle)/float64(dTotal)) * 100.0
				}

				lblCPU.SetText(fmt.Sprintf(
					"CPU\n%s\n%.1f%%",
					barra(cpu, 18), cpu,
				))

				// ── RAM ──────────────────────────────────────────
				if usadoMB, libreMB, ok := leerRAM(); ok {
					lblRAM.SetText(fmt.Sprintf(
						"RAM\nusada:  %.0f MB\nlibre:  %.0f MB",
						usadoMB, libreMB,
					))
				}

				// ── Disco ─────────────────────────────────────────
				if usadoGB, totalGB, ok := leerDisco(); ok {
					lblDisco.SetText(fmt.Sprintf(
						"DISCO\nusada: %.1f GB\ntotal: %.1f GB\nlibre: %.1f GB",
						usadoGB, totalGB, totalGB-usadoGB,
					))
				}
			}
		}
	}()
}

// barra genera una barra ASCII proporcional al porcentaje dado.
func barra(pct float64, ancho int) string {
	llenos := int(pct / 100.0 * float64(ancho))
	if llenos > ancho {
		llenos = ancho
	}
	return "[" + strings.Repeat("█", llenos) + strings.Repeat("░", ancho-llenos) + "]"
}