// volcli: gerenciador de áudio para o terminal.
//
// O main só faz duas coisas: confere se o pactl existe e inicia a interface.
// Toda a lógica está em:
//   - internal/pulse: conversa com o servidor de áudio (via pactl)
//   - internal/ui:    tela, teclas e navegação (Bubble Tea)
package main

import (
	"fmt"
	"os"

	"volcli/internal/pulse"
	"volcli/internal/ui"
)

func main() {
	if err := pulse.Available(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := ui.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
