package red_cliente

// metricas.go
// Parsea las líneas <<<METRICS:...>>> que llegan del servidor y
// despacha los valores al callback OnMetrics de la UI.

import (
	"strconv"
	"strings"
)

// DespachadorMetricas lee el canal Metricas y llama a OnMetrics.
// Debe lanzarse en su propio goroutine.
// Se detiene cuando se recibe en el canal stop o cuando el canal Metricas
// se cierra (conexión caída).
func (c *ConexionServidor) DespachadorMetricas(stop <-chan struct{}) {
	go func() {
		for {
			select {
			case <-stop:
				return
			case linea, ok := <-c.Metricas:
				if !ok {
					return
				}
				if c.OnMetrics != nil {
					if cpu, ru, rf, du, dt, ok2 := parsearMetrics(linea); ok2 {
						c.OnMetrics(cpu, ru, rf, du, dt)
					}
				}
			}
		}
	}()
}

// parsearMetrics extrae los 5 valores numéricos de una línea con formato:
//
//	<<<METRICS:CPU=12.3|RAM_USED=1024|RAM_FREE=2048|DISK_USED=40.1|DISK_TOTAL=100.0>>>
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
