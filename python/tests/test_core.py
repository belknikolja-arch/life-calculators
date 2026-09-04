# -*- coding: utf-8 -*-
"""Тесты Python-порта: ядро + каталог, сверено с оригиналом (JS) 1-в-1."""
import os
import sys
import unittest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from life_calculators import catalog as CAT
from life_calculators.core import (annuity, deposit_fv, vat_extract, vat_add, tax_calc,
                                   discount, room_geom, paint_needed, flooring_needed,
                                   appliance_cost, serving_nutrition, cups_convert,
                                   stats_of, date_diff, bmi_of, tdee, one_rm, tip_split,
                                   fuel_cost, met_calories, due_date, pregnancy_weeks,
                                   tv_diag, unit_price, watts_amps, savings_monthly,
                                   rule72_calc, wallpaper, dilute_calc, unit_convert,
                                   temp_conv, pace_calc, hr_zones, recipe_scale)
from life_calculators.formatting import fmt_num, fmt_r


def rows(id_, **kv):
    """Запуск карточки: defaults + переопределения."""
    c = next(x for x in CAT.CALCS if x['id'] == id_)
    vals = {}
    for inp in c['inputs']:
        d = inp['defv']
        if callable(d):
            vals[inp['k']] = d()
        elif inp.get('type') == 'date' and d is None:
            vals[inp['k']] = ''
        else:
            vals[inp['k']] = '' if d is None else str(d)
    vals.update(kv)
    return c['calc'](vals)


class CoreTest(unittest.TestCase):
    def approx(self, a, b, eps=1e-6):
        self.assertTrue(abs(a - b) <= eps, '{} != {}'.format(a, b))

    # ---- деньги ----
    def test_loan(self):
        r = rows('loan', p='2000000', r='21', m='48')
        self.assertEqual(r[0][1], fmt_r(annuity(2000000, 21, 48)['pmt']))
        self.assertEqual(len(r), 4)
        self.assertIsNone(rows('loan', p='0'))

    def test_annuity(self):
        r = annuity(1000000, 19, 60)
        self.assertIsNotNone(r)
        self.approx(r['total'] - r['pmt'] * 60, 0)

    def test_deposit_vs_js(self):
        # эталон из оригинала: P=1e6, 18%, 12 мес, взнос 0
        r = deposit_fv(1000000, 18, 12, 0)
        self.approx(r['end'], 1195618.1714, 0.01)
        self.approx(r['interest'], 195618.1714, 0.01)

    def test_vat(self):
        self.approx(vat_extract(120000, 20)['base'], 100000)
        self.approx(vat_add(100000, 20)['vat'], 20000)

    def test_tax(self):
        r = tax_calc(0, 100000, 13)
        self.approx(r['gross'], 114942.5287, 0.01)

    def test_percent_rows(self):
        r = rows('percent', x='1000', p='12.5', a='40', b='250')
        self.assertEqual(len(r), 4)
        self.assertEqual(r[0][1], '125')

    # ---- дом ----
    def test_room(self):
        r = rows('room', l='4', w='3', h='2.7', o='2')
        self.assertIn('м²', r[0][1])
        self.assertEqual(r[-1][0], 'Стены за вычетом проёмов')

    def test_paint(self):
        r = rows('paint', a='60', c='10', n='2', can='2.5')
        self.assertEqual(r[0][1], '12 л')
        self.assertEqual(r[1][1], 5)  # 12/2.5 = 4.8 -> 5 банок

    def test_floor(self):
        self.assertEqual(flooring_needed(20, 8, 2)['packs'], 11)

    def test_appliance(self):
        self.approx(appliance_cost(2000, 2, 30, 6)['cost'], 720)

    # ---- еда ----
    def test_serving(self):
        r = serving_nutrition(dict(kcal=200, p=4, f=3, c=10), 50)
        self.approx(r['kcal'], 100)

    def test_cups(self):
        r = cups_convert('flour', 1, None)
        self.assertEqual(r['dir'], 'cups->g')
        self.approx(r['g'], 130)

    def test_tip(self):
        r = tip_split(3500, 10, 2)
        self.approx(r['per'], 1925)

    # ---- поездки ----
    def test_fuel(self):
        r = fuel_cost(450, 8, 60)
        self.approx(r['liters'], 36)
        self.approx(r['cost'], 2160)

    # ---- математика ----
    def test_stats(self):
        s = stats_of([120, 95, 110, 130, 98])
        self.assertEqual(s['n'], 5)
        self.approx(s['mean'], 110.6)

    def test_dates(self):
        r = rows('dates', a='2026-01-05', b='2026-01-31')
        self.assertEqual(r[0][1], 26)
        # 05.01 пн .. 31.01 сб включительно: будние
        self.assertEqual(r[-1][1], 20)

    def test_datediff_core(self):
        d = date_diff('2026-01-05', '2026-01-31')
        self.assertEqual(d['days'], 26)
        self.assertEqual(d['workdays'], 20)

    # ---- здоровье ----
    def test_bmi(self):
        self.approx(bmi_of(176, 80)['v'], 25.83, 0.01)

    def test_tdee_rows(self):
        r = rows('tdee', age='30', h='176', w='80', act='1.375')
        self.assertEqual(len(r), 5)

    def test_one_rm(self):
        self.approx(one_rm(100, 5)['v'], 116.6667, 0.01)

    def test_met(self):
        self.approx(met_calories(75, 6, 1)['kcal'], 450)

    def test_due(self):
        r = due_date('2025-11-01', 30)
        self.assertEqual(r['due'], '2026-08-10')
        self.assertEqual(r['shift'], 2)

    # ---- v1.3 ----
    def test_v13(self):
        self.approx(savings_monthly(500000, 18, 24, 100000)['pmt'], 12469.64, 0.02)
        self.approx(rule72_calc(12, 10)['growth'], 3.1058, 0.001)
        self.assertEqual(wallpaper(14, 2.5, 0.53, 10, 0, 0)['rolls'], 7)
        self.approx(dilute_calc(70, 9, 100, None)['water'], 677.78, 0.01)
        self.approx(unit_convert('length', 'миля', 'км', 1), 1.609344)
        self.approx(temp_conv(180, None)['f'], 356)
        self.assertEqual(pace_calc(None, 6, 0)['paceM'], 6)
        self.assertEqual(hr_zones(30, None)['max'], 190)

    def test_formatting(self):
        self.assertEqual(fmt_num(1234567.891, 2), '1\u00a0234\u00a0567,89')
        self.assertEqual(fmt_num(1000, 0), '1\u00a0000')
        self.assertEqual(fmt_r(1234.5), '1\u00a0234,5 ₽')
        self.assertEqual(fmt_num(0.5, 0), '1')  # round half up как JS

    def test_watts_cable(self):
        # 3500 Вт / (220 В * 0.9) ≈ 17.68 А => ближайший стандартный автомат 25 А
        r = watts_amps(3500, 220, 0.9)
        self.assertEqual(r['breaker'], 25)
        self.assertEqual(r['cable'], '4 мм²')

    def test_recipe_scale(self):
        r = recipe_scale(4, 8, [dict(name='Мука', amount=250, unit='г')])
        self.assertEqual(r['k'], 2)
        self.approx(r['items'][0]['scaled'], 500)


if __name__ == '__main__':
    unittest.main()
