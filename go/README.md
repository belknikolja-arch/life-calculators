# life-calculators · Go-порт

Консольная версия «Калькуляторов на все случаи жизни» (v1.3, 40 калькуляторов) на Go.
Без внешних зависимостей (только стандартная библиотека).

## Сборка и запуск

```bash
go build -o life-calculators-go .
./life-calculators-go list                     # список калькуляторов
./life-calculators-go                          # интерактивное меню
./life-calculators-go run loan p=2000000 r=21 m=48
./life-calculators-go run dates a=2026-01-05 b=2026-01-31
./life-calculators-go run recipe from=4 to=8 i0="Мука;250;г" i1="Сахар;150;г"
./life-calculators-go run units kind=mass from=кг to=г value=2
```

или без сборки: `go run . list` / `go run . run loan p=…`.

## Тесты

```bash
go test ./...
```

## Структура

- `core.go` — чистое ядро: все формулы + форматирование (1-в-1 с JS-ядром `index.html`)
- `calcs.go` — каталог 40 калькуляторов: поля, подсказки, функции результата
- `main.go` — CLI (интерактив, `list`, `run`, спец-режимы рецепта и конвертера)
- `core_test.go` — юнит-тесты

## Корректность

Численные результаты сверены с оригинальным JS (Node) и Python-портом по всем
калькуляторам. Отличия от веб-версии — только намеренные, в местах багов
оригинала (см. `python/README.md`).
