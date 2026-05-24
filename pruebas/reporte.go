package main

// reporte.go
// Monitor del sistema — solo Linux.
// Funcion bloqueante con time.Sleep. Se lanza con go reporte(...) desde
// ventana.go y red_cliente.go. Sin goroutines internas.

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

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
		idleT  := p(campos[4])
		iowait := p(campos[5])
		idle   = idleT + iowait
		total  = p(campos[1]) + p(campos[2]) + p(campos[3]) +
		         idleT + iowait + p(campos[6]) + p(campos[7])
		return
	}
	return
}

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
	total := v["MemTotal"]
	if total == 0 {
		return
	}
	libre := v["MemFree"] + v["Buffers"] + v["Cached"] + v["SReclaimable"]
	usadoMB = float64(total-libre) / 1024.0
	libreMB  = float64(libre) / 1024.0
	ok = true
	return
}

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
	tot, e1 := strconv.ParseUint(campos[0], 10, 64)
	uso, e2 := strconv.ParseUint(campos[1], 10, 64)
	if e1 != nil || e2 != nil || tot == 0 {
		return
	}
	totalGB = float64(tot) / (1 << 30)
	usadoGB = float64(uso) / (1 << 30)
	ok = true
	return
}

// reporte es una funcion bloqueante. Lanzar con: go reporte(...)
// Usa time.Sleep de 5 segundos entre cada lectura. Sin goroutines internas.
func reporte(lblCPU, lblRAM, lblDisco *widget.Label, stop <-chan struct{}) {
	var prevIdle, prevTotal uint64
	prevIdle, prevTotal = leerStatCPU()

	for {
		// Verificar si hay que parar antes de dormir
		select {
		case <-stop:
			return
		default:
		}

		time.Sleep(5 * time.Second)

		// Verificar de nuevo tras el sleep
		select {
		case <-stop:
			return
		default:
		}

		// CPU
		idle, total := leerStatCPU()
		dIdle  := idle - prevIdle
		dTotal := total - prevTotal
		prevIdle  = idle
		prevTotal = total

		var cpu float64
		if dTotal > 0 {
			cpu = (1.0 - float64(dIdle)/float64(dTotal)) * 100.0
		}

		cpuText := fmt.Sprintf("CPU: %.1f%%", cpu)
		var ramText, discoText string

		if usadoMB, libreMB, ok := leerRAM(); ok {
			ramText = fmt.Sprintf("RAM — usada: %.0f MB  |  libre: %.0f MB", usadoMB, libreMB)
		}
		if usadoGB, totalGB, ok := leerDisco(); ok {
			discoText = fmt.Sprintf("DISCO — usado: %.1f GB  |  total: %.1f GB  |  libre: %.1f GB",
				usadoGB, totalGB, totalGB-usadoGB)
		}

		// fyne.Do garantiza ejecucion en el hilo principal de la UI
		fyne.Do(func() {
			lblCPU.SetText(cpuText)
			if ramText != "" {
				lblRAM.SetText(ramText)
			}
			if discoText != "" {
				lblDisco.SetText(discoText)
			}
		})
	}
}