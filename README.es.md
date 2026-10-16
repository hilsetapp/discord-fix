[Русский](README.md) | [English](README.en.md) | **Español** | [Português](README.pt.md)

# Discord Fix

Un clic y Discord vuelve a funcionar bajo el bloqueo DPI. El resto de la red no se toca: solo los dominios de Discord pasan por el desvio.

Esta app esta hecha para el publico rusoparlante y el entorno de bloqueos ruso. La documentacion completa y las capturas estan en la version rusa: [README.md](README.md).

## Que es

Una pequeña app de Windows con una ventana oscura y un boton grande. Al pulsarlo, la app prueba varias estrategias de desbloqueo conocidas y comprueba despues de cada una si discord.com abre de verdad. Cuando algo funciona, lo deja activo. Vuelven el sitio, la app de escritorio y los canales de voz.

Motores usados, sin modificar, tal como los publican sus autores:

- [zapret](https://github.com/bol-van/zapret) (winws) de bol-van: modo principal, soporta listas de dominios y UDP (voz)
- [GoodbyeDPI](https://github.com/ValdikSS/GoodbyeDPI) de ValdikSS: modo de respaldo simple, solo TCP (sin voz)

## Descarga

Descarga `discord-fix-windows.zip` en [Releases](../../releases), descomprime y ejecuta `discord-fix.exe`, luego pulsa el boton grande. Si SmartScreen se queja, pulsa "Mas informacion" y "Ejecutar de todas formas": el binario lo compila GitHub Actions desde este mismo repositorio.

## Alcance, con honestidad

- No es una VPN, ni un proxy, ni una herramienta de anonimato.
- Solo la lista de dominios de Discord recibe el desvio (ver `lists/discord.txt`). Tu banco, tus juegos y todo lo demas siguen por la conexion normal.
- Los paquetes de voz UDP de Discord tambien se manejan (en los modos basados en zapret).
- La app no envia nada a ningun servidor; solo consulta dos endpoints publicos de Discord para comprobar si el desbloqueo funciono.

## Compilar

Windows, Go 1.24+:

```
go build -ldflags "-H=windowsgui -s -w" -o discord-fix.exe ./cmd/discord-fix
```

Modos sin interfaz: `-check` (ver si Discord responde), `-selftest` (comprobar los hashes de los motores), `-run=auto` (sin ventana, elige estrategia solo).

## Licencia

Codigo de la app: MIT. Los motores mantienen sus licencias (zapret MIT, GoodbyeDPI Apache-2.0, WinDivert LGPLv3): ver `THIRD_PARTY_NOTICES.md` y `licenses/`.
