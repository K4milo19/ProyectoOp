package monitor

// ram.go
// Lectura del uso de RAM desde /proc/meminfo (solo Linux).

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// LeerRAM lee /proc/meminfo y devuelve RAM usada y libre en MB.
// ok=false si el archivo no existe o no contiene MemTotal.
func LeerRAM() (usadoMB, libreMB float64, ok bool) {
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
