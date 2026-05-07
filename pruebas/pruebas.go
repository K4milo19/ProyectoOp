package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ─── Validación ───────────────────────────────────────────────────────────────

func validarCredenciales(usuario, contrasena string) (bool, error) {
	archivo, err := os.Open("contraseña.txt")
	if err != nil {
		return false, fmt.Errorf("no se pudo abrir el archivo: %w", err)
	}
	defer archivo.Close()

	hashIngresado := fmt.Sprintf("%x", sha256.Sum256([]byte(contrasena)))

	scanner := bufio.NewScanner(archivo)
	for scanner.Scan() {
		linea := strings.TrimSpace(scanner.Text())
		if linea == "" {
			continue
		}
		partes := strings.SplitN(linea, ":", 2)
		if len(partes) != 2 {
			continue
		}
		if strings.TrimSpace(partes[0]) == usuario &&
			strings.TrimSpace(partes[1]) == hashIngresado {
			return true, nil
		}
	}
	return false, scanner.Err()
}

// ─── Ejecutar comando shell ───────────────────────────────────────────────────

// ejecutarComando procesa el comando recibido y devuelve:
//   - la salida (stdout+stderr combinados o mensaje propio)
//   - el nuevo directorio de trabajo actual
//   - un booleano que indica si la aplicación debe cerrarse
func ejecutarComando(comando string) (salida string, quit bool) {
	comando = strings.TrimSpace(comando)
	if comando == "" {
		return "", false
	}

	// Comando de salida
	if comando == "bye" {
		return "Gracias por ejecutar la shell…", true
	}

	slcomando := strings.Fields(comando)

	// Cambio de directorio (cd)
	if slcomando[0] == "cd" {
		if len(slcomando) > 1 {
			if err := os.Chdir(slcomando[1]); err != nil {
				return fmt.Sprintf("cd: %v", err), false
			}
			pwd, _ := os.Getwd()
			return "→ " + pwd, false
		}
		return "cd: debe especificar el directorio.", false
	}

	// Cualquier otro comando: ejecutar con bash
	shell := exec.Command("bash", "-c", comando)
	// CombinedOutput captura tanto stdout como stderr
	out, err := shell.CombinedOutput()
	resultado := strings.TrimRight(string(out), "\n")
	if err != nil && resultado == "" {
		resultado = fmt.Sprintf("error: %v", err)
	}
	return resultado, false
}

// ─── Ventana principal (terminal GUI) ────────────────────────────────────────

func mostrarVentanaPrincipal(a fyne.App) {
	w := a.NewWindow("ShellOS — Terminal")
	w.Resize(fyne.NewSize(900, 560))

	// ── historial de salida (scroll) ──
	historial := widget.NewLabel("")
	historial.Wrapping = fyne.TextWrapWord
	// Fuente monoespaciada para que parezca una terminal
	historial.TextStyle = fyne.TextStyle{Monospace: true}

	scrollHistorial := container.NewVScroll(historial)
	scrollHistorial.SetMinSize(fyne.NewSize(860, 400))

	lineasHistorial := []string{}

	agregarLinea := func(linea string) {
		lineasHistorial = append(lineasHistorial, linea)
		historial.SetText(strings.Join(lineasHistorial, "\n"))
		scrollHistorial.ScrollToBottom()
	}

	// Mensaje de bienvenida
	agregarLinea("╔══════════════════════════════════════╗")
	agregarLinea("║       Bienvenido a ShellOS GUI       ║")
	agregarLinea("║  Escribe 'bye' para salir             ║")
	agregarLinea("╚══════════════════════════════════════╝")

	// ── entrada de comando ──
	entrada := widget.NewEntry()
	entrada.SetPlaceHolder("Escribe un comando…")

	// Función compartida entre botón y tecla Enter
	procesarEntrada := func() {
		texto := strings.TrimSpace(entrada.Text)
		entrada.SetText("")
		if texto == "" {
			return
		}

		// Mostrar el prompt + comando escrito
		pwd, _ := os.Getwd()
		agregarLinea(fmt.Sprintf("%s → ShellOS# %s", pwd, texto))

		// Ejecutar y mostrar salida
		salida, quit := ejecutarComando(texto)
		if salida != "" {
			agregarLinea(salida)
		}
		agregarLinea("") // línea en blanco entre comandos

		if quit {
			a.Quit()
		}
	}

	// Detectar tecla Enter en el campo de texto
	entrada.OnSubmitted = func(_ string) {
		procesarEntrada()
	}

	botonEnviar := widget.NewButtonWithIcon("Enviar", theme.MailSendIcon(), func() {
		procesarEntrada()
	})

	barraInferior := container.NewBorder(nil, nil, nil, botonEnviar, entrada)

	w.SetContent(container.NewBorder(nil, barraInferior, nil, nil, scrollHistorial))
	w.Show()
}

// ─── Login ────────────────────────────────────────────────────────────────────

func main() {
	intentos := 0

	a := app.New()
	w := a.NewWindow("Login — ShellOS")
	w.Resize(fyne.NewSize(800, 600))

	inpUser := widget.NewEntry()
	inpUser.SetPlaceHolder("Ingrese su usuario")

	inpPass := widget.NewPasswordEntry()
	inpPass.SetPlaceHolder("Ingrese su contraseña")

	message  := widget.NewLabel("Digite su usuario y contraseña")
	message2 := widget.NewLabel("Intentos: 0")
	message3 := widget.NewLabel("❌ Usuario sospechoso — Plataforma cerrada ❌")

	button := widget.NewButton("Iniciar sesión", func() {
		valido, err := validarCredenciales(inpUser.Text, inpPass.Text)
		if err != nil {
			message.SetText("Error al leer credenciales: " + err.Error())
			return
		}
		if valido {
			w.Hide()
			mostrarVentanaPrincipal(a)
		} else {
			intentos++
			message2.SetText("Intentos: " + strconv.Itoa(intentos))
			message.SetText("⚠ Usuario o contraseña incorrectos")
			if intentos >= 3 {
				w.SetContent(container.NewCenter(container.NewVBox(message3)))
			}
		}
	})

	// Permitir login con Enter desde el campo de contraseña
	inpPass.OnSubmitted = func(_ string) {
		button.OnTapped()
	}

	w.SetContent(container.NewCenter(container.NewVBox(
		message, inpUser, inpPass, button, message2,
	)))
	w.ShowAndRun()
}