package main

// CLI калькуляторов (Go-порт). Запуск:
//   go run .                  интерактивное меню
//   go run . list             список
//   go run . run loan p=2000000 r=21 m=48
//   go run . run dates a=2026-01-05 b=2026-01-31

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

var catNames = map[string]string{
	"money": "💰 Деньги", "home": "🏠 Дом", "food": "🍽️ Еда",
	"math": "📐 Математика", "health": "❤️ Здоровье", "travel": "🚗 Поездки",
}

func printRows(rows []Row) {
	if rows == nil {
		fmt.Println("  (нет результата — проверьте ввод)")
		return
	}
	for _, rw := range rows {
		star := " "
		if rw.Big {
			star = "★"
		}
		fmt.Printf(" %s %-42s %s\n", star, rw.Label, rw.Value)
	}
}

func findCalc(id string) *Calc {
	for i := range CALCS {
		if CALCS[i].ID == id {
			return &CALCS[i]
		}
	}
	return nil
}

func resolveMap(c *Calc, kv map[string]string) map[string]string {
	v := map[string]string{}
	for _, in := range c.Inputs {
		if in.Type == "date" && in.Defv == "" {
			v[in.K] = todayYMD()
		} else if in.Defv != "" {
			v[in.K] = in.Defv
		} else {
			v[in.K] = ""
		}
	}
	for k, val := range kv {
		v[k] = val
	}
	return v
}

// ---------- спец-режимы ----------
func runRecipe(kv map[string]string) {
	from := numOr(kv["from"], 4)
	to := numOr(kv["to"], 8)
	var items []ing
	keys := []string{}
	for k := range kv {
		if strings.HasPrefix(k, "i") && len(k) > 1 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		parts := strings.Split(kv[k], ";")
		if len(parts) < 2 {
			continue
		}
		am, ok := numStr(strings.TrimSpace(parts[1]))
		if !ok || am <= 0 {
			continue
		}
		unit := ""
		if len(parts) > 2 {
			unit = strings.TrimSpace(parts[2])
		}
		items = append(items, ing{strings.TrimSpace(parts[0]), am, unit})
	}
	k, scaled, ok := recipeScale(from, to, items)
	if !ok || len(scaled) == 0 {
		fmt.Println("  Укажите порции и хотя бы один ингредиент с количеством.")
		return
	}
	fmt.Printf("  Пересчёт рецепта (%.4g -> %.4g порций):\n", from, to)
	printRows([]Row{{"Коэффициент", "×" + fmtNum(k, 2), true}})
	for _, it := range scaled {
		unit := it.unit
		printRows([]Row{{it.name + " · " + fmtQty(it.amount) + " " + unit, fmtQty(it.amount*k) + " " + unit, false}})
	}
	if to > 0 {
		for _, it := range scaled {
			per := it.amount * k / to
			printRows([]Row{{"На порцию: " + it.name, fmtQty(per) + " " + it.unit, false}})
		}
	}
}

func fmtQty(x float64) string {
	if x == 0 {
		return "0"
	}
	if math.Trunc(x) == x {
		return it(int(x))
	}
	return strings.TrimRight(strconv.FormatFloat(x, 'f', 1, 64), "0")
}

func runUnits(kv map[string]string) {
	kind := kv["kind"]
	if kind == "" {
		kind = "length"
	}
	from := kv["from"]
	if from == "" {
		from = "м"
	}
	to := kv["to"]
	if to == "" {
		to = "см"
	}
	val := kv["value"]
	if val == "" {
		fmt.Println("  Укажите value=число.")
		return
	}
	n, ok := numStr(val)
	if !ok {
		fmt.Println("  Не число.")
		return
	}
	res, ok2 := unitConvert(kind, from, to, n)
	if !ok2 {
		fmt.Println("  Неизвестные единицы.")
		return
	}
	printRows([]Row{{fmtNum(n, 6) + " " + from + " =", fmtNum(res, 6) + " " + to, true}})
}

func runNoninteractive(id string, kv map[string]string) {
	c := findCalc(id)
	if c == nil {
		fmt.Println("Калькулятор не найден: " + id)
		os.Exit(1)
	}
	fmt.Printf("== %s %s ==\n", c.Emoji, c.Title)
	if id == "recipe" {
		runRecipe(kv)
		return
	}
	if id == "units" {
		runUnits(kv)
		return
	}
	v := resolveMap(c, kv)
	printRows(c.Run(v))
}

// ---------- интерактив ----------
func ask(reader *bufio.Reader, in Input) string {
	if in.Type == "select" {
		fmt.Println("  Выберите: " + in.L)
		for i, o := range in.Opts {
			fmt.Printf("    %d. %s\n", i+1, o.L)
		}
		fmt.Printf("    номер [%s]: ", in.Defv)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			return in.Defv
		}
		n, err := strconv.Atoi(line)
		if err != nil || n < 1 || n > len(in.Opts) {
			return in.Defv
		}
		return in.Opts[n-1].V
	}
	def := in.Defv
	if in.Type == "date" && def == "" {
		def = todayYMD()
	}
	prompt := "  " + in.L + ":"
	if def != "" {
		prompt += " [" + def + "]"
	}
	prompt += " "
	fmt.Print(prompt)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

func runInteractive(reader *bufio.Reader, id string) {
	c := findCalc(id)
	if c == nil {
		return
	}
	fmt.Printf("== %s %s ==\n", c.Emoji, c.Title)
	if c.Hint != "" {
		fmt.Println("  ⓘ " + c.Hint)
	}
	if id == "recipe" {
		fmt.Println("  Ингредиенты (пустое название — конец):")
		from := numOr(askLine(reader, "  Порций было [4]"), 4)
		to := numOr(askLine(reader, "  Порций нужно [8]"), 8)
		var items []ing
		for {
			name := askLine(reader, "    Название: ")
			if name == "" {
				break
			}
			am := numOr(askLine(reader, "    Кол-во: "), 0)
			unit := askLine(reader, "    Единица (г/мл): ")
			if am > 0 {
				items = append(items, ing{name, am, unit})
			}
		}
		runRecipeKV(from, to, items)
		return
	}
	if id == "units" {
		kind := askSelectUnit(reader)
		runUnitsInteractive(reader, kind)
		return
	}
	v := map[string]string{}
	for _, in := range c.Inputs {
		v[in.K] = ask(reader, in)
	}
	printRows(c.Run(v))
}

func askLine(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	s, _ := reader.ReadString('\n')
	return strings.TrimSpace(s)
}

func runRecipeKV(from, to float64, items []ing) {
	k, scaled, ok := recipeScale(from, to, items)
	if !ok || len(scaled) == 0 {
		fmt.Println("  Укажите порции и хотя бы один ингредиент.")
		return
	}
	fmt.Printf("  Коэффициент ×%s\n", fmtNum(k, 2))
	for _, it := range scaled {
		fmt.Printf("    %s: %s %s\n", it.name, fmtQty(it.amount), it.unit)
	}
	if to > 0 {
		fmt.Println("  На одну порцию:")
		for _, it := range scaled {
			fmt.Printf("    %s: %s %s\n", it.name, fmtQty(it.amount*k/to), it.unit)
		}
	}
}

func askSelectUnit(reader *bufio.Reader) string {
	kinds := []string{"length", "mass", "volume"}
	names := map[string]string{"length": "Длина", "mass": "Масса", "volume": "Объём"}
	fmt.Println("  Величина:")
	for i, k := range kinds {
		fmt.Printf("    %d. %s\n", i+1, names[k])
	}
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if n, err := strconv.Atoi(line); err == nil && n >= 1 && n <= len(kinds) {
		return kinds[n-1]
	}
	return "length"
}

func runUnitsInteractive(reader *bufio.Reader, kind string) {
	units := []string{}
	for u := range unitKinds[kind] {
		units = append(units, u)
	}
	sort.Strings(units)
	fmt.Println("  Единицы: " + strings.Join(units, ", "))
	from := askLine(reader, "  Из единицы [м]: ")
	if from == "" {
		from = "м"
	}
	to := askLine(reader, "  В единицу [см]: ")
	if to == "" {
		to = "см"
	}
	val := askLine(reader, "  Значение: ")
	runUnits(map[string]string{"kind": kind, "from": from, "to": to, "value": val})
}

func cmdList() {
	prev := ""
	for i := range CALCS {
		c := CALCS[i]
		cn := catNames[c.Cat]
		if cn == "" {
			cn = c.Cat
		}
		if cn != prev {
			fmt.Printf("\n%s\n", cn)
			prev = cn
		}
		fmt.Printf("  %-10s %s %s\n", c.ID, c.Emoji, c.Title)
	}
	fmt.Printf("\nВсего: %d калькуляторов.\n", len(CALCS))
}

func main() {
	args := os.Args[1:]
	if len(args) >= 1 && args[0] == "list" {
		cmdList()
		return
	}
	if len(args) >= 1 && (args[0] == "run" || args[0] == "calc") {
		if len(args) < 2 {
			fmt.Println("run <id> [k=v ...]")
			os.Exit(1)
		}
		id := args[1]
		kv := map[string]string{}
		for _, a := range args[2:] {
			if i := strings.Index(a, "="); i > 0 {
				kv[a[:i]] = a[i+1:]
			}
		}
		runNoninteractive(id, kv)
		return
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("🧮 Калькуляторы на все случаи жизни — Go-порт v1.3")
	fmt.Println("Введите id калькулятора (см. `list`), 'list' — список, 'q' — выход.")
	cmdList()
	for {
		fmt.Print("\n> ")
		line, _ := reader.ReadString('\n')
		line = strings.ToLower(strings.TrimSpace(line))
		switch line {
		case "q", "quit", "exit":
			return
		case "list", "l":
			cmdList()
			continue
		}
		if line == "" {
			continue
		}
		runInteractive(reader, line)
	}
}
