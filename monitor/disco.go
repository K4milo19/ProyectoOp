package monitor

// disco.go
// Lectura del uso de disco con "df" (solo Linux).

import (
	"os/exec"
	"strconv"
	"strings"
)

// LeerDisco ejecuta "df -B1 --output=size,used /" y devuelve
// el espacio usado y total del disco raíz en GB.
// ok=false si el comando falla o la salida no es parseable.
func LeerDisco() (usadoGB, totalGB float64, ok bool) {
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
