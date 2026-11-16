package pulse

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// MaxVolume é o limite do ajuste de volume (acima de 100% é amplificação).
	MaxVolume = 150

	// Depois de trocar o perfil de uma placa, o dispositivo novo leva um instante
	// para aparecer. Esperamos até waitTries * waitStep (3 segundos).
	waitTries = 20
	waitStep  = 150 * time.Millisecond
)

// Load lê o estado atual do áudio e devolve as listas de saída e entrada.
func Load() (Snapshot, error) {
	var st state
	for _, k := range Kinds {
		devs, err := listDevices(k)
		if err != nil {
			return Snapshot{}, err
		}
		st.devs[k] = devs
		st.defaults[k], _ = pactl(cmds[k].getDefault)
	}

	// Sem as placas ainda dá para listar os dispositivos ativos, então um
	// erro aqui não é fatal.
	_ = pactlJSON(&st.cards, "list", "cards")

	var snap Snapshot
	for _, k := range Kinds {
		snap[k] = buildEntries(st, k)
	}
	return snap, nil
}

// listDevices lista os sinks ou sources ativos, sem os ".monitor"
// (fontes que apenas "escutam" o que uma saída está tocando).
func listDevices(k Kind) ([]device, error) {
	var all []device
	if err := pactlJSON(&all, "list", cmds[k].list); err != nil {
		return nil, err
	}
	devs := all[:0]
	for _, d := range all {
		if !strings.HasSuffix(d.Name, ".monitor") {
			devs = append(devs, d)
		}
	}
	return devs, nil
}

// Activate passa a usar a entrada escolhida:
//  1. se a porta está num perfil inativo, troca o perfil da placa;
//  2. escolhe a porta no dispositivo;
//  3. define o dispositivo como padrão;
//  4. move os áudios que já estão em andamento para ele.
func Activate(k Kind, e Entry) error {
	c := cmds[k]

	dev := e.Device
	if !e.Live() {
		var err error
		if dev, err = switchProfile(k, e); err != nil {
			return err
		}
	}
	if e.Port != "" {
		if _, err := pactl(c.setPort, dev, e.Port); err != nil {
			return err
		}
	}
	if _, err := pactl(c.setDefault, dev); err != nil {
		return err
	}
	moveStreams(k, dev)
	return nil
}

// switchProfile ativa o perfil necessário e espera o dispositivo aparecer.
func switchProfile(k Kind, e Entry) (string, error) {
	if _, err := pactl("set-card-profile", e.Card, e.Profile); err != nil {
		return "", err
	}
	for i := 0; i < waitTries; i++ {
		time.Sleep(waitStep)
		devs, err := listDevices(k)
		if err != nil {
			return "", err
		}
		if d, ok := findLive(devs, e.Card, e.Port); ok {
			return d.Name, nil
		}
	}
	return "", errors.New("o dispositivo não apareceu depois de trocar o perfil")
}

// moveStreams leva os áudios em andamento (app tocando ou gravando) para o
// novo dispositivo. É "melhor esforço": se algum falhar, os outros seguem.
func moveStreams(k Kind, dev string) {
	var streams []struct {
		Index int `json:"index"`
	}
	if pactlJSON(&streams, "list", cmds[k].streams) != nil {
		return
	}
	for _, s := range streams {
		pactl(cmds[k].moveStream, strconv.Itoa(s.Index), dev)
	}
}

// SetVolume define o volume do dispositivo em porcentagem (limitado a 0..MaxVolume).
func SetVolume(k Kind, device string, percent int) error {
	percent = min(MaxVolume, max(0, percent))
	_, err := pactl(cmds[k].setVolume, device, fmt.Sprintf("%d%%", percent))
	return err
}

// ToggleMute liga ou desliga o mudo do dispositivo.
func ToggleMute(k Kind, device string) error {
	_, err := pactl(cmds[k].setMute, device, "toggle")
	return err
}
