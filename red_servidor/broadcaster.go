package red_servidor

// broadcaster.go
// Goroutine que cada 5 s lee las métricas locales del servidor
// (CPU, RAM, Disco) y las empuja a todos los clientes conectados.
//
// Formato de la línea enviada:
//   <<<METRICS:CPU=12.3|RAM_USED=1024|RAM_FREE=2048|DISK_USED=40.1|DISK_TOTAL=100.0>>>

import (
	"fmt"
	"time"

	"shellos/monitor"
)

// IniciarBroadcasterMetricas lanza la goroutine del broadcaster.
// Se detiene cuando se cierra el canal stop.
func IniciarBroadcasterMetricas(stop <-chan struct{}) {
	prevIdle, prevTotal := monitor.LeerStatCPU()

	go func() {
		for {
			select {
			case <-stop:
				return
			case <-time.After(5 * time.Second):
			}

			// CPU
			idle, total := monitor.LeerStatCPU()
			dIdle  := idle - prevIdle
			dTotal := total - prevTotal
			prevIdle  = idle
			prevTotal = total

			var cpu float64
			if dTotal > 0 {
				cpu = (1.0 - float64(dIdle)/float64(dTotal)) * 100.0
			}

			// RAM
			ramUsed, ramFree := 0.0, 0.0
			if u, f, ok := monitor.LeerRAM(); ok {
				ramUsed, ramFree = u, f
			}

			// Disco
			diskUsed, diskTotal := 0.0, 0.0
			if u, t, ok := monitor.LeerDisco(); ok {
				diskUsed, diskTotal = u, t
			}

			linea := fmt.Sprintf(
				"<<<METRICS:CPU=%.1f|RAM_USED=%.0f|RAM_FREE=%.0f|DISK_USED=%.1f|DISK_TOTAL=%.1f>>>",
				cpu, ramUsed, ramFree, diskUsed, diskTotal,
			)
			broadcastMetrics(linea)
		}
	}()
}
