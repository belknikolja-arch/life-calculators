package main

import (
	"math"
	"testing"
)

func approx(a, b, eps float64) bool { return math.Abs(a-b) <= eps }

func TestAnnuity(t *testing.T) {
	r, ok := annuity(1000000, 19, 60)
	if !ok {
		t.Fatal("expected ok")
	}
	if !approx(r.total, r.pmt*60, 1e-6) {
		t.Errorf("total mismatch %v vs %v", r.total, r.pmt*60)
	}
}

func TestDeposit(t *testing.T) {
	r, ok := depositFV(1000000, 18, 12, 0)
	if !ok || !approx(r.end, 1195618.1714, 0.01) {
		t.Errorf("end=%v", r.end)
	}
}

func TestVatTax(t *testing.T) {
	if b, _ := vatExtract(120000, 20); !approx(b.base, 100000, 1e-6) {
		t.Error("vatExtract")
	}
	if b, _ := vatAdd(100000, 20); !approx(b.vat, 20000, 1e-6) {
		t.Error("vatAdd")
	}
	tr, ok := taxCalc(0, 100000, 13)
	if !ok || !approx(tr.gross, 114942.5287, 0.01) {
		t.Errorf("tax gross=%v", tr.gross)
	}
}

func TestWallpaperAndDilute(t *testing.T) {
	w, ok := wallpaper(14, 2.5, 0.53, 10, 0, 0)
	if !ok || w.rolls != 7 || w.stripsTotal != 27 || w.stripsPer != 4 {
		t.Errorf("wallpaper %+v", w)
	}
	d, ok := diluteCalc(70, 9, 100, 0, true, false)
	if !ok || !approx(d.water, 677.7778, 0.01) {
		t.Errorf("dilute %+v", d)
	}
}

func TestUnitsTempPace(t *testing.T) {
	if u, _ := unitConvert("length", "миля", "км", 1); !approx(u, 1.609344, 1e-9) {
		t.Error("mile")
	}
	if r, _ := tempConv(180, 0, true, false); !approx(r.f, 356, 1e-9) {
		t.Error("temp")
	}
	p, ok := paceCalc(0, 6, 0, false, true)
	if !ok || p.paceM != 6 || p.paceS != 0 || !approx(p.spd, 10, 0.01) {
		t.Errorf("pace %+v", p)
	}
}

func TestStats(t *testing.T) {
	s, ok := statsOf([]float64{120, 95, 110, 130, 98})
	if !ok || s.n != 5 || !approx(s.mean, 110.6, 1e-9) {
		t.Errorf("stats %+v", s)
	}
}

func TestDateDiff(t *testing.T) {
	d, ok := dateDiff("2026-01-05", "2026-01-31")
	if !ok || d.days != 26 || d.workdays != 20 {
		t.Errorf("dateDiff %+v", d)
	}
}

func TestZonesAndV13(t *testing.T) {
	z := hrZones(30, 0, false)
	if z.max != 190 || z.zones[0].lo != 95 {
		t.Errorf("hrzones %+v", z)
	}
	s, ok := savingsMonthly(500000, 18, 24, 100000)
	if !ok || !approx(s.pmt, 12469.64, 0.02) {
		t.Errorf("savings %+v", s)
	}
	_, _, ex, ok := rule72Calc(12, 10)
	if !ok || !approx(ex, 6.1163, 0.01) {
		t.Errorf("rule72 %v", ex)
	}
}

func TestFormatting(t *testing.T) {
	if fmtNum(1234567.891, 2) != "1\u00a0234\u00a0567,89" {
		t.Errorf("fmtNum got %q", fmtNum(1234567.891, 2))
	}
	if fmtR(1234.5) != "1\u00a0234,5 ₽" {
		t.Errorf("fmtR got %q", fmtR(1234.5))
	}
	if fmtNum(0.5, 0) != "1" {
		t.Errorf("fmtNum round got %q", fmtNum(0.5, 0))
	}
}

func TestCatalogRows(t *testing.T) {
	cases := []struct {
		id   string
		kv   map[string]string
		want string
	}{
		{"loan", map[string]string{"p": "1000000", "r": "19", "m": "60"}, "25\u00a0940,55 ₽"},
		{"wallpaper", map[string]string{"per": "14", "h": "2.5", "rw": "0.53", "rl": "10"}, "7 шт"},
		{"bmi", map[string]string{"h": "176", "w": "80"}, "25,8"},
		{"dates", map[string]string{"a": "2026-01-05", "b": "2026-01-31"}, "26"},
	}
	for _, tc := range cases {
		c := findCalc(tc.id)
		if c == nil {
			t.Fatalf("calc %s not found", tc.id)
		}
		v := resolveMap(c, tc.kv)
		rows := c.Run(v)
		if rows == nil || len(rows) == 0 {
			t.Errorf("%s: no rows", tc.id)
			continue
		}
		if rows[0].Value != tc.want {
			t.Errorf("%s: first row value %q, want %q", tc.id, rows[0].Value, tc.want)
		}
	}
}

func TestAllCalcsRunWithDefaults(t *testing.T) {
	for i := range CALCS {
		c := &CALCS[i]
		if c.Run == nil { // recipe / units — спец-режимы
			continue
		}
		v := resolveMap(c, nil)
		rows := c.Run(v)
		if c.ID == "tax" || c.ID == "dilute" || c.ID == "cups" || c.ID == "vts" || c.ID == "pace" {
			continue // требуют ввода одного из полей — не обязаны давать результат по умолчанию
		}
		if rows == nil {
			t.Errorf("calc %s вернул nil с дефолтами", c.ID)
		}
	}
}
