package red_servidor

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
	initOnce   sync.Once
)

func resolverRuta() {
	initOnce.Do(func() {
		wd, err := os.Getwd()
		if err != nil {
			rutaLogAbs = ArchivoLog
			return
		}
		rutaLogAbs = filepath.Join(wd, ArchivoLog)
		fmt.Println("DEBUG log: ruta fijada en:", rutaLogAbs)
	})
}

func EscribirLog(cliente, comando, salida string) {
	resolverRuta()

	logMu.Lock()
	defer logMu.Unlock()

	f, err := os.OpenFile(rutaLogAbs, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("ERROR abriendo log:", err)
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
	_, err = f.WriteString(entrada)
	if err != nil {
		fmt.Println("ERROR escribiendo log:", err)
	}
}