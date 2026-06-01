package red_servidor

import (
	"fmt"
	"net"
)

var PuertoServidor string

func IniciarServidor(onLog func(string), stop <-chan struct{}) error {
	CargarConfig()

	PuertoServidor = GetPuerto()

	ln, err := net.Listen("tcp", PuertoServidor)
	if err != nil {
		return fmt.Errorf("no se pudo iniciar el servidor en %s: %w", PuertoServidor, err)
	}

	onLog(fmt.Sprintf("◈  Servidor escuchando en %s", PuertoServidor))
	onLog(fmt.Sprintf("◈  Log guardado en: %s", ArchivoLog))
	onLog("◈  Broadcast de métricas cada 5 s activado")

	IniciarBroadcasterMetricas(stop)

	go func() {
		<-stop
		ln.Close()
	}()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-stop:
					return
				default:
					onLog(fmt.Sprintf("[ ERROR accept ]  %v", err))
					continue
				}
			}
			if !IPEstaPermitida(conn.RemoteAddr().String()) {
				onLog(fmt.Sprintf("[ BLOQUEADO ]  %s no está en la lista de IPs permitidas", conn.RemoteAddr().String()))
				RechazarCliente(conn)
				continue
			}
			go ManejarCliente(conn, onLog)
		}
	}()

	return nil
}

func ObtenerIPs() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return []string{"(error obteniendo IPs)"}
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}
	if len(ips) == 0 {
		return []string{"127.0.0.1 (solo loopback)"}
	}
	return ips
}