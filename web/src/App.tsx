import { useEffect, useMemo, useRef, useState } from 'react'
import { CALCS, CATS, CalcDef, Row } from './catalog'
import * as C from './core'

/* ------------------------------------------------------------------ */
/* Лёгкие вспомогательные компоненты                                  */
/* ------------------------------------------------------------------ */

const THEME_KEY = 'lifeCalcThemeReact'

function useTheme() {
  const [theme, setTheme] = useState<'dark' | 'light'>(() => {
    try {
      return (localStorage.getItem(THEME_KEY) as 'dark' | 'light') || 'dark'
    } catch {
      return 'dark'
    }
  })
  useEffect(() => {
    document.documentElement.dataset.theme = theme
    try {
      localStorage.setItem(THEME_KEY, theme)
    } catch {
      /* ignore */
    }
  }, [theme])
  return { theme, toggle: () => setTheme((t) => (t === 'dark' ? 'light' : 'dark')) }
}

function ResultRows({ rows }: { rows: Row[] | null }) {
  if (!rows) return null
  return (
    <div className="results" aria-live="polite">
      {rows.map((r, i) => (
        <div className={'res' + (r.b ? ' big' : '')} key={i}>
          <span className="res-l">{r.l}</span>
          <span className="res-v">{r.v}</span>
        </div>
      ))}
    </div>
  )
}

function Err() {
  return <div className="err">Проверьте ввод — не хватает значений или они некорректны.</div>
}

/* ------------------------------------------------------------------ */
/* Общая карточка (простые калькуляторы)                              */
/* ------------------------------------------------------------------ */
function fieldDefault(f: CalcDef['fields'][number]): string {
  if (f.def != null) return String(f.def)
  if (f.defFn) return f.defFn()
  return ''
}

function GenericCard({ calc }: { calc: CalcDef }) {
  const memKey = 'mem_' + calc.id
  const [vals, setVals] = useState<Record<string, string>>(() => {
    const base: Record<string, string> = {}
    for (const f of calc.fields) base[f.k] = fieldDefault(f)
    try {
      const saved = JSON.parse(localStorage.getItem(memKey) || 'null')
      if (saved) for (const f of calc.fields) if (saved[f.k] != null && saved[f.k] !== '') base[f.k] = String(saved[f.k])
    } catch {
      /* ignore */
    }
    return base
  })
  const [rows, setRows] = useState<Row[] | null>(null)
  const [err, setErr] = useState(false)
  const [mounted, setMounted] = useState(false)

  const run = (e?: React.FormEvent) => {
    e?.preventDefault()
    const out = calc.calc(vals)
    if (!out) {
      setRows(null)
      setErr(true)
    } else {
      setRows(out)
      setErr(false)
    }
    try {
      localStorage.setItem(memKey, JSON.stringify(vals))
    } catch {
      /* ignore */
    }
  }

  useEffect(() => {
    const t = requestAnimationFrame(() => setMounted(true))
    return () => cancelAnimationFrame(t)
  }, [])

  const set = (k: string, v: string) => setVals((s) => ({ ...s, [k]: v }))
  const reset = () => {
    const base: Record<string, string> = {}
    for (const f of calc.fields) base[f.k] = fieldDefault(f)
    setVals(base)
    setRows(null)
    setErr(false)
  }

  return (
    <article className={'card' + (mounted ? ' in' : '')}>
      <header className="card-head">
        <span className="emoji" aria-hidden>
          {calc.emoji}
        </span>
        <h3>{calc.title}</h3>
      </header>

      <form onSubmit={run} className="fields">
        {calc.fields.map((f) => (
          <label className="field" key={f.k}>
            <span className="field-l">{f.l}</span>
            {f.type === 'select' ? (
              <select value={vals[f.k]} onChange={(e) => set(f.k, e.target.value)}>
                {(f.opts || []).map(([val, label]) => (
                  <option value={val} key={val}>
                    {label}
                  </option>
                ))}
              </select>
            ) : f.type === 'date' ? (
              <input type="date" value={vals[f.k]} onChange={(e) => set(f.k, e.target.value)} />
            ) : (
              <input
                inputMode={calc.id === 'avg' ? 'text' : 'decimal'}
                placeholder={f.ph ?? f.def != null ? undefined : '—'}
                value={vals[f.k]}
                onChange={(e) => set(f.k, e.target.value)}
              />
            )}
          </label>
        ))}

        <div className="card-actions">
          <button type="submit" className="btn primary">
            Рассчитать
          </button>
          <button type="button" className="btn ghost" onClick={reset} title="Сброс к значениям по умолчанию">
            Сброс
          </button>
        </div>
      </form>

      {err && <Err />}
      <ResultRows rows={rows} />

      {calc.hint && <p className="hint">💡 {calc.hint}</p>}
    </article>
  )
}

/* ------------------------------------------------------------------ */
/* Рецепт — динамический список ингредиентов                          */
/* ------------------------------------------------------------------ */
function RecipeCard() {
  const [from, setFrom] = useState('4')
  const [to, setTo] = useState('8')
  const [rows, setRows] = useState<Row[] | null>(null)
  const [items, setItems] = useState<Array<{ name: string; amount: string; unit: string }>>([
    { name: 'Мука', amount: '250', unit: 'г' },
    { name: 'Сахар', amount: '150', unit: 'г' },
    { name: 'Масло', amount: '100', unit: 'г' },
  ])

  const run = (e: React.FormEvent) => {
    e.preventDefault()
    const parsed = items
      .map((it) => ({ name: it.name.trim(), amount: C.parseNumRaw(it.amount), unit: it.unit.trim() }))
      .filter((x) => x.name && x.amount && x.amount > 0) as Array<{ name: string; amount: number; unit: string }>
    const fp = C.parseNumRaw(from)
    const tp = C.parseNumRaw(to)
    const r = C.recipeScale(fp || 0, tp || 0, parsed)
    if (!r || !r.items.length) {
      setRows(null)
      return
    }
    const out: Row[] = [{ l: 'Коэффициент', v: '×' + C.fmtNum(r.k, 2), b: true }]
    for (const it of r.items) {
      const qty = C.fmtNum(it.scaled, it.scaled % 1 ? 1 : 0)
      out.push({ l: `${it.name || '—'} · ${C.fmtNum(it.amount, it.amount % 1 ? 1 : 0)} ${it.unit || ''}`, v: `${qty} ${it.unit || ''}` })
    }
    if (tp && tp > 0) {
      out.push({ l: '—', v: '', b: false })
      for (const it of r.items) {
        const per = it.scaled / tp
        out.push({ l: `На порцию: ${it.name || '—'}`, v: `${C.fmtNum(per, per % 1 ? 1 : 0)} ${it.unit || ''}` })
      }
    }
    setRows(out)
  }

  const setItem = (i: number, patch: Partial<(typeof items)[number]>) =>
    setItems((arr) => arr.map((it, idx) => (idx === i ? { ...it, ...patch } : it)))

  return (
    <article className="card in">
      <header className="card-head">
        <span className="emoji" aria-hidden>
          🧑‍🍳
        </span>
        <h3>Рецепт: пересчёт порций</h3>
      </header>
      <form onSubmit={run}>
        <div className="two">
          <label className="field">
            <span className="field-l">Порций было</span>
            <input inputMode="decimal" value={from} onChange={(e) => setFrom(e.target.value)} />
          </label>
          <label className="field">
            <span className="field-l">Порций нужно</span>
            <input inputMode="decimal" value={to} onChange={(e) => setTo(e.target.value)} />
          </label>
        </div>
        <div className="ing-list">
          {items.map((it, i) => (
            <div className="ing-row" key={i}>
              <input placeholder="название" value={it.name} onChange={(e) => setItem(i, { name: e.target.value })} />
              <input placeholder="кол-во" inputMode="decimal" value={it.amount} onChange={(e) => setItem(i, { amount: e.target.value })} />
              <input placeholder="г/мл" value={it.unit} onChange={(e) => setItem(i, { unit: e.target.value })} />
              <button
                type="button"
                className="ing-del"
                title="Убрать ингредиент"
                disabled={items.length <= 1}
                onClick={() => setItems((arr) => arr.filter((_, idx) => idx !== i))}
              >
                ×
              </button>
            </div>
          ))}
        </div>
        <div className="card-actions">
          <button type="button" className="btn mini" onClick={() => setItems((arr) => [...arr, { name: '', amount: '', unit: '' }])}>
            ＋ Ингредиент
          </button>
        </div>
        <div className="card-actions">
          <button type="submit" className="btn primary">
            Рассчитать
          </button>
        </div>
      </form>
      <ResultRows rows={rows} />
      <p className="hint">💡 Все ингредиенты умножаются на одно и то же число — пропорции сохраняются.</p>
    </article>
  )
}

/* ------------------------------------------------------------------ */
/* Конвертер единиц                                                   */
/* ------------------------------------------------------------------ */
const KIND_LABEL: Record<string, string> = { length: '📏 Длина', mass: '⚖️ Масса', volume: '🧴 Объём' }

function UnitsCard() {
  const [kind, setKind] = useState('length')
  const [from, setFrom] = useState('м')
  const [to, setTo] = useState('см')
  const [value, setValue] = useState('1')
  const [rows, setRows] = useState<Row[] | null>(null)

  const units = Object.keys(C.UNITS[kind])
  const switchKind = (k: string) => {
    const prefs: Record<string, [string, string]> = { length: ['м', 'см'], mass: ['кг', 'г'], volume: ['л', 'мл'] }
    const [a, b] = prefs[k] || ['м', 'см']
    setKind(k)
    setFrom(a)
    setTo(b)
  }

  const run = (e: React.FormEvent) => {
    e.preventDefault()
    const n = C.parseNumRaw(value)
    const out = n != null ? C.unitConvert(kind, from, to, n) : null
    if (out == null) {
      setRows(null)
      return
    }
    setRows([{ l: `${C.fmtNum(n!, 6)} ${from} =`, v: `${C.fmtNum(out, 6)} ${to}`, b: true }])
  }

  return (
    <article className="card in">
      <header className="card-head">
        <span className="emoji" aria-hidden>
          🔄
        </span>
        <h3>Конвертер единиц</h3>
      </header>
      <form onSubmit={run}>
        <div className="seg" role="group" aria-label="Величина">
          {Object.keys(KIND_LABEL).map((k) => (
            <button
              type="button"
              className={'seg-btn' + (kind === k ? ' on' : '')}
              key={k}
              onClick={() => switchKind(k)}
            >
              {KIND_LABEL[k]}
            </button>
          ))}
        </div>
        <div className="two">
          <label className="field">
            <span className="field-l">Значение</span>
            <input inputMode="decimal" value={value} onChange={(e) => setValue(e.target.value)} />
          </label>
        </div>
        <div className="two">
          <label className="field">
            <span className="field-l">Из единицы</span>
            <select value={from} onChange={(e) => setFrom(e.target.value)}>
              {units.map((u) => (
                <option value={u} key={u}>
                  {u}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            <span className="field-l">В единицу</span>
            <select value={to} onChange={(e) => setTo(e.target.value)}>
              {units.map((u) => (
                <option value={u} key={u}>
                  {u}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div className="card-actions">
          <button type="submit" className="btn primary">
            Перевести
          </button>
        </div>
      </form>
      <ResultRows rows={rows} />
      <p className="hint">💡 Например: 1 м = 100 см, 1 кг = 1000 г, 1 л = 1000 мл.</p>
    </article>
  )
}

/* ------------------------------------------------------------------ */
/* Приложение                                                         */
/* ------------------------------------------------------------------ */
const CAT_EMOJI: Record<string, string> = {
  all: '🧮',
  money: '💰',
  home: '🏠',
  food: '🍽️',
  travel: '🚗',
  math: '📐',
  health: '❤️',
}

export default function App() {
  const { theme, toggle } = useTheme()
  const [cat, setCat] = useState<string>('all')
  const [q, setQ] = useState('')
  const gridRef = useRef<HTMLDivElement>(null)

  const visible = useMemo(() => {
    const query = q.trim().toLowerCase()
    return CALCS.filter((c) => {
      const okCat = cat === 'all' || c.cat === cat
      const okQ = !query || (c.title + ' ' + c.kw).toLowerCase().includes(query)
      return okCat && okQ
    })
  }, [cat, q])

  const onPickCat = (id: string) => {
    setCat(id)
    gridRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  return (
    <div className="app">
      <div className="bg-orbs" aria-hidden />
      <header className="topbar">
        <div className="brand">
          <span className="logo">🧮</span>
          <div>
            <h1>Калькуляторы на все случаи жизни</h1>
            <p className="sub">
              <b>40</b> бытовых калькуляторов · деньги · дом · еда · поездки · математика · здоровье
            </p>
          </div>
        </div>
        <button className="theme-btn" onClick={toggle} title="Переключить тему">
          {theme === 'dark' ? '☀️ Светлая' : '🌙 Тёмная'}
        </button>
      </header>

      <div className="toolbar">
        <label className="search">
          <span aria-hidden>🔍</span>
          <input
            type="text"
            placeholder="Поиск: кредит, ИМТ, краска, темп бега, обои…"
            value={q}
            onChange={(e) => setQ(e.target.value)}
          />
        </label>
      </div>

      <div className="chips">
        {CATS.map((c) => (
          <button
            key={c.id}
            className={'chip' + (cat === c.id ? ' on' : '')}
            onClick={() => onPickCat(c.id)}
          >
            <span aria-hidden>{CAT_EMOJI[c.id]}</span> {c.name}
          </button>
        ))}
      </div>

      <p className="count" role="status">
        Показано: <b>{visible.length}</b> из {CALCS.length}
        {q.trim() && <span className="hint-inline"> · поиск «{q.trim()}»</span>}
      </p>

      <div className="grid" ref={gridRef}>
        {visible.map((c) =>
          c.special === 'recipe' ? (
            <RecipeCard key={c.id} />
          ) : c.special === 'units' ? (
            <UnitsCard key={c.id} />
          ) : (
            <GenericCard calc={c} key={c.id} />
          ),
        )}
        {!visible.length && <p className="empty">Ничего не нашлось по запросу «{q}».</p>}
      </div>

      <footer>
        <p>
          React + TypeScript версия · все расчёты выполняются локально в браузере · значения запоминаются · работает офлайн.
        </p>
        <p className="disclaimer">
          Результаты — ориентиры для бытовых задач, не финансовая, медицинская или налоговая консультация.
        </p>
      </footer>
    </div>
  )
}
