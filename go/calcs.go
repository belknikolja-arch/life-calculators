package main

// Каталог калькуляторов (Go-порт CALCS из index.html).

import (
	"math"
	"strconv"
	"time"
)

type Row struct {
	Label string
	Value string
	Big   bool
}
type Opt struct{ V, L string }
type Input struct {
	K, L, Defv string
	Type       string
	Opts       []Opt
	// для date: Defv "" означает «сегодня»
}
type Calc struct {
	ID, Cat, Emoji, Title, Kw, Hint string
	Inputs                          []Input
	Run                             func(map[string]string) []Row // nil для special
}

func inp(k, l, defv string) Input { return Input{K: k, L: l, Defv: defv} }
func inpSel(k, l, defv string, o []Opt) Input {
	return Input{K: k, L: l, Defv: defv, Type: "select", Opts: o}
}
func inpDate(k, l string) Input { return Input{K: k, L: l, Type: "date"} }

func todayYMD() string {
	t := time.Now()
	return t.Format("2006-01-02")
}

// helpers row
func r(l, v string) Row  { return Row{l, v, false} }
func rb(l, v string) Row { return Row{l, v, true} }

var CALCS []Calc

func add(c Calc) { CALCS = append(CALCS, c) }

// дата по-русски: "10 августа 2026 г."
var monthsRU = []string{"января", "февраля", "марта", "апреля", "мая", "июня",
	"июля", "августа", "сентября", "октября", "ноября", "декабря"}

func fmtDateRu(s string) string {
	t, ok := parseYMD(s)
	if !ok {
		return s
	}
	return strconv.Itoa(t.Day()) + " " + monthsRU[int(t.Month())-1] + " " + strconv.Itoa(t.Year()) + " г."
}

func f1(x float64) string { return fmtNum(x, 1) }
func f2(x float64) string { return fmtNum(x, 2) }
func it(x int) string     { return strconv.Itoa(x) }
func fnum(x float64) string { // как JS String(number) для целых/полуцелых значений
	if x == math.Trunc(x) {
		return it(int(x))
	}
	return strconv.FormatFloat(x, 'f', -1, 64)
}

func init() {
	// ================= ДЕНЬГИ =================
	add(Calc{"loan", "money", "🏦", "Кредит / кредитка", "ипотека банковский платёж аннуитет переплата",
		"Аннуитетный платёж: одинаковый каждый месяц. Сначала банк «съедает» проценты, к концу срока — тело долга.",
		[]Input{inp("p", "Сумма, ₽", "1000000"), inp("r", "Ставка, % годовых", "19"), inp("m", "Срок, месяцев", "60")},
		func(v map[string]string) []Row {
			p := numOr(v["p"], 0)
			months, _ := parseIntRounded(v["m"])
			rr, ok := annuity(p, numOr(v["r"], 0), months)
			if !ok || p == 0 {
				return nil
			}
			return []Row{rb("Ежемесячный платёж", fmtR(rr.pmt)), r("Всего выплатите", fmtR(rr.total)),
				r("Переплата (проценты)", fmtR(rr.interest)), r("Переплата, % от суммы", f1(rr.interest/p*100)+" %")}
		}})

	add(Calc{"deposit", "money", "🏛️", "Депозит", "вклад процент капиталация накопления",
		"Капитализация ежемесячная: проценты прибавляются к телу, и дальше процент идёт и на проценты.",
		[]Input{inp("p", "Начальная сумма, ₽", "1000000"), inp("r", "Ставка, % годовых", "18"),
			inp("m", "Срок, месяцев", "12"), inp("add", "Взнос каждый месяц, ₽ (0 — без)", "0")},
		func(v map[string]string) []Row {
			months, _ := parseIntRounded(v["m"])
			rr, ok := depositFV(numOr(v["p"], 0), numOr(v["r"], 0), months, numOr(v["add"], 0))
			if !ok {
				return nil
			}
			mm := numOr(v["m"], 1)
			if mm < 1 {
				mm = 1
			}
			return []Row{rb("Будет на счёте", fmtR(rr.end)), r("Из них ваши вложения", fmtR(rr.contrib)),
				r("Проценты", fmtR(rr.interest)), r("Прибыль в месяц", fmtR(rr.interest/mm)),
				r("Прибыль, % от вложений", f1(rr.interest/(rr.contrib*pos1(rr.contrib))*100)+" %")}
		}})

	add(Calc{"vat1", "money", "🧾", "НДС: выделить из цены", "налог добавленная стоимость 20 извлечь", "",
		[]Input{inp("a", "Сумма с НДС, ₽", "120000"), inp("r", "Ставка НДС, %", "20")},
		func(v map[string]string) []Row {
			rate := numOr(v["r"], 20)
			rr, ok := vatExtract(numOr(v["a"], 0), rate)
			if !ok {
				return nil
			}
			return []Row{rb("Цена без НДС", fmtR(rr.base)), r("НДС", fmtR(rr.vat)),
				r("НДС, % от базы", f1(rr.vat/(rr.base)*100)+" %")}
		}})

	add(Calc{"vat2", "money", "➕", "НДС: наложить сверху", "считать цену с налогом наценка", "",
		[]Input{inp("a", "Цена без НДС, ₽", "100000"), inp("r", "Ставка НДС, %", "20")},
		func(v map[string]string) []Row {
			rate := numOr(v["r"], 20)
			rr, ok := vatAdd(numOr(v["a"], 0), rate)
			if !ok {
				return nil
			}
			return []Row{rb("Цена с НДС", fmtR(rr.total)), r("НДС", fmtR(rr.vat)), r("Без НДС", fmtR(rr.base))}
		}})

	add(Calc{"tax", "money", "💼", "НДФЛ: brutto/netto", "налог зарплата на руки отчёт",
		"Заполните только одно поле: brutto (до налога) или netto (на руки) — второе посчитается.",
		[]Input{inp("g", "До вычета (brutto), ₽ — заполнить один из двух", ""),
			inp("n", "После вычета (netto), ₽", "100000"), inp("r", "Ставка, %", "13")},
		func(v map[string]string) []Row {
			rr, ok := taxCalc(numOr(v["g"], 0), numOr(v["n"], 0), numOr(v["r"], 13))
			if !ok {
				return nil
			}
			return []Row{rb("Brutto (до вычета)", fmtR(rr.gross)), r("НДФЛ", fmtR(rr.tax)),
				r("Netto (на руки)", fmtR(rr.net)), r("На руки — % от brutto", f1(rr.net/(rr.gross)*100)+" %")}
		}})

	add(Calc{"percent", "money", "➗", "Проценты", "процент от числа сколько составляет скидка наценка", "",
		[]Input{inp("x", "Число", "200"), inp("p", "…процентов от него, %", "15"),
			inp("a", "Число А (сколько % от Б?)", "30"), inp("b", "Число Б", "200")},
		func(v map[string]string) []Row {
			x, hx := numStr(v["x"])
			p, hp := numStr(v["p"])
			aa, ha := numStr(v["a"])
			bb, hb := numStr(v["b"])
			var rows []Row
			if hx && hp {
				px := f2(p)
				rows = append(rows, rb(px+"% от "+f2(x), f2(pctOf(x, p))))
				rows = append(rows, r("Добавить "+px+"%: "+f2(x)+" →", f2(x+pctOf(x, p))))
				rows = append(rows, r("Отнять "+px+"%: "+f2(x)+" →", f2(x-pctOf(x, p))))
			}
			if ha && hb && bb != 0 {
				rows = append(rows, r("Число А — это % от Б", f2(pctOfBase(aa, bb))+" %"))
			}
			if len(rows) == 0 {
				return nil
			}
			return rows
		}})

	add(Calc{"discount", "money", "🏷️", "Скидка и итоговая цена", "скидка акция цена магазин", "",
		[]Input{inp("p", "Цена, ₽", "2990"), inp("d", "Скидка, %", "25")},
		func(v map[string]string) []Row {
			rr, ok := discount(numOr(v["p"], 0), numOr(v["d"], 0))
			if !ok {
				return nil
			}
			return []Row{rb("Итоговая цена", fmtR(rr.final)), r("Выгода", fmtR(rr.saved)),
				r("Это % от исходной цены", f1(rr.final/(rr.price)*100)+" %")}
		}})

	add(Calc{"salary", "money", "⏱️", "Зарплата во времени", "час день смена ставка почасово", "В месяце в среднем 4.33 недели.",
		[]Input{inp("m", "Зарплата в месяц, ₽", "200000"), inp("w", "Часов в неделю", "40")},
		func(v map[string]string) []Row {
			rr, ok := salaryBreakdown(numOr(v["m"], 0), numOr(v["w"], 0))
			if !ok {
				return nil
			}
			return []Row{rb("В час", fmtR(rr.hourly)), r("За день (8 ч)", fmtR(rr.day8)),
				r("За неделю", fmtR(rr.week)), r("Оvertime (час по 1,5×)", fmtR(rr.hourly*1.5)), r("В год", fmtR(rr.year))}
		}})

	add(Calc{"inflation", "money", "📉", "Инфляция: сила денег", "обесценивание рублей цены годы", "",
		[]Input{inp("a", "Сумма сегодня, ₽", "100000"), inp("r", "Инфляция, % в год", "8"), inp("y", "Годов", "10")},
		func(v map[string]string) []Row {
			a := numOr(v["a"], 0)
			rr, ok := inflationPower(a, numOr(v["r"], 0), numOr(v["y"], 0))
			if !ok {
				return nil
			}
			rows := []Row{rb("Такая же покупательная способность будет", fmtR(rr.value)), r("Потеряно", fmtR(rr.lost))}
			if rr.value > 0 {
				rows = append(rows, r("Деньги станут слабее в", f2(a/rr.value)+" раза"))
			}
			return rows
		}})

	add(Calc{"unitprice", "money", "🛒", "Цена за 100 г / кг", "весовое сравнить магазин сырок упаковка", "",
		[]Input{inp("p", "Цена упаковки, ₽", "189"), inp("g", "Вес, г", "250"), inp("kg", "Сколько купить, кг (опционально)", "")},
		func(v map[string]string) []Row {
			rr, ok := unitPrice(numOr(v["p"], 0), numOr(v["g"], 0))
			if !ok {
				return nil
			}
			rows := []Row{rb("За 100 г", fmtR(rr.per100)), r("За 1 кг", fmtR(rr.per1000))}
			if kg, h := numStr(v["kg"]); h && kg > 0 {
				rows = append(rows, r("Стоимость "+f2(kg)+" кг", fmtR(rr.per1000*kg)))
			}
			return rows
		}})

	add(Calc{"savings", "money", "🐷", "Копилка: накопить цель", "откладывать копить накопления цель вклад каждый месяц",
		"Обратная задача к депозиту: сколько откладывать в месяц, чтобы через M месяцев выйти на цель (с учётом процентов и уже накопленного).",
		[]Input{inp("goal", "Цель, ₽", "500000"), inp("s", "Уже отложено, ₽", "100000"),
			inp("r", "Ставка по накоплениям, % годовых", "18"), inp("m", "Срок, месяцев", "24")},
		func(v map[string]string) []Row {
			m, _ := parseIntRounded(v["m"])
			s, ok := savingsMonthly(numOr(v["goal"], 0), numOr(v["r"], 0), m, numOr(v["s"], 0))
			if !ok {
				return nil
			}
			if s.pmt <= 0 {
				return []Row{rb("Откладывать в месяц", fmtR(0)+" ₽"), r("Уже накоплено достаточно — цель достижима", "✓")}
			}
			return []Row{rb("Откладывать в месяц", fmtR(s.pmt)), r("Всего внесёте сами", fmtR(s.paid)),
				r("Прирост за счёт процентов", fmtR(s.interest))}
		}})

	add(Calc{"rule72", "money", "📈", "Правило 72: рост и удвоение", "срок удвоения сложный процент во сколько раз вырастет инвестиции доходность",
		"72 ÷ ставка ≈ годы, за которые капитал удвоится (надёжно для ставок до ~15 %). Точный срок — логарифмический.",
		[]Input{inp("r", "Доходность, % годовых", "12"), inp("y", "Горизонт, лет", "10")},
		func(v map[string]string) []Row {
			g, d72, ex, ok := rule72Calc(numOr(v["r"], 0), numOr(v["y"], 0))
			if !ok {
				return nil
			}
			return []Row{rb("Капитал вырастет в … раз", f1(g)+"×"), r("Срок удвоения (правило 72)", f1(d72)+" лет"),
				r("Срок удвоения (точно)", f1(ex)+" лет")}
		}})

	// ================= ДОМ =================
	add(Calc{"room", "home", "📏", "Комната: площадь и стены", "ремонт метры квадратная стены потолки", "",
		[]Input{inp("l", "Длина, м", "4"), inp("w", "Ширина, м", "3"), inp("h", "Высота потолка, м", "2.7"),
			inp("o", "Двери и окна, м² (опционально)", "")},
		func(v map[string]string) []Row {
			rr, ok := roomGeom(numOr(v["l"], 0), numOr(v["w"], 0), numOr(v["h"], 0))
			if !ok {
				return nil
			}
			rows := []Row{rb("Площадь пола", f2(rr.area)+" м²"), r("Периметр", f2(rr.per)+" м"),
				r("Площадь стен", f2(rr.walls)+" м²"), r("Потолок", f2(rr.ceil)+" м²")}
			if o, h := numStr(v["o"]); h && o > 0 {
				rows = append(rows, r("Стены за вычетом проёмов", f2(math.Max(0, rr.walls-o))+" м²"))
			}
			return rows
		}})

	add(Calc{"paint", "home", "🎨", "Краска / шпаклёвка", "краска банки литраж укрывистость слои",
		"Возьмите с запасом ~10%: стены «пьют» по-разному.",
		[]Input{inp("a", "Площадь, м²", "60"), inp("c", "Расход: м² на литр", "10"), inp("n", "Слоёв", "2"),
			inp("can", "Объём банки, л (0 — без)", "2.5"), inp("pr", "Цена, ₽/л (опционально)", "")},
		func(v map[string]string) []Row {
			can := numOr(v["can"], 0)
			rr, ok := paintNeeded(numOr(v["a"], 0), numOr(v["c"], 0), numOr(v["n"], 1), can)
			if !ok {
				return nil
			}
			rows := []Row{rb("Нужно краски", f2(rr.liters)+" л")}
			if rr.hasCans {
				rows = append(rows, r("Банок по "+f1(can)+" л", it(rr.cans)))
			}
			if pr, h := numStr(v["pr"]); h && pr > 0 {
				rows = append(rows, r("Стоимость ≈", fmtR(rr.liters*pr)))
			}
			return rows
		}})

	add(Calc{"floor", "home", "🪵", "Ламинат / плитка / линолеум", "напольное покрытие запас упаковка подрезка", "",
		[]Input{inp("a", "Площадь пола, м²", "20"), inp("l", "Запас на подрезку, %", "8"),
			inp("p", "м² в упаковке (0 — без)", "2"), inp("pr", "Цена, ₽/м² (опционально)", "")},
		func(v map[string]string) []Row {
			rr, ok := flooringNeeded(numOr(v["a"], 0), numOr(v["l"], 0), numOr(v["p"], 0))
			if !ok {
				return nil
			}
			rows := []Row{rb("Купить с запасом", f2(rr.need)+" м²")}
			if rr.hasPacks {
				rows = append(rows, r("Упаковок", it(rr.packs)))
			}
			if pr, h := numStr(v["pr"]); h && pr > 0 {
				rows = append(rows, r("Стоимость ≈", fmtR(rr.need*pr)))
			}
			return rows
		}})

	add(Calc{"watt", "home", "⚡", "Электричество: сколько стоит прибор", "ватт киловатты чайник расход света", "",
		[]Input{inp("w", "Мощность, Вт", "2000"), inp("h", "Часов в день", "2"),
			inp("d", "Дней", "30"), inp("r", "Тариф, ₽ за кВт·ч", "6")},
		func(v map[string]string) []Row {
			rr, ok := applianceCost(numOr(v["w"], 0), numOr(v["h"], 0), numOr(v["d"], 0), numOr(v["r"], 0))
			if !ok {
				return nil
			}
			return []Row{rb("Расход", f2(rr.kwh)+" кВт·ч"), r("Стоимость", fmtR(rr.cost)), r("За год (×12)", fmtR(rr.cost*12))}
		}})

	add(Calc{"tv", "home", "📺", "Диагональ ТВ: дюймы ↔ см", "телевизор диагональ размер дюйм", "",
		[]Input{inp("i", "Дюймы (заполните одно из двух)", "55"), inp("c", "…или сантиметры", "")},
		func(v map[string]string) []Row {
			cm, hasCm := numStr(v["c"])
			ii, hasIn := numStr(v["i"])
			cmv, inchv, ok := tvDiag(ii, cm, hasIn, hasCm)
			if !ok {
				return nil
			}
			rows := []Row{rb("Дюймов", f1(inchv)), r("Сантиметров", f1(cmv)+" см")}
			mn, mx := tvViewDist(inchv)
			rows = append(rows, r("Дистанция просмотра (HD)", f1(mn)+" – "+f1(mx)+" м"))
			return rows
		}})

	add(Calc{"watts", "home", "🔌", "Ватты ↔ амперы, автомат", "розетка автомат пробки ток нагрузка",
		"A = Вт/(В × cos φ). «Автомат» — ближайший стандартный номинал (6…100 А), «Кабель» — примерное сечение меди для скрытой проводки.",
		[]Input{inp("w", "Мощность прибора, Вт", "2000"), inp("v", "Напряжение, В (по умолчанию 220)", "220"),
			inp("cp", "cos φ (только для двигателей/инверторов, по умолчанию 1)", "")},
		func(v map[string]string) []Row {
			cp, hasCp := numStr(v["cp"])
			rr, ok := wattsAmps(numOr(v["w"], 0), numOr(v["v"], 0), cp, hasCp)
			if !ok {
				return nil
			}
			rows := []Row{rb("Ток", f2(rr.amps)+" А")}
			if rr.breaker > 0 {
				rows = append(rows, r("Ближайший автомат", it(int(rr.breaker))+" А"))
				rows = append(rows, r("Кабель (медь, скрытая прокладка)", rr.cable))
			}
			return rows
		}})

	add(Calc{"wallpaper", "home", "🧱", "Обои: сколько рулонов", "поклейка ремонт рулон комната стены периметр",
		"Полосы считаются по периметру: из рулона выходит столько полос, сколько раз в его длине умещается высота стены плюс раппорт.",
		[]Input{inp("per", "Периметр комнаты (сумма длин стен), м", "14"), inp("h", "Высота потолка, м", "2.5"),
			inp("rw", "Ширина рулона, м", "0.53"), inp("rl", "Длина рулона, м", "10"),
			inp("pat", "Подгонка рисунка (раппорт), см", ""), inp("skip", "Полос, закрытых проёмами", "")},
		func(v map[string]string) []Row {
			skip := 0
			if s, h := parseIntRounded(v["skip"]); h {
				skip = s
			}
			rr, ok := wallpaper(numOr(v["per"], 0), numOr(v["h"], 0), numOr(v["rw"], 0.53),
				numOr(v["rl"], 10), numOr(v["pat"], 0), skip)
			if !ok {
				return nil
			}
			return []Row{rb("Нужно рулонов", it(rr.rolls)+" шт"), r("Всего полос по периметру", it(rr.stripsTotal)+" шт"),
				r("Полос из одного рулона", it(rr.stripsPer)+" шт"), r("Купленная площадь (с запасом)", f1(rr.area)+" м²")}
		}})

	add(Calc{"temp", "home", "🌡️", "Температура: °C ↔ °F", "духовка цельсий фаренгейт градусы рецепт выпечка",
		"Заполните °C или °F. Духовка (примерно): 160 °C = 320 °F, 180 °C = 350 °F, 200 °C = 400 °F.",
		[]Input{inp("c", "Градусы Цельсия, °C", "180"), inp("f", "Градусы Фаренгейта, °F", "")},
		func(v map[string]string) []Row {
			c, hc := numStr(v["c"])
			f, hf := numStr(v["f"])
			rr, ok := tempConv(c, f, hc, hf)
			if !ok {
				return nil
			}
			return []Row{rb("По Цельсию", fmtNum(rr.c, 0)+" °C"), r("По Фаренгейту", fmtNum(rr.f, 0)+" °F")}
		}})

	// ================= ЕДА =================
	add(Calc{"recipe", "food", "🧑‍🍳", "Рецепт: пересчёт порций", "порции масштабировать блюдо ингредиент",
		"Все ингредиенты умножаются на одно и то же число — пропорции сохраняются.",
		[]Input{inp("from", "Порций было", "4"), inp("to", "Порций нужно", "8")}, nil})

	add(Calc{"calories", "food", "🔥", "Калории порции", "бжу ккал упаковка на 100 грамм", "",
		[]Input{inp("k", "Ккал на 100 г", "200"), inp("p", "Белки, г на 100 г", "4"), inp("f", "Жиры, г на 100 г", "3"),
			inp("c", "Углеводы, г на 100 г", "10"), inp("g", "Ваша порция, г", "50")},
		func(v map[string]string) []Row {
			rr, ok := servingNutrition(numOr(v["k"], 0), numOr(v["p"], 0), numOr(v["f"], 0), numOr(v["c"], 0), numOr(v["g"], 0))
			if !ok {
				return nil
			}
			return []Row{rb("В порции", f1(rr.kcal)+" ккал"), r("Белки", f1(rr.p)+" г"), r("Жиры", f1(rr.f)+" г"),
				r("Углеводы", f1(rr.c)+" г"), r("% от дневных 2000 ккал", fmtNum(rr.kcal/2000*100, 0)+" %")}
		}})

	var cupOpts []Opt
	for _, k := range []string{"flour", "sugar", "butter", "milk", "cocoa", "choc", "rice", "oil"} {
		cupOpts = append(cupOpts, Opt{k, cupsName[k]})
	}
	add(Calc{"cups", "food", "🥣", "Стаканы ↔ граммы (выпечка)", "чашки стаканы мука сахар выпечка конвертер", "",
		[]Input{inpSel("prod", "Продукт", "flour", cupOpts), inp("cups", "Стаканов (если считаете стаканы)", ""),
			inp("grams", "…или граммов (заполните одно из двух)", "130")},
		func(v map[string]string) []Row {
			cups, hc := numStr(v["cups"])
			gr, hg := numStr(v["grams"])
			rr, ok := cupsConvert(v["prod"], cups, gr)
			if !ok || !(hc || hg) {
				return nil
			}
			var rows []Row
			if rr.toGrams {
				rows = append(rows, rb("В граммах", f1(rr.g)+" г"), r("Стаканов", f2(rr.cups)))
			} else {
				rows = append(rows, rb("В стаканах", f2(rr.cups)+" стак."), r("Граммов", f1(rr.g)+" г"))
			}
			if per, has := cupsG[v["prod"]]; has {
				rows = append(rows, r("1 стакан =", fmtNum(per, 0)+" г"))
			}
			return rows
		}})

	add(Calc{"tip", "food", "🍽️", "Чаевые и делёж", "чаевые ресторан счёт пополам делим", "",
		[]Input{inp("b", "Счёт, ₽", "3500"), inp("p", "Чаевые, %", "10"), inp("n", "Человек делим (1 — без делёжа)", "2")},
		func(v map[string]string) []Row {
			n := 1
			if nn, h := parseIntRounded(v["n"]); h {
				n = nn
			}
			rr, ok := tipSplit(numOr(v["b"], 0), numOr(v["p"], 10), n)
			if !ok {
				return nil
			}
			rows := []Row{rb("Итого с чаевыми", fmtR(rr.total)), r("Чаевые", fmtR(rr.tip))}
			if rr.per < rr.total {
				rows = append(rows, r("С каждого человека", fmtR(rr.per)))
				rnd := math.Ceil(rr.per/50) * 50
				if rnd > rr.per {
					rows = append(rows, r("Ровно, округлив до 50 ₽", "по "+fmtR(rnd)+" — итого "+fmtR(rnd*float64(n))))
				}
			}
			return rows
		}})

	add(Calc{"dilute", "food", "🍶", "Разбавление: уксус и спирт", "уксусная эссенция развести спирт самогон вода концентрация 70 в 9",
		"Заполните объём исходного ИЛИ нужный объём готового. Пример: 100 мл 70%-й эссенции → 9%-й уксус — добавить ~678 мл воды.",
		[]Input{inp("c1", "Концентрация исходная, %", "70"), inp("c2", "Нужная концентрация, %", "9"),
			inp("v1", "Объём исходного, мл", ""), inp("vt", "Сколько нужно готового раствора, мл", "1000")},
		func(v map[string]string) []Row {
			v1, hv1 := numStr(v["v1"])
			vt, hvt := numStr(v["vt"])
			rr, ok := diluteCalc(numOr(v["c1"], 0), numOr(v["c2"], 0), v1, vt, hv1, hvt)
			if !ok {
				return nil
			}
			var rows []Row
			if !hv1 {
				rows = append(rows, r("Взять исходного", f1(rr.src)+" мл"))
			}
			rows = append(rows, rb("Добавить воды", f1(rr.water)+" мл"), r("Итоговый объём", f1(rr.total)+" мл"))
			return rows
		}})

	// ================= ПОЕЗДКИ =================
	add(Calc{"fuel", "travel", "⛽", "Топливо на поездку", "бензин дизель расход литры маршрут", "",
		[]Input{inp("km", "Километров", "450"), inp("l", "Расход, л/100 км", "8"), inp("p", "Цена, ₽/л", "60")},
		func(v map[string]string) []Row {
			price, hp := numStr(v["p"])
			rr, ok := fuelCost(numOr(v["km"], 0), numOr(v["l"], 0), price, hp)
			if !ok {
				return nil
			}
			rows := []Row{rb("Нужно топлива", f1(rr.liters)+" л")}
			if rr.hasCost {
				rows = append(rows, r("Стоимость", fmtR(rr.cost)), r("Туда и обратно (×2)", fmtR(rr.cost*2)))
			}
			rows = append(rows, r("CO₂ в атмосферу ≈", f1(rr.liters*2.31)+" кг"))
			return rows
		}})

	// ================= МАТЕМАТИКА =================
	var shapeOpts = []Opt{{"circleR", "Круг (по радиусу)"}, {"circleD", "Круг (по диаметре)"}, {"rect", "Прямоугольник"},
		{"tri", "Прямоугольный треугольник"}, {"cyl", "Цилиндр"}, {"sphere", "Сфера"}}
	add(Calc{"shapes", "math", "📐", "Площади и объёмы", "круг прямоугольник треугольник цилиндр сфера", "",
		[]Input{inpSel("k", "Фигура", "circleR", shapeOpts), inp("a", "Значение A (r / сторона / a)", "2"),
			inp("b", "Значение B (если нужно: h / вторая сторона)", "5")},
		func(v map[string]string) []Row {
			a := numOr(v["a"], 0)
			b := numOr(v["b"], 0)
			rr, ok := shapeCalc(v["k"], a, b)
			if !ok {
				return nil
			}
			rows := []Row{rb(rr.name, f3d(rr.value)+" "+rr.unit)}
			switch v["k"] {
			case "circleR":
				rows = append(rows, r("Длина окружности (2πr)", fmtNum(2*math.Pi*a, 3)))
			case "circleD":
				rows = append(rows, r("Длина окружности (πd)", fmtNum(math.Pi*a, 3)))
			case "rect":
				rows = append(rows, r("Периметр (2(a+b))", fmtNum(2*(a+b), 3)))
			case "cyl":
				if b > 0 {
					rows = append(rows, r("Боковая площадь (2πrh)", fmtNum(2*math.Pi*a*b, 3)))
				}
			}
			return rows
		}})

	add(Calc{"avg", "math", "🧮", "Среднее из списка чисел", "медиана сумма минимум максимум статистика", "",
		[]Input{inp("list", "Числа (через пробел или запятую)", "120 95 110 130 98")},
		func(v map[string]string) []Row {
			s, ok := statsOf(parseNumbers(v["list"]))
			if !ok {
				return nil
			}
			return []Row{r("Чисел", it(s.n)), r("Сумма", f2(s.sum)), rb("Среднее", f2(s.mean)),
				r("Медиана", f2(s.median)), r("Минимум", f2(s.min)), r("Максимум", f2(s.max)),
				r("Размах (max − min)", f2(s.max-s.min))}
		}})

	add(Calc{"vts", "math", "🚗", "Скорость · время · путь", "расстояние часы км в час маршрут", "Заполните любые два поля из трёх — третье посчитается.",
		[]Input{inp("s", "Скорость (км/ч)", "60"), inp("t", "Время (ч)", "2"), inp("d", "Путь (км) — оставьте пустым то, что ищем", "")},
		func(v map[string]string) []Row {
			s, hs := numStr(v["s"])
			t, ht := numStr(v["t"])
			d, hd := numStr(v["d"])
			rr, ok := vts(s, t, d, hs, ht, hd)
			if !ok {
				return nil
			}
			switch rr.mode {
			case 0:
				return []Row{rb("Путь", f2(rr.dist)+" км")}
			case 1:
				return []Row{rb("Время", f2(rr.time)+" ч"), r("Минут", fmtNum(rr.time*60, 0))}
			default:
				return []Row{rb("Скорость", f2(rr.speed)+" км/ч"), r("В м/с", f2(rr.speed/3.6))}
			}
		}})

	add(Calc{"dates", "math", "📅", "Между датами", "дней сколько разницы дата возраст", "",
		[]Input{inpDate("a", "Дата 1"), inpDate("b", "Дата 2")},
		func(v map[string]string) []Row {
			a := v["a"]
			b := v["b"]
			if a == "" {
				a = todayYMD()
			}
			if b == "" {
				b = todayYMD()
			}
			rr, ok := dateDiff(a, b)
			if !ok {
				return nil
			}
			days := float64(rr.days)
			return []Row{rb("Дней", it(rr.days)), r("Часов", fmtNum(days*24, 0)), r("Недель", f1(days/7)),
				r("≈ месяцев", f1(days/30.44)), r("≈ лет", f2(days/365.25)), r("Будних дней (пн–пт, с датами)", it(rr.workdays))}
		}})

	add(Calc{"units", "math", "🔄", "Конвертер единиц", "перевод длина масса объём см м кг грамм литры дюймы футы мили галлон",
		"Например: 1 м = 100 см, 1 кг = 1000 г, 1 л = 1000 мл.", nil, nil})

	// ================= ЗДОРОВЬЕ =================
	add(Calc{"bmi", "health", "⚖️", "ИМТ (индекс массы тела)", "имт вес рост категория воз",
		"По классификации ВОЗ. ИМТ не видит распределение мышц/жиров — это ориентир, не диагноз.",
		[]Input{inp("h", "Рост, см", "176"), inp("w", "Вес, кг", "80")},
		func(v map[string]string) []Row {
			h := numOr(v["h"], 0)
			rr, ok := bmiOf(h, numOr(v["w"], 0))
			if !ok {
				return nil
			}
			hm := h / 100
			return []Row{rb("ИМТ", f1(rr.v)), r("Категория (ВОЗ)", rr.cat),
				r("Вес при ИМТ 18.5–24.9", f1(18.5*hm*hm)+" – "+f1(24.9*hm*hm)+" кг")}
		}})

	add(Calc{"tdee", "health", "🔥", "Калории в день", "тdee bmr метаболизм похудеть набрать миффлин",
		"Формула Миффлина–Сан Жеора. −20% — плавное похудение, +10% — набор.",
		[]Input{inpSel("sex", "Пол", "m", []Opt{{"m", "Мужской"}, {"f", "Женский"}}),
			inp("age", "Возраст, лет", "30"), inp("h", "Рост, см", "176"), inp("w", "Вес, кг", "80"),
			inpSel("act", "Активность", "1.375", []Opt{{"1.2", "Диван (сидячая работа)"}, {"1.375", "Лёгкая (1–3 тренировки/нед)"},
				{"1.55", "Средняя (3–5 тренировок/нед)"}, {"1.72", "Высокая (6–7 тренировок/нед)"}, {"1.9", "Спортсмен / физ. работа"}})},
		func(v map[string]string) []Row {
			w := numOr(v["w"], 0)
			rr, ok := tdee(v["sex"], numOr(v["age"], 0), numOr(v["h"], 0), w, numOr(v["act"], 1.2))
			if !ok {
				return nil
			}
			return []Row{r("Обмен веществ (BMR)", fmtNum(rr.bmr, 0)+" ккал"), rb("Норма (TDEE)", fmtNum(rr.tdee, 0)+" ккал"),
				r("Похудение (−20%)", fmtNum(rr.cut, 0)+" ккал"), r("Набор (+10%)", fmtNum(rr.gain, 0)+" ккал"),
				r("Белок, ориентир (1,6 г/кг)", fmtNum(w*1.6, 0)+" г/день")}
		}})

	add(Calc{"macros", "health", "🥩", "БЖУ из калорий", "белки жиры углеводы граммы соотношение",
		"Белок и углеводы — 4 ккал/г, жиры — 9 ккал/г. Доли в сумме лучше = 100%.",
		[]Input{inp("k", "Калорий в день", "2200"), inp("p", "Белки, %", "30"), inp("f", "Жиры, %", "20"),
			inp("c", "Углеводы, %", "50")},
		func(v map[string]string) []Row {
			rr, ok := macros(numOr(v["k"], 0), numOr(v["p"], 0), numOr(v["f"], 0), numOr(v["c"], 0))
			if !ok {
				return nil
			}
			return []Row{rb("Белки", f1(rr.p)+" г"), r("Жиры", f1(rr.f)+" г"), r("Углеводы", f1(rr.c)+" г"),
				r("На один приём (3 приёма)", f1(rr.p/3)+" / "+f1(rr.f/3)+" / "+f1(rr.c/3)+" г")}
		}})

	add(Calc{"water", "health", "💧", "Норма воды", "литры жидкость пить активность", "",
		[]Input{inp("w", "Вес, кг", "75"), inpSel("a", "Есть тренировки (35 мл/кг вместо 30)", "0", []Opt{{"0", "Нет"}, {"1", "Да"}})},
		func(v map[string]string) []Row {
			lit, ok := waterLiters(numOr(v["w"], 0), v["a"] == "1")
			if !ok {
				return nil
			}
			rows := []Row{rb("Воды в день", f2(lit)+" л"), r("Стаканов по 0.3 л", f1(lit/0.3))}
			if v["a"] != "1" {
				rows = append(rows, r("С тренировками (+500 мл)", f2(lit+0.5)+" л"))
			}
			return rows
		}})

	add(Calc{"oner", "health", "🏋️", "Рабочий 1ПМ (Эйпли)", "один повтор максимум силовой тренажёр",
		"1ПМ = вес × (1 + повторы/30). Формула Эйпли — оценка, не мера точности.",
		[]Input{inp("w", "Вес, кг", "100"), inp("r", "Сколько раз сделали", "5")},
		func(v map[string]string) []Row {
			x, ok := oneRm(numOr(v["w"], 0), numOr(v["r"], 0))
			if !ok {
				return nil
			}
			return []Row{rb("Примерный 1ПМ", f1(x)+" кг"), r("10 повторов ≈ 70% 1ПМ", f1(x*0.7)+" кг"),
				r("6 повторов ≈ 80% 1ПМ", f1(x*0.8)+" кг"), r("3 повтора ≈ 90% 1ПМ", f1(x*0.9)+" кг")}
		}})

	var metOpts = []Opt{{"2.5", "Ходьба 5 км/ч"}, {"4", "Быстрая ходьба 6,5 км/ч"}, {"6", "Бег 8 км/ч"},
		{"7", "Плавание"}, {"8", "Велосипед 16 км/ч"}, {"10", "Тяжёлая тренировка / уборка"}}
	add(Calc{"metcal", "health", "🏃", "Калории за активность", "мет меты сожечь ходьба бег тренировка",
		"Ккал = MET × вес × время. MET — коэффициент метаболического эквивалента.",
		[]Input{inp("w", "Вес, кг", "75"), inpSel("m", "Активность", "2.5", metOpts), inp("h", "Часов", "0.5")},
		func(v map[string]string) []Row {
			kcal, ok := metCalories(numOr(v["w"], 0), numOr(v["m"], 0), numOr(v["h"], 0))
			if !ok {
				return nil
			}
			perHour := kcal / math.Max(numOr(v["h"], 0), 1)
			rows := []Row{rb("Сожжено", fmtNum(kcal, 0)+" ккал")}
			if perHour > 0 {
				rows = append(rows, r("Чтобы сжечь 1 кг жира (7 700 ккал)", f1(7700/perHour)+" ч такой активности"))
			}
			return rows
		}})

	add(Calc{"due", "health", "🤰", "Роды: предполагаемая дата", "беременность срок гестация последний день",
		"Формула Нагеле: +280 дней (40 недель от последней менструации). При цикле ≠28 дней дата сдвигается.",
		[]Input{inpDate("l", "Первый день последней менструации"), inp("c", "Длина цикла, дней (по умолчанию 28)", "")},
		func(v map[string]string) []Row {
			l := v["l"]
			if l == "" {
				l = todayYMD()
			}
			cl, hc := numStr(v["c"])
			due, base, shift, ok := dueDate(l, cl, hc)
			if !ok {
				return nil
			}
			rows := []Row{rb("Ожидаемая дата", fmtDateRu(due))}
			if shift != 0 {
				sg := "+"
				if shift < 0 {
					sg = ""
				}
				rows = append(rows, r("Нагеле (цикл 28 дн)", fmtDateRu(base)+" ("+sg+it(shift)+" дн.)"))
			}
			if w, _, lab, ok2 := pregnancyWeeks(l, ""); ok2 {
				_ = w
				rows = append(rows, rb("Срок сейчас", lab))
			}
			// до даты
			dDue, _ := parseYMD(due)
			nowT := time.Now()
			today := time.Date(nowT.Year(), nowT.Month(), nowT.Day(), 0, 0, 0, 0, time.UTC)
			left := int(dDue.Sub(today).Hours() / 24)
			switch {
			case left > 0:
				rows = append(rows, r("До предполагаемой даты", it(left)+" дн."))
			case left == 0:
				rows = append(rows, r("Сегодня", "предполагаемая дата 🎉"))
			default:
				rows = append(rows, r("Срок перешагнут", "на "+it(-left)+" дн."))
			}
			return rows
		}})

	add(Calc{"pace", "health", "🏃", "Темп бега: мин/км ↔ км/ч", "бег марафон скорость время дистанция тренировка",
		"Заполните скорость ИЛИ темп (темп приоритетнее). Темп 6:00 мин/км ≈ 10 км/ч.",
		[]Input{inp("spd", "Скорость, км/ч", ""), inp("pmin", "Темп: минут на км", "6"), inp("psec", "Темп: секунд на км", "0"),
			inp("dist", "Своя дистанция, км (для времени)", "")},
		func(v map[string]string) []Row {
			kmh, hk := numStr(v["spd"])
			pmin, hp := numStr(v["pmin"])
			psec, _ := numStr(v["psec"])
			rr, ok := paceCalc(kmh, pmin, psec, hk, hp)
			if !ok {
				return nil
			}
			tm := it(rr.paceM) + ":"
			if rr.paceS < 10 {
				tm += "0"
			}
			tm += it(rr.paceS)
			rows := []Row{rb("Темп", tm+" мин/км"), r("Скорость", f1(rr.spd)+" км/ч")}
			if d, h := numStr(v["dist"]); h && d > 0 {
				rows = append(rows, r("Время на "+f2(d)+" км", fmtDur(rr.paceMin*d)))
			}
			rows = append(rows, r("Время на 5 км", fmtDur(rr.paceMin*5)), r("Время на 10 км", fmtDur(rr.paceMin*10)),
				r("Время на 21,1 км", fmtDur(rr.paceMin*21.0975)), r("Время на 42,2 км", fmtDur(rr.paceMin*42.195)))
			return rows
		}})

	add(Calc{"hrzones", "health", "💓", "Пульсовые зоны", "сердце пульс тренировка зоны чсс кардио бег возраст интенсивность",
		"Пусто в поле пульса = 220 − возраст. Зоны считаются как % от максимального пульса.",
		[]Input{inp("age", "Возраст, лет", "30"), inp("mx", "Максимальный пульс, уд/мин", "")},
		func(v map[string]string) []Row {
			age := numOr(v["age"], 0)
			mx, hm := numStr(v["mx"])
			rr := hrZones(age, mx, hm)
			rows := []Row{rb("Максимальный пульс", it(rr.max)+" уд/мин")}
			for _, z := range rr.zones {
				rows = append(rows, r(z.pct+" — "+z.name, it(z.lo)+"–"+it(z.hi)+" уд/мин"))
			}
			return rows
		}})
}

func pos1(x float64) float64 {
	if x == 0 {
		return 1
	}
	return x
}

func f3d(x float64) string { return fmtNum(x, 3) }
