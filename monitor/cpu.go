package monitor

// cpu.go
// Lectura del uso de CPU desde /proc/stat (solo Linux).
// Calcula tiempo idle y tiempo total para que el llamador pueda
// calcular el porcentaje de uso entre dos muestras consecutivas.

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// LeerStatCPU lee /proc/stat y devuelve los ticks idle y totales
// de la primera línea (cpu agregada de todos los núcleos).
func LeerStatCPU() (idle, total uint64) {
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
