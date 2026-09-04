# -*- coding: utf-8 -*-
"""Чистое ядро калькуляторов — Python-порт JS-ядра из index.html.
Функции чисто математические (без ввода/вывода), повторяют поведение оригинала
1-в-1 (валидация, границы, округления). Ниже — комментарии-соответствия.
"""
from __future__ import annotations
import math
from datetime import date, datetime, timedelta

# ---------------------------------------------------------------------------
# деньги  (соответствует: annuity, depositFV, vatExtract, vatAdd, taxCalc,
#           pctOf, pctOfBase, discount, salaryBreakdown, inflationPower)
# ---------------------------------------------------------------------------

def annuity(P, rate_pct, months):
    if not (P and P > 0) or not (months >= 1) or not (rate_pct >= 0):
        return None
    r = rate_pct / 100 / 12
    if r == 0:
        pmt = P / months
    else:
        pmt = P * r / (1 - math.pow(1 + r, -months))
    return dict(pmt=pmt, total=pmt * months, interest=pmt * months - P)

def deposit_fv(P, rate_pct, months, monthly):
    if not (P >= 0) or not (months >= 1) or not (rate_pct >= 0):
        return None
    r = rate_pct / 100 / 12
    add = monthly or 0
    bal = P
    for _ in range(months):
        bal = bal * (1 + r) + add
    contrib = P + add * months
    return dict(end=bal, contrib=contrib, interest=bal - contrib)

def vat_extract(amount, rate=None):
    if not (amount and amount > 0):
        return None
    r = 20 if rate is None else rate
    base = amount / (1 + r / 100)
    return dict(base=base, vat=amount - base)

def vat_add(base, rate=None):
    if not (base and base >= 0):
        return None
    r = 20 if rate is None else rate
    vat = base * r / 100
    return dict(total=base + vat, vat=vat)

def tax_calc(gross, net, rate=None):
    r = 13 if rate is None else rate
    if not (0 < r < 100):
        return None
    if gross and gross > 0:
        tax = gross * r / 100
        return dict(gross=gross, tax=tax, net=gross - tax)
    if net and net > 0:
        gross = net / (1 - r / 100)
        return dict(gross=gross, tax=gross - net, net=net)
    return None

def pct_of(x, p):
    return x * p / 100

def pct_of_base(a, b):
    return a / b * 100 if b else None

def discount(price, p):
    if not (price and price > 0) or not (0 <= p <= 100):
        return None
    final = price * (1 - p / 100)
    return dict(final=final, saved=price - final, price=price)

def salary_breakdown(monthly, hw):
    if not (monthly and monthly > 0):
        return None
    w = hw if hw and hw > 0 else 40
    per_month_hours = w * 4.33
    hourly = monthly / per_month_hours
    return dict(hourly=hourly, day8=hourly * 8, week=monthly * w / 40, year=monthly * 12)

def inflation_power(amount, rate_pct, years):
    if not (amount and amount >= 0) or rate_pct < 0 or not (years and years >= 0):
        return None
    k = math.pow(1 + rate_pct / 100, years)
    value = amount / k
    return dict(value=value, lost=amount - value)

# ---------------------------------------------------------------------------
# дом  (roomGeom, paintNeeded, flooringNeeded, applianceCost)
# ---------------------------------------------------------------------------

def _rnd6(x):
    return round(x * 1e6) / 1e6   # JS: Math.round(x*1e6)/1e6

def room_geom(l, w, h):
    if not (l and l > 0 and w and w > 0):
        return None
    hh = h if h and h > 0 else 0
    return dict(area=_rnd6(l * w), per=_rnd6(2 * (l + w)),
                walls=_rnd6(2 * (l + w) * hh), ceil=_rnd6(l * w))

def paint_needed(area, coverage, coats, can_l):
    if not (area and area > 0 and coverage and coverage > 0):
        return None
    c = coats if coats and coats > 0 else 1
    liters = area * c / coverage
    cans = math.ceil(liters / can_l - 1e-9) if can_l and can_l > 0 else None
    return dict(liters=liters, cans=cans)

def flooring_needed(area, loss_pct, pack_m):
    if not (area and area > 0) or (loss_pct and loss_pct < 0):
        return None
    need = area * (1 + loss_pct / 100)
    packs = math.ceil(need / pack_m - 1e-9) if pack_m and pack_m > 0 else None
    return dict(need=need, packs=packs)

def appliance_cost(watts, hours_day, days, rate):
    if not (watts and watts > 0 and hours_day >= 0 and days >= 0):
        return None
    kwh = watts * hours_day * days / 1000
    return dict(kwh=kwh, cost=kwh * (rate or 0))

# ---------------------------------------------------------------------------
# еда  (recipeScale, servingNutrition, CUPS, cupsConvert)
# ---------------------------------------------------------------------------

def recipe_scale(from_p, to_p, items):
    if not (from_p and from_p > 0 and to_p and to_p > 0):
        return None
    k = to_p / from_p
    scaled = []
    for it in items:
        if it.get('amount', 0) and it['amount'] > 0:
            scaled.append(dict(name=it.get('name', ''), amount=it['amount'],
                               unit=it.get('unit', ''), scaled=it['amount'] * k))
    return dict(k=k, items=scaled)

def serving_nutrition(per100, grams):
    if not (grams and grams > 0):
        return None
    k = grams / 100
    return dict(kcal=(per100.get('kcal') or 0) * k,
                p=(per100.get('p') or 0) * k,
                f=(per100.get('f') or 0) * k,
                c=(per100.get('c') or 0) * k)

CUPS = {
    'flour': dict(g=130, name='Мука'), 'sugar': dict(g=200, name='Сахар'),
    'butter': dict(g=227, name='Масло слив.'), 'milk': dict(g=240, name='Молоко (мл)'),
    'cocoa': dict(g=100, name='Какао-порошок'), 'choc': dict(g=170, name='Шоколад'),
    'rice': dict(g=180, name='Рис (сухой)'), 'oil': dict(g=240, name='Масло раст. (мл)'),
}

def cups_convert(product, cups, grams):
    c = CUPS.get(product)
    if not c:
        return None
    if cups is not None and cups > 0:
        return dict(g=cups * c['g'], cups=cups, dir='cups->g')
    if grams is not None and grams > 0:
        return dict(g=grams, cups=grams / c['g'], dir='g->cups')
    return None

# ---------------------------------------------------------------------------
# математика  (shapeCalc, parseNumbers, statsOf, vts, dateDiff)
# ---------------------------------------------------------------------------

def shape_calc(kind, a, b):
    if not (a and a > 0):
        return None
    if kind == 'circleR':
        return dict(name='Площадь круга (r=' + _fs(a) + ')', value=math.pi * a * a, unit='кв. ед.')
    if kind == 'circleD':
        return dict(name='Площадь круга (d=' + _fs(a) + ')', value=math.pi * (a / 2) ** 2, unit='кв. ед.')
    if kind == 'rect':
        if not (b and b > 0):
            return None
        return dict(name='Площадь прямоугольника', value=a * b, unit='кв. ед.')
    if kind == 'tri':
        if not (b and b > 0):
            return None
        return dict(name='Площадь прямоугольного треугольника', value=a * b / 2, unit='кв. ед.')
    if kind == 'cyl':
        if not (b and b > 0):
            return None
        return dict(name='Объём цилиндра (r=' + _fs(a) + ', h=' + _fs(b) + ')',
                    value=math.pi * a * a * b, unit='куб. ед.')
    if kind == 'sphere':
        return dict(name='Объём сферы (r=' + _fs(a) + ')', value=4 / 3 * math.pi * a ** 3, unit='куб. ед.')
    return None

def _fs(x):
    return str(int(x)) if float(x).is_integer() else str(x)

def parse_numbers(s):
    if s is None:
        return []
    return [float(t.replace(',', '.')) for t in
            [p for p in __import__('re').split(r'[\s,;]+', str(s)) if p != '']
            if _isnum(t.replace(',', '.'))]

def _isnum(x):
    try:
        float(x)
        return True
    except Exception:
        return False

def stats_of(nums):
    if not nums:
        return None
    s = sorted(nums)
    n = len(s)
    total = sum(s)
    median = s[(n - 1) // 2] if n % 2 else (s[n // 2 - 1] + s[n // 2]) / 2
    return dict(n=n, sum=total, mean=total / n, median=median, min=s[0], max=s[-1])

def vts(speed, time, dist):
    if speed and speed > 0 and time and time > 0:
        return dict(dist=speed * time)
    if speed and speed > 0 and dist and dist > 0:
        return dict(time=dist / speed)
    if time and time > 0 and dist and dist > 0:
        return dict(speed=dist / time)
    return None

def _parse_date(yyyymmdd):
    try:
        return datetime.strptime(str(yyyymmdd), '%Y-%m-%d').date()
    except Exception:
        return None

def date_diff(a, b):
    d1 = _parse_date(a)
    d2 = _parse_date(b)
    if d1 is None or d2 is None:
        return None
    days = abs((d2 - d1).days)
    lo, hi = min(d1, d2), max(d1, d2)
    workdays = 0
    t = lo
    while t <= hi:
        if t.weekday() < 5:  # 0=пн..4=пт (JS getDay 1..5)
            workdays += 1
        t += timedelta(days=1)
    return dict(days=days, weeks=days / 7, months=days / 30.44, years=days / 365.25,
                hours=days * 24, workdays=workdays)

# ---------------------------------------------------------------------------
# здоровье  (bmiOf, tdee, macros, waterLiters, oneRm)
# ---------------------------------------------------------------------------

def bmi_of(h_cm, w_kg):
    if not (h_cm and h_cm > 0 and w_kg and w_kg > 0):
        return None
    h = h_cm / 100
    v = w_kg / (h * h)
    if v < 16:
        cat = 'выраженный дефицит'
    elif v < 18.5:
        cat = 'дефицит веса'
    elif v < 25:
        cat = 'норма'
    elif v < 30:
        cat = 'избыточный вес'
    else:
        cat = 'ожирение'
    return dict(v=v, cat=cat)

def tdee(sex, age, h_cm, w_kg, activity):
    if not (age and age > 0 and h_cm and h_cm > 0 and w_kg and w_kg > 0):
        return None
    base = 10 * w_kg + 6.25 * h_cm - 5 * age + (5 if sex == 'm' else -161)
    act = activity or 1.2
    return dict(bmr=base, tdee=base * act, cut=base * act * 0.8, gain=base * act * 1.1)

def macros(kcal, pct_p, pct_f, pct_c):
    if not (kcal and kcal > 0):
        return None
    return dict(p=kcal * pct_p / 100 / 4, f=kcal * pct_f / 100 / 9, c=kcal * pct_c / 100 / 4)

def water_liters(weight_kg, active):
    if not (weight_kg and weight_kg > 0):
        return None
    return dict(liters=weight_kg * (0.035 if active else 0.03))

def one_rm(weight, reps):
    if not (weight and weight > 0 and reps and reps > 0):
        return None
    return dict(v=weight * (1 + reps / 30))

# ---------------------------------------------------------------------------
# v1.1  (tipSplit, fuelCost, metCalories, dueDate, pregnancyWeeks,
#        tvDiag, tvViewDist, unitPrice, CABLE_CU, wattsAmps)
# ---------------------------------------------------------------------------

def tip_split(bill, pct, people):
    if not (bill and bill > 0) or pct is None or pct < 0:
        return None
    tip = bill * pct / 100
    total = bill + tip
    n = people if people and people > 1 else 1
    return dict(tip=tip, total=total, per=total / n, tipPer=tip / n)

def fuel_cost(km, l100, price):
    if not (km and km > 0 and l100 and l100 > 0):
        return None
    liters = km * l100 / 100
    return dict(liters=liters, cost=liters * price if price is not None else None)

def met_calories(weight_kg, met, hours):
    if not (weight_kg and weight_kg > 0 and met and met > 0 and hours and hours > 0):
        return None
    return dict(kcal=met * weight_kg * hours)

def _fmt_iso(d):
    return d.strftime('%Y-%m-%d')

def due_date(lmp_str, cycle_len):
    d = _parse_date(lmp_str)
    if d is None:
        return None
    base = d + timedelta(days=280)
    shift = (cycle_len - 28) if (cycle_len and 21 < cycle_len < 40) else 0
    due = base + timedelta(days=shift)
    return dict(due=_fmt_iso(due), base=_fmt_iso(base), shift=shift, weeksTotal=40)

def pregnancy_weeks(lmp_str, now_str=None):
    l = _parse_date(lmp_str)
    if l is None:
        return None
    n = _parse_date(now_str) if now_str else date.today()
    days = (n - l).days
    if days < 0 or days > 420:
        return None
    w = days // 7
    rem = days % 7
    label = str(w) + ' нед.' + ((' ' + str(rem) + ' дн.') if rem else '')
    return dict(weeks=w, days=rem, label=label)

def tv_diag(inch, cm):
    if inch is not None and inch > 0:
        return dict(cm=inch * 2.54, inch=inch)
    if cm is not None and cm > 0:
        return dict(cm=cm, inch=cm / 2.54)
    return None

def tv_view_dist(diag_inch):
    if not (diag_inch and diag_inch > 0):
        return None
    return dict(min_m=diag_inch * 1.5 * 2.54 / 100, max_m=diag_inch * 2.5 * 2.54 / 100)

def unit_price(price, amount):
    if not (price and price > 0 and amount and amount > 0):
        return None
    per100 = price / amount * 100
    return dict(per100=per100, per1000=per100 * 10)

CABLE_CU = {6: '1 мм²', 10: '1,5 мм²', 16: '2,5 мм²', 25: '4 мм²', 32: '6 мм²',
            40: '10 мм²', 50: '10 мм²', 63: '16 мм²', 80: '16 мм²', 100: '25 мм²'}

def watts_amps(watts, volts, cos_phi):
    v = volts if volts and volts > 0 else 220
    cp = 1 if cos_phi is None or cos_phi <= 0 else min(cos_phi, 1)
    if not (watts and watts > 0):
        return None
    a = watts / (v * cp)
    breaker = None
    for b in (6, 10, 16, 25, 32, 40, 50, 63, 80, 100):
        if a <= b:
            breaker = b
            break
    return dict(amps=a, watts=watts, volts=v, cosPhi=cp,
                breaker=breaker, cable=CABLE_CU.get(breaker))

# ---------------------------------------------------------------------------
# v1.3  (savingsMonthly, rule72Calc, wallpaper, diluteCalc, UNITS,
#        unitConvert, tempConv, paceCalc, hrZones)
# ---------------------------------------------------------------------------

def savings_monthly(goal, rate_pct, months, start):
    g = goal if goal and goal > 0 else 0
    s = start if start and start > 0 else 0
    m = round(months) if months and months > 0 else 0
    if not m or g <= s:
        return None
    i = rate_pct / 100 / 12 if rate_pct and rate_pct > 0 else 0
    if i == 0:
        pmt = (g - s) / m
        return dict(pmt=pmt, paid=pmt * m, interest=0)
    fv_s = s * math.pow(1 + i, m)
    pmt = (g - fv_s) * i / (math.pow(1 + i, m) - 1)
    paid = pmt * m
    return dict(pmt=pmt, paid=paid, interest=g - s - paid)

def rule72_calc(rate_pct, years):
    r = rate_pct if rate_pct and rate_pct > 0 else 0
    y = years if years and years > 0 else 0
    if not r:
        return None
    return dict(growth=math.pow(1 + r / 100, y), d72=72 / r,
                dexact=math.log(2) / math.log(1 + r / 100))

def wallpaper(perimeter, wall_h, roll_w, roll_l, pattern_cm, skip_strips):
    P = perimeter if perimeter and perimeter > 0 else 0
    h = wall_h if wall_h and wall_h > 0 else 0
    rw = roll_w if roll_w and roll_w > 0 else 0.53
    rl = roll_l if roll_l and roll_l > 0 else 10
    if not P or not h:
        return None
    step = h + (pattern_cm / 100 if pattern_cm and pattern_cm > 0 else 0)
    strips_total = max(0, math.ceil(P / rw) - (round(skip_strips) if skip_strips and skip_strips > 0 else 0))
    strips_per = max(1, math.floor(rl / step)) if rl >= step else 1
    rolls = max(0, math.ceil(strips_total / strips_per))
    return dict(stripsTotal=strips_total, stripsPer=strips_per, rolls=rolls, area=rolls * rl * rw)

def dilute_calc(c1, c2, v1, vt):
    a = c1 if c1 and 0 < c1 <= 100 else 0
    b = c2 if c2 and 0 < c2 <= 100 else 0
    if not a or not b or b > a:
        return None
    has_v1 = v1 is not None and v1 > 0
    has_vt = vt is not None and vt > 0
    if not has_v1 and not has_vt:
        return None
    if has_v1:
        src = v1
        total = v1 * a / b
        water = total - src
    else:
        total = vt
        src = vt * b / a
        water = total - src
    return dict(src=src, water=water, total=total)

UNITS = {
    'length': {'мм': 0.001, 'см': 0.01, 'м': 1, 'км': 1000, 'дюйм': 0.0254,
               'фут': 0.3048, 'ярд': 0.9144, 'миля': 1609.344},
    'mass': {'мг': 0.001, 'г': 1, 'кг': 1000, 'т': 1000000, 'унция': 28.349523125,
             'фунт': 453.59237},
    'volume': {'мл': 0.001, 'л': 1, 'м³': 1000, 'галлон (США)': 3.785411784,
               'кварта (США)': 0.946352946, 'пинта (США)': 0.473176473,
               'чашка (240 мл)': 0.24},
}

def unit_convert(kind, frm, to, value):
    u = UNITS.get(kind)
    if not u or frm not in u or to not in u:
        return None
    if not isinstance(value, (int, float)) or isinstance(value, bool):
        return None
    return value * u[frm] / u[to]

def temp_conv(c, f):
    if c is not None:
        try:
            c = float(c)
            return dict(c=c, f=c * 9 / 5 + 32)
        except Exception:
            return None
    if f is not None:
        try:
            f = float(f)
            return dict(c=(f - 32) * 5 / 9, f=f)
        except Exception:
            return None
    return None

def pace_calc(kmh, pmin, psec):
    use_tempo = pmin is not None and pmin >= 0
    if use_tempo:
        s = pmin + ((psec if psec and psec > 0 else 0) / 60)
    else:
        s = 60 / kmh if kmh and kmh > 0 else None
    if not (s and s > 0):
        return None
    pace_m = math.floor(s)
    pace_s = round((s - pace_m) * 60)
    if pace_s >= 60:
        pace_m += 1
        pace_s -= 60
    return dict(spd=60 / s, paceM=pace_m, paceS=pace_s, paceMin=s)

def hr_zones(age, max_hr):
    if max_hr and max_hr > 0:
        mx = max_hr
    elif age and age > 0:
        mx = 220 - age
    else:
        mx = 220
    defs = [(50, 60, 'Восстановление'), (60, 70, 'Базовая, жиросжигание'),
            (70, 80, 'Аэробная, темп'), (80, 90, 'Пороговая'), (90, 100, 'Максимальная')]
    zones = [dict(pct=str(a) + '–' + str(b) + ' %', name=name,
                  lo=round(mx * a / 100), hi=round(mx * b / 100))
             for a, b, name in defs]
    return dict(max=mx, zones=zones)
