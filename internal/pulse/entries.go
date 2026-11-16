package pulse

import "strings"

// Entry é uma linha da lista na tela: uma porta de um dispositivo
// (ex.: "Áudio interno · Fones de ouvido").
type Entry struct {
	Card    string // nome da placa dona da porta (vazio se o dispositivo não tem placa)
	Port    string // nome técnico da porta
	Desc    string // texto principal (nome da placa ou do dispositivo)
	Sub     string // texto da porta (Fones de ouvido, HDMI...)
	Device  string // sink/source ativo que expõe essa porta; vazio se a porta está num perfil inativo
	Profile string // perfil que precisa ser ativado para a porta existir (só quando Device está vazio)
	Avail   string // disponibilidade informada pelo driver

	Percent    int
	Mute       bool
	IsDefault  bool // o dispositivo é o padrão do sistema
	PortActive bool // a porta é a que o dispositivo está usando agora
	Builtin    bool // alto-falante ou microfone interno
}

// Live: a porta já existe no sistema (não precisa trocar de perfil).
func (e Entry) Live() bool { return e.Device != "" }

// Current: é exatamente isto que está em uso agora (marcado com ● na tela).
func (e Entry) Current() bool { return e.Live() && e.IsDefault && e.PortActive }

// Controllable: dá para ajustar volume e mudo (só vale para a porta ativa).
func (e Entry) Controllable() bool { return e.Live() && e.PortActive }

// Unavailable: o driver diz que não há nada conectado (ex.: HDMI sem cabo).
// Portas embutidas nunca contam como desconectadas: alguns drivers marcam o
// alto-falante como indisponível quando há algo no P2, e ele não deve sumir.
func (e Entry) Unavailable() bool { return e.Avail == "not available" && !e.Builtin }

// Snapshot é a foto do estado do áudio: as entradas de cada direção.
// Indexe com Output ou Input.
type Snapshot [2][]Entry

// state é a leitura bruta do pactl, antes de virar Entry.
type state struct {
	devs     [2][]device
	defaults [2]string
	cards    []card
}

// buildEntries monta a lista de uma direção. Ideia central do app:
//
//  1. Para cada porta de cada placa, procura um dispositivo ativo que a ofereça.
//     Se achar, a entrada é "viva" (tem volume e mudo).
//  2. Se não achar, a porta está num perfil inativo. Guardamos qual perfil
//     ativar, para que um Enter seja suficiente para usá-la.
//  3. Dispositivos sem placa (virtuais, etc.) entram como entradas simples.
func buildEntries(st state, k Kind) []Entry {
	devs, def, other := st.devs[k], st.defaults[k], st.devs[k.Other()]
	var out []Entry
	claimed := map[string]bool{} // dispositivos já representados por portas de placas

	for _, c := range st.cards {
		keep := activePorts(c, other)

		for _, name := range c.sortedPortNames() {
			if pk, ok := portKind(name); !ok || pk != k {
				continue // porta da outra direção
			}
			p := c.Ports[name]
			e := Entry{
				Card:    c.Name,
				Port:    name,
				Desc:    c.description(),
				Sub:     p.Description,
				Avail:   p.Availability,
				Builtin: isBuiltin(name, p),
			}

			if d, ok := findLive(devs, c.Name, name); ok {
				e.Device, e.Percent, e.Mute = d.Name, d.percent(), d.Mute
				e.IsDefault, e.PortActive = d.Name == def, d.ActivePort == name
				claimed[d.Name] = true
			} else {
				e.Profile = pickProfile(c, p, keep)
				if e.Profile == "" {
					continue // nenhum perfil oferece essa porta: não há como ativar
				}
			}
			out = append(out, e)
		}
	}

	for _, d := range devs {
		if claimed[d.Name] {
			continue
		}
		base := Entry{Device: d.Name, Desc: d.Description, Percent: d.percent(), Mute: d.Mute, IsDefault: d.Name == def, PortActive: true}
		if len(d.Ports) == 0 {
			out = append(out, base)
			continue
		}
		for _, p := range d.Ports {
			e := base
			e.Port, e.PortActive = p.Name, p.Name == d.ActivePort
			out = append(out, e)
		}
	}
	return out
}

// findLive procura um dispositivo ativo da placa que ofereça a porta.
func findLive(devs []device, cardName, port string) (device, bool) {
	for _, d := range devs {
		if belongs(d, cardName) && d.hasPort(port) {
			return d, true
		}
	}
	return device{}, false
}

// activePorts lista as portas em uso, na direção oposta, nos dispositivos da placa.
// Servem para escolher um perfil que não as derrube.
func activePorts(c card, other []device) []string {
	var ports []string
	for _, d := range other {
		if belongs(d, c.Name) && d.ActivePort != "" {
			ports = append(ports, d.ActivePort)
		}
	}
	return ports
}

// pickProfile escolhe o perfil a ativar para que a porta p passe a existir.
//
// Entre os perfis que oferecem p, vence o que mantém mais portas de `keep`
// (o que você já usa na outra direção, ex.: o microfone P2 ao ativar o HDMI).
// Em caso de empate, vale a prioridade que o próprio driver atribui ao perfil.
// Devolve "" se nenhum perfil serve.
func pickProfile(c card, p cardPort, keep []string) string {
	best, bestScore := "", -1.0
	for _, name := range p.Profiles {
		prof, ok := c.Profiles[name]
		if !ok || !prof.usable() {
			continue
		}
		kept := 0
		for _, kp := range keep {
			for _, pn := range c.Ports[kp].Profiles {
				if pn == name {
					kept++
					break
				}
			}
		}
		// O peso 1e6 garante que "manter mais portas" sempre ganha da prioridade.
		if score := float64(kept)*1e6 + prof.Priority; score > bestScore {
			best, bestScore = name, score
		}
	}
	return best
}

// isBuiltin identifica alto-falante e microfone interno pelo nome ou tipo da porta.
func isBuiltin(name string, p cardPort) bool {
	tag := strings.ToLower(name + " " + p.Type)
	return strings.Contains(tag, "speaker") || strings.Contains(tag, "internal")
}
