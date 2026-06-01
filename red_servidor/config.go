package red_servidor

// config.go
// Configuración del servidor: puerto de escucha y lista de IPs permitidas.
// Se carga desde config.txt al arrancar el servidor.
// Si el archivo no existe se usan los valores por defecto.
//
// Formato de config.txt:
//
//   puerto=9000
//   ips_permitidas=192.168.1.10,192.168.1.20,192.168.1.30
//
// Si ips_permitidas está vacío o no aparece, se aceptan todas las conexiones.

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// Config contiene los parámetros de configuración del servidor.
type Config struct {
	Puerto        string   // ejemplo: ":9000"
	IPsPermitidas []string // vacío = todas permitidas
}

// configActual es la configuración cargada al iniciar.
var configActual = Config{
	Puerto:        ":9000",
	IPsPermitidas: []string{},
}

// CargarConfig lee config.txt y actualiza configActual.
// Si el archivo no existe usa los valores por defecto y lo crea como plantilla.
func CargarConfig() {
	f, err := os.Open("config.txt")
	if err != nil {
		// No existe: crear plantilla y usar valores por defecto
		crearConfigPorDefecto()
		fmt.Println("INFO config: no se encontró config.txt, se creó uno con valores por defecto")
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		linea := strings.TrimSpace(scanner.Text())
		if linea == "" || strings.HasPrefix(linea, "#") {
			continue
		}
		partes := strings.SplitN(linea, "=", 2)
		if len(partes) != 2 {
			continue
		}
		clave := strings.TrimSpace(partes[0])
		valor := strings.TrimSpace(partes[1])

		switch clave {
		case "puerto":
			if valor != "" {
				if !strings.HasPrefix(valor, ":") {
					valor = ":" + valor
				}
				configActual.Puerto = valor
			}
		case "ips_permitidas":
			if valor != "" {
				for _, ip := range strings.Split(valor, ",") {
					ip = strings.TrimSpace(ip)
					if ip != "" {
						configActual.IPsPermitidas = append(configActual.IPsPermitidas, ip)
					}
				}
			}
		}
	}

	fmt.Printf("INFO config: puerto=%s  ips_permitidas=%v\n",
		configActual.Puerto, configActual.IPsPermitidas)
}

// IPEstaPermitida devuelve true si la IP de la conexión entrante está
// en la lista de permitidas. Si la lista está vacía acepta todas.
func IPEstaPermitida(remoto string) bool {
	if len(configActual.IPsPermitidas) == 0 {
		return true
	}
	// remoto tiene formato "192.168.1.10:56234", extraer solo la IP
	ip, _, err := net.SplitHostPort(remoto)
	if err != nil {
		return false
	}
	for _, permitida := range configActual.IPsPermitidas {
		if ip == permitida {
			return true
		}
	}
	return false
}

// GetPuerto devuelve el puerto configurado.
func GetPuerto() string {
	return configActual.Puerto
}

// crearConfigPorDefecto escribe un config.txt de plantilla en el directorio actual.
func crearConfigPorDefecto() {
	contenido := `# config.txt — Configuración del servidor ShellOS
# Líneas que empiezan con # son comentarios

# Puerto en el que escucha el servidor
puerto=9000

# IPs que tienen permiso de conectarse
# Separar con comas, sin espacios entre ellas
# Si se deja vacío se aceptan TODAS las conexiones
ips_permitidas=
`
	os.WriteFile("config.txt", []byte(contenido), 0644)
}