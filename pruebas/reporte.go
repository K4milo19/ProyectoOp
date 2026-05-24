package main

// reporte.go
// Monitor del sistema — solo Linux.
// Una única goroutine se ejecuta cada 5 segundos y actualiza en pantalla:
//   - CPU  : porcentaje de uso
//   - RAM  : MB ocupados vs MB disponibles
//   - Disco: GB usados vs GB totales en /

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2/canvas"
)

// ── Variables de estado para el delta de CPU ──────────────────────────────────

var (
	prevIdle  uint64
	prevTotal uint64
	cpuMu     sync.Mutex
)

// ── Lecturas ──────────────────────────────────────────────────────────────────

// leerCPU devuelve el % de uso de CPU calculado entre dos lecturas consecutivas.
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
		p := func(s string) uint64 { v, _ := strconv.ParseUint(s, 10, 64); return v }

		idle    := p(campos[4]) + p(campos[5])
		total   := p(campos[1]) + p(campos[2]) + p(campos[3]) +
		           p(campos[4]) + p(campos[5]) + p(campos[6]) + p(campos[7])

		cpuMu.Lock()
		dIdle  := idle - prevIdle
		dTotal := total - prevTotal
		prevIdle  = idle
		prevTotal = total
		cpuMu.Unlock()

		if dTotal == 0 {
			return 0
		}
		return (1.0 - float64(dIdle)/float64(dTotal)) * 100.0
	}
	return 0
}

// leerRAM devuelve MB usados y MB disponibles desde /proc/meminfo.
func leerRAM() (usadoMB, disponibleMB float64) {
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

	// disponible = lo que el kernel puede dar a procesos nuevos
	disp := v["MemFree"] + v["Buffers"] + v["Cached"] + v["SReclaimable"]
	usado := v["MemTotal"] - disp

	usadoMB     = float64(usado) / 1024.0
	disponibleMB = float64(disp)  / 1024.0
	return
}

// leerDisco devuelve GB usados y GB totales del sistema de archivos raíz
// usando el comando df, que está disponible en cualquier Linux sin imports nativos.
func leerDisco() (usadoGB, totalGB float64) {
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
	total, _ := strconv.ParseUint(campos[0], 10, 64)
	usado, _  := strconv.ParseUint(campos[1], 10, 64)
	totalGB = float64(total) / (1 << 30)
	usadoGB = float64(usado) / (1 << 30)
	return
}

// ── Goroutine principal ───────────────────────────────────────────────────────

// reporte lanza una goroutine que cada 5 segundos lee CPU, RAM y disco
// y actualiza los tres canvas.Text recibidos.
// Se detiene limpiamente cuando se cierra el canal stop.
func reporte(lblCPU, lblRAM, lblDisco *canvas.Text, stop <-chan struct{}) {
	leerCPU() // inicializa prevIdle/prevTotal para que el primer delta sea válido

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				cpu                   := leerCPU()
				ramUsado, ramDisp     := leerRAM()
				discoUsado, discoTotal := leerDisco()

				lblCPU.Text = fmt.Sprintf(
					"CPU    %s  %.1f%%",
					barra(cpu, 24), cpu,
				)
				lblCPU.Refresh()

				lblRAM.Text = fmt.Sprintf(
					"RAM    usada: %6.0f MB   disponible: %6.0f MB",
					ramUsado, ramDisp,
				)
				lblRAM.Refresh()

				lblDisco.Text = fmt.Sprintf(
					"DISCO  usada: %5.1f GB   total:      %5.1f GB   libre: %.1f GB",
					discoUsado, discoTotal, discoTotal-discoUsado,
				)
				lblDisco.Refresh()
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