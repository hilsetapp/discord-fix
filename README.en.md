# Discord Fix

One click, and Discord works again under a DPI block. Everything else on the network is left untouched: only Discord domains are routed through the bypass.

**Language:** this app is built for the Russian-speaking audience and the Russian blocking environment. The main documentation is in [README.md](README.md) (RU).

## What it is

A small Windows app with a dark window and one big button. Press it, and the app tries several known-good unblocking strategies, checking after each one whether discord.com actually opens. When something works, it keeps it running. Website, desktop app and voice channels all come back.

Engines under the hood, unmodified, taken from the authors' official releases:

- [zapret](https://github.com/bol-van/zapret) (winws) by bol-van - primary, supports domain lists and UDP (voice)
- [GoodbyeDPI](https://github.com/ValdikSS/GoodbyeDPI) by ValdikSS - simple fallback mode, TCP only (no voice)

## Download

Grab `discord-fix-windows.zip` from [Releases](../../releases), unpack, run `discord-fix.exe`, press the big button. If SmartScreen complains, click "More info" then "Run anyway" - the binary is built by GitHub Actions from this exact repository.

## Scope, honestly

- Not a VPN, not a proxy, no anonymity features.
- Only the Discord domain list gets the bypass treatment (see `lists/discord.txt`). Your banks, games and everything else stay on the plain connection.
- UDP voice packets for Discord calls are handled too (in zapret-based modes).
- The app phones nothing home; it only probes two public Discord endpoints to detect whether the unblock worked.

## Building

Windows, Go 1.24+:

```
go build -ldflags "-H=windowsgui -s -w" -o discord-fix.exe ./cmd/discord-fix
```

Headless modes: `-check` (is Discord reachable), `-selftest` (verify pinned engine hashes), `-run=auto` (no window, pick strategy automatically).

## License

App code: MIT. Engines keep their own licenses (zapret MIT, GoodbyeDPI Apache-2.0, WinDivert LGPLv3) - see `THIRD_PARTY_NOTICES.md` and `licenses/`.
