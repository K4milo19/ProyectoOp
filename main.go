package main

import (
    "crypto/sha256"
    "fmt"
    "os"
    "bufio"
    "os/exec"
    "strings"
    "golang.org/x/term"
)

func main() {
    hdbPasswdBytes, _ := os.ReadFile("contraseña.txt")
    hdbPasswd := strings.TrimSpace(string(hdbPasswdBytes)) 
    var uPasswd string
    intentos := 0
    fmt.Println("Contraseña almacenada:")

    // Iniciar el ciclo de intentos de login
    for {
        fmt.Println("Digite su password: ")
        sluPasswd, _ := term.ReadPassword(int(os.Stdin.Fd()))
        uPasswd = string(sluPasswd)

        vUserHash := sha256.Sum256([]byte(uPasswd))
        huPasswd := fmt.Sprintf("%x", vUserHash)

        if huPasswd == hdbPasswd {
            fmt.Println("<<<<<<< LOGIN CORRECTO >>>>>>>>>\n")
            for {
                pwd, _ := os.Getwd()
                var comando string
                fmt.Print(pwd, "-> ShellOS# ")

                lector := bufio.NewReader(os.Stdin)
                comando, _ = lector.ReadString('\n')
                comando = strings.TrimRight(comando, "\n")

                slcomando := strings.Fields(comando)

                if comando == "bye" {
                    fmt.Println("Gracias por ejecutar la shell so...")
                    break
                } else if slcomando[0] == "cd" {
                    // Cambiar directorio si el comando es 'cd'
                    if len(slcomando) > 1 {
                        os.Chdir(slcomando[1])
                    } else {
                        fmt.Println("Debe especificar el directorio.")
                    }
                }

                shell := exec.Command("bash", "-c", comando)
                salida, _ := shell.Output()
                fmt.Print(string(salida))
            }
            break
        } else {
            fmt.Println("XXXXXXXXX LOGIN INCORRECTO XXXXXXXXXX\n")
            intentos++
            if intentos >= 3 {
                fmt.Println("Demasiados intentos fallidos. Saliendo...")
                break
            }
        }
    }

    fmt.Println("Gracias por usar loginOP >>>>>>>>>\n")
}
