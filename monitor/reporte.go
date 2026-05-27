package monitor

// reporte.go
// Bucle de actualización de métricas para la UI local.
// Función bloqueante; lanzar con: go monitor.Reporte(...)

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// Reporte actualiza los tres labels de la UI con los datos del sistema
// cada 5 segundos hasta que se cierra el canal stop.
// Debe lanzarse en su propio goroutine.
func Reporte(lblCPU, lblRAM, lblDisco *widget.Label, stop <-chan struct{}) {
	prevIdle, prevTotal := LeerStatCPU()

	for {
		select {
		case <-stop:
			return
		default:
		}

		time.Sleep(5 * time.Second)

		select {
		case <-stop:
			return
		default:
		}

		// CPU
		idle, total := LeerStatCPU()
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
		if usadoMB, libreMB, ok := LeerRAM(); ok {
			ramText = fmt.Sprintf("RAM — usada: %.0f MB  |  libre: %.0f MB", usadoMB, libreMB)
		}
		if usadoGB, totalGB, ok := LeerDisco(); ok {
			discoText = fmt.Sprintf(
				"DISCO — usado: %.1f GB  |  total: %.1f GB  |  libre: %.1f GB",
				usadoGB, totalGB, totalGB-usadoGB,
			)
		}

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
