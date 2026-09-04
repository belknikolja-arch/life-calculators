# -*- coding: utf-8 -*-
"""Форматирование чисел — 1-в-1 как в JS: fmtNum/fmtR/fmtDur.
NBSP = U+00A0 как разделитель тысяч; запятая как десятичный разделитель.
"""
from __future__ import annotations
import math

NBSP = '\u00A0'
MINUS = '\u2212'  # '−'

def _round_js(x):
    """Math.round из JS: 0.5 округляется вверх (в сторону +∞)."""
    return math.floor(x + 0.5)

def fmt_num(n, max_frac=2):
    """JS fmtNum: округление до max_frac знаков, тысячи NBSP, запятая-разделитель."""
    if n is None:
        return ''
    neg = n < 0
    n = abs(_round_js(n * (10 ** max_frac)) / (10 ** max_frac))
    # toFixed(max_frac) с последующим срезанием хвостовых нулей у дроби
    scaled = _round_js(n * (10 ** max_frac))
    int_part = scaled // (10 ** max_frac)
    frac = scaled % (10 ** max_frac)
    int_str = str(int_part)
    # вставить NBSP каждые 3 разряда (JS /\B(?=(\d{3})+(?!\d))/g)
    grouped = ''
    for i, ch in enumerate(int_str):
        if i and (len(int_str) - i) % 3 == 0:
            grouped += NBSP
        grouped += ch
    if max_frac and frac:
        # обрезать хвостовые нули у дробной части
        frac_str = str(frac).rjust(max_frac, '0').rstrip('0')
        out = grouped + (',' + frac_str if frac_str else '')
    else:
        out = grouped
    return (MINUS if neg else '') + out

def fmt_units(n, unit, max_frac=2, space=True):
    sep = ' ' if space else ''
    return fmt_num(n, max_frac) + sep + unit

def fmt_r(n):
    """JS fmtR: рубли."""
    return fmt_num(n, 2) + ' ₽'

def fmt_dur(minutes):
    """JS fmtDur: '30 мин', '2 ч', '1 ч 5 мин'."""
    if minutes is None:
        return '—'
    m = round(minutes)
    if m < 60:
        return str(m) + ' мин'
    h = m // 60
    mm = m % 60
    if mm:
        return str(h) + ' ч ' + str(mm) + ' мин'
    return str(h) + ' ч'
