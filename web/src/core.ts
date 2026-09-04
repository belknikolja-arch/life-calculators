/* ============================================================
   ЧИСТОЕ ЯДРО — TypeScript-порт JS-ядра из index.html (v1.3)
   Все функции чисто математические. Ниже 1-в-1 повторяют оригинал.
   ============================================================ */

// ---------------- числа / форматирование (fmtNum, fmtR, fmtDur) ----------------
export const NBSP = '\u00a0'
export const MINUS = '\u2212'

/** Math.round из JS: 0.5 — вверх (отрицательные к +∞). */
export function jsRound(x: number): number {
  return Math.floor(x + 0.5)
}

export function fmtNum(n: number, maxFrac = 2): string {
  const neg = n < 0
  n = Math.abs(jsRound(n * Math.pow(10, maxFrac)) / Math.pow(10, maxFrac))
  const scaled = jsRound(n * Math.pow(10, maxFrac))
  const den = Math.pow(10, maxFrac)
  const intPart = Math.trunc(scaled / den)
  const frac = scaled % den
  const intStr = String(intPart)
  const intFormatted = intStr.replace(/\B(?=(\d{3})+(?!\d))/g, NBSP)
  let fracStr = frac > 0 ? String(frac).padStart(maxFrac, '0').replace(/0+$/, '') : ''
  if (maxFrac <= 0 || frac === 0) fracStr = ''
  return (neg ? MINUS : '') + intFormatted + (fracStr ? ',' + fracStr : '')
}

export function fmtR(n: number): string {
  return fmtNum(n, 2) + ' ₽'
}

export function fmtDur(min: number | null): string {
  if (min == null || !isFinite(min)) return '—'
  const m = jsRound(min)
  if (m < 60) return m + ' мин'
  const h = Math.floor(m / 60)
  const mm = m % 60
  return h + ' ч' + (mm ? ' ' + mm + ' мин' : '')
}

/** JS num(): строка -> число|null (учитывает NBSP, пробелы, запятую). */
export function parseNumRaw(s: string | undefined | null): number | null {
  if (s == null || s === '') return null
  const n = Number(String(s).replace(/\s/g, '').replace(',', '.'))
  return isFinite(n) ? n : null
}

export function todayStr(): string {
  const d = new Date()
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0')
}

export function fmtISO(d: Date): string {
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0')
}

// ---------------- деньги ----------------
export function annuity(P: number, ratePct: number, months: number) {
  if (!(P > 0) || !(months >= 1) || !(ratePct >= 0)) return null
  const r = ratePct / 100 / 12
  const pmt = r === 0 ? P / months : (P * r) / (1 - Math.pow(1 + r, -months))
  return { pmt, total: pmt * months, interest: pmt * months - P }
}

export function depositFV(P: number, ratePct: number, months: number, monthly: number) {
  if (!(P >= 0) || !(months >= 1) || !(ratePct >= 0)) return null
  const r = ratePct / 100 / 12
  const add = monthly || 0
  let bal = P
  for (let i = 0; i < months; i++) bal = bal * (1 + r) + add
  const contrib = P + add * months
  return { end: bal, contrib, interest: bal - contrib }
}

export function vatExtract(amount: number, rate: number | null) {
  if (!(amount > 0)) return null
  const r = rate == null ? 20 : rate
  const base = amount / (1 + r / 100)
  return { base, vat: amount - base }
}

export function vatAdd(base: number, rate: number | null) {
  if (!(base >= 0)) return null
  const r = rate == null ? 20 : rate
  const vat = (base * r) / 100
  return { total: base + vat, vat }
}

export function taxCalc(gross: number, net: number, rate: number | null) {
  const r = rate == null ? 13 : rate
  if (!(r > 0 && r < 100)) return null
  if (gross > 0) return { gross, tax: (gross * r) / 100, net: gross - (gross * r) / 100 }
  if (net > 0) return { gross: net / (1 - r / 100), tax: net / (1 - r / 100) - net, net }
  return null
}

export function pctOf(x: number, p: number): number {
  return (x * p) / 100
}
export function pctOfBase(a: number, b: number): number | null {
  return b ? (a / b) * 100 : null
}

export function discount(price: number, p: number) {
  if (!(price > 0) || p < 0 || p > 100) return null
  const final = price * (1 - p / 100)
  return { final, saved: price - final, price }
}

export function salaryBreakdown(monthly: number, hw: number | null) {
  if (!(monthly > 0)) return null
  const w = hw && hw > 0 ? hw : 40
  const perMonthHours = w * 4.33
  const hourly = monthly / perMonthHours
  return { hourly, day8: hourly * 8, week: (monthly * w) / 40, year: monthly * 12 }
}

export function inflationPower(amount: number, ratePct: number, years: number) {
  if (!(amount >= 0) || ratePct < 0 || !(years >= 0)) return null
  const k = Math.pow(1 + ratePct / 100, years)
  const value = amount / k
  return { value, lost: amount - value }
}

// ---------------- дом ----------------
export function roomGeom(l: number, w: number, h: number) {
  if (!(l > 0 && w > 0)) return null
  const hh = h > 0 ? h : 0
  const rnd = (x: number) => Math.round(x * 1e6) / 1e6
  return { area: rnd(l * w), per: rnd(2 * (l + w)), walls: rnd(2 * (l + w) * hh), ceil: rnd(l * w) }
}

export function paintNeeded(area: number, coverage: number, coats: number | null, canL: number | null) {
  if (!(area > 0 && coverage > 0)) return null
  const c = coats && coats > 0 ? coats : 1
  const liters = (area * c) / coverage
  const cans = canL && canL > 0 ? Math.ceil(liters / canL - 1e-9) : null
  return { liters, cans }
}

export function flooringNeeded(area: number, lossPct: number, packM: number | null) {
  if (!(area > 0) || lossPct < 0) return null
  const need = area * (1 + lossPct / 100)
  const packs = packM && packM > 0 ? Math.ceil(need / packM - 1e-9) : null
  return { need, packs }
}

export function applianceCost(watts: number, hoursDay: number, days: number, rate: number) {
  if (!(watts > 0 && hoursDay >= 0 && days >= 0)) return null
  const kwh = (watts * hoursDay * days) / 1000
  return { kwh, cost: kwh * (rate || 0) }
}

// ---------------- еда ----------------
export interface RecipeItem {
  name: string
  amount: number
  unit: string
}
export function recipeScale(fromP: number, toP: number, items: RecipeItem[]) {
  if (!(fromP > 0 && toP > 0)) return null
  const k = toP / fromP
  const filtered = items.filter((it) => it.amount > 0)
  return { k, items: filtered.map((it) => ({ name: it.name, amount: it.amount, unit: it.unit, scaled: it.amount * k })) }
}

export function servingNutrition(per100: { kcal: number; p: number; f: number; c: number }, grams: number) {
  if (!(grams > 0)) return null
  const k = grams / 100
  return { kcal: per100.kcal * k, p: per100.p * k, f: per100.f * k, c: per100.c * k }
}

export const CUPS: Record<string, { g: number; name: string }> = {
  flour: { g: 130, name: 'Мука' },
  sugar: { g: 200, name: 'Сахар' },
  butter: { g: 227, name: 'Масло слив.' },
  milk: { g: 240, name: 'Молоко (мл)' },
  cocoa: { g: 100, name: 'Какао-порошок' },
  choc: { g: 170, name: 'Шоколад' },
  rice: { g: 180, name: 'Рис (сухой)' },
  oil: { g: 240, name: 'Масло раст. (мл)' },
}

export function cupsConvert(product: string, cups: number | null, grams: number | null) {
  const c = CUPS[product]
  if (!c) return null
  if (cups != null && cups > 0) return { g: cups * c.g, cups, dir: 'cups->g' as const }
  if (grams != null && grams > 0) return { g: grams, cups: grams / c.g, dir: 'g->cups' as const }
  return null
}

export function tipSplit(bill: number, pct: number, people: number) {
  if (!(bill > 0) || pct == null || pct < 0) return null
  const tip = (bill * pct) / 100
  const total = bill + tip
  const n = people > 1 ? people : 1
  return { tip, total, per: total / n, tipPer: tip / n }
}

export function fuelCost(km: number, l100: number, price: number | null) {
  if (!(km > 0 && l100 > 0)) return null
  const liters = (km * l100) / 100
  return { liters, cost: price != null ? liters * price : null }
}

export function metCalories(weightKg: number, met: number, hours: number) {
  if (!(weightKg > 0 && met > 0 && hours > 0)) return null
  return { kcal: met * weightKg * hours }
}

// ---------------- математика ----------------
export type ShapeKind = 'circleR' | 'circleD' | 'rect' | 'tri' | 'cyl' | 'sphere'

export function shapeCalc(kind: ShapeKind, a: number, b: number) {
  if (!(a > 0)) return null
  const fs = (x: number) => (Number.isInteger(x) ? String(x) : String(x))
  switch (kind) {
    case 'circleR':
      return { name: `Площадь круга (r=${fs(a)})`, value: Math.PI * a * a, unit: 'кв. ед.' }
    case 'circleD':
      return { name: `Площадь круга (d=${fs(a)})`, value: Math.PI * Math.pow(a / 2, 2), unit: 'кв. ед.' }
    case 'rect':
      if (!(b > 0)) return null
      return { name: 'Площадь прямоугольника', value: a * b, unit: 'кв. ед.' }
    case 'tri':
      if (!(b > 0)) return null
      return { name: 'Площадь прямоугольного треугольника', value: (a * b) / 2, unit: 'кв. ед.' }
    case 'cyl':
      if (!(b > 0)) return null
      return { name: `Объём цилиндра (r=${fs(a)}, h=${fs(b)})`, value: Math.PI * a * a * b, unit: 'куб. ед.' }
    case 'sphere':
      return { name: `Объём сферы (r=${fs(a)})`, value: (4 / 3) * Math.PI * Math.pow(a, 3), unit: 'куб. ед.' }
  }
  return null
}

export function parseNumbers(str: string | null): number[] {
  return String(str ?? '')
    .split(/[\s,;]+/)
    .filter((x) => x !== '')
    .map((x) => Number(String(x).replace(',', '.')))
    .filter((x) => isFinite(x))
}

export function statsOf(nums: number[]) {
  if (!nums.length) return null
  const s = [...nums].sort((a, b) => a - b)
  const sum = s.reduce((a, b) => a + b, 0)
  const median = s.length % 2 ? s[(s.length - 1) / 2] : (s[s.length / 2 - 1] + s[s.length / 2]) / 2
  return { n: s.length, sum, mean: sum / s.length, median, min: s[0], max: s[s.length - 1] }
}

export function vts(speed: number | null, time: number | null, dist: number | null) {
  if (speed != null && speed > 0 && time != null && time > 0) return { dist: speed * time }
  if (speed != null && speed > 0 && dist != null && dist > 0) return { time: dist / speed }
  if (time != null && time > 0 && dist != null && dist > 0) return { speed: dist / time }
  return null
}

export function dateDiff(a: string, b: string) {
  const d1 = new Date(a + 'T00:00:00')
  const d2 = new Date(b + 'T00:00:00')
  if (isNaN(d1.getTime()) || isNaN(d2.getTime())) return null
  const days = Math.round(Math.abs(d2.getTime() - d1.getTime()) / 86400000)
  const lo = Math.min(d1.getTime(), d2.getTime())
  const hi = Math.max(d1.getTime(), d2.getTime())
  let workdays = 0
  for (let t = lo; t <= hi; t += 86400000) {
    const wd = new Date(t).getDay()
    if (wd !== 0 && wd !== 6) workdays++
  }
  return { days, weeks: days / 7, months: days / 30.44, years: days / 365.25, hours: days * 24, workdays }
}

// ---------------- здоровье ----------------
export function bmiOf(hCm: number, wKg: number) {
  if (!(hCm > 0 && wKg > 0)) return null
  const h = hCm / 100
  const v = wKg / (h * h)
  const cat = v < 16 ? 'выраженный дефицит' : v < 18.5 ? 'дефицит веса' : v < 25 ? 'норма' : v < 30 ? 'избыточный вес' : 'ожирение'
  return { v, cat }
}

export function tdee(sex: 'm' | 'f', age: number, hCm: number, wKg: number, activity: number | null) {
  if (!(age > 0 && hCm > 0 && wKg > 0)) return null
  const base = 10 * wKg + 6.25 * hCm - 5 * age + (sex === 'm' ? 5 : -161)
  const act = activity || 1.2
  return { bmr: base, tdee: base * act, cut: base * act * 0.8, gain: base * act * 1.1 }
}

export function macros(kcal: number, pctP: number, pctF: number, pctC: number) {
  if (!(kcal > 0)) return null
  return { p: (kcal * pctP) / 100 / 4, f: (kcal * pctF) / 100 / 9, c: (kcal * pctC) / 100 / 4 }
}

export function waterLiters(weightKg: number, active: boolean) {
  if (!(weightKg > 0)) return null
  return { liters: weightKg * (active ? 0.035 : 0.03) }
}

export function oneRm(weight: number, reps: number) {
  if (!(weight > 0 && reps > 0)) return null
  return { v: weight * (1 + reps / 30) }
}

// ---------------- v1.1 ----------------
export function dueDate(lmpStr: string, cycleLen: number | null) {
  const d = new Date(lmpStr + 'T00:00:00')
  if (isNaN(d.getTime())) return null
  const base = new Date(d)
  base.setDate(base.getDate() + 280)
  const shift = cycleLen != null && cycleLen > 21 && cycleLen < 40 ? cycleLen - 28 : 0
  const due = new Date(base)
  due.setDate(due.getDate() + shift)
  return { due: fmtISO(due), base: fmtISO(base), shift, weeksTotal: 40 }
}

export function pregnancyWeeks(lmpStr: string, nowStr: string | null) {
  const l = new Date(lmpStr + 'T00:00:00')
  const n = nowStr ? new Date(nowStr + 'T00:00:00') : new Date()
  if (isNaN(l.getTime())) return null
  const days = Math.floor((n.getTime() - l.getTime()) / 86400000)
  if (days < 0 || days > 420) return null
  const w = Math.floor(days / 7)
  const rem = days % 7
  const label = w + ' нед.' + (rem ? ' ' + rem + ' дн.' : '')
  return { weeks: w, days: rem, label }
}

export function tvDiag(inch: number | null, cm: number | null) {
  if (inch != null && inch > 0) return { cm: inch * 2.54, inch }
  if (cm != null && cm > 0) return { cm, inch: cm / 2.54 }
  return null
}

export function tvViewDist(diagInch: number) {
  if (!(diagInch > 0)) return null
  return { minM: (diagInch * 1.5 * 2.54) / 100, maxM: (diagInch * 2.5 * 2.54) / 100 }
}

export function unitPrice(price: number, amount: number) {
  if (!(price > 0 && amount > 0)) return null
  const per100 = (price / amount) * 100
  return { per100, per1000: per100 * 10 }
}

export const CABLE_CU: Record<number, string> = {
  6: '1 мм²', 10: '1,5 мм²', 16: '2,5 мм²', 25: '4 мм²', 32: '6 мм²',
  40: '10 мм²', 50: '10 мм²', 63: '16 мм²', 80: '16 мм²', 100: '25 мм²',
}

export function wattsAmps(watts: number, volts: number, cosPhi: number | null) {
  const v = volts > 0 ? volts : 220
  const cp = cosPhi == null || cosPhi <= 0 ? 1 : Math.min(cosPhi, 1)
  if (!(watts > 0)) return null
  const a = watts / (v * cp)
  const breaks = [6, 10, 16, 25, 32, 40, 50, 63, 80, 100]
  let breaker: number | null = null
  for (const b of breaks) {
    if (a <= b) {
      breaker = b
      break
    }
  }
  return { amps: a, watts, volts: v, cosPhi: cp, breaker, cable: breaker ? CABLE_CU[breaker] : null }
}

// ================= v1.3 =================
export function savingsMonthly(goal: number, ratePct: number, months: number, start: number) {
  const g = goal > 0 ? goal : 0
  const s = start > 0 ? start : 0
  const m = months > 0 ? Math.round(months) : 0
  if (!m || g <= s) return null
  const i = ratePct > 0 ? ratePct / 100 / 12 : 0
  if (i === 0) {
    const pmt = (g - s) / m
    return { pmt, paid: pmt * m, interest: 0 }
  }
  const fvS = s * Math.pow(1 + i, m)
  const pmt = ((g - fvS) * i) / (Math.pow(1 + i, m) - 1)
  const paid = pmt * m
  return { pmt, paid, interest: g - s - paid }
}

export function rule72Calc(ratePct: number, years: number) {
  const r = ratePct > 0 ? ratePct : 0
  const y = years > 0 ? years : 0
  if (!r) return null
  return { growth: Math.pow(1 + r / 100, y), d72: 72 / r, dexact: Math.log(2) / Math.log(1 + r / 100) }
}

export function wallpaper(perimeter: number, wallH: number, rollW: number, rollL: number, patternCm: number, skipStrips: number) {
  const P = perimeter > 0 ? perimeter : 0
  const h = wallH > 0 ? wallH : 0
  const rw = rollW > 0 ? rollW : 0.53
  const rl = rollL > 0 ? rollL : 10
  if (!P || !h) return null
  const step = h + (patternCm > 0 ? patternCm / 100 : 0)
  const stripsTotal = Math.max(0, Math.ceil(P / rw) - (skipStrips > 0 ? Math.round(skipStrips) : 0))
  const stripsPer = rl >= step ? Math.max(1, Math.floor(rl / step)) : 1
  const rolls = Math.max(0, Math.ceil(stripsTotal / stripsPer))
  return { stripsTotal, stripsPer, rolls, area: rolls * rl * rw }
}

export function diluteCalc(c1: number, c2: number, v1: number | null, vt: number | null) {
  const a = c1 > 0 && c1 <= 100 ? c1 : 0
  const b = c2 > 0 && c2 <= 100 ? c2 : 0
  if (!a || !b || b > a) return null
  const hasV1 = v1 != null && v1 > 0
  const hasVt = vt != null && vt > 0
  if (!hasV1 && !hasVt) return null
  let src: number, water: number, total: number
  if (hasV1) {
    src = v1!
    total = (v1! * a) / b
    water = total - src
  } else {
    total = vt!
    src = (vt! * b) / a
    water = total - src
  }
  return { src, water, total }
}

export const UNITS: Record<string, Record<string, number>> = {
  length: { 'мм': 0.001, 'см': 0.01, 'м': 1, 'км': 1000, 'дюйм': 0.0254, 'фут': 0.3048, 'ярд': 0.9144, 'миля': 1609.344 },
  mass: { 'мг': 0.001, 'г': 1, 'кг': 1000, 'т': 1000000, 'унция': 28.349523125, 'фунт': 453.59237 },
  volume: { 'мл': 0.001, 'л': 1, 'м³': 1000, 'галлон (США)': 3.785411784, 'кварта (США)': 0.946352946, 'пинта (США)': 0.473176473, 'чашка (240 мл)': 0.24 },
}

export function unitConvert(kind: string, from: string, to: string, value: number) {
  const u = UNITS[kind]
  if (!u || u[from] == null || u[to] == null) return null
  if (!isFinite(value)) return null
  return (value * u[from]) / u[to]
}

export function tempConv(c: number | null, f: number | null) {
  if (c != null && isFinite(c)) return { c, f: (c * 9) / 5 + 32 }
  if (f != null && isFinite(f)) return { c: ((f - 32) * 5) / 9, f }
  return null
}

export function paceCalc(kmh: number | null, pmin: number | null, psec: number | null) {
  const useTempo = pmin != null && pmin >= 0
  const s = useTempo ? pmin! + ((psec && psec > 0 ? psec : 0) / 60) : kmh && kmh > 0 ? 60 / kmh : null
  if (!(s && s > 0)) return null
  let paceM = Math.floor(s)
  let paceS = Math.round((s - paceM) * 60)
  if (paceS >= 60) {
    paceM += 1
    paceS -= 60
  }
  return { spd: 60 / s, paceM, paceS, paceMin: s }
}

export function hrZones(age: number, maxHR: number | null) {
  const mx = maxHR && maxHR > 0 ? maxHR : age > 0 ? 220 - age : 220
  const defs: Array<[number, number, string]> = [
    [50, 60, 'Восстановление'],
    [60, 70, 'Базовая, жиросжигание'],
    [70, 80, 'Аэробная, темп'],
    [80, 90, 'Пороговая'],
    [90, 100, 'Максимальная'],
  ]
  return { max: mx, zones: defs.map((d) => ({ pct: `${d[0]}–${d[1]} %`, name: d[2], lo: Math.round((mx * d[0]) / 100), hi: Math.round((mx * d[1]) / 100) })) }
}
