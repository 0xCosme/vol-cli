package pulse

import "testing"

// Placa de exemplo parecida com a de um notebook: analógica + HDMI.
func testCard() card {
	return card{
		Name:       "alsa_card.pci-0000_00_1f.3",
		Properties: map[string]interface{}{"device.description": "Áudio interno"},
		Profiles: map[string]cardProfile{
			"output:analog-stereo+input:analog-stereo": {Priority: 6500, Available: true},
			"output:hdmi-stereo":                       {Priority: 5900, Available: true},
			"output:hdmi-stereo+input:analog-stereo":   {Priority: 5400, Available: true},
		},
		Ports: map[string]cardPort{
			"analog-output-speaker": {Description: "Alto-falantes", Type: "Speaker", Availability: "not available",
				Profiles: []string{"output:analog-stereo+input:analog-stereo"}},
			"hdmi-output-0": {Description: "HDMI", Type: "HDMI", Availability: "available",
				Profiles: []string{"output:hdmi-stereo", "output:hdmi-stereo+input:analog-stereo"}},
			"analog-input-mic": {Description: "Microfone", Type: "Mic", Availability: "available",
				Profiles: []string{"output:analog-stereo+input:analog-stereo", "output:hdmi-stereo+input:analog-stereo"}},
		},
	}
}

// Ao ativar o HDMI, o perfil escolhido deve manter o microfone que já está em uso.
func TestPickProfileKeepsOtherDirection(t *testing.T) {
	c := testCard()
	got := pickProfile(c, c.Ports["hdmi-output-0"], []string{"analog-input-mic"})
	if want := "output:hdmi-stereo+input:analog-stereo"; got != want {
		t.Fatalf("perfil = %q, esperado %q", got, want)
	}
}

// Sem nada para manter, vale a maior prioridade.
func TestPickProfileFallsBackToPriority(t *testing.T) {
	c := testCard()
	got := pickProfile(c, c.Ports["hdmi-output-0"], nil)
	if want := "output:hdmi-stereo"; got != want {
		t.Fatalf("perfil = %q, esperado %q", got, want)
	}
}

// Porta ativa vira entrada "viva"; porta de perfil inativo guarda o perfil a ativar.
func TestBuildEntriesLiveAndInactive(t *testing.T) {
	var st state
	st.cards = []card{testCard()}
	st.devs[Output] = []device{{
		Name:        "alsa_output.pci-0000_00_1f.3.analog-stereo",
		Description: "Áudio interno",
		ActivePort:  "analog-output-speaker",
		Ports:       []devicePort{{Name: "analog-output-speaker"}},
		Volume:      map[string]channelVolume{"front-left": {Value: 32768}},
	}}
	st.defaults[Output] = st.devs[Output][0].Name

	got := buildEntries(st, Output)
	if len(got) != 2 {
		t.Fatalf("esperadas 2 entradas de saída, vieram %d", len(got))
	}
	byPort := map[string]Entry{}
	for _, e := range got {
		byPort[e.Port] = e
	}

	spk := byPort["analog-output-speaker"]
	if !spk.Live() || !spk.Current() || spk.Percent != 50 {
		t.Errorf("alto-falante deveria estar ativo a 50%%: %+v", spk)
	}
	hdmi := byPort["hdmi-output-0"]
	if hdmi.Live() || hdmi.Profile == "" {
		t.Errorf("HDMI deveria exigir troca de perfil: %+v", hdmi)
	}
}

// O alto-falante nunca pode ser tratado como "desconectado", mesmo que o driver diga isso.
func TestBuiltinNeverUnavailable(t *testing.T) {
	c := testCard()
	spk := Entry{Avail: "not available", Builtin: isBuiltin("analog-output-speaker", c.Ports["analog-output-speaker"])}
	if spk.Unavailable() {
		t.Error("alto-falante embutido não deveria ficar indisponível")
	}
	if !(Entry{Avail: "not available"}).Unavailable() {
		t.Error("porta externa sem nada conectado deveria ficar indisponível")
	}
}

func TestBelongs(t *testing.T) {
	d := device{Name: "alsa_output.pci-0000_00_1f.3.analog-stereo"}
	if !belongs(d, "alsa_card.pci-0000_00_1f.3") {
		t.Error("o dispositivo deveria pertencer à placa")
	}
	if belongs(d, "alsa_card.usb-Outra_Placa-00") {
		t.Error("o dispositivo não deveria pertencer a outra placa")
	}
}
