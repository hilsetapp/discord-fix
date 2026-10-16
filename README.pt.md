[English](README.en.md) | [Русский](README.md) | [Español](README.es.md) | **Português**

# Discord Fix

Um clique e o Discord volta a funcionar sob bloqueio de DPI. O resto da rede fica intocado: so os dominios do Discord passam pelo desvio.

Este app foi feito para o publico falante de russo e para o ambiente de bloqueios russo. A documentacao completa esta em [README.md](README.md) (russo).

## O que e

Um pequeno app para Windows com janela escura e um botao grande. Ao apertar, o app testa varias estrategias de desbloqueio conhecidas e verifica depois de cada uma se discord.com abre de verdade. Quando algo funciona, ele mantem ligado. Voltam o site, o app de desktop e os canais de voz.

Motores usados, sem modificacao, como publicados pelos autores:

- [zapret](https://github.com/bol-van/zapret) (winws) de bol-van: modo principal, suporta listas de dominio e UDP (voz)
- [GoodbyeDPI](https://github.com/ValdikSS/GoodbyeDPI) de ValdikSS: modo de reserva simples, so TCP (sem voz)

## Download

Baixe `discord-fix-windows.zip` em [Releases](../../releases), descompacte e rode `discord-fix.exe`, depois aperte o botao grande. Se o SmartScreen reclamar, clique em "Mais informacoes" e "Executar assim mesmo": o binario e compilado pelo GitHub Actions a partir deste repositorio.

## Escopo, com honestidade

- Nao e VPN, nao e proxy e nao tem recurso de anonimato.
- Apenas a lista de dominios do Discord recebe o desvio (veja `lists/discord.txt`). Seu banco, seus jogos e todo o resto seguem na conexao normal.
- Pacotes de voz UDP do Discord tambem sao tratados (nos modos baseados em zapret).
- O app nao envia nada para casa; ele so consulta dois endpoints publicos do Discord para saber se o desbloqueio funcionou.

## Compilar

Windows, Go 1.24+:

```
go build -ldflags "-H=windowsgui -s -w" -o discord-fix.exe ./cmd/discord-fix
```

Modos sem interface: `-check` (o Discord responde?), `-selftest` (confere os hashes dos motores), `-run=auto` (sem janela, escolhe a estrategia sozinho).

## Licenca

Codigo do app: MIT. Os motores mantem as proprias licencas (zapret MIT, GoodbyeDPI Apache-2.0, WinDivert LGPLv3): veja `THIRD_PARTY_NOTICES.md` e `licenses/`.
