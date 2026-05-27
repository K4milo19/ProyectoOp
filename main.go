package main

// main.go
// Punto de entrada de ShellOS.
// Solo orquesta el flujo de pantallas: selector → login → ventana final.
// Toda la lógica está en los paquetes internos.

import (
	"fyne.io/fyne/v2/app"

	"shellos/tema"
	"shellos/ui/login"
	"shellos/ui/selector"
	"shellos/ui/servidor"
	"shellos/ui/terminal_cliente"
)

func main() {
	a := app.New()
	a.Settings().SetTheme(&tema.ShellTheme{})

	selector.MostrarSeleccionModo(a, func(modo selector.ModoApp) {
		login.MostrarLogin(a, modo, func() {
			switch modo {
			case selector.ModoServidor:
				servidor.MostrarVentanaServidor(a)
			case selector.ModoCliente:
				terminal_cliente.MostrarVentanaConexion(a)
			}
		})
	})
}
