# life-calculators · Python-порт

Консольная версия «Калькуляторов на все случаи жизни» (v1.3, 40 калькуляторов).
Чистый Python 3, без зависимостей.

## Запуск

```bash
python3 -m life_calculators.cli            # интерактивное меню
python3 -m life_calculators.cli list       # список всех калькуляторов
python3 -m life_calculators.cli run loan p=2000000 r=21 m=48
python3 -m life_calculators.cli run dates a=2026-01-05 b=2026-01-31
python3 -m life_calculators.cli run recipe from=4 to=8 i0="Мука;250;г" i1="Сахар;150;г"
python3 -m life_calculators.cli run units kind=mass from=кг to=г value=2
```

## Тесты

```bash
python3 -m unittest discover -s tests -v
```

## Структура

- `life_calculators/core.py` — чистое ядро (все формулы; 1-в-1 с JS-ядром `index.html`)
- `life_calculators/formatting.py` — форматирование чисел как в оригинале
- `life_calculators/catalog.py` — 40 калькуляторов: поля, подсказки, функции результата
- `life_calculators/cli.py` — консольный интерфейс
- `tests/` — юнит-тесты (сверка с эталоном JS)

## Корректность

Вывод сверен автоматически с оригинальным JS (Node) по всем 40 калькуляторам:
формулы и строки совпадают 1-в-1. Три отличия — намеренные (в оригинале это баги):

- «Роды» — сдвиг Нагеле показывается в строке (в веб-версии прятался в невидимом параметре);
- «Краска» — объём банки показывается реальный (в веб-версии баг «Банок по 0 л»);
- «Стаканы» — подсказка «1 стакан = …» выводится отдельной строкой (в веб-версии терялась).

Числа при этом совпадают с оригиналом везде.
