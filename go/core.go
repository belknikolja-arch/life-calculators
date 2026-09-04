package main

// Чистое ядро — Go-порт JS-ядра (index.html). Все функции возвращают (данные, ok).

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ---------------- форматирование (fmtNum/fmtR/fmtDur из JS) ---------------
const nb = "\u00a0" // NBSP
const minus = "\u2212"

func jsRound(x float64) float64 { // Math.round: 0.5 -> вверх
	return math.Floor(x + 0.5)
}

func fmtNum(n float64, maxFrac int) string {
	neg := n < 0
	a := math.Abs(jsRound(n*math.Pow10(maxFrac)) / math.Pow10(maxFrac))
	scaled := int64(jsRound(a * math.Pow10(maxFrac)))
	den := int64(1)
	for i := 0; i < maxFrac; i++ {
		den *= 10
	}
	intPart := scaled / den
	frac := scaled % den
	intStr := strconv.FormatInt(intPart, 10)
	// тысячи NBSP
	var b strings.Builder
	for i, ch := range intStr {
		if i > 0 && (len(intStr)-i)%3 == 0 {
			b.WriteString(nb)
		}
		b.WriteRune(ch)
	}
	out := b.String()
	if maxFrac > 0 && frac > 0 {
		fr := strings.TrimRight(fmtIntPad(frac, maxFrac), "0")
		if fr != "" {
			out += "," + fr
		}
	}
	if neg {
		return minus + out
	}
	return out
}

func fmtIntPad(v int64, width int) string {
	s := strconv.FormatInt(v, 10)
	for len(s) < width {
		s = "0" + s
	}
	return s
}

func fmtR(n float64) string { return fmtNum(n, 2) + " ₽" }

func fmtDur(min float64) string {
	m := int(jsRound(min))
	if m < 60 {
		return strconv.Itoa(m) + " мин"
	}
	h, mm := m/60, m%60
	if mm != 0 {
		return strconv.Itoa(h) + " ч " + strconv.Itoa(mm) + " мин"
	}
	return strconv.Itoa(h) + " ч"
}

// ---------------- вспомогательные ----------------
func numStr(v string) (float64, bool) { // аналог num() из JS
	v = strings.ReplaceAll(v, nb, "")
	sp := regexp.MustCompile(`\s`)
	v = sp.ReplaceAllString(v, "")
	v = strings.ReplaceAll(v, ",", ".")
	n, err := strconv.ParseFloat(v, 64)
	if err != nil || math.IsInf(n, 0) || math.IsNaN(n) {
		return 0, false
	}
	return n, true
}

func numOr(v string, def float64) float64 {
	if n, ok := numStr(v); ok {
		return n
	}
	return def
}

func parseIntRounded(v string) (int, bool) {
	n, ok := numStr(v)
	if !ok {
		return 0, false
	}
	return int(jsRound(n)), true
}

// ---------------- деньги ----------------
type loanR struct{ pmt, total, interest float64 }

func annuity(P, ratePct float64, months int) (loanR, bool) {
	if !(P > 0) || !(months >= 1) || !(ratePct >= 0) {
		return loanR{}, false
	}
	r := ratePct / 100 / 12
	var pmt float64
	if r == 0 {
		pmt = P / float64(months)
	} else {
		pmt = P * r / (1 - math.Pow(1+r, -float64(months)))
	}
	return loanR{pmt, pmt * float64(months), pmt*float64(months) - P}, true
}

type depR struct{ end, contrib, interest float64 }

func depositFV(P, ratePct float64, months int, monthly float64) (depR, bool) {
	if !(P >= 0) || !(months >= 1) || !(ratePct >= 0) {
		return depR{}, false
	}
	r := ratePct / 100 / 12
	bal := P
	for i := 0; i < months; i++ {
		bal = bal*(1+r) + monthly
	}
	contrib := P + monthly*float64(months)
	return depR{bal, contrib, bal - contrib}, true
}

type vatR struct{ base, vat, total float64 }

func vatExtract(amount, rate float64) (vatR, bool) {
	if !(amount > 0) {
		return vatR{}, false
	}
	base := amount / (1 + rate/100)
	return vatR{base, amount - base, amount}, true
}
func vatAdd(base, rate float64) (vatR, bool) {
	if !(base >= 0) {
		return vatR{}, false
	}
	vat := base * rate / 100
	return vatR{base, vat, base + vat}, true
}

type taxR struct{ gross, tax, net float64 }

func taxCalc(gross, net, rate float64) (taxR, bool) {
	if !(rate > 0 && rate < 100) {
		return taxR{}, false
	}
	if gross > 0 {
		tax := gross * rate / 100
		return taxR{gross, tax, gross - tax}, true
	}
	if net > 0 {
		g := net / (1 - rate/100)
		return taxR{g, g - net, net}, true
	}
	return taxR{}, false
}

func pctOf(x, p float64) float64 { return x * p / 100 }
func pctOfBase(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b * 100
}

type discR struct{ final, saved, price float64 }

func discount(price, p float64) (discR, bool) {
	if !(price > 0) || p < 0 || p > 100 {
		return discR{}, false
	}
	final := price * (1 - p/100)
	return discR{final, price - final, price}, true
}

type salR struct{ hourly, day8, week, year float64 }

func salaryBreakdown(monthly, hw float64) (salR, bool) {
	if !(monthly > 0) {
		return salR{}, false
	}
	w := hw
	if !(w > 0) {
		w = 40
	}
	ph := w * 4.33
	hourly := monthly / ph
	return salR{hourly, hourly * 8, monthly * w / 40, monthly * 12}, true
}

type infR struct{ value, lost float64 }

func inflationPower(amount, ratePct, years float64) (infR, bool) {
	if !(amount >= 0) || ratePct < 0 || !(years >= 0) {
		return infR{}, false
	}
	k := math.Pow(1+ratePct/100, years)
	v := amount / k
	return infR{v, amount - v}, true
}

// ---------------- дом ----------------
type roomR struct{ area, per, walls, ceil float64 }

func round6(x float64) float64 { return jsRound(x*1e6) / 1e6 }

func roomGeom(l, w, h float64) (roomR, bool) {
	if !(l > 0 && w > 0) {
		return roomR{}, false
	}
	hh := h
	if !(hh > 0) {
		hh = 0
	}
	return roomR{round6(l * w), round6(2 * (l + w)), round6(2 * (l + w) * hh), round6(l * w)}, true
}

type paintR struct {
	liters  float64
	cans    int
	hasCans bool
}

func paintNeeded(area, coverage, coats, canL float64) (paintR, bool) {
	if !(area > 0 && coverage > 0) {
		return paintR{}, false
	}
	c := coats
	if !(c > 0) {
		c = 1
	}
	liters := area * c / coverage
	if canL > 0 {
		return paintR{liters, int(math.Ceil(liters/canL - 1e-9)), true}, true
	}
	return paintR{liters: liters}, true
}

type floorR struct {
	need     float64
	packs    int
	hasPacks bool
}

func flooringNeeded(area, lossPct, packM float64) (floorR, bool) {
	if !(area > 0) || lossPct < 0 {
		return floorR{}, false
	}
	need := area * (1 + lossPct/100)
	if packM > 0 {
		return floorR{need, int(math.Ceil(need/packM - 1e-9)), true}, true
	}
	return floorR{need: need}, true
}

type appR struct{ kwh, cost float64 }

func applianceCost(watts, hoursDay, days, rate float64) (appR, bool) {
	if !(watts > 0 && hoursDay >= 0 && days >= 0) {
		return appR{}, false
	}
	kwh := watts * hoursDay * days / 1000
	return appR{kwh, kwh * rate}, true
}

// ---------------- еда ----------------
type ing struct {
	name   string
	amount float64
	unit   string
}

func recipeScale(fromP, toP float64, items []ing) (k float64, out []ing, ok bool) {
	if !(fromP > 0 && toP > 0) {
		return 0, nil, false
	}
	k = toP / fromP
	for _, it := range items {
		if it.amount > 0 {
			out = append(out, ing{it.name, it.amount * k, it.unit})
		}
	}
	return k, out, true
}

type nutR struct{ kcal, p, f, c float64 }

func servingNutrition(kcal, p, f, c, grams float64) (nutR, bool) {
	if !(grams > 0) {
		return nutR{}, false
	}
	kk := grams / 100
	return nutR{kcal * kk, p * kk, f * kk, c * kk}, true
}

var cupsG = map[string]float64{ // 1 стакан = N граммов/мл
	"flour": 130, "sugar": 200, "butter": 227, "milk": 240,
	"cocoa": 100, "choc": 170, "rice": 180, "oil": 240,
}
var cupsName = map[string]string{
	"flour": "Мука", "sugar": "Сахар", "butter": "Масло слив.", "milk": "Молоко (мл)",
	"cocoa": "Какао-порошок", "choc": "Шоколад", "rice": "Рис (сухой)", "oil": "Масло раст. (мл)",
}

type cupR struct {
	g, cups float64
	toGrams bool
} // toGrams=true: cups->g

func cupsConvert(prod string, cups, grams float64) (cupR, bool) {
	g, ok := cupsG[prod]
	if !ok {
		return cupR{}, false
	}
	if cups > 0 {
		return cupR{cups * g, cups, true}, true
	}
	if grams > 0 {
		return cupR{grams, grams / g, false}, true
	}
	return cupR{}, false
}

type tipR struct{ tip, total, per float64 }

func tipSplit(bill, pct float64, people int) (tipR, bool) {
	if !(bill > 0) || pct < 0 {
		return tipR{}, false
	}
	tip := bill * pct / 100
	total := bill + tip
	n := people
	if n <= 1 {
		n = 1
	}
	return tipR{tip, total, total / float64(n)}, true
}

type fuelR struct {
	liters, cost float64
	hasCost      bool
}

func fuelCost(km, l100 float64, price float64, hasPrice bool) (fuelR, bool) {
	if !(km > 0 && l100 > 0) {
		return fuelR{}, false
	}
	liters := km * l100 / 100
	if hasPrice {
		return fuelR{liters, liters * price, true}, true
	}
	return fuelR{liters: liters}, true
}

func metCalories(weightKg, met, hours float64) (float64, bool) {
	if !(weightKg > 0 && met > 0 && hours > 0) {
		return 0, false
	}
	return met * weightKg * hours, true
}

// ---------------- математика ----------------
type shapeR struct {
	name, unit string
	value      float64
}

func shapeCalc(kind string, a, b float64) (shapeR, bool) {
	if !(a > 0) {
		return shapeR{}, false
	}
	f := func(x float64) string {
		if x == math.Floor(x) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	}
	switch kind {
	case "circleR":
		return shapeR{"Площадь круга (r=" + f(a) + ")", "кв. ед.", math.Pi * a * a}, true
	case "circleD":
		return shapeR{"Площадь круга (d=" + f(a) + ")", "кв. ед.", math.Pi * (a / 2) * (a / 2)}, true
	case "rect":
		if !(b > 0) {
			return shapeR{}, false
		}
		return shapeR{"Площадь прямоугольника", "кв. ед.", a * b}, true
	case "tri":
		if !(b > 0) {
			return shapeR{}, false
		}
		return shapeR{"Площадь прямоугольного треугольника", "кв. ед.", a * b / 2}, true
	case "cyl":
		if !(b > 0) {
			return shapeR{}, false
		}
		return shapeR{"Объём цилиндра (r=" + f(a) + ", h=" + f(b) + ")", "куб. ед.", math.Pi * a * a * b}, true
	case "sphere":
		return shapeR{"Объём сферы (r=" + f(a) + ")", "куб. ед.", 4.0 / 3.0 * math.Pi * a * a * a}, true
	}
	return shapeR{}, false
}

var sepRe = regexp.MustCompile(`[\s,;]+`)

func parseNumbers(s string) []float64 {
	parts := sepRe.Split(s, -1)
	out := []float64{}
	for _, p := range parts {
		if p == "" {
			continue
		}
		p = strings.ReplaceAll(p, ",", ".")
		if n, err := strconv.ParseFloat(p, 64); err == nil {
			out = append(out, n)
		}
	}
	return out
}

type statsR struct {
	n                           int
	sum, mean, median, min, max float64
}

func statsOf(nums []float64) (statsR, bool) {
	if len(nums) == 0 {
		return statsR{}, false
	}
	// копия + сортировка
	s := append([]float64{}, nums...)
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
	n := len(s)
	var sum float64
	for _, x := range s {
		sum += x
	}
	var median float64
	if n%2 == 1 {
		median = s[(n-1)/2]
	} else {
		median = (s[n/2-1] + s[n/2]) / 2
	}
	return statsR{n, sum, sum / float64(n), median, s[0], s[n-1]}, true
}

type vtsR struct {
	dist, time, speed float64
	mode              int
} // mode: 0 dist,1 time,2 speed

func vts(speed, time, dist float64, hasS, hasT, hasD bool) (vtsR, bool) {
	if hasS && speed > 0 && hasT && time > 0 {
		return vtsR{dist: speed * time, mode: 0}, true
	}
	if hasS && speed > 0 && hasD && dist > 0 {
		return vtsR{time: dist / speed, mode: 1}, true
	}
	if hasT && time > 0 && hasD && dist > 0 {
		return vtsR{speed: dist / time, mode: 2}, true
	}
	return vtsR{}, false
}

func parseYMD(s string) (time.Time, bool) {
	t, err := time.Parse("2006-01-02", s)
	return t, err == nil
}

type dateR struct {
	days     int
	workdays int
}

func dateDiff(a, b string) (dateR, bool) {
	d1, ok1 := parseYMD(a)
	d2, ok2 := parseYMD(b)
	if !ok1 || !ok2 {
		return dateR{}, false
	}
	days := int(math.Abs(d2.Sub(d1).Hours()) / 24)
	lo, hi := d1, d2
	if lo.After(hi) {
		lo, hi = hi, lo
	}
	wd := 0
	for t := lo; !t.After(hi); t = t.AddDate(0, 0, 1) {
		if w := int(t.Weekday()); w >= 1 && w <= 5 { // пн..пт
			wd++
		}
	}
	return dateR{days, wd}, true
}

// ---------------- здоровье ----------------
type bmiR struct {
	v   float64
	cat string
}

func bmiOf(hCm, wKg float64) (bmiR, bool) {
	if !(hCm > 0 && wKg > 0) {
		return bmiR{}, false
	}
	h := hCm / 100
	v := wKg / (h * h)
	cat := "ожирение"
	switch {
	case v < 16:
		cat = "выраженный дефицит"
	case v < 18.5:
		cat = "дефицит веса"
	case v < 25:
		cat = "норма"
	case v < 30:
		cat = "избыточный вес"
	}
	return bmiR{v, cat}, true
}

type tdeeR struct{ bmr, tdee, cut, gain float64 }

func tdee(sex string, age, hCm, wKg, activity float64) (tdeeR, bool) {
	if !(age > 0 && hCm > 0 && wKg > 0) {
		return tdeeR{}, false
	}
	base := 10*wKg + 6.25*hCm - 5*age
	if sex == "m" {
		base += 5
	} else {
		base -= 161
	}
	act := activity
	if act == 0 {
		act = 1.2
	}
	return tdeeR{base, base * act, base * act * 0.8, base * act * 1.1}, true
}

type macroR struct{ p, f, c float64 }

func macros(kcal, pctP, pctF, pctC float64) (macroR, bool) {
	if !(kcal > 0) {
		return macroR{}, false
	}
	return macroR{kcal * pctP / 100 / 4, kcal * pctF / 100 / 9, kcal * pctC / 100 / 4}, true
}

func waterLiters(wKg float64, active bool) (float64, bool) {
	if !(wKg > 0) {
		return 0, false
	}
	if active {
		return wKg * 0.035, true
	}
	return wKg * 0.03, true
}

func oneRm(weight, reps float64) (float64, bool) {
	if !(weight > 0 && reps > 0) {
		return 0, false
	}
	return weight * (1 + reps/30), true
}

// ---------------- v1.1 ----------------
func tvDiag(inch float64, cm float64, hasIn, hasCm bool) (cmv float64, inchv float64, ok bool) {
	if hasIn && inch > 0 {
		return inch * 2.54, inch, true
	}
	if hasCm && cm > 0 {
		return cm, cm / 2.54, true
	}
	return 0, 0, false
}
func tvViewDist(diag float64) (minM, maxM float64) {
	return diag * 1.5 * 2.54 / 100, diag * 2.5 * 2.54 / 100
}

type priceR struct{ per100, per1000 float64 }

func unitPrice(price, amount float64) (priceR, bool) {
	if !(price > 0 && amount > 0) {
		return priceR{}, false
	}
	per100 := price / amount * 100
	return priceR{per100, per100 * 10}, true
}

var breakerList = []float64{6, 10, 16, 25, 32, 40, 50, 63, 80, 100}
var cableMap = map[float64]string{
	6: "1 мм²", 10: "1,5 мм²", 16: "2,5 мм²", 25: "4 мм²", 32: "6 мм²",
	40: "10 мм²", 50: "10 мм²", 63: "16 мм²", 80: "16 мм²", 100: "25 мм²",
}

type wattsR struct {
	amps, v float64
	breaker float64
	cable   string
}

func wattsAmps(watts, volts, cosPhi float64, hasCp bool) (wattsR, bool) {
	v := volts
	if !(v > 0) {
		v = 220
	}
	cp := 1.0
	if hasCp && cosPhi > 0 {
		cp = math.Min(cosPhi, 1)
	}
	if !(watts > 0) {
		return wattsR{}, false
	}
	a := watts / (v * cp)
	var br float64
	for _, b := range breakerList {
		if a <= b {
			br = b
			break
		}
	}
	return wattsR{a, v, br, cableMap[br]}, true
}

// ---------------- v1.3 ----------------
type saveR struct{ pmt, paid, interest float64 }

func savingsMonthly(goal, ratePct float64, months int, start float64) (saveR, bool) {
	g, s := goal, start
	if !(g > 0) {
		g = 0
	}
	if !(s > 0) {
		s = 0
	}
	if months <= 0 || g <= s {
		return saveR{}, false
	}
	i := ratePct / 100 / 12
	if i == 0 {
		pmt := (g - s) / float64(months)
		return saveR{pmt, pmt * float64(months), 0}, true
	}
	fvS := s * math.Pow(1+i, float64(months))
	pmt := (g - fvS) * i / (math.Pow(1+i, float64(months)) - 1)
	paid := pmt * float64(months)
	return saveR{pmt, paid, g - s - paid}, true
}

func rule72Calc(ratePct, years float64) (growth, d72, dexact float64, ok bool) {
	r, y := ratePct, years
	if !(r > 0) {
		r = 0
	}
	if !(y > 0) {
		y = 0
	}
	if r == 0 {
		return 0, 0, 0, false
	}
	return math.Pow(1+r/100, y), 72 / r, math.Log(2) / math.Log(1+r/100), true
}

type wallR struct {
	stripsTotal, stripsPer, rolls int
	area                          float64
}

func wallpaper(perimeter, wallH, rollW, rollL, patternCm float64, skipStrips int) (wallR, bool) {
	P, h := perimeter, wallH
	rw, rl := rollW, rollL
	if !(P > 0) {
		P = 0
	}
	if !(h > 0) {
		h = 0
	}
	if !(rw > 0) {
		rw = 0.53
	}
	if !(rl > 0) {
		rl = 10
	}
	if P == 0 || h == 0 {
		return wallR{}, false
	}
	step := h
	if patternCm > 0 {
		step += patternCm / 100
	}
	stripsTotal := int(math.Ceil(P/rw)) - skipStrips
	if stripsTotal < 0 {
		stripsTotal = 0
	}
	stripsPer := 1
	if rl >= step {
		stripsPer = int(math.Floor(rl / step))
		if stripsPer < 1 {
			stripsPer = 1
		}
	}
	rolls := int(math.Ceil(float64(stripsTotal) / float64(stripsPer)))
	if rolls < 0 {
		rolls = 0
	}
	return wallR{stripsTotal, stripsPer, rolls, float64(rolls) * rl * rw}, true
}

type dilR struct{ src, water, total float64 }

func diluteCalc(c1, c2 float64, v1, vt float64, hasV1, hasVt bool) (dilR, bool) {
	a, b := c1, c2
	if !(a > 0 && a <= 100) {
		a = 0
	}
	if !(b > 0 && b <= 100) {
		b = 0
	}
	if a == 0 || b == 0 || b > a {
		return dilR{}, false
	}
	if !hasV1 && !hasVt {
		return dilR{}, false
	}
	if hasV1 {
		total := v1 * a / b
		return dilR{v1, total - v1, total}, true
	}
	src := vt * b / a
	return dilR{src, vt - src, vt}, true
}

type unitsTable map[string]float64

var unitKinds = map[string]unitsTable{
	"length": {"мм": 0.001, "см": 0.01, "м": 1, "км": 1000, "дюйм": 0.0254, "фут": 0.3048, "ярд": 0.9144, "миля": 1609.344},
	"mass":   {"мг": 0.001, "г": 1, "кг": 1000, "т": 1000000, "унция": 28.349523125, "фунт": 453.59237},
	"volume": {"мл": 0.001, "л": 1, "м³": 1000, "галлон (США)": 3.785411784, "кварта (США)": 0.946352946, "пинта (США)": 0.473176473, "чашка (240 мл)": 0.24},
}

func unitConvert(kind, from, to string, value float64) (float64, bool) {
	u, ok := unitKinds[kind]
	if !ok {
		return 0, false
	}
	f, ok1 := u[from]
	t, ok2 := u[to]
	if !ok1 || !ok2 {
		return 0, false
	}
	return value * f / t, true
}

type tempR struct{ c, f float64 }

func tempConv(c, f float64, hasC, hasF bool) (tempR, bool) {
	if hasC {
		return tempR{c, c*9/5 + 32}, true
	}
	if hasF {
		return tempR{(f - 32) * 5 / 9, f}, true
	}
	return tempR{}, false
}

type paceR struct {
	spd          float64
	paceM, paceS int
	paceMin      float64
}

func paceCalc(kmh float64, pmin float64, psec float64, hasKmh, hasPmin bool) (paceR, bool) {
	var s float64
	useTempo := hasPmin && pmin >= 0
	if useTempo {
		s = pmin
		if psec > 0 {
			s += psec / 60
		}
	} else if hasKmh && kmh > 0 {
		s = 60 / kmh
	}
	if !(s > 0) {
		return paceR{}, false
	}
	paceM := int(math.Floor(s))
	paceS := int(jsRound((s - float64(paceM)) * 60))
	if paceS >= 60 {
		paceM++
		paceS -= 60
	}
	return paceR{60 / s, paceM, paceS, s}, true
}

type zone struct {
	pct, name string
	lo, hi    int
}
type hrR struct {
	max   int
	zones []zone
}

func hrZones(age float64, maxHR float64, hasMax bool) hrR {
	mx := maxHR
	if !(hasMax && mx > 0) {
		if age > 0 {
			mx = 220 - age
		} else {
			mx = 220
		}
	}
	defs := [][3]interface{}{
		{50, 60, "Восстановление"},
		{60, 70, "Базовая, жиросжигание"},
		{70, 80, "Аэробная, темп"},
		{80, 90, "Пороговая"},
		{90, 100, "Максимальная"},
	}
	zs := []zone{}
	for _, d := range defs {
		a := d[0].(int)
		b := d[1].(int)
		zs = append(zs, zone{
			pct:  strconv.Itoa(a) + "–" + strconv.Itoa(b) + " %",
			name: d[2].(string),
			lo:   int(jsRound(float64(mx) * float64(a) / 100)),
			hi:   int(jsRound(float64(mx) * float64(b) / 100)),
		})
	}
	return hrR{int(mx), zs}
}

// ---------------- даты (v1.1) ----------------
func fmtISO(t time.Time) string { return t.Format("2006-01-02") }

func dueDate(lmp string, cycleLen float64, hasCycle bool) (due string, base string, shift int, ok bool) {
	d, ok1 := parseYMD(lmp)
	if !ok1 {
		return "", "", 0, false
	}
	baseD := d.AddDate(0, 0, 280)
	shift = 0
	if hasCycle && cycleLen > 21 && cycleLen < 40 {
		shift = int(jsRound(cycleLen - 28))
	}
	dueD := baseD.AddDate(0, 0, shift)
	return fmtISO(dueD), fmtISO(baseD), shift, true
}

func pregnancyWeeks(lmp string, now string) (weeks int, rem int, label string, ok bool) {
	l, ok1 := parseYMD(lmp)
	if !ok1 {
		return 0, 0, "", false
	}
	var n time.Time
	if now != "" {
		n, ok = parseYMD(now)
		if !ok {
			return 0, 0, "", false
		}
	} else {
		nowT := time.Now()
		n = time.Date(nowT.Year(), nowT.Month(), nowT.Day(), 0, 0, 0, 0, time.UTC)
	}
	days := int(n.Sub(l).Hours() / 24)
	if days < 0 || days > 420 {
		return 0, 0, "", false
	}
	w := days / 7
	r := days % 7
	label = strconv.Itoa(w) + " нед."
	if r != 0 {
		label += " " + strconv.Itoa(r) + " дн."
	}
	return w, r, label, true
}
