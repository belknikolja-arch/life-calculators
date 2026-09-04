/* ============================================================
   Каталог калькуляторов (40) — TypeScript-порт CALCS из index.html.
   Каждый элемент: поля, подсказка и функция calc(values)->строки.
   values: Record<ключ поля, строка ввода> ('' = пусто).
   ============================================================ */
import * as C from './core'

export type Cat = 'money' | 'home' | 'food' | 'math' | 'health' | 'travel'
export type FieldType = 'text' | 'date' | 'select'

export interface Field {
  k: string
  l: string
  type?: FieldType
  opts?: Array<[string, string]>
  def?: string | number
  defFn?: () => string
  ph?: string
}

export type Row = { l: string; v: string; b?: boolean }

export interface CalcDef {
  id: string
  cat: Cat
  emoji: string
  title: string
  kw: string
  hint?: string
  fields: Field[]
  calc: (v: Record<string, string>) => Row[] | null
  special?: 'recipe' | 'units'
}

export const CATS: Array<{ id: Cat | 'all'; name: string }> = [
  { id: 'all', name: 'Все' },
  { id: 'money', name: 'Деньги' },
  { id: 'home', name: 'Дом' },
  { id: 'food', name: 'Еда' },
  { id: 'travel', name: 'Поездки' },
  { id: 'math', name: 'Математика' },
  { id: 'health', name: 'Здоровье' },
]

export const num = C.parseNumRaw
const fmtNum = C.fmtNum
const fmtR = C.fmtR
const fmtDur = C.fmtDur

// ——— helpers для результата ———
const R = (l: string, v: string | number, b = false): Row => ({ l, v: String(v), b })
const RB = (l: string, v: string | number): Row => R(l, v, true)

// ——— каталог ———
export const CALCS: CalcDef[] = []

function def(c: Omit<CalcDef, 'calc' | 'fields'> & Partial<Pick<CalcDef, 'fields' | 'calc'>>): void {
  CALCS.push({
    id: c.id,
    cat: c.cat,
    emoji: c.emoji,
    title: c.title,
    kw: c.kw,
    hint: c.hint,
    special: c.special,
    fields: (c.fields ?? []) as Field[],
    calc: (c.calc ?? (() => null)) as CalcDef['calc'],
  })
}

/* ================= ДЕНЬГИ ================= */
def({
  id: 'loan', cat: 'money', emoji: '🏦', title: 'Кредит / кредитка',
  kw: 'ипотека банковский платёж аннуитет переплата',
  hint: 'Аннуитетный платёж: одинаковый каждый месяц. Сначала банк «съедает» проценты, к концу срока — тело долга.',
  fields: [
    { k: 'p', l: 'Сумма, ₽', def: 1000000 },
    { k: 'r', l: 'Ставка, % годовых', def: 19 },
    { k: 'm', l: 'Срок, месяцев', def: 60 },
  ],
  calc: (v) => {
    const r = C.annuity(num(v.p)!, num(v.r)!, Math.round(num(v.m) || 0))
    if (!r) return null
    return [
      RB('Ежемесячный платёж', fmtR(r.pmt)),
      R('Всего выплатите', fmtR(r.total)),
      R('Переплата (проценты)', fmtR(r.interest)),
      R('Переплата, % от суммы', fmtNum((r.interest / num(v.p)!) * 100, 1) + ' %'),
    ]
  },
})
def({
  id: 'deposit', cat: 'money', emoji: '🏛️', title: 'Депозит',
  kw: 'вклад процент капитализация накопления',
  hint: 'Капитализация ежемесячная: проценты прибавляются к телу, и дальше процент идёт и на проценты.',
  fields: [
    { k: 'p', l: 'Начальная сумма, ₽', def: 1000000 },
    { k: 'r', l: 'Ставка, % годовых', def: 18 },
    { k: 'm', l: 'Срок, месяцев', def: 12 },
    { k: 'add', l: 'Взнос каждый месяц, ₽ (0 — без)', def: 0 },
  ],
  calc: (v) => {
    const r = C.depositFV(num(v.p) || 0, num(v.r)!, Math.round(num(v.m) || 0), num(v.add) || 0)
    if (!r) return null
    const mm = num(v.m) || 1
    return [
      RB('Будет на счёте', fmtR(r.end)),
      R('Из них ваши вложения', fmtR(r.contrib)),
      R('Проценты', fmtR(r.interest)),
      R('Прибыль в месяц', fmtR(r.interest / mm)),
      R('Прибыль, % от вложений', fmtNum((r.interest / (r.contrib || 1)) * 100, 1) + ' %'),
    ]
  },
})
def({
  id: 'vat1', cat: 'money', emoji: '🧾', title: 'НДС: выделить из цены',
  kw: 'налог добавленная стоимость 20 извлечь',
  fields: [
    { k: 'a', l: 'Сумма с НДС, ₽', def: 120000 },
    { k: 'r', l: 'Ставка НДС, %', def: 20 },
  ],
  calc: (v) => {
    const r = C.vatExtract(num(v.a)!, num(v.r) == null ? 20 : num(v.r))
    if (!r) return null
    return [
      RB('Цена без НДС', fmtR(r.base)),
      R('НДС', fmtR(r.vat)),
      R('НДС, % от базы', fmtNum((r.vat / (r.base || 1)) * 100, 1) + ' %'),
    ]
  },
})
def({
  id: 'vat2', cat: 'money', emoji: '➕', title: 'НДС: наложить сверху',
  kw: 'считать цену с налогом наценка',
  fields: [
    { k: 'a', l: 'Цена без НДС, ₽', def: 100000 },
    { k: 'r', l: 'Ставка НДС, %', def: 20 },
  ],
  calc: (v) => {
    const r = C.vatAdd(num(v.a) || 0, num(v.r) == null ? 20 : num(v.r))
    if (!r) return null
    return [
      RB('Цена с НДС', fmtR(r.total)),
      R('НДС', fmtR(r.vat)),
      R('Без НДС', fmtR(r.total - r.vat)),
    ]
  },
})
def({
  id: 'tax', cat: 'money', emoji: '💼', title: 'НДФЛ: brutto/netto',
  kw: 'налог зарплата на руки отчёт',
  hint: 'Заполните только одно поле: brutto (до налога) или netto (на руки) — второе посчитается.',
  fields: [
    { k: 'g', l: 'До вычета (brutto), ₽ — заполнить одно из двух' },
    { k: 'n', l: 'После вычета (netto), ₽', def: 100000 },
    { k: 'r', l: 'Ставка, %', def: 13 },
  ],
  calc: (v) => {
    const r = C.taxCalc(num(v.g) || 0, num(v.n) || 0, num(v.r) == null ? 13 : num(v.r))
    if (!r) return null
    return [
      RB('Brutto (до вычета)', fmtR(r.gross)),
      R('НДФЛ', fmtR(r.tax)),
      R('Netto (на руки)', fmtR(r.net)),
      R('На руки — % от brutto', fmtNum((r.net / (r.gross || 1)) * 100, 1) + ' %'),
    ]
  },
})
def({
  id: 'percent', cat: 'money', emoji: '➗', title: 'Проценты',
  kw: 'процент от числа сколько составляет скидка наценка',
  fields: [
    { k: 'x', l: 'Число', def: 200 },
    { k: 'p', l: '…процентов от него, %', def: 15 },
    { k: 'a', l: 'Число А (сколько % от Б?)', def: 30 },
    { k: 'b', l: 'Число Б', def: 200 },
  ],
  calc: (v) => {
    const x = num(v.x), p = num(v.p), a = num(v.a), b = num(v.b)
    const rows: Row[] = []
    if (x != null && p != null) {
      rows.push(RB(`${fmtNum(p, 2)}% от ${fmtNum(x, 2)}`, fmtNum(C.pctOf(x, p), 2)))
      rows.push(R(`Добавить ${fmtNum(p, 2)}%: ${fmtNum(x, 2)} →`, fmtNum(x + C.pctOf(x, p), 2)))
      rows.push(R(`Отнять ${fmtNum(p, 2)}%: ${fmtNum(x, 2)} →`, fmtNum(x - C.pctOf(x, p), 2)))
    }
    if (a != null && b != null && b !== 0) rows.push(R('Число А — это % от Б', fmtNum(C.pctOfBase(a, b)!, 2) + ' %'))
    return rows.length ? rows : null
  },
})
def({
  id: 'discount', cat: 'money', emoji: '🏷️', title: 'Скидка и итоговая цена',
  kw: 'скидка акция цена магазин',
  fields: [
    { k: 'p', l: 'Цена, ₽', def: 2990 },
    { k: 'd', l: 'Скидка, %', def: 25 },
  ],
  calc: (v) => {
    const r = C.discount(num(v.p)!, num(v.d) == null ? 0 : num(v.d)!)
    if (!r) return null
    return [
      RB('Итоговая цена', fmtR(r.final)),
      R('Выгода', fmtR(r.saved)),
      R('Это % от исходной цены', fmtNum((r.final / (r.price || 1)) * 100, 1) + ' %'),
    ]
  },
})
def({
  id: 'salary', cat: 'money', emoji: '⏱️', title: 'Зарплата во времени',
  kw: 'час день смена ставка почасово',
  hint: 'В месяце в среднем 4.33 недели.',
  fields: [
    { k: 'm', l: 'Зарплата в месяц, ₽', def: 200000 },
    { k: 'w', l: 'Часов в неделю', def: 40 },
  ],
  calc: (v) => {
    const r = C.salaryBreakdown(num(v.m)!, num(v.w))
    if (!r) return null
    return [
      RB('В час', fmtR(r.hourly)),
      R('За день (8 ч)', fmtR(r.day8)),
      R('За неделю', fmtR(r.week)),
      R('Оvertime (час по 1,5×)', fmtR(r.hourly * 1.5)),
      R('В год', fmtR(r.year)),
    ]
  },
})
def({
  id: 'inflation', cat: 'money', emoji: '📉', title: 'Инфляция: сила денег',
  kw: 'обесценивание рублей цены годы',
  fields: [
    { k: 'a', l: 'Сумма сегодня, ₽', def: 100000 },
    { k: 'r', l: 'Инфляция, % в год', def: 8 },
    { k: 'y', l: 'Годов', def: 10 },
  ],
  calc: (v) => {
    const a = num(v.a) || 0
    const r = C.inflationPower(a, num(v.r) || 0, num(v.y) || 0)
    if (!r) return null
    const rows: Row[] = [RB('Такая же покупательная способность будет', fmtR(r.value)), R('Потеряно', fmtR(r.lost))]
    if (r.value > 0) rows.push(R('Деньги станут слабее в', fmtNum(a / r.value, 2) + ' раза'))
    return rows
  },
})
def({
  id: 'unitprice', cat: 'money', emoji: '🛒', title: 'Цена за 100 г / кг',
  kw: 'весовое сравнить магазин сырок упаковка',
  fields: [
    { k: 'p', l: 'Цена упаковки, ₽', def: 189 },
    { k: 'g', l: 'Вес, г', def: 250 },
    { k: 'kg', l: 'Сколько купить, кг (опционально)' },
  ],
  calc: (v) => {
    const r = C.unitPrice(num(v.p)!, num(v.g)!)
    if (!r) return null
    const rows: Row[] = [RB('За 100 г', fmtR(r.per100)), R('За 1 кг', fmtR(r.per1000))]
    const kg = num(v.kg)
    if (kg && kg > 0) rows.push(R(`Стоимость ${fmtNum(kg, 2)} кг`, fmtR(r.per1000 * kg)))
    return rows
  },
})
def({
  id: 'savings', cat: 'money', emoji: '🐷', title: 'Копилка: накопить цель',
  kw: 'откладывать копить накопления цель вклад каждый месяц',
  hint: 'Обратная задача к депозиту: сколько откладывать в месяц, чтобы через N месяцев выйти на цель (с учётом процентов и уже накопленного).',
  fields: [
    { k: 'goal', l: 'Цель, ₽', def: 500000 },
    { k: 's', l: 'Уже отложено, ₽', def: 100000 },
    { k: 'r', l: 'Ставка по накоплениям, % годовых', def: 18 },
    { k: 'm', l: 'Срок, месяцев', def: 24 },
  ],
  calc: (v) => {
    const r = C.savingsMonthly(num(v.goal)!, num(v.r)!, Math.round(num(v.m) || 0), num(v.s) || 0)
    if (!r) return null
    if (r.pmt <= 0)
      return [RB('Откладывать в месяц', fmtR(0)), R('Уже накоплено достаточно — цель достижима', '✓')]
    return [
      RB('Откладывать в месяц', fmtR(r.pmt)),
      R('Всего внесёте сами', fmtR(r.paid)),
      R('Прирост за счёт процентов', fmtR(r.interest)),
    ]
  },
})
def({
  id: 'rule72', cat: 'money', emoji: '📈', title: 'Правило 72: рост и удвоение',
  kw: 'срок удвоения сложный процент во сколько раз вырастет инвестиции доходность',
  hint: '72 ÷ ставка ≈ годы, за которые капитал удвоится (надёжно для ставок до ~15 %). Рядом — точный логарифмический срок.',
  fields: [
    { k: 'r', l: 'Доходность, % годовых', def: 12 },
    { k: 'y', l: 'Горизонт, лет', def: 10 },
  ],
  calc: (v) => {
    const r = C.rule72Calc(num(v.r) || 0, num(v.y) || 0)
    if (!r) return null
    return [
      RB('Капитал вырастет в … раз', fmtNum(r.growth, 1) + '×'),
      R('Срок удвоения (правило 72)', fmtNum(r.d72, 1) + ' лет'),
      R('Срок удвоения (точно)', fmtNum(r.dexact, 1) + ' лет'),
    ]
  },
})

/* ================= ДОМ ================= */
def({
  id: 'room', cat: 'home', emoji: '📏', title: 'Комната: площадь и стены',
  kw: 'ремонт метры квадратная стены потолки',
  fields: [
    { k: 'l', l: 'Длина, м', def: 4 },
    { k: 'w', l: 'Ширина, м', def: 3 },
    { k: 'h', l: 'Высота потолка, м', def: 2.7 },
    { k: 'o', l: 'Двери и окна, м² (опционально)' },
  ],
  calc: (v) => {
    const r = C.roomGeom(num(v.l)!, num(v.w)!, num(v.h) || 0)
    if (!r) return null
    const rows: Row[] = [
      RB('Площадь пола', fmtNum(r.area, 2) + ' м²'),
      R('Периметр', fmtNum(r.per, 2) + ' м'),
      R('Площадь стен', fmtNum(r.walls, 2) + ' м²'),
      R('Потолок', fmtNum(r.ceil, 2) + ' м²'),
    ]
    const o = num(v.o)
    if (o && o > 0) rows.push(R('Стены за вычетом проёмов', fmtNum(Math.max(0, r.walls - o), 2) + ' м²'))
    return rows
  },
})
def({
  id: 'paint', cat: 'home', emoji: '🎨', title: 'Краска / шпаклёвка',
  kw: 'краска банки литраж укрывистость слои',
  hint: 'Возьмите с запасом ~10%: стены «пьют» по-разному.',
  fields: [
    { k: 'a', l: 'Площадь, м²', def: 60 },
    { k: 'c', l: 'Расход: м² на литр', def: 10 },
    { k: 'n', l: 'Слоёв', def: 2 },
    { k: 'can', l: 'Объём банки, л (0 — без)', def: 2.5 },
    { k: 'pr', l: 'Цена, ₽/л (опционально)' },
  ],
  calc: (v) => {
    const r = C.paintNeeded(num(v.a)!, num(v.c)!, num(v.n) || 1, num(v.can) || 0)
    if (!r) return null
    const rows: Row[] = [RB('Нужно краски', fmtNum(r.liters, 2) + ' л')]
    if (r.cans != null) rows.push(R(`Банок по ${fmtNum(num(v.can) || 0, 1)} л`, String(r.cans)))
    const pr = num(v.pr)
    if (pr && pr > 0) rows.push(R('Стоимость ≈', fmtR(r.liters * pr)))
    return rows
  },
})
def({
  id: 'floor', cat: 'home', emoji: '🪵', title: 'Ламинат / плитка / линолеум',
  kw: 'напольное покрытие запас упаковка подрезка',
  fields: [
    { k: 'a', l: 'Площадь пола, м²', def: 20 },
    { k: 'l', l: 'Запас на подрезку, %', def: 8 },
    { k: 'p', l: 'м² в упаковке (0 — без)', def: 2 },
    { k: 'pr', l: 'Цена, ₽/м² (опционально)' },
  ],
  calc: (v) => {
    const r = C.flooringNeeded(num(v.a)!, num(v.l) || 0, num(v.p) || 0)
    if (!r) return null
    const rows: Row[] = [RB('Купить с запасом', fmtNum(r.need, 2) + ' м²')]
    if (r.packs != null) rows.push(R('Упаковок', String(r.packs)))
    const pr = num(v.pr)
    if (pr && pr > 0) rows.push(R('Стоимость ≈', fmtR(r.need * pr)))
    return rows
  },
})
def({
  id: 'watt', cat: 'home', emoji: '⚡', title: 'Электричество: сколько стоит прибор',
  kw: 'ватт киловатты чайник расход света',
  fields: [
    { k: 'w', l: 'Мощность, Вт', def: 2000 },
    { k: 'h', l: 'Часов в день', def: 2 },
    { k: 'd', l: 'Дней', def: 30 },
    { k: 'r', l: 'Тариф, ₽ за кВт·ч', def: 6 },
  ],
  calc: (v) => {
    const r = C.applianceCost(num(v.w)!, num(v.h) || 0, num(v.d) || 0, num(v.r) || 0)
    if (!r) return null
    return [
      RB('Расход', fmtNum(r.kwh, 2) + ' кВт·ч'),
      R('Стоимость', fmtR(r.cost)),
      R('За год (×12)', fmtR(r.cost * 12)),
    ]
  },
})
def({
  id: 'tv', cat: 'home', emoji: '📺', title: 'Диагональ ТВ: дюймы ↔ см',
  kw: 'телевизор диагональ размер дюйм',
  fields: [
    { k: 'i', l: 'Дюймы (заполните одно из двух)', def: 55 },
    { k: 'c', l: '…или сантиметры' },
  ],
  calc: (v) => {
    const r = C.tvDiag(num(v.i), num(v.c))
    if (!r) return null
    const rows: Row[] = [RB('Дюймов', fmtNum(r.inch, 1)), R('Сантиметров', fmtNum(r.cm, 1) + ' см')]
    const vd = C.tvViewDist(r.inch)
    if (vd) rows.push(R('Дистанция просмотра (HD)', fmtNum(vd.minM, 1) + ' – ' + fmtNum(vd.maxM, 1) + ' м'))
    return rows
  },
})
def({
  id: 'watts', cat: 'home', emoji: '🔌', title: 'Ватты ↔ амперы, автомат',
  kw: 'розетка автомат пробки ток нагрузка',
  hint: 'A = Вт/(В × cos φ). «Автомат» — ближайший стандартный номинал, «Кабель» — примерное сечение меди для скрытой проводки.',
  fields: [
    { k: 'w', l: 'Мощность прибора, Вт', def: 2000 },
    { k: 'v', l: 'Напряжение, В (по умолчанию 220)', def: 220 },
    { k: 'cp', l: 'cos φ (для двигателей/инверторов, по умолчанию 1)' },
  ],
  calc: (v) => {
    const r = C.wattsAmps(num(v.w)!, num(v.v)!, num(v.cp))
    if (!r) return null
    const rows: Row[] = [RB('Ток', fmtNum(r.amps, 2) + ' А')]
    if (r.breaker) rows.push(R('Ближайший автомат', r.breaker + ' А'))
    if (r.cable) rows.push(R('Кабель (медь, скрытая прокладка)', r.cable))
    return rows
  },
})
def({
  id: 'wallpaper', cat: 'home', emoji: '🧱', title: 'Обои: сколько рулонов',
  kw: 'поклейка ремонт рулон комната стены периметр',
  hint: 'Полосы считаются по периметру: из рулона выходит столько полос, сколько раз в его длине умещается высота стены плюс раппорт.',
  fields: [
    { k: 'per', l: 'Периметр комнаты, м', def: 14 },
    { k: 'h', l: 'Высота потолка, м', def: 2.5 },
    { k: 'rw', l: 'Ширина рулона, м', def: 0.53 },
    { k: 'rl', l: 'Длина рулона, м', def: 10 },
    { k: 'pat', l: 'Подгонка рисунка (раппорт), см' },
    { k: 'skip', l: 'Полос, закрытых проёмами (дверь ≈ 1)' },
  ],
  calc: (v) => {
    const r = C.wallpaper(num(v.per)!, num(v.h)!, num(v.rw) || 0.53, num(v.rl) || 10, num(v.pat) || 0, num(v.skip) || 0)
    if (!r) return null
    return [
      RB('Нужно рулонов', r.rolls + ' шт'),
      R('Всего полос по периметру', r.stripsTotal + ' шт'),
      R('Полос из одного рулона', r.stripsPer + ' шт'),
      R('Купленная площадь (с запасом)', fmtNum(r.area, 1) + ' м²'),
    ]
  },
})
def({
  id: 'temp', cat: 'home', emoji: '🌡️', title: 'Температура: °C ↔ °F',
  kw: 'духовка цельсий фаренгейт градусы рецепт выпечка',
  hint: 'Заполните °C или °F. Духовка (примерно): 160 °C = 320 °F, 180 °C = 350 °F, 200 °C = 400 °F.',
  fields: [
    { k: 'c', l: 'Градусы Цельсия, °C', def: 180 },
    { k: 'f', l: 'Градусы Фаренгейта, °F' },
  ],
  calc: (v) => {
    const r = C.tempConv(num(v.c), num(v.f))
    if (!r) return null
    return [RB('По Цельсию', fmtNum(r.c, 0) + ' °C'), R('По Фаренгейту', fmtNum(r.f, 0) + ' °F')]
  },
})

/* ================= ЕДА ================= */
def({
  id: 'recipe', cat: 'food', emoji: '🧑‍🍳', title: 'Рецепт: пересчёт порций',
  kw: 'порции масштабировать блюдо ингредиент',
  special: 'recipe',
  fields: [
    { k: 'from', l: 'Порций было', def: 4 },
    { k: 'to', l: 'Порций нужно', def: 8 },
  ],
})
def({
  id: 'calories', cat: 'food', emoji: '🔥', title: 'Калории порции',
  kw: 'бжу ккал упаковка на 100 грамм',
  fields: [
    { k: 'k', l: 'Ккал на 100 г', def: 200 },
    { k: 'p', l: 'Белки, г на 100 г', def: 4 },
    { k: 'f', l: 'Жиры, г на 100 г', def: 3 },
    { k: 'c', l: 'Углеводы, г на 100 г', def: 10 },
    { k: 'g', l: 'Ваша порция, г', def: 50 },
  ],
  calc: (v) => {
    const r = C.servingNutrition(
      { kcal: num(v.k) || 0, p: num(v.p) || 0, f: num(v.f) || 0, c: num(v.c) || 0 },
      num(v.g)!,
    )
    if (!r) return null
    return [
      RB('В порции', fmtNum(r.kcal, 1) + ' ккал'),
      R('Белки', fmtNum(r.p, 1) + ' г'),
      R('Жиры', fmtNum(r.f, 1) + ' г'),
      R('Углеводы', fmtNum(r.c, 1) + ' г'),
      R('% от дневных 2000 ккал', fmtNum((r.kcal / 2000) * 100, 0) + ' %'),
    ]
  },
})
def({
  id: 'cups', cat: 'food', emoji: '🥣', title: 'Стаканы ↔ граммы (выпечка)',
  kw: 'чашки стаканы мука сахар выпечка конвертер',
  fields: [
    {
      k: 'prod', l: 'Продукт', type: 'select',
      opts: Object.keys(C.CUPS).map((k) => [k, C.CUPS[k].name] as [string, string]),
      def: 'flour',
    },
    { k: 'cups', l: 'Стаканов (если считаете стаканы)' },
    { k: 'grams', l: '…или граммов (заполните одно из двух)', def: 130 },
  ],
  calc: (v) => {
    const r = C.cupsConvert(v.prod, num(v.cups), num(v.grams))
    if (!r) return null
    const per = C.CUPS[v.prod]?.g
    const extra: Row[] = per ? [R('1 стакан =', fmtNum(per, 0) + ' г')] : []
    return r.dir === 'cups->g'
      ? [RB('В граммах', fmtNum(r.g, 1) + ' г'), R('Стаканов', fmtNum(r.cups, 2)), ...extra]
      : [RB('В стаканах', fmtNum(r.cups, 2) + ' стак.'), R('Граммов', fmtNum(r.g, 1) + ' г'), ...extra]
  },
})
def({
  id: 'tip', cat: 'food', emoji: '🍽️', title: 'Чаевые и делёж',
  kw: 'чаевые ресторан счёт пополам делим',
  fields: [
    { k: 'b', l: 'Счёт, ₽', def: 3500 },
    { k: 'p', l: 'Чаевые, %', def: 10 },
    { k: 'n', l: 'Человек делим (1 — без делёжа)', def: 2 },
  ],
  calc: (v) => {
    const r = C.tipSplit(num(v.b)!, num(v.p) == null ? 10 : num(v.p)!, Math.round(num(v.n) || 1))
    if (!r) return null
    const rows: Row[] = [RB('Итого с чаевыми', fmtR(r.total)), R('Чаевые', fmtR(r.tip))]
    if (r.per < r.total) {
      rows.push(R('С каждого человека', fmtR(r.per)))
      const n = Math.round(num(v.n) || 1)
      const rnd = Math.ceil(r.per / 50) * 50
      if (rnd > r.per) rows.push(R('Ровно, округлив до 50 ₽', `по ${fmtR(rnd)} — итого ${fmtR(rnd * n)}`))
    }
    return rows
  },
})
def({
  id: 'dilute', cat: 'food', emoji: '🍶', title: 'Разбавление: уксус и спирт',
  kw: 'уксусная эссенция развести спирт самогон вода концентрация 70 в 9',
  hint: 'Заполните объём исходного ИЛИ нужный объём готового. Пример: 100 мл 70%-й эссенции → 9%-й уксус — добавить 100×(70/9−1) ≈ 678 мл воды.',
  fields: [
    { k: 'c1', l: 'Концентрация исходная, %', def: 70 },
    { k: 'c2', l: 'Нужная концентрация, %', def: 9 },
    { k: 'v1', l: 'Объём исходного, мл' },
    { k: 'vt', l: 'Сколько нужно готового раствора, мл', def: 1000 },
  ],
  calc: (v) => {
    const r = C.diluteCalc(num(v.c1)!, num(v.c2)!, num(v.v1), num(v.vt))
    if (!r) return null
    const rows: Row[] = []
    if (num(v.v1) == null) rows.push(R('Взять исходного', fmtNum(r.src, 1) + ' мл'))
    rows.push(RB('Добавить воды', fmtNum(r.water, 1) + ' мл'), R('Итоговый объём', fmtNum(r.total, 1) + ' мл'))
    return rows
  },
})

/* ================= ПОЕЗДКИ ================= */
def({
  id: 'fuel', cat: 'travel', emoji: '⛽', title: 'Топливо на поездку',
  kw: 'бензин дизель расход литры маршрут',
  fields: [
    { k: 'km', l: 'Километров', def: 450 },
    { k: 'l', l: 'Расход, л/100 км', def: 8 },
    { k: 'p', l: 'Цена, ₽/л', def: 60 },
  ],
  calc: (v) => {
    const r = C.fuelCost(num(v.km)!, num(v.l)!, num(v.p) == null ? 0 : num(v.p)!)
    if (!r) return null
    const rows: Row[] = [RB('Нужно топлива', fmtNum(r.liters, 1) + ' л')]
    if (r.cost != null) {
      rows.push(R('Стоимость', fmtR(r.cost)), R('Туда и обратно (×2)', fmtR(r.cost * 2)))
    }
    rows.push(R('CO₂ в атмосферу ≈', fmtNum(r.liters * 2.31, 1) + ' кг'))
    return rows
  },
})

/* ================= МАТЕМАТИКА ================= */
def({
  id: 'shapes', cat: 'math', emoji: '📐', title: 'Площади и объёмы',
  kw: 'круг прямоугольник треугольник цилиндр сфера',
  fields: [
    {
      k: 'k', l: 'Фигура', type: 'select',
      opts: [
        ['circleR', 'Круг (по радиусу)'],
        ['circleD', 'Круг (по диаметру)'],
        ['rect', 'Прямоугольник'],
        ['tri', 'Прямоугольный треугольник'],
        ['cyl', 'Цилиндр'],
        ['sphere', 'Сфера'],
      ],
      def: 'circleR',
    },
    { k: 'a', l: 'Значение A (r / сторона)', def: 2 },
    { k: 'b', l: 'Значение B (h / вторая сторона)', def: 5 },
  ],
  calc: (v) => {
    const r = C.shapeCalc(v.k as C.ShapeKind, num(v.a)!, num(v.b)!)
    if (!r) return null
    const a = num(v.a)!, b = num(v.b)!
    const rows: Row[] = [RB(r.name, fmtNum(r.value, 3) + ' ' + r.unit)]
    if (v.k === 'circleR') rows.push(R('Длина окружности (2πr)', fmtNum(2 * Math.PI * a, 3)))
    if (v.k === 'circleD') rows.push(R('Длина окружности (πd)', fmtNum(Math.PI * a, 3)))
    if (v.k === 'rect') rows.push(R('Периметр (2(a+b))', fmtNum(2 * (a + b), 3)))
    if (v.k === 'cyl' && b > 0) rows.push(R('Боковая площадь (2πrh)', fmtNum(2 * Math.PI * a * b, 3)))
    return rows
  },
})
def({
  id: 'avg', cat: 'math', emoji: '🧮', title: 'Среднее из списка чисел',
  kw: 'медиана сумма минимум максимум статистика',
  fields: [{ k: 'list', l: 'Числа (через пробел или запятую)', def: '120 95 110 130 98' }],
  calc: (v) => {
    const s = C.statsOf(C.parseNumbers(v.list))
    if (!s) return null
    return [
      R('Чисел', s.n),
      R('Сумма', fmtNum(s.sum, 2)),
      RB('Среднее', fmtNum(s.mean, 2)),
      R('Медиана', fmtNum(s.median, 2)),
      R('Минимум', fmtNum(s.min, 2)),
      R('Максимум', fmtNum(s.max, 2)),
      R('Размах (max − min)', fmtNum(s.max - s.min, 2)),
    ]
  },
})
def({
  id: 'vts', cat: 'math', emoji: '🚗', title: 'Скорость · время · путь',
  kw: 'расстояние часы км в час маршрут',
  hint: 'Заполните любые два поля из трёх — третье посчитается.',
  fields: [
    { k: 's', l: 'Скорость (км/ч)', def: 60 },
    { k: 't', l: 'Время (ч)', def: 2 },
    { k: 'd', l: 'Путь (км) — оставьте пустым то, что ищем' },
  ],
  calc: (v) => {
    const r = C.vts(num(v.s), num(v.t), num(v.d))
    if (!r) return null
    if (r.dist != null) return [RB('Путь', fmtNum(r.dist, 2) + ' км')]
    if (r.time != null) return [RB('Время', fmtNum(r.time, 2) + ' ч'), R('Минут', fmtNum(r.time * 60, 0))]
    return [RB('Скорость', fmtNum(r.speed!, 2) + ' км/ч'), R('В м/с', fmtNum(r.speed! / 3.6, 2))]
  },
})
def({
  id: 'dates', cat: 'math', emoji: '📅', title: 'Между датами',
  kw: 'дней сколько разницы дата возраст',
  fields: [
    { k: 'a', l: 'Дата 1', type: 'date', defFn: C.todayStr },
    { k: 'b', l: 'Дата 2', type: 'date', defFn: C.todayStr },
  ],
  calc: (v) => {
    const r = C.dateDiff(v.a, v.b)
    if (!r) return null
    return [
      RB('Дней', r.days),
      R('Часов', fmtNum(r.hours, 0)),
      R('Недель', fmtNum(r.weeks, 1)),
      R('≈ месяцев', fmtNum(r.months, 1)),
      R('≈ лет', fmtNum(r.years, 2)),
      R('Будних дней (пн–пт, с датами)', r.workdays),
    ]
  },
})
def({
  id: 'units', cat: 'math', emoji: '🔄', title: 'Конвертер единиц',
  kw: 'перевод длина масса объём см м кг грамм литры дюймы футы мили галлон',
  special: 'units',
  fields: [],
})

/* ================= ЗДОРОВЬЕ ================= */
def({
  id: 'bmi', cat: 'health', emoji: '⚖️', title: 'ИМТ (индекс массы тела)',
  kw: 'имт вес рост категория воз',
  hint: 'По классификации ВОЗ. ИМТ не видит распределение мышц/жиров — это ориентир, не диагноз.',
  fields: [
    { k: 'h', l: 'Рост, см', def: 176 },
    { k: 'w', l: 'Вес, кг', def: 80 },
  ],
  calc: (v) => {
    const r = C.bmiOf(num(v.h)!, num(v.w)!)
    if (!r) return null
    const h = num(v.h)! / 100
    return [
      RB('ИМТ', fmtNum(r.v, 1)),
      R('Категория (ВОЗ)', r.cat),
      R('Вес при ИМТ 18.5–24.9', fmtNum(18.5 * h * h, 1) + ' – ' + fmtNum(24.9 * h * h, 1) + ' кг'),
    ]
  },
})
def({
  id: 'tdee', cat: 'health', emoji: '🔥', title: 'Калории в день',
  kw: 'tdee bmr метаболизм похудеть набрать миффлин',
  hint: 'Формула Миффлина–Сан Жеора. −20% — плавное похудение, +10% — набор.',
  fields: [
    { k: 'sex', l: 'Пол', type: 'select', opts: [['m', 'Мужской'], ['f', 'Женский']], def: 'm' },
    { k: 'age', l: 'Возраст, лет', def: 30 },
    { k: 'h', l: 'Рост, см', def: 176 },
    { k: 'w', l: 'Вес, кг', def: 80 },
    {
      k: 'act', l: 'Активность', type: 'select',
      opts: [
        ['1.2', 'Диван (сидячая работа)'],
        ['1.375', 'Лёгкая (1–3 тренировки/нед)'],
        ['1.55', 'Средняя (3–5 тренировок/нед)'],
        ['1.72', 'Высокая (6–7 тренировок/нед)'],
        ['1.9', 'Спортсмен / физ. работа'],
      ],
      def: '1.375',
    },
  ],
  calc: (v) => {
    const r = C.tdee(v.sex as 'm' | 'f', num(v.age)!, num(v.h)!, num(v.w)!, num(v.act))
    if (!r) return null
    return [
      R('Обмен веществ (BMR)', fmtNum(r.bmr, 0) + ' ккал'),
      RB('Норма (TDEE)', fmtNum(r.tdee, 0) + ' ккал'),
      R('Похудение (−20%)', fmtNum(r.cut, 0) + ' ккал'),
      R('Набор (+10%)', fmtNum(r.gain, 0) + ' ккал'),
      R('Белок, ориентир (1,6 г/кг)', fmtNum((num(v.w) || 0) * 1.6, 0) + ' г/день'),
    ]
  },
})
def({
  id: 'macros', cat: 'health', emoji: '🥩', title: 'БЖУ из калорий',
  kw: 'белки жиры углеводы граммы соотношение',
  hint: 'Белок и углеводы — 4 ккал/г, жиры — 9 ккал/г. Доли в сумме лучше = 100%.',
  fields: [
    { k: 'k', l: 'Калорий в день', def: 2200 },
    { k: 'p', l: 'Белки, %', def: 30 },
    { k: 'f', l: 'Жиры, %', def: 20 },
    { k: 'c', l: 'Углеводы, %', def: 50 },
  ],
  calc: (v) => {
    const r = C.macros(num(v.k)!, num(v.p) || 0, num(v.f) || 0, num(v.c) || 0)
    if (!r) return null
    return [
      RB('Белки', fmtNum(r.p, 1) + ' г'),
      R('Жиры', fmtNum(r.f, 1) + ' г'),
      R('Углеводы', fmtNum(r.c, 1) + ' г'),
      R('На один приём (3 приёма)', `${fmtNum(r.p / 3, 1)} / ${fmtNum(r.f / 3, 1)} / ${fmtNum(r.c / 3, 1)} г`),
    ]
  },
})
def({
  id: 'water', cat: 'health', emoji: '💧', title: 'Норма воды',
  kw: 'литры жидкость пить активность',
  fields: [
    { k: 'w', l: 'Вес, кг', def: 75 },
    { k: 'a', l: 'Есть тренировки (35 мл/кг вместо 30)', type: 'select', opts: [['0', 'Нет'], ['1', 'Да']], def: '0' },
  ],
  calc: (v) => {
    const r = C.waterLiters(num(v.w)!, v.a === '1')
    if (!r) return null
    const rows: Row[] = [RB('Воды в день', fmtNum(r.liters, 2) + ' л'), R('Стаканов по 0.3 л', fmtNum(r.liters / 0.3, 1))]
    if (v.a !== '1') rows.push(R('С тренировками (+500 мл)', fmtNum(r.liters + 0.5, 2) + ' л'))
    return rows
  },
})
def({
  id: 'oner', cat: 'health', emoji: '🏋️', title: 'Рабочий 1ПМ (Эйпли)',
  kw: 'один повтор максимум силовой тренажёр',
  hint: '1ПМ = вес × (1 + повторы/30). Формула Эйпли — оценка, не мера точности.',
  fields: [
    { k: 'w', l: 'Вес, кг', def: 100 },
    { k: 'r', l: 'Сколько раз сделали', def: 5 },
  ],
  calc: (v) => {
    const r = C.oneRm(num(v.w)!, num(v.r)!)
    if (!r) return null
    return [
      RB('Примерный 1ПМ', fmtNum(r.v, 1) + ' кг'),
      R('10 повторов ≈ 70% 1ПМ', fmtNum(r.v * 0.7, 1) + ' кг'),
      R('6 повторов ≈ 80% 1ПМ', fmtNum(r.v * 0.8, 1) + ' кг'),
      R('3 повтора ≈ 90% 1ПМ', fmtNum(r.v * 0.9, 1) + ' кг'),
    ]
  },
})
def({
  id: 'metcal', cat: 'health', emoji: '🏃', title: 'Калории за активность',
  kw: 'мет меты сожечь ходьба бег тренировка',
  hint: 'Ккал = MET × вес × время. MET — коэффициент метаболического эквивалента.',
  fields: [
    { k: 'w', l: 'Вес, кг', def: 75 },
    {
      k: 'm', l: 'Активность', type: 'select',
      opts: [
        ['2.5', 'Ходьба 5 км/ч'],
        ['4', 'Быстрая ходьба 6,5 км/ч'],
        ['6', 'Бег 8 км/ч'],
        ['7', 'Плавание'],
        ['8', 'Велосипед 16 км/ч'],
        ['10', 'Тяжёлая тренировка / уборка'],
      ],
      def: '2.5',
    },
    { k: 'h', l: 'Часов', def: 0.5 },
  ],
  calc: (v) => {
    const r = C.metCalories(num(v.w)!, num(v.m)!, num(v.h)!)
    if (!r) return null
    const perHour = r.kcal / ((num(v.h) || 0) > 0 ? num(v.h)! : 1)
    const rows: Row[] = [RB('Сожжено', fmtNum(r.kcal, 0) + ' ккал')]
    if (perHour > 0) rows.push(R('Чтобы сжечь 1 кг жира (7 700 ккал)', fmtNum(7700 / perHour, 1) + ' ч такой активности'))
    return rows
  },
})
def({
  id: 'due', cat: 'health', emoji: '🤰', title: 'Роды: предполагаемая дата',
  kw: 'беременность срок гестация последний день',
  hint: 'Формула Нагеле: +280 дней (40 недель от последней менструации). При цикле ≠28 дней дата сдвигается. Это ориентир — точную дату даёт врач по УЗИ.',
  fields: [
    { k: 'l', l: 'Первый день последней менструации', type: 'date', defFn: C.todayStr },
    { k: 'c', l: 'Длина цикла, дней (по умолчанию 28)' },
  ],
  calc: (v) => {
    const r = C.dueDate(v.l, num(v.c))
    if (!r) return null
    const rows: Row[] = [RB('Ожидаемая дата', new Date(r.due + 'T00:00:00').toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' }))]
    if (r.shift !== 0)
      rows.push(
        R(
          'Нагеле (цикл 28 дн)',
          new Date(r.base + 'T00:00:00').toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' }) +
            (r.shift > 0 ? ' +' : ' ') + r.shift + ' дн.',
        ),
      )
    const pw = C.pregnancyWeeks(v.l, null)
    if (pw) rows.push(RB('Срок сейчас', pw.label))
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    const left = Math.round((new Date(r.due + 'T00:00:00').getTime() - today.getTime()) / 86400000)
    rows.push(left > 0 ? R('До предполагаемой даты', left + ' дн.') : left === 0 ? R('Сегодня', 'предполагаемая дата 🎉') : R('Срок перешагнут', 'на ' + -left + ' дн.'))
    return rows
  },
})
def({
  id: 'pace', cat: 'health', emoji: '🏃', title: 'Темп бега: мин/км ↔ км/ч',
  kw: 'бег марафон скорость время дистанция тренировка',
  hint: 'Заполните скорость ИЛИ темп (темп приоритетнее). Темп 6:00 мин/км ≈ 10 км/ч. Внизу — время на стандартные дистанции.',
  fields: [
    { k: 'spd', l: 'Скорость, км/ч' },
    { k: 'pmin', l: 'Темп: минут на км', def: 6 },
    { k: 'psec', l: 'Темп: секунд на км', def: 0 },
    { k: 'dist', l: 'Своя дистанция, км (для времени)' },
  ],
  calc: (v) => {
    const r = C.paceCalc(num(v.spd), num(v.pmin), num(v.psec))
    if (!r) return null
    const rows: Row[] = [RB('Темп', `${r.paceM}:${String(r.paceS).padStart(2, '0')} мин/км`), R('Скорость', fmtNum(r.spd, 1) + ' км/ч')]
    const d = num(v.dist)
    if (d && d > 0) rows.push(R(`Время на ${fmtNum(d, 2)} км`, fmtDur(r.paceMin * d)))
    rows.push(R('Время на 5 км', fmtDur(r.paceMin * 5)), R('Время на 10 км', fmtDur(r.paceMin * 10)))
    rows.push(R('Время на 21,1 км', fmtDur(r.paceMin * 21.0975)), R('Время на 42,2 км', fmtDur(r.paceMin * 42.195)))
    return rows
  },
})
def({
  id: 'hrzones', cat: 'health', emoji: '💓', title: 'Пульсовые зоны',
  kw: 'сердце пульс тренировка зоны чсс кардио бег возраст интенсивность',
  hint: 'Пусто в поле пульса = 220 − возраст. Зоны считаются как % от максимального пульса. Z1–Z2 — восстановление и база, Z4–Z5 — короткие интервалы.',
  fields: [
    { k: 'age', l: 'Возраст, лет', def: 30 },
    { k: 'mx', l: 'Максимальный пульс, уд/мин' },
  ],
  calc: (v) => {
    const r = C.hrZones(num(v.age) || 0, num(v.mx))
    const rows: Row[] = [RB('Максимальный пульс', r.max + ' уд/мин')]
    for (const z of r.zones) rows.push(R(`${z.pct} — ${z.name}`, `${z.lo}–${z.hi} уд/мин`))
    return rows
  },
})
