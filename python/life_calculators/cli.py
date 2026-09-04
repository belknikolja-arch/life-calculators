# -*- coding: utf-8 -*-
"""Консольный интерфейс Python-порта калькуляторов.

Примеры:
    python -m life_calculators.cli                 # интерактивное меню
    python -m life_calculators.cli list            # список калькуляторов
    python -m life_calculators.cli run loan p=2000000 r=21 m=48
    python -m life_calculators.cli run dates a=2026-01-01 b=2026-09-03
    python -m life_calculators.cli run recipe from=4 to=8 i0="Мука;250;г" i1="Сахар;150;г"
    python -m life_calculators.cli run units kind=mass from=кг to=г value=2
"""
from __future__ import annotations
import sys
from datetime import date
from . import catalog as CAT
from .catalog import num, today_str

CATNAME = {c[0]: c[1] for c in CAT.CATS}

def default_value(inp):
    d = inp['defv']
    if callable(d):
        return d()
    if inp.get('type') == 'date' and d is None:
        return today_str()
    return '' if d is None else str(d)

def resolve(item, kv):
    """kv: dict ключ->строка; недостающие ключи = значения по умолчанию."""
    v = {}
    for inp in item['inputs']:
        k = inp['k']
        v[k] = kv.get(k, default_value(inp))
    return v

def print_rows(rows):
    if not rows:
        print('  (нет результата — проверьте ввод)')
        return
    for r in rows:
        label, value = r[0], r[1]
        big = r[2] if len(r) > 2 else None
        star = '★' if big else ' '
        if isinstance(value, float) and value.is_integer():
            value = int(value)
        print(' {} {:<42} {}'.format(star, label, value))

def ask_input(inp):
    label = inp['l']
    if inp.get('type') == 'select':
        opts = inp['opts']
        print('  Выберите: ' + label)
        for i, o in enumerate(opts, 1):
            print('    {}. {}'.format(i, o[1]))
        ans = input('    номер [{}]: '.format(inp['defv'])).strip()
        if ans == '':
            sel = inp['defv']
        else:
            try:
                sel = opts[int(ans) - 1][0]
            except (ValueError, IndexError):
                print('  Не понял — беру по умолчанию.')
                sel = inp['defv']
        return str(sel)
    dv = default_value(inp)
    ph = ' [{}]'.format(dv) if dv != '' else ''
    ans = input('  {}:{} '.format(label, ph)).strip()
    return ans if ans != '' else dv

def special_recipe(kv):
    """Рецепт из kv: from,to + i0..iN = 'название;кол-во;единица'."""
    try:
        from_p = num(kv.get('from', '4')) or 0
        to_p = num(kv.get('to', '8')) or 0
    except Exception:
        return
    items = []
    for k in sorted(kv):
        if k.startswith('i') and k[1:].isdigit():
            parts = [p.strip() for p in kv[k].split(';')]
            name = parts[0]
            amount = num(parts[1]) if len(parts) > 1 else None
            unit = parts[2] if len(parts) > 2 else ''
            if amount and amount > 0:
                items.append(dict(name=name, amount=amount, unit=unit))
    rows = CAT.recipe_rows(from_p, to_p, items)
    print('  Пересчёт рецепта ({} -> {} порций):'.format(from_p, to_p))
    print_rows(rows)

def special_units(kv):
    kind = kv.get('kind', 'length')
    frm = kv.get('from', 'м')
    to = kv.get('to', 'см')
    val = kv.get('value', '')
    if val == '':
        print('  Укажите value=число.')
        return
    rows = CAT.units_flow(kind, frm, to, val)
    print_rows(rows)

def run_interactive(item):
    v = resolve(item, {})
    if item['id'] == 'recipe':
        print('== {} {} =='.format(item['emoji'], item['title']))
        print('  Введите ингредиенты (пустое название — конец):')
        items = []
        while True:
            name = input('    Название: ').strip()
            if not name:
                break
            amount = input('    Кол-во: ').strip()
            unit = input('    Единица (г/мл): ').strip()
            am = num(amount)
            if am:
                items.append(dict(name=name, amount=am, unit=unit))
        from_p = num(input('  Порций было [4]: ') or '4')
        to_p = num(input('  Порций нужно [8]: ') or '8')
        rows = CAT.recipe_rows(from_p or 0, to_p or 0, items)
        print_rows(rows)
        return
    if item['id'] == 'units':
        print('== {} {} =='.format(item['emoji'], item['title']))
        kinds = list(CAT.C.UNITS.keys())
        names = {'length': 'Длина', 'mass': 'Масса', 'volume': 'Объём'}
        print('  Величина:', ', '.join('{}. {}'.format(i + 1, names[k]) for i, k in enumerate(kinds)))
        try:
            kind = kinds[int(input('  Номер: ').strip()) - 1]
        except Exception:
            kind = 'length'
        units = list(CAT.C.UNITS[kind].keys())
        print('  Единицы:', ', '.join(units))
        frm = input('  Из единицы [м]: ').strip() or 'м'
        to = input('  В единицу [см]: ').strip() or 'см'
        val = input('  Значение: ').strip()
        print_rows(CAT.units_flow(kind, frm, to, val))
        return
    print('== {} {} =='.format(item['emoji'], item['title']))
    if item.get('hint'):
        print('  ⓘ ' + item['hint'])
    for inp in item['inputs']:
        v[inp['k']] = ask_input(inp)
    rows = item['calc'](v)
    print_rows(rows)

def run_noninteractive(item, kv):
    if item['id'] == 'recipe':
        special_recipe(kv)
        return
    if item['id'] == 'units':
        special_units(kv)
        return
    v = resolve(item, kv)
    print('== {} {} =='.format(item['emoji'], item['title']))
    rows = item['calc'](v)
    print_rows(rows)

def find_by_id(i):
    for c in CAT.CALCS:
        if c['id'] == i:
            return c
    return None

def cmd_list():
    prev = None
    for c in CAT.CALCS:
        cat = CATNAME.get(c['cat'], c['cat'])
        if cat != prev:
            print('\n{}'.format(cat))
            prev = cat
        print('  {:<10} {}{}'.format(c['id'], c['emoji'], c['title']))
    print('\nВсего: {} калькуляторов.'.format(len(CAT.CALCS)))

def interactive():
    print('🧮 Калькуляторы на все случаи жизни — Python-порт v1.3')
    print('Введите номер калькулятора, "list" — список, "q" — выход.')
    while True:
        order = [(i + 1, c) for i, c in enumerate(CAT.CALCS)]
        print('\nДоступно: {}. Показать?'.format(len(order)))
        try:
            line = input('> ').strip().lower()
        except (EOFError, KeyboardInterrupt):
            print('\nПока!')
            return
        if line in ('q', 'quit', 'exit'):
            return
        if line in ('list', 'l', 'список'):
            cmd_list()
            continue
        if line == 'help':
            print('Номер из меню / list / q')
            continue
        if not line:
            continue
        if line.isdigit():
            n = int(line)
            item = order[n - 1][1] if 1 <= n <= len(order) else None
            if not item:
                print('Нет такого номера.')
                continue
            run_interactive(item)
            continue
        item = find_by_id(line)
        if item:
            run_interactive(item)
        else:
            print('Неизвестно: ' + line)

def main(argv=None):
    argv = argv if argv is not None else sys.argv[1:]
    if argv and argv[0] == 'list':
        cmd_list()
        return 0
    if argv and argv[0] in ('run', 'calc'):
        ident = argv[1] if len(argv) > 1 else ''
        item = find_by_id(ident)
        if not item:
            print('Калькулятор не найден: ' + ident)
            return 1
        kv = {}
        for tok in argv[2:]:
            if '=' in tok:
                k, val = tok.split('=', 1)
                kv[k.strip()] = val
        run_noninteractive(item, kv)
        return 0
    interactive()
    return 0

if __name__ == '__main__':
    sys.exit(main())
