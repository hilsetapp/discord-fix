# Сторонние компоненты

В релизном архиве рядом с приложением лежат готовые бинари сторонних проектов.
Они взяты из официальных релизов авторов без изменений, их контрольные суммы
зашиты в приложение и проверяются скриптами из `scripts/verify-cores`.

## zapret (winws.exe + конфигурационные payload-файлы)

- Автор: bol-van
- Источник: https://github.com/bol-van/zapret (v72.13)
- Лицензия: MIT - текст в `licenses/zapret-MIT.txt`

Используется как основной двигатель обхода: фильтрация по списку доменов
и DPI-desync для TCP и UDP.

## GoodbyeDPI (goodbyedpi.exe)

- Автор: ValdikSS
- Источник: https://github.com/ValdikSS/GoodbyeDPI (0.2.2)
- Лицензия: Apache-2.0 - текст в `licenses/goodbyedpi-APACHE-2.0.txt`

Запасной простой режим обхода без обработки UDP.

## WinDivert (WinDivert.dll, WinDivert64.sys)

- Автор: basil (Diablo-D3 и сообщество)
- Источник: https://reqrypt.org/windivert.html
- Лицензия: GNU LGPL v3 - текст в `licenses/windivert-LGPLv3.txt`

Системный драйвер перехвата пакетов. Оба движка загружают его динамически,
как отдельную библиотеку - форма использования, предусмотренная LGPL.
Исходный код WinDivert доступен на сайте автора.

## uthash, getline

- Внутренние вспомогательные библиотеки zapret, их лицензии -
  в `licenses/uthash-BSD.txt` и `licenses/getline-MIT.txt`.

## Шрифт Inter

- Автор: Rasmus Andersson (rsms)
- Источник: https://github.com/rsms/inter (v4.1)
- Лицензия: SIL Open Font License 1.1 - https://openfontlicense.org

Шрифт интерфейса (переменный Inter с кириллицей), встроен в приложение.

## Gio

- Источник: https://gioui.org
- Лицензия: MIT + Unlicense (отдельные части)

GUI-фреймворк, на котором нарисовано окно приложения.
