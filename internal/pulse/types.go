package pulse

import (
	"sort"
	"strings"
)

// Este arquivo só descreve os dados que o pactl devolve em JSON.
// Nada aqui é exportado: o resto do programa enxerga apenas Entry.

// volumeNorm é o valor que o PulseAudio usa para representar 100% de volume.
const volumeNorm = 65536

type devicePort struct {
	Name string `json:"name"`
}

type channelVolume struct {
	Value int `json:"value"`
}

// device é um sink (saída) ou source (entrada) ativo no momento.
type device struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Mute        bool                     `json:"mute"`
	ActivePort  string                   `json:"active_port"`
	Ports       []devicePort             `json:"ports"`
	Volume      map[string]channelVolume `json:"volume"` // um valor por canal (esquerdo, direito...)
}

// percent converte o volume bruto para porcentagem (usa o canal mais alto).
func (d device) percent() int {
	peak := 0
	for _, v := range d.Volume {
		if v.Value > peak {
			peak = v.Value
		}
	}
	return int(float64(peak)*100/volumeNorm + 0.5)
}

func (d device) hasPort(name string) bool {
	for _, p := range d.Ports {
		if p.Name == name {
			return true
		}
	}
	return false
}

// cardPort é uma porta física/lógica de uma placa (alto-falante, P2, HDMI...).
// Profiles lista os perfis da placa em que essa porta existe.
type cardPort struct {
	Description  string   `json:"description"`
	Type         string   `json:"type"`
	Availability string   `json:"availability"` // "available", "not available" ou "availability unknown"
	Profiles     []string `json:"profiles"`
}

// cardProfile é um modo de funcionamento da placa. Cada perfil expõe um
// conjunto diferente de portas, por isso um HDMI pode não aparecer enquanto
// outro perfil estiver ativo.
type cardProfile struct {
	Priority  float64     `json:"priority"`
	Available interface{} `json:"available"` // bool na prática; tratado em usable()
}

func (p cardProfile) usable() bool {
	v, ok := p.Available.(bool)
	return !ok || v
}

// card é uma placa de som (interna, USB, Bluetooth...).
type card struct {
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties"`
	Profiles   map[string]cardProfile `json:"profiles"`
	Ports      map[string]cardPort    `json:"ports"`
}

func (c card) description() string {
	if s, ok := c.Properties["device.description"].(string); ok && s != "" {
		return s
	}
	return c.Name
}

// sortedPortNames devolve as portas em ordem fixa. Sem isso a ordem do map do
// Go muda a cada leitura e a lista ficaria "pulando" na tela.
func (c card) sortedPortNames() []string {
	names := make([]string, 0, len(c.Ports))
	for n := range c.Ports {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// belongs diz se o dispositivo pertence à placa.
//
// O pactl não informa isso diretamente, mas o identificador da placa aparece
// no nome do dispositivo:
//
//	placa:        alsa_card.pci-0000_00_1f.3
//	dispositivo:  alsa_output.pci-0000_00_1f.3.analog-stereo
//
// Então basta ver se o trecho depois do primeiro "." da placa está no nome.
func belongs(d device, cardName string) bool {
	id := cardName
	if i := strings.Index(cardName, "."); i >= 0 {
		id = cardName[i+1:]
	}
	return strings.Contains(d.Name, id)
}
