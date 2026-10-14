# Проверяет sha256 бинарей ядер против таблицы официальных релизов.
# Запуск: powershell -File scripts\verify-cores.ps1
$ErrorActionPreference = "Stop"
$ok = $true

function Check-Core($Path, $Want) {
    if (-not (Test-Path $Path)) {
        Write-Host "[ПЛОХО] $Path - файл не найден" -ForegroundColor Red
        $script:ok = $false
        return
    }
    $got = (Get-FileHash $Path -Algorithm SHA256).Hash.ToLower()
    if ($got -eq $Want) {
        Write-Host "[ок] $Path" -ForegroundColor Green
    } else {
        Write-Host "[ПЛОХО] $Path - хеш не совпал" -ForegroundColor Red
        Write-Host "  ждали: $Want"
        Write-Host "  есть:  $got"
        $script:ok = $false
    }
}

Check-Core "cores/zapret/winws.exe"       "a14bff1df6234ea555d2e0c61b589f0707c0b12d6c9b7eeccda76012154996e8"
Check-Core "cores/zapret/WinDivert.dll"   "c1e060ee19444a259b2162f8af0f3fe8c4428a1c6f694dce20de194ac8d7d9a2"
Check-Core "cores/zapret/WinDivert64.sys" "8da085332782708d8767bcace5327a6ec7283c17cfb85e40b03cd2323a90ddc2"
Check-Core "cores/gdpi/goodbyedpi.exe"    "331ac6c1d22ba5a0a217f3f27d0d823051869cafc8b8ef7f2002fa2accebc74e"
Check-Core "cores/gdpi/WinDivert.dll"     "a97859785a2df1d4462e7d48d33ccbd89fedd40dac4970f4afd89e63f59ee1ec"
Check-Core "cores/gdpi/WinDivert64.sys"   "53ab28ec00be6e6f8aefa9ee76fc2735e94d7f3f9dbc06eb2b7ac8cd3084a6af"

if ($ok) {
    Write-Host "всё совпало - бинари официальные" -ForegroundColor Green
} else {
    Write-Host "есть расхождения - не запускай, напиши нам issue" -ForegroundColor Red
    exit 1
}
