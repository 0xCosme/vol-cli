package pulse

import "strings"

// Kind é a direção do áudio: saída (caixas, fones) ou entrada (microfones).
//
// No pactl, saída = "sink" e entrada = "source". Os comandos são quase iguais
// (set-sink-volume x set-source-volume), então o resto do código trabalha com
// Kind e este arquivo traduz para o nome certo de cada comando.
type Kind int

const (
	Output Kind = iota
	Input
)

// Kinds permite percorrer as duas direções com um for.
var Kinds = [...]Kind{Output, Input}

// commands guarda os nomes dos subcomandos do pactl de cada direção.
type commands struct {
	title      string // nome mostrado na aba
	list       string // pactl list <list>
	getDefault string // descobre o dispositivo padrão
	setDefault string // define o dispositivo padrão
	setPort    string // escolhe a porta (alto-falante, fone, HDMI...)
	setVolume  string
	setMute    string
	streams    string // áudios em andamento (apps tocando ou gravando)
	moveStream string // move um áudio em andamento para outro dispositivo
}

var cmds = [...]commands{
	Output: {"Saída", "sinks", "get-default-sink", "set-default-sink", "set-sink-port", "set-sink-volume", "set-sink-mute", "sink-inputs", "move-sink-input"},
	Input:  {"Entrada", "sources", "get-default-source", "set-default-source", "set-source-port", "set-source-volume", "set-source-mute", "source-outputs", "move-source-output"},
}

// Title é o nome da aba na interface.
func (k Kind) Title() string { return cmds[k].title }

// Other devolve a direção oposta (saída <-> entrada).
func (k Kind) Other() Kind { return 1 - k }

// portKind descobre a direção de uma porta pelo nome
// (ex.: "analog-output-speaker" é saída, "analog-input-mic" é entrada).
func portKind(name string) (Kind, bool) {
	switch {
	case strings.Contains(name, "output"):
		return Output, true
	case strings.Contains(name, "input"):
		return Input, true
	}
	return 0, false
}
