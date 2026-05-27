package red_servidor

// log.go
// Escritura thread-safe al archivo logs.txt.
// Cada entrada registra timestamp, IP del cliente, comando y salida.

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const ArchivoLog = "logs.txt"

var logMu sync.Mutex

// EscribirLog añade una entrada al archivo logs.txt.
// Es seguro llamarlo desde múltiples goroutines.
func EscribirLog(cliente, comando, salida string) {
	logMu.Lock()
	defer logMu.Unlock()

	f, err := os.OpenFile(ArchivoLog, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
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
