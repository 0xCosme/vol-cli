# volcli

Gerenciador de áudio para o terminal (Linux). Lista **todas** as saídas e entradas de áudio, deixa escolher qual usar e ajustar volume e captação do microfone. Os perfis das placas são trocados automaticamente, sem você precisar mexer neles.

Funciona com PulseAudio e PipeWire (via `pipewire-pulse`).

## Requisitos

- Go 1.21+
- `pactl` (pacote `pulseaudio-utils`, ou `libpulse` no Arch)
- PulseAudio 16+ ou PipeWire (o app usa `pactl --format=json`)

## Instalação

```bash
go mod tidy          # baixa as dependências (Bubble Tea e Lip Gloss)
go build -o volcli .
sudo mv volcli /usr/local/bin/
```

Para só testar: `go run .`

Testes da lógica de áudio: `go test ./...`

## Uso

Abra com `volcli`. Há duas abas: **Saída** (caixas, fones, HDMI, Bluetooth) e **Entrada** (microfones).

| Tecla | Ação |
|---|---|
| `Tab` | Alterna entre Saída e Entrada |
| `↑` `↓` / `k` `j` | Escolhe o dispositivo |
| `←` `→` / `h` `l` | Volume (ou ganho do microfone) de 5 em 5%, até 150% |
| `Enter` | Usa o dispositivo como padrão |
| `m` | Liga/desliga o mudo |
| `a` | Mostra/oculta dispositivos desconectados |
| `q` | Sai |

O `●` marca o dispositivo padrão atual. A lista atualiza sozinha a cada 2 segundos.

### Exemplo: ouvir no Bluetooth e falar no microfone do fone P2

1. Aba **Saída**: selecione o fone Bluetooth e aperte `Enter`.
2. Aba **Entrada**: selecione "Headset Microphone" da placa interna e aperte `Enter`.

Saída e entrada padrão são independentes, então os dois ficam ativos ao mesmo tempo.

## Como funciona

- Cada placa de som tem **perfis** que definem quais portas (alto-falante, P2, HDMI...) ficam disponíveis. O app lê as portas de todos os perfis, então mostra tudo, inclusive o que está num perfil inativo.
- Ao dar `Enter` num item de outro perfil, o app escolhe um perfil que ofereça aquela porta e, se possível, mantenha a porta que você já usa na outra direção. Depois espera o dispositivo aparecer, define a porta e o padrão, e move os áudios que já estão tocando.
- Tudo é feito chamando o `pactl`. Não há dependência de cgo.

## Estrutura do projeto

```
main.go                      ponto de entrada (confere o pactl e abre a interface)
internal/pulse/              tudo que fala com o áudio, sem nenhuma dependência de tela
  pactl.go                   chama o pactl (texto e JSON)
  kind.go                    saída x entrada e a tabela de comandos de cada uma
  types.go                   tipos que espelham o JSON do pactl
  entries.go                 monta a lista de itens e escolhe o perfil a ativar
  actions.go                 Load, Activate, SetVolume, ToggleMute
  entries_test.go            testes da lógica
internal/ui/                 interface de terminal (Bubble Tea)
  model.go                   estado da tela e tratamento de teclas
  cmds.go                    tarefas assíncronas e atualização a cada 2 s
  view.go                    desenho da tela
```

O pacote `ui` só usa a API exportada de `pulse` (`Load`, `Activate`, `SetVolume`, `ToggleMute`). Para trocar a interface (por exemplo, uma GUI), basta reescrever `internal/ui`.

## Limitações

- Se você escolher o microfone do próprio fone Bluetooth, ele troca para o perfil de headset (HSP/HFP) e o som da saída fica abafado. É uma limitação do Bluetooth.
- Os dispositivos são associados às placas pelo nome. Funciona com ALSA e Bluetooth, mas dispositivos com nomes fora do padrão podem aparecer sem a troca automática de perfil.
- Itens desconectados (ex.: HDMI sem cabo) ficam ocultos por padrão. Alto-falantes e microfones internos nunca são ocultados.

## Solução de problemas

- **"pactl não encontrado"**: instale `pulseaudio-utils` (ou `pipewire-pulse`).
- **Erro ao falar com o servidor de áudio**: confirme que `pactl info` funciona.
- **Dispositivo não aparece**: aperte `a` e confira `pactl list short sinks` e `pactl list short sources`.
