package red_servidor

// log.go
// Escritura thread-safe al archivo logs.txt.
// La ruta absoluta se fija al iniciar el servidor, antes de cualquier cd.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const ArchivoLog = "logs.txt"

var (
	logMu      sync.Mutex
	rutaLogAbs string
)

// InitLog captura el directorio de trabajo actual como raíz permanente del log.
// Debe llamarse al arrancar el servidor, antes de aceptar cualquier cliente.
func InitLog() {
	wd, err := os.Getwd()
	if err != nil {
		rutaLogAbs = ArchivoLog
		return
	}
	rutaLogAbs = filepath.Join(wd, ArchivoLog)
}

// EscribirLog añade una entrada al archivo logs.txt.
// Es seguro llamarlo desde múltiples goroutines.
func EscribirLog(cliente, comando, salida string) {
	logMu.Lock()
	defer logMu.Unlock()

	f, err := os.OpenFile(rutaLogAbs, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	marca := time.Now().Format("2006-01-02 15:04:05")
	entrada := fmt.Sprintf(
		"[%s] CLIENTE=%s\n  CMD : %s\n  OUT : %s\n%s\n",
		marca, cliente, comando,
		strings.ReplaceAll(strings.TrimSpace(salida), "\n", "\n       "),
		strings.Repeat("─", 60),
	)
	f.WriteString(entrada)
}