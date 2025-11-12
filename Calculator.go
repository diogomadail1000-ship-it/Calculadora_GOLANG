// calculadora.go
// Uma calculadora de linha de comando em Go com REPL,
// suporte a + - * / ^, parênteses, funções e constantes.
// Funções: sin, cos, tan, sqrt, log (base 10), ln, abs, floor, ceil, round, max, min
// Constantes: pi, e
// Variável especial: ans (resultado anterior)
package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type tokenType int

const (
	tNumber tokenType = iota
	tOp
	tLParen
	tRParen
	tFunc
	tComma
	tIdent
)

type token struct {
	typ tokenType
	val string
}

var ops = map[string]struct {
	prec       int
	rightAssoc bool
	unary      bool
	fn         func(a, b float64) float64
}{
	"+":  {prec: 1, rightAssoc: false, unary: false, fn: func(a, b float64) float64 { return a + b }},
	"-":  {prec: 1, rightAssoc: false, unary: false, fn: func(a, b float64) float64 { return a - b }},
	"*":  {prec: 2, rightAssoc: false, unary: false, fn: func(a, b float64) float64 { return a * b }},
	"/":  {prec: 2, rightAssoc: false, unary: false, fn: func(a, b float64) float64 { return a / b }},
	"^":  {prec: 3, rightAssoc: true, unary: false, fn: func(a, b float64) float64 { return math.Pow(a, b) }},
	"u-": {prec: 4, rightAssoc: true, unary: true, fn: func(_, b float64) float64 { return -b }}, // unário menos
	"u+": {prec: 4, rightAssoc: true, unary: true, fn: func(_, b float64) float64 { return +b }},
}

var functions = map[string]func(args ...float64) (float64, error){
	"sin": func(a ...float64) (float64, error) { return math.Sin(a[0]), nil },
	"cos": func(a ...float64) (float64, error) { return math.Cos(a[0]), nil },
	"tan": func(a ...float64) (float64, error) { return math.Tan(a[0]), nil },
	"sqrt": func(a ...float64) (float64, error) {
		if a[0] < 0 {
			return 0, errors.New("sqrt de número negativo")
		}
		return math.Sqrt(a[0]), nil
	},
	"log":   func(a ...float64) (float64, error) { return math.Log10(a[0]), nil },
	"ln":    func(a ...float64) (float64, error) { return math.Log(a[0]), nil },
	"abs":   func(a ...float64) (float64, error) { return math.Abs(a[0]), nil },
	"floor": func(a ...float64) (float64, error) { return math.Floor(a[0]), nil },
	"ceil":  func(a ...float64) (float64, error) { return math.Ceil(a[0]), nil },
	"round": func(a ...float64) (float64, error) { return math.Round(a[0]), nil },
	"max": func(a ...float64) (float64, error) {
		if len(a) < 2 {
			return 0, errors.New("max precisa de 2 argumentos")
		}
		if a[0] > a[1] {
			return a[0], nil
		}
		return a[1], nil
	},
	"min": func(a ...float64) (float64, error) {
		if len(a) < 2 {
			return 0, errors.New("min precisa de 2 argumentos")
		}
		if a[0] < a[1] {
			return a[0], nil
		}
		return a[1], nil
	},
}

var constants = map[string]float64{
	"pi": math.Pi,
	"e":  math.E,
}

func isIdentStart(r rune) bool { return unicode.IsLetter(r) || r == '_' }
func isIdent(r rune) bool      { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' }

func tokenize(input string) ([]token, error) {
	var toks []token
	s := strings.TrimSpace(input)
	i := 0
	prevType := tOp // como se começasse com operador
	for i < len(s) {
		ch := rune(s[i])
		if unicode.IsSpace(ch) {
			i++
			continue
		}
		if unicode.IsDigit(ch) || ch == '.' {
			j := i + 1
			hasE := false
			for j < len(s) {
				r := rune(s[j])
				if unicode.IsDigit(r) || r == '.' {
					j++
				} else if (r == 'e' || r == 'E') && !hasE {
					hasE = true
					j++
					if j < len(s) && (s[j] == '+' || s[j] == '-') {
						j++
					}
				} else {
					break
				}
			}
			toks = append(toks, token{tNumber, s[i:j]})
			prevType = tNumber
			i = j
			continue
		}
		switch ch {
		case '+', '-':
			op := string(ch)
			if prevType == tOp || prevType == tLParen || len(toks) == 0 {
				if op == "-" {
					op = "u-"
				} else {
					op = "u+"
				}
			}
			toks = append(toks, token{tOp, op})
			prevType = tOp
			i++
		case '*', '/', '^':
			toks = append(toks, token{tOp, string(ch)})
			prevType = tOp
			i++
		case '(':
			toks = append(toks, token{tLParen, "("})
			prevType = tLParen
			i++
		case ')':
			toks = append(toks, token{tRParen, ")"})
			prevType = tRParen
			i++
		case ',':
			toks = append(toks, token{tComma, ","})
			prevType = tComma
			i++
		default:
			if isIdentStart(ch) {
				j := i + 1
				for j < len(s) && isIdent(rune(s[j])) {
					j++
				}
				id := s[i:j]
				low := strings.ToLower(id)
				if _, ok := functions[low]; ok {
					toks = append(toks, token{tFunc, low})
				} else if _, ok := constants[low]; ok || low == "ans" {
					toks = append(toks, token{tIdent, low})
				} else {
					return nil, fmt.Errorf("identificador desconhecido: %s", id)
				}
				prevType = tIdent
				i = j
			} else {
				return nil, fmt.Errorf("caractere inválido: %q", ch)
			}
		}
	}
	return toks, nil
}

func shuntingYard(toks []token) ([]token, error) {
	var output []token
	var stack []token
	arity := map[string]int{
		"sin": 1, "cos": 1, "tan": 1, "sqrt": 1, "log": 1, "ln": 1,
		"abs": 1, "floor": 1, "ceil": 1, "round": 1, "max": 2, "min": 2,
	}
	for _, t := range toks {
		switch t.typ {
		case tNumber, tIdent:
			output = append(output, t)
		case tFunc:
			stack = append(stack, t)
		case tComma:
			for len(stack) > 0 && stack[len(stack)-1].typ != tLParen {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, errors.New("vírgula fora de função")
			}
		case tOp:
			for len(stack) > 0 && stack[len(stack)-1].typ == tOp {
				top := stack[len(stack)-1].val
				curr := t.val
				if (ops[top].prec > ops[curr].prec) ||
					(ops[top].prec == ops[curr].prec && !ops[curr].rightAssoc) {
					output = append(output, stack[len(stack)-1])
					stack = stack[:len(stack)-1]
				} else {
					break
				}
			}
			stack = append(stack, t)
		case tLParen:
			stack = append(stack, t)
		case tRParen:
			for len(stack) > 0 && stack[len(stack)-1].typ != tLParen {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, errors.New("parênteses desbalanceados")
			}
			stack = stack[:len(stack)-1]
			if len(stack) > 0 && stack[len(stack)-1].typ == tFunc {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
		}
	}
	for len(stack) > 0 {
		if stack[len(stack)-1].typ == tLParen {
			return nil, errors.New("parênteses desbalanceados")
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}
	for _, t := range output {
		if t.typ == tFunc {
			if _, ok := arity[t.val]; !ok {
				return nil, fmt.Errorf("função não suportada: %s", t.val)
			}
		}
	}
	return output, nil
}

func evalRPN(rpn []token, lastAns float64) (float64, error) {
	var st []float64
	for _, t := range rpn {
		switch t.typ {
		case tNumber:
			v, err := strconv.ParseFloat(t.val, 64)
			if err != nil {
				return 0, err
			}
			st = append(st, v)
		case tIdent:
			if t.val == "ans" {
				st = append(st, lastAns)
			} else if c, ok := constants[t.val]; ok {
				st = append(st, c)
			} else {
				return 0, fmt.Errorf("identificador desconhecido: %s", t.val)
			}
		case tOp:
			if ops[t.val].unary {
				if len(st) < 1 {
					return 0, errors.New("operador unário sem operando")
				}
				b := st[len(st)-1]
				st = st[:len(st)-1]
				res := ops[t.val].fn(0, b)
				st = append(st, res)
			} else {
				if len(st) < 2 {
					return 0, errors.New("operador binário com poucos operandos")
				}
				b := st[len(st)-1]
				a := st[len(st)-2]
				st = st[:len(st)-2]
				if t.val == "/" && b == 0 {
					return 0, errors.New("divisão por zero")
				}
				res := ops[t.val].fn(a, b)
				st = append(st, res)
			}
		case tFunc:
			var nargs int
			switch t.val {
			case "max", "min":
				nargs = 2
			default:
				nargs = 1
			}
			if len(st) < nargs {
				return 0, fmt.Errorf("função %s com poucos argumentos", t.val)
			}
			args := st[len(st)-nargs:]
			st = st[:len(st)-nargs]
			fn := functions[t.val]
			res, err := fn(args...)
			if err != nil {
				return 0, err
			}
			st = append(st, res)
		}
	}
	if len(st) != 1 {
		return 0, errors.New("expressão inválida")
	}
	return st[0], nil
}

func evalExpr(expr string, lastAns float64) (float64, error) {
	toks, err := tokenize(expr)
	if err != nil {
		return 0, err
	}
	rpn, err := shuntingYard(toks)
	if err != nil {
		return 0, err
	}
	return evalRPN(rpn, lastAns)
}

func printHelp() {
	fmt.Println("Calculadora Go — exemplos:")
	fmt.Println("  2+2*3")
	fmt.Println("  (1+2)^3/9")
	fmt.Println("  sqrt(2), log(100), ln(e), abs(-3.5)")
	fmt.Println("  sin(pi/2), cos(0), tan(pi/4)")
	fmt.Println("  max(3, 9), min(4, -2)")
	fmt.Println("  Use ans para o último resultado, ex.: 1+ans")
	fmt.Println("  Comandos: :quit para sair, :help para ajuda, :const para listar constantes, :func para listar funções")
}

func main() {
	fmt.Println("Calculadora em Go — REPL (:help para ajuda)")
	in := bufio.NewScanner(os.Stdin)
	lastAns := 0.0
	for {
		fmt.Print("> ")
		if !in.Scan() {
			break
		}
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, ":") {
			switch strings.ToLower(line) {
			case ":quit", ":q", ":exit":
				return
			case ":help", ":h":
				printHelp()
			case ":const":
				fmt.Println("Constantes:")
				for k, v := range constants {
					fmt.Printf("  %s = %.15g\n", k, v)
				}
			case ":func":
				fmt.Println("Funções: sin, cos, tan, sqrt, log, ln, abs, floor, ceil, round, max(a,b), min(a,b)")
			default:
				fmt.Println("Comando desconhecido. Use :help")
			}
			continue
		}
		res, err := evalExpr(line, lastAns)
		if err != nil {
			fmt.Println("Erro:", err)
			continue
		}
		lastAns = res
		fmt.Printf("= %.15g\n", res)
	}
}
