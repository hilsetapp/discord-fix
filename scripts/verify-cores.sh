#!/usr/bin/env bash
# Проверяет, что бинари ядер лежат ровно те, что положены -
# сверяет sha256 с таблицей официальных релизов.
# Запускать из корня репозитория: ./scripts/verify-cores.sh
set -u

ok=1
check() {
  local path="$1" want="$2"
  if [ ! -f "$path" ]; then
    echo "[ПЛОХО] $path - файл не найден"
    ok=0
    return
  fi
  local got
  got=$(sha256sum "$path" | awk '{print $1}')
  if [ "$got" = "$want" ]; then
    echo "[ок] $path"
  else
    echo "[ПЛОХО] $path - хеш не совпал"
    echo "  ждали: $want"
    echo "  есть:  $got"
    ok=0
  fi
}

check cores/zapret/winws.exe       a14bff1df6234ea555d2e0c61b589f0707c0b12d6c9b7eeccda76012154996e8
check cores/zapret/WinDivert.dll   c1e060ee19444a259b2162f8af0f3fe8c4428a1c6f694dce20de194ac8d7d9a2
check cores/zapret/WinDivert64.sys 8da085332782708d8767bcace5327a6ec7283c17cfb85e40b03cd2323a90ddc2
check cores/zapret/cygwin1.dll     103104a52e5293ce418944725df19e2bf81ad9269b9a120d71d39028e821499b
check cores/gdpi/goodbyedpi.exe    331ac6c1d22ba5a0a217f3f27d0d823051869cafc8b8ef7f2002fa2accebc74e
check cores/gdpi/WinDivert.dll     a97859785a2df1d4462e7d48d33ccbd89fedd40dac4970f4afd89e63f59ee1ec
check cores/gdpi/WinDivert64.sys   53ab28ec00be6e6f8aefa9ee76fc2735e94d7f3f9dbc06eb2b7ac8cd3084a6af

if [ "$ok" = "1" ]; then
  echo "всё совпало - бинари официальные"
else
  echo "есть расхождения - не запускай, напиши нам issue"
  exit 1
fi
