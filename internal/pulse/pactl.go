// Package pulse conversa com o servidor de áudio (PulseAudio ou PipeWire)
// chamando o comando `pactl`. Nenhum outro pacote do projeto executa comandos
// externos: tudo passa por aqui.
package pulse

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Available confere se o pactl está instalado.
func Available() error {
	if _, err := exec.LookPath("pactl"); err != nil {
		return errors.New("pactl não encontrado: instale pulseaudio-utils (ou pipewire-pulse)")
	}
	return nil
}

// pactl executa `pactl <args>` e devolve a saída em texto.
// Se o comando falhar, o erro inclui a mensagem que o próprio pactl imprimiu.
func pactl(args ...string) (string, error) {
	out, err := exec.Command("pactl", args...).Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return "", fmt.Errorf("pactl %s: %s", args[0], strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// pactlJSON executa `pactl --format=json <args>` e decodifica o resultado em v.
// O formato JSON exige PulseAudio 16+ ou PipeWire.
func pactlJSON(v any, args ...string) error {
	out, err := exec.Command("pactl", append([]string{"--format=json"}, args...)...).Output()
	if err != nil {
		return err
	}
	return json.Unmarshal(out, v)
}
