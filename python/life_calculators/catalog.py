# -*- coding: utf-8 -*-
"""Каталог калькуляторов (40) — Python-порт массива CALCS из index.html.
Каждый элемент: dict(id, cat, emoji, title, kw, inputs, hint, calc).
calc(values) -> список строк результата вида (label, value[, big]) | None.
value уже отформатировано (как в JS-рендере).
"""
from __future__ import annotations
import math
from datetime import date, datetime
from . import core as C
from .formatting import fmt_num, fmt_r, fmt_dur

def today_str():
    return date.today().strftime('%Y-%m-%d')

def num(v):
    if v is None or v == '':
        return None
    try:
        n = float(str(v).replace('\u00a0', '').replace(' ', '').replace(',', '.'))
    except (ValueError, TypeError):
        return None
    return n if math.isfinite(n) else None

MONTHS_RU = ['января', 'февраля', 'марта', 'апреля', 'мая', 'июня',
             'июля', 'августа', 'сентября', 'октября', 'ноября', 'декабря']

def _fmt_date(d):
    """Эквивалент toLocaleDateString('ru-RU', day/month/year)."""
    return '{} {} {} г.'.format(d.day, MONTHS_RU[d.month - 1], d.year)

def _today():
    return date.today()

CATS = [
    ('all', 'Все', ''),
    ('money', '💰 Деньги', ''),
    ('home', '🏠 Дом', ''),
    ('food', '🍽️ Еда', ''),
    ('math', '📐 Математика', ''),
    ('health', '❤️ Здоровье', ''),
    ('travel', '🚗 Поездки', ''),
]

CALCS = []

def calc(id, cat, emoji, title, kw, inputs, fn, hint=None):
    CALCS.append(dict(id=id, cat=cat, emoji=emoji, title=title, kw=kw,
                      inputs=inputs, calc=fn, hint=hint))
    return fn

def input(k, l, defv='', type_='text', opts=None):
    return dict(k=k, l=l, defv=defv, type=type_, opts=opts or [])

def calccalc(fn):  # noqa - декоратор регистрирует функцию ниже
    return fn

# ============================ ДЕНЬГИ =====================================

def _calc_loan(v):
    p, r_, m = num(v['p']), num(v['r']), round(num(v['m']) or 0)
    r = C.annuity(p, r_, m)
    if not r or not p:
        return None
    return [['Ежемесячный платёж', fmt_r(r['pmt']), 1],
            ['Всего выплатите', fmt_r(r['total'])],
            ['Переплата (проценты)', fmt_r(r['interest'])],
            ['Переплата, % от суммы', fmt_num(r['interest'] / p * 100, 1) + ' %']]
calc('loan', 'money', '🏦', 'Кредит / кредитка', 'ипотека банковский платёж аннуитет переплата',
     [input('p', 'Сумма, ₽', '1000000'), input('r', 'Ставка, % годовых', '19'),
      input('m', 'Срок, месяцев', '60')], _calc_loan,
     'Аннуитетный платёж: одинаковый каждый месяц. Сначала банк «съедает» проценты, к концу срока — тело долга.')

def _calc_deposit(v):
    p, r_, m, add = num(v['p']) or 0, num(v['r']), round(num(v['m']) or 0), num(v['add']) or 0
    r = C.deposit_fv(p, r_, m, add)
    if not r:
        return None
    mm = num(v['m']) or 1
    return [['Будет на счёте', fmt_r(r['end']), 1],
            ['Из них ваши вложения', fmt_r(r['contrib'])],
            ['Проценты', fmt_r(r['interest'])],
            ['Прибыль в месяц', fmt_r(r['interest'] / mm)],
            ['Прибыль, % от вложений', fmt_num(r['interest'] / (r['contrib'] or 1) * 100, 1) + ' %']]
calc('deposit', 'money', '🏛️', 'Депозит', 'вклад процент капиталация накопления',
     [input('p', 'Начальная сумма, ₽', '1000000'), input('r', 'Ставка, % годовых', '18'),
      input('m', 'Срок, месяцев', '12'), input('add', 'Взнос каждый месяц, ₽ (0 — без)', '0')],
     _calc_deposit, 'Капитализация ежемесячная: проценты прибавляются к телу, и дальше процент идёт и на проценты.')

def _calc_vat1(v):
    rate = num(v['r'])
    r = C.vat_extract(num(v['a']), 20 if rate is None else rate)
    if not r:
        return None
    return [['Цена без НДС', fmt_r(r['base']), 1], ['НДС', fmt_r(r['vat'])],
            ['НДС, % от базы', fmt_num(r['vat'] / (r['base'] or 1) * 100, 1) + ' %']]
calc('vat1', 'money', '🧾', 'НДС: выделить из цены', 'налог добавленная стоимость 20 извлечь',
     [input('a', 'Сумма с НДС, ₽', '120000'), input('r', 'Ставка НДС, %', '20')], _calc_vat1)

def _calc_vat2(v):
    rate = num(v['r'])
    r = C.vat_add(num(v['a']) or 0, 20 if rate is None else rate)
    if not r:
        return None
    return [['Цена с НДС', fmt_r(r['total']), 1], ['НДС', fmt_r(r['vat'])],
            ['Без НДС', fmt_r(r['total'] - r['vat'])]]
calc('vat2', 'money', '➕', 'НДС: наложить сверху', 'считать цену с налогом наценка',
     [input('a', 'Цена без НДС, ₽', '100000'), input('r', 'Ставка НДС, %', '20')], _calc_vat2)

def _calc_tax(v):
    rate = num(v['r'])
    r = C.tax_calc(num(v['g']) or 0, num(v['n']) or 0, 13 if rate is None else rate)
    if not r:
        return None
    return [['Brutto (до вычета)', fmt_r(r['gross']), 1], ['НДФЛ', fmt_r(r['tax'])],
            ['Netto (на руки)', fmt_r(r['net'])],
            ['На руки — % от brutto', fmt_num(r['net'] / (r['gross'] or 1) * 100, 1) + ' %']]
calc('tax', 'money', '💼', 'НДФЛ: brutto/netto', 'налог зарплата на руки отчёт',
     [input('g', 'До вычета (brutto), ₽ — заполнить один из двух', ''),
      input('n', 'После вычета (netto), ₽', '100000'), input('r', 'Ставка, %', '13')],
     _calc_tax, 'Заполните только одно поле: brutto (до налога) или netto (на руки) — второе посчитается.')

def _calc_percent(v):
    x, p, a, b = num(v['x']), num(v['p']), num(v['a']), num(v['b'])
    rows = []
    if x is not None and p is not None:
        px = fmt_num(p, 2)
        rows.append(['{}% от {}'.format(px, fmt_num(x, 2)), fmt_num(C.pct_of(x, p), 2), 1])
        rows.append(['Добавить {}%: {} →'.format(px, fmt_num(x, 2)), fmt_num(x + C.pct_of(x, p), 2)])
        rows.append(['Отнять {}%: {} →'.format(px, fmt_num(x, 2)), fmt_num(x - C.pct_of(x, p), 2)])
    if a is not None and b is not None and b != 0:
        rows.append(['Число А — это % от Б', fmt_num(C.pct_of_base(a, b)) + ' %'])
    return rows or None
calc('percent', 'money', '➗', 'Проценты', 'процент от числа сколько составляет скидка наценка',
     [input('x', 'Число', '200'), input('p', '…процентов от него, %', '15'),
      input('a', 'Число А (сколько % от Б?)', '30'), input('b', 'Число Б', '200')], _calc_percent)

def _calc_discount(v):
    p = num(v['p'])
    r = C.discount(p, 0 if num(v['d']) is None else num(v['d']))
    if not r:
        return None
    return [['Итоговая цена', fmt_r(r['final']), 1], ['Выгода', fmt_r(r['saved'])],
            ['Это % от исходной цены', fmt_num(r['final'] / (r['price'] or 1) * 100, 1) + ' %']]
calc('discount', 'money', '🏷️', 'Скидка и итоговая цена', 'скидка акция цена магазин',
     [input('p', 'Цена, ₽', '2990'), input('d', 'Скидка, %', '25')], _calc_discount)

def _calc_salary(v):
    r = C.salary_breakdown(num(v['m']), num(v['w']))
    if not r:
        return None
    return [['В час', fmt_r(r['hourly']), 1], ['За день (8 ч)', fmt_r(r['day8'])],
            ['За неделю', fmt_r(r['week'])], ['Оvertime (час по 1,5×)', fmt_r(r['hourly'] * 1.5)],
            ['В год', fmt_r(r['year'])]]
calc('salary', 'money', '⏱️', 'Зарплата во времени', 'час день смена ставка почасово',
     [input('m', 'Зарплата в месяц, ₽', '200000'), input('w', 'Часов в неделю', '40')],
     _calc_salary, 'В месяце в среднем 4.33 недели.')

def _calc_inflation(v):
    a = num(v['a']) or 0
    r = C.inflation_power(a, num(v['r']) or 0, num(v['y']) or 0)
    if not r:
        return None
    rows = [['Такая же покупательная способность будет', fmt_r(r['value']), 1],
            ['Потеряно', fmt_r(r['lost'])]]
    if r['value'] > 0:
        rows.append(['Деньги станут слабее в', fmt_num(a / r['value'], 2) + ' раза'])
    return rows
calc('inflation', 'money', '📉', 'Инфляция: сила денег', 'обесценивание рублей цены годы',
     [input('a', 'Сумма сегодня, ₽', '100000'), input('r', 'Инфляция, % в год', '8'),
      input('y', 'Годов', '10')], _calc_inflation)

def _calc_unitprice(v):
    r = C.unit_price(num(v['p']), num(v['g']))
    if not r:
        return None
    rows = [['За 100 г', fmt_r(r['per100']), 1], ['За 1 кг', fmt_r(r['per1000'])]]
    kg = num(v['kg'])
    if kg and kg > 0:
        rows.append(['Стоимость ' + fmt_num(kg, 2) + ' кг', fmt_r(r['per1000'] * kg)])
    return rows
calc('unitprice', 'money', '🛒', 'Цена за 100 г / кг', 'весовое сравнить магазин сырок упаковка',
     [input('p', 'Цена упаковки, ₽', '189'), input('g', 'Вес, г', '250'),
      input('kg', 'Сколько купить, кг (опционально)', '')], _calc_unitprice)

def _calc_savings(v):
    months = round(num(v['m']) or 0)
    r = C.savings_monthly(num(v['goal']), num(v['r']), months, num(v['s']) or 0)
    if not r:
        return None
    if r['pmt'] <= 0:
        return [['Откладывать в месяц', fmt_r(0) + ' ₽', 1],
                ['Уже накоплено достаточно — цель достижима', '✓']]
    return [['Откладывать в месяц', fmt_r(r['pmt']), 1],
            ['Всего внесёте сами', fmt_r(r['paid'])],
            ['Прирост за счёт процентов', fmt_r(r['interest'])]]
calc('savings', 'money', '🐷', 'Копилка: накопить цель',
     'откладывать копить накопления цель вклад каждый месяц',
     [input('goal', 'Цель, ₽', '500000'), input('s', 'Уже отложено, ₽', '100000'),
      input('r', 'Ставка по накоплениям, % годовых', '18'), input('m', 'Срок, месяцев', '24')],
     _calc_savings, 'Обратная задача к депозиту: сколько откладывать в месяц, чтобы через M месяцев выйти на цель.')

def _calc_rule72(v):
    r = C.rule72_calc(num(v['r']) or 0, num(v['y']) or 0)
    if not r:
        return None
    return [['Капитал вырастет в … раз', fmt_num(r['growth'], 1) + '×', 1],
            ['Срок удвоения (правило 72)', fmt_num(r['d72'], 1) + ' лет'],
            ['Срок удвоения (точно)', fmt_num(r['dexact'], 1) + ' лет']]
calc('rule72', 'money', '📈', 'Правило 72: рост и удвоение',
     'срок удвоения сложный процент во сколько раз вырастет инвестиции доходность',
     [input('r', 'Доходность, % годовых', '12'), input('y', 'Горизонт, лет', '10')],
     _calc_rule72, '72 ÷ ставка ≈ годы, за которые капитал удвоится (надёжно для ставок до ~15 %).')

# ============================ ДОМ ========================================

def _calc_room(v):
    r = C.room_geom(num(v['l']), num(v['w']), num(v['h']) or 0)
    if not r:
        return None
    rows = [['Площадь пола', fmt_num(r['area'], 2) + ' м²', 1],
            ['Периметр', fmt_num(r['per'], 2) + ' м'],
            ['Площадь стен', fmt_num(r['walls'], 2) + ' м²'],
            ['Потолок', fmt_num(r['ceil'], 2) + ' м²']]
    o = num(v['o'])
    if o and o > 0:
        rows.append(['Стены за вычетом проёмов', fmt_num(max(0, r['walls'] - o), 2) + ' м²'])
    return rows
calc('room', 'home', '📏', 'Комната: площадь и стены', 'ремонт метры квадратная стены потолки',
     [input('l', 'Длина, м', '4'), input('w', 'Ширина, м', '3'), input('h', 'Высота потолка, м', '2.7'),
      input('o', 'Двери и окна, м² (опционально)', '')], _calc_room)

def _calc_paint(v):
    r = C.paint_needed(num(v['a']), num(v['c']), num(v['n']) or 1, num(v['can']) or 0)
    if not r:
        return None
    rows = [['Нужно краски', fmt_num(r['liters'], 2) + ' л', 1]]
    if r['cans'] is not None:
        rows.append(['Банок по ' + fmt_num(num(v['can']) or 0, 1) + ' л', r['cans']])
    pr = num(v['pr'])
    if pr and pr > 0:
        rows.append(['Стоимость ≈', fmt_r(r['liters'] * pr)])
    return rows
calc('paint', 'home', '🎨', 'Краска / шпаклёвка', 'краска банки литраж укрывистость слои',
     [input('a', 'Площадь, м²', '60'), input('c', 'Расход: м² на литр', '10'),
      input('n', 'Слоёв', '2'), input('can', 'Объём банки, л (0 — без)', '2.5'),
      input('pr', 'Цена, ₽/л (опционально)', '')], _calc_paint,
     'Возьмите с запасом ~10%: стены «пьют» по-разному.')

def _calc_floor(v):
    r = C.flooring_needed(num(v['a']), num(v['l']) or 0, num(v['p']) or 0)
    if not r:
        return None
    rows = [['Купить с запасом', fmt_num(r['need'], 2) + ' м²', 1]]
    if r['packs'] is not None:
        rows.append(['Упаковок', r['packs']])
    pr = num(v['pr'])
    if pr and pr > 0:
        rows.append(['Стоимость ≈', fmt_r(r['need'] * pr)])
    return rows
calc('floor', 'home', '🪵', 'Ламинат / плитка / линолеум', 'напольное покрытие запас упаковка подрезка',
     [input('a', 'Площадь пола, м²', '20'), input('l', 'Запас на подрезку, %', '8'),
      input('p', 'м² в упаковке (0 — без)', '2'), input('pr', 'Цена, ₽/м² (опционально)', '')], _calc_floor)

def _calc_watt(v):
    r = C.appliance_cost(num(v['w']), num(v['h']) or 0, num(v['d']) or 0, num(v['r']) or 0)
    if not r:
        return None
    return [['Расход', fmt_num(r['kwh'], 2) + ' кВт·ч', 1], ['Стоимость', fmt_r(r['cost'])],
            ['За год (×12)', fmt_r(r['cost'] * 12)]]
calc('watt', 'home', '⚡', 'Электричество: сколько стоит прибор', 'ватт киловатты чайник расход света',
     [input('w', 'Мощность, Вт', '2000'), input('h', 'Часов в день', '2'),
      input('d', 'Дней', '30'), input('r', 'Тариф, ₽ за кВт·ч', '6')], _calc_watt)

def _calc_tv(v):
    r = C.tv_diag(num(v['i']), num(v['c']))
    if not r:
        return None
    rows = [['Дюймов', fmt_num(r['inch'], 1), 1], ['Сантиметров', fmt_num(r['cm'], 1) + ' см']]
    vd = C.tv_view_dist(r['inch'])
    if vd:
        rows.append(['Дистанция просмотра (HD)',
                     fmt_num(vd['min_m'], 1) + ' – ' + fmt_num(vd['max_m'], 1) + ' м'])
    return rows
calc('tv', 'home', '📺', 'Диагональ ТВ: дюймы ↔ см', 'телевизор диагональ размер дюйм',
     [input('i', 'Дюймы (заполните одно из двух)', '55'), input('c', '…или сантиметры', '')], _calc_tv)

def _calc_watts(v):
    r = C.watts_amps(num(v['w']), num(v['v']), num(v['cp']))
    if not r:
        return None
    rows = [['Ток', fmt_num(r['amps'], 2) + ' А', 1]]
    if r['breaker']:
        rows.append(['Ближайший автомат', str(r['breaker']) + ' А'])
    if r['cable']:
        rows.append(['Кабель (медь, скрытая прокладка)', r['cable']])
    return rows
calc('watts', 'home', '🔌', 'Ватты ↔ амперы, автомат', 'розетка автомат пробки ток нагрузка',
     [input('w', 'Мощность прибора, Вт', '2000'), input('v', 'Напряжение, В (по умолчанию 220)', '220'),
      input('cp', 'cos φ (только для двигателей/инверторов, по умолчанию 1)', '')], _calc_watts,
     'A = Вт/(В × cos φ). «Автомат» — ближайший стандартный номинал (6…100 А), «Кабель» — примерное сечение меди для скрытой проводки.')

def _calc_wallpaper(v):
    r = C.wallpaper(num(v['per']), num(v['h']), num(v['rw']), num(v['rl']),
                    num(v['pat']), num(v['skip']))
    if not r:
        return None
    return [['Нужно рулонов', str(r['rolls']) + ' шт', 1],
            ['Всего полос по периметру', str(r['stripsTotal']) + ' шт'],
            ['Полос из одного рулона', str(r['stripsPer']) + ' шт'],
            ['Купленная площадь (с запасом)', fmt_num(r['area'], 1) + ' м²']]
calc('wallpaper', 'home', '🧱', 'Обои: сколько рулонов', 'поклейка ремонт рулон комната стены периметр',
     [input('per', 'Периметр комнаты (сумма длин стен), м', '14'),
      input('h', 'Высота потолка, м', '2.5'), input('rw', 'Ширина рулона, м', '0.53'),
      input('rl', 'Длина рулона, м', '10'), input('pat', 'Подгонка рисунка (раппорт), см', ''),
      input('skip', 'Полос, закрытых проёмами (дверь ≈ 1, окно ≈ 1–2)', '')], _calc_wallpaper,
     'Полосы считаются по периметру: из рулона выходит столько полос, сколько раз в его длине умещается высота стены плюс раппорт.')

def _calc_temp(v):
    r = C.temp_conv(num(v['c']), num(v['f']))
    if not r:
        return None
    return [['По Цельсию', fmt_num(r['c'], 0) + ' °C', 1],
            ['По Фаренгейту', fmt_num(r['f'], 0) + ' °F']]
calc('temp', 'home', '🌡️', 'Температура: °C ↔ °F', 'духовка цельсий фаренгейт градусы рецепт выпечка',
     [input('c', 'Градусы Цельсия, °C', '180'), input('f', 'Градусы Фаренгейта, °F', '')], _calc_temp,
     'Заполните °C или °F. Духовка (примерно): 160 °C = 320 °F, 180 °C = 350 °F, 200 °C = 400 °F.')

# ============================ ЕДА ========================================
def _calc_recipe(v):
    return None  # используется build_recipe_flow ниже (динамические ингредиенты)
calc('recipe', 'food', '🧑‍🍳', 'Рецепт: пересчёт порций', 'порции масштабировать блюдо ингредиент',
     [input('from', 'Порций было', '4'), input('to', 'Порций нужно', '8')], _calc_recipe,
     'Все ингредиенты умножаются на одно и то же число — пропорции сохраняются.')

def _calc_calories(v):
    r = C.serving_nutrition(dict(kcal=num(v['k']) or 0, p=num(v['p']) or 0,
                                 f=num(v['f']) or 0, c=num(v['c']) or 0), num(v['g']))
    if not r:
        return None
    return [['В порции', fmt_num(r['kcal'], 1) + ' ккал', 1], ['Белки', fmt_num(r['p'], 1) + ' г'],
            ['Жиры', fmt_num(r['f'], 1) + ' г'], ['Углеводы', fmt_num(r['c'], 1) + ' г'],
            ['% от дневных 2000 ккал', fmt_num(r['kcal'] / 2000 * 100, 0) + ' %']]
calc('calories', 'food', '🔥', 'Калории порции', 'бжу ккал упаковка на 100 грамм',
     [input('k', 'Ккал на 100 г', '200'), input('p', 'Белки, г на 100 г', '4'),
      input('f', 'Жиры, г на 100 г', '3'), input('c', 'Углеводы, г на 100 г', '10'),
      input('g', 'Ваша порция, г', '50')], _calc_calories)

def _calc_cups(v):
    prod = v['prod']
    r = C.cups_convert(prod, num(v['cups']), num(v['grams']))
    if not r:
        return None
    extra = [['1 стакан =', fmt_num(C.CUPS[prod]['g'], 0) + ' г']] if prod in C.CUPS else []
    if r['dir'] == 'cups->g':
        return [['В граммах', fmt_num(r['g'], 1) + ' г', 1], ['Стаканов', fmt_num(r['cups'], 2)]] + extra
    return [['В стаканах', fmt_num(r['cups'], 2) + ' стак.', 1],
            ['Граммов', fmt_num(r['g'], 1) + ' г']] + extra
calc('cups', 'food', '🥣', 'Стаканы ↔ граммы (выпечка)', 'чашки стаканы мука сахар выпечка конвертер',
     [dict(k='prod', l='Продукт', type='select',
           opts=[[k, C.CUPS[k]['name']] for k in C.CUPS], defv='flour'),
      input('cups', 'Стаканов (если считаете стаканы)', ''), input('grams', '…или граммов', '130')],
     _calc_cups)

def _calc_tip(v):
    pct = num(v['p'])
    n = round(num(v['n']) or 1)
    r = C.tip_split(num(v['b']), 10 if pct is None else pct, n)
    if not r:
        return None
    rows = [['Итого с чаевыми', fmt_r(r['total']), 1], ['Чаевые', fmt_r(r['tip'])]]
    if r['per'] < r['total']:
        rows.append(['С каждого человека', fmt_r(r['per'])])
        rnd = math.ceil(r['per'] / 50) * 50
        if rnd > r['per']:
            rows.append(['Ровно, округлив до 50 ₽', 'по ' + fmt_r(rnd) + ' — итого ' + fmt_r(rnd * n)])
    return rows
calc('tip', 'food', '🍽️', 'Чаевые и делёж', 'чаевые ресторан счёт пополам делим',
     [input('b', 'Счёт, ₽', '3500'), input('p', 'Чаевые, %', '10'),
      input('n', 'Человек делим (1 — без делёжа)', '2')], _calc_tip)

def _calc_dilute(v):
    v1 = num(v['v1'])
    r = C.dilute_calc(num(v['c1']), num(v['c2']), v1, num(v['vt']))
    if not r:
        return None
    rows = []
    if v1 is None and r['src'] is not None:
        rows.append(['Взять исходного', fmt_num(r['src'], 1) + ' мл'])
    rows.append(['Добавить воды', fmt_num(r['water'], 1) + ' мл', 1])
    rows.append(['Итоговый объём', fmt_num(r['total'], 1) + ' мл'])
    return rows
calc('dilute', 'food', '🍶', 'Разбавление: уксус и спирт', 'уксусная эссенция развести спирт самогон вода концентрация 70 в 9',
     [input('c1', 'Концентрация исходная, %', '70'), input('c2', 'Нужная концентрация, %', '9'),
      input('v1', 'Объём исходного, мл', ''), input('vt', 'Сколько нужно готового раствора, мл', '1000')],
     _calc_dilute, 'Заполните объём исходного ИЛИ нужный объём готового. Пример: 100 мл 70%-й эссенции → 9%-й уксус — добавить ~678 мл воды.')

# ============================ ПОЕЗДКИ ====================================
def _calc_fuel(v):
    price = num(v['p'])
    r = C.fuel_cost(num(v['km']), num(v['l']), 0 if price is None else price)
    if not r:
        return None
    rows = [['Нужно топлива', fmt_num(r['liters'], 1) + ' л', 1]]
    if r['cost'] is not None:
        rows += [['Стоимость', fmt_r(r['cost'])], ['Туда и обратно (×2)', fmt_r(r['cost'] * 2)]]
    rows.append(['CO₂ в атмосферу ≈', fmt_num(r['liters'] * 2.31, 1) + ' кг'])
    return rows
calc('fuel', 'travel', '⛽', 'Топливо на поездку', 'бензин дизель расход литры маршрут',
     [input('km', 'Километров', '450'), input('l', 'Расход, л/100 км', '8'),
      input('p', 'Цена, ₽/л', '60')], _calc_fuel)

# ============================ МАТЕМАТИКА =================================
def _calc_shapes(v):
    a, b = num(v['a']), num(v['b'])
    r = C.shape_calc(v['k'], a, b)
    if not r:
        return None
    rows = [[r['name'], fmt_num(r['value'], 3) + ' ' + r['unit'], 1]]
    if v['k'] == 'circleR' and a:
        rows.append(['Длина окружности (2πr)', fmt_num(2 * math.pi * a, 3)])
    elif v['k'] == 'circleD' and a:
        rows.append(['Длина окружности (πd)', fmt_num(math.pi * a, 3)])
    elif v['k'] == 'rect' and a and b:
        rows.append(['Периметр (2(a+b))', fmt_num(2 * (a + b), 3)])
    elif v['k'] == 'cyl' and a and b and b > 0:
        rows.append(['Боковая площадь (2πrh)', fmt_num(2 * math.pi * a * b, 3)])
    return rows
calc('shapes', 'math', '📐', 'Площади и объёмы', 'круг прямоугольник треугольник цилиндр сфера',
     [dict(k='k', l='Фигура', type='select',
           opts=[['circleR', 'Круг (по радиусу)'], ['circleD', 'Круг (по диаметре)'],
                 ['rect', 'Прямоугольник'], ['tri', 'Прямоугольный треугольник'],
                 ['cyl', 'Цилиндр'], ['sphere', 'Сфера']], defv='circleR'),
      input('a', 'Значение A (r / сторона / a)', '2'),
      input('b', 'Значение B (если нужно: h / вторая сторона)', '5')], _calc_shapes)

def _calc_avg(v):
    s = C.stats_of(C.parse_numbers(v['list']))
    if not s:
        return None
    return [['Чисел', s['n']], ['Сумма', fmt_num(s['sum'], 2)],
            ['Среднее', fmt_num(s['mean'], 2), 1], ['Медиана', fmt_num(s['median'], 2)],
            ['Минимум', fmt_num(s['min'], 2)], ['Максимум', fmt_num(s['max'], 2)],
            ['Размах (max − min)', fmt_num(s['max'] - s['min'], 2)]]
calc('avg', 'math', '🧮', 'Среднее из списка чисел', 'медиана сумма минимум максимум статистика',
     [input('list', 'Числа (через пробел или запятую)', '120 95 110 130 98', type_='text')], _calc_avg)

def _calc_vts(v):
    r = C.vts(num(v['s']), num(v['t']), num(v['d']))
    if not r:
        return None
    if r.get('dist') is not None:
        return [['Путь', fmt_num(r['dist'], 2) + ' км', 1]]
    if r.get('time') is not None:
        return [['Время', fmt_num(r['time'], 2) + ' ч', 1], ['Минут', fmt_num(r['time'] * 60, 0)]]
    return [['Скорость', fmt_num(r['speed'], 2) + ' км/ч', 1], ['В м/с', fmt_num(r['speed'] / 3.6, 2)]]
calc('vts', 'math', '🚗', 'Скорость · время · путь', 'расстояние часы км в час маршрут',
     [input('s', 'Скорость (км/ч)', '60'), input('t', 'Время (ч)', '2'),
      input('d', 'Путь (км) — оставьте пустым то, что ищем', '')], _calc_vts,
     'Заполните любые два поля из трёх — третье посчитается.')

def _calc_dates(v):
    r = C.date_diff(v['a'], v['b'])
    if not r:
        return None
    return [['Дней', r['days'], 1], ['Часов', fmt_num(r['hours'], 0)],
            ['Недель', fmt_num(r['weeks'], 1)], ['≈ месяцев', fmt_num(r['months'], 1)],
            ['≈ лет', fmt_num(r['years'], 2)], ['Будних дней (пн–пт, с датами)', r['workdays']]]
calc('dates', 'math', '📅', 'Между датами', 'дней сколько разницы дата возраст',
     [dict(k='a', l='Дата 1', type='date', defv=None),
      dict(k='b', l='Дата 2', type='date', defv=None)], _calc_dates)

def _calc_units(v):
    return None  # интерактивный конвертер — отдельный flow в CLI
calc('units', 'math', '🔄', 'Конвертер единиц', 'перевод длина масса объём см м кг грамм литры дюймы футы мили галлон',
     [], _calc_units, 'Например: 1 м = 100 см, 1 кг = 1000 г, 1 л = 1000 мл.')

# ============================ ЗДОРОВЬЕ ===================================
def _calc_bmi(v):
    h = num(v['h'])
    r = C.bmi_of(h, num(v['w']))
    if not r:
        return None
    hm = h / 100
    return [['ИМТ', fmt_num(r['v'], 1), 1], ['Категория (ВОЗ)', r['cat']],
            ['Вес при ИМТ 18.5–24.9', fmt_num(18.5 * hm * hm, 1) + ' – ' + fmt_num(24.9 * hm * hm, 1) + ' кг']]
calc('bmi', 'health', '⚖️', 'ИМТ (индекс массы тела)', 'имт вес рост категория воз',
     [input('h', 'Рост, см', '176'), input('w', 'Вес, кг', '80')], _calc_bmi,
     'По классификации ВОЗ. ИМТ не видит распределение мышц/жиров — это ориентир, не диагноз.')

def _calc_tdee(v):
    w = num(v['w'])
    r = C.tdee(v['sex'], num(v['age']), num(v['h']), w, num(v['act']))
    if not r:
        return None
    return [['Обмен веществ (BMR)', fmt_num(r['bmr'], 0) + ' ккал'],
            ['Норма (TDEE)', fmt_num(r['tdee'], 0) + ' ккал', 1],
            ['Похудение (−20%)', fmt_num(r['cut'], 0) + ' ккал'],
            ['Набор (+10%)', fmt_num(r['gain'], 0) + ' ккал'],
            ['Белок, ориентир (1,6 г/кг)', fmt_num((w or 0) * 1.6, 0) + ' г/день']]
calc('tdee', 'health', '🔥', 'Калории в день', 'тdee bmr метаболизм похудеть набрать миффлин',
     [dict(k='sex', l='Пол', type='select', opts=[['m', 'Мужской'], ['f', 'Женский']], defv='m'),
      input('age', 'Возраст, лет', '30'), input('h', 'Рост, см', '176'), input('w', 'Вес, кг', '80'),
      dict(k='act', l='Активность', type='select',
           opts=[['1.2', 'Диван (сидячая работа)'], ['1.375', 'Лёгкая (1–3 тренировки/нед)'],
                 ['1.55', 'Средняя (3–5 тренировок/нед)'], ['1.72', 'Высокая (6–7 тренировок/нед)'],
                 ['1.9', 'Спортсмен / физ. работа']], defv='1.375')], _calc_tdee,
     'Формула Миффлина–Сан Жеора. −20% — плавное похудение, +10% — набор.')

def _calc_macros(v):
    r = C.macros(num(v['k']), num(v['p']) or 0, num(v['f']) or 0, num(v['c']) or 0)
    if not r:
        return None
    return [['Белки', fmt_num(r['p'], 1) + ' г', 1], ['Жиры', fmt_num(r['f'], 1) + ' г'],
            ['Углеводы', fmt_num(r['c'], 1) + ' г'],
            ['На один приём (3 приёма)',
             fmt_num(r['p'] / 3, 1) + ' / ' + fmt_num(r['f'] / 3, 1) + ' / ' + fmt_num(r['c'] / 3, 1) + ' г']]
calc('macros', 'health', '🥩', 'БЖУ из калорий', 'белки жиры углеводы граммы соотношение',
     [input('k', 'Калорий в день', '2200'), input('p', 'Белки, %', '30'),
      input('f', 'Жиры, %', '20'), input('c', 'Углеводы, %', '50')], _calc_macros,
     'Белок и углеводы — 4 ккал/г, жиры — 9 ккал/г. Доли в сумме лучше = 100%.')

def _calc_water(v):
    r = C.water_liters(num(v['w']), v['a'] == '1')
    if not r:
        return None
    rows = [['Воды в день', fmt_num(r['liters'], 2) + ' л', 1],
            ['Стаканов по 0.3 л', fmt_num(r['liters'] / 0.3, 1)]]
    if v['a'] != '1':
        rows.append(['С тренировками (+500 мл)', fmt_num(r['liters'] + 0.5, 2) + ' л'])
    return rows
calc('water', 'health', '💧', 'Норма воды', 'литры жидкость пить активность',
     [input('w', 'Вес, кг', '75'),
      dict(k='a', l='Есть тренировки (35 мл/кг вместо 30)', type='select',
           opts=[['0', 'Нет'], ['1', 'Да']], defv='0')], _calc_water)

def _calc_oner(v):
    r = C.one_rm(num(v['w']), num(v['r']))
    if not r:
        return None
    return [['Примерный 1ПМ', fmt_num(r['v'], 1) + ' кг', 1],
            ['10 повторов ≈ 70% 1ПМ', fmt_num(r['v'] * 0.7, 1) + ' кг'],
            ['6 повторов ≈ 80% 1ПМ', fmt_num(r['v'] * 0.8, 1) + ' кг'],
            ['3 повтора ≈ 90% 1ПМ', fmt_num(r['v'] * 0.9, 1) + ' кг']]
calc('oner', 'health', '🏋️', 'Рабочий 1ПМ (Эйпли)', 'один повтор максимум силовой тренажёр',
     [input('w', 'Вес, кг', '100'), input('r', 'Сколько раз сделали', '5')], _calc_oner,
     '1ПМ = вес × (1 + повторы/30). Формула Эйпли — оценка, не мера точности.')

def _calc_metcal(v):
    r = C.met_calories(num(v['w']), num(v['m']), num(v['h']))
    if not r:
        return None
    h = num(v['h'])
    per_hour = r['kcal'] / (h if (h or 0) > 0 else 1)
    rows = [['Сожжено', fmt_num(r['kcal'], 0) + ' ккал', 1]]
    if per_hour > 0:
        rows.append(['Чтобы сжечь 1 кг жира (7 700 ккал)',
                     fmt_num(7700 / per_hour, 1) + ' ч такой активности'])
    return rows
calc('metcal', 'health', '🏃', 'Калории за активность', 'мет меты сожечь ходьба бег тренировка',
     [input('w', 'Вес, кг', '75'),
      dict(k='m', l='Активность', type='select',
           opts=[['2.5', 'Ходьба 5 км/ч'], ['4', 'Быстрая ходьба 6,5 км/ч'], ['6', 'Бег 8 км/ч'],
                 ['7', 'Плавание'], ['8', 'Велосипед 16 км/ч'], ['10', 'Тяжёлая тренировка / уборка']],
           defv='2.5'),
      input('h', 'Часов', '0.5')], _calc_metcal,
     'Ккал = MET × вес × время. MET — коэффициент метаболического эквивалента.')

def _calc_due(v):
    r = C.due_date(v['l'], num(v['c']))
    if not r:
        return None
    d = datetime.strptime(r['due'], '%Y-%m-%d').date()
    rows = [['Ожидаемая дата', _fmt_date(d), 1]]
    if r['shift'] != 0:
        base_d = datetime.strptime(r['base'], '%Y-%m-%d').date()
        shift = int(r['shift'])
        sign = '+' if shift > 0 else ''
        rows.append(['Нагеле (цикл 28 дн)',
                     _fmt_date(base_d) + ' (сдвиг ' + sign + str(shift) + ' дн.)'])
    pw = C.pregnancy_weeks(v['l'])
    if pw:
        rows.append(['Срок сейчас', pw['label'], 1])
    left = (d - _today()).days
    if left > 0:
        rows.append(['До предполагаемой даты', str(left) + ' дн.'])
    elif left == 0:
        rows.append(['Сегодня', 'предполагаемая дата 🎉'])
    else:
        rows.append(['Срок перешагнут', 'на ' + str(-left) + ' дн.'])
    return rows
calc('due', 'health', '🤰', 'Роды: предполагаемая дата', 'беременность срок гестация последний день',
     [dict(k='l', l='Первый день последней менструации', type='date', defv=None),
      input('c', 'Длина цикла, дней (опционально, по умолчанию 28)', '')], _calc_due,
     'Формула Нагеле: +280 дней (40 недель от последней менструации). При цикле ≠28 дней дата сдвигается.')

def _calc_pace(v):
    r = C.pace_calc(num(v['spd']), num(v['pmin']), num(v['psec']))
    if not r:
        return None
    rows = [['Темп', '{}:{:02d} мин/км'.format(r['paceM'], r['paceS']), 1],
            ['Скорость', fmt_num(r['spd'], 1) + ' км/ч']]
    d = num(v['dist'])
    if d and d > 0:
        rows.append(['Время на ' + fmt_num(d, 2) + ' км', fmt_dur(r['paceMin'] * d)])
    rows.append(['Время на 5 км', fmt_dur(r['paceMin'] * 5)])
    rows.append(['Время на 10 км', fmt_dur(r['paceMin'] * 10)])
    rows.append(['Время на 21,1 км', fmt_dur(r['paceMin'] * 21.0975)])
    rows.append(['Время на 42,2 км', fmt_dur(r['paceMin'] * 42.195)])
    return rows
calc('pace', 'health', '🏃', 'Темп бега: мин/км ↔ км/ч', 'бег марафон скорость время дистанция тренировка',
     [input('spd', 'Скорость, км/ч', ''), input('pmin', 'Темп: минут на км', '6'),
      input('psec', 'Темп: секунд на км', '0'), input('dist', 'Своя дистанция, км (для времени)', '')],
     _calc_pace, 'Заполните скорость ИЛИ темп (темп приоритетнее). Темп 6:00 мин/км ≈ 10 км/ч.')

def _calc_hrzones(v):
    r = C.hr_zones(num(v['age']) or 0, num(v['mx']))
    mx = r['max']
    if float(mx).is_integer():
        mx = int(mx)
    rows = [['Максимальный пульс', str(mx) + ' уд/мин', 1]]
    for z in r['zones']:
        rows.append([z['pct'] + ' — ' + z['name'], str(z['lo']) + '–' + str(z['hi']) + ' уд/мин'])
    return rows
calc('hrzones', 'health', '💓', 'Пульсовые зоны', 'сердце пульс тренировка зоны чсс кардио бег возраст интенсивность',
     [input('age', 'Возраст, лет', '30'), input('mx', 'Максимальный пульс, уд/мин', '')], _calc_hrzones,
     'Пусто в поле пульса = 220 − возраст. Зоны считаются как % от максимального пульса.')

# recipe / units — специальные потоки
def recipe_rows(from_p, to_p, items):
    """Из общего recipe_scale: пересчёт + построчные результаты."""
    r = C.recipe_scale(from_p, to_p, items)
    if not r or not r['items']:
        return None
    rows = [['Коэффициент', '×' + fmt_num(r['k'], 2), 1]]
    for it in r['items']:
        amount_str = fmt_num(it['amount'], 1 if it['amount'] % 1 else 0)
        scaled_str = fmt_num(it['scaled'], 1 if it['scaled'] % 1 else 0)
        name = it['name'] or '—'
        unit = it['unit']
        rows.append(['{} · {} {}'.format(name, amount_str, unit),
                     '{} {}'.format(scaled_str, unit)])
    if to_p > 0:
        for it in r['items']:
            per = it['scaled'] / to_p
            per_str = fmt_num(per, 1 if per % 1 else 0)
            rows.append(['На порцию: ' + (it['name'] or '—'), per_str + ' ' + (it['unit'] or '')])
    return rows

def units_flow(kind, frm, to, value):
    """Конвертер единиц; возвращает строку результата."""
    if num(value) is None:
        return None
    r = C.unit_convert(kind, frm, to, num(value))
    if r is None:
        return None
    return [['{} {} ='.format(fmt_num(num(value), 6), frm), '{} {}'.format(fmt_num(r, 6), to), 1]]
