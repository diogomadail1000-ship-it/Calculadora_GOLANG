# 🧮 Calculadora em Go (REPL)

Uma **calculadora de linha de comando** escrita em **Golang**, com suporte a expressões matemáticas completas,
funções, constantes e uma variável `ans` para guardar o último resultado.

---

## 🚀 Funcionalidades

✅ Operadores aritméticos: `+`, `-`, `*`, `/`, `^`  
✅ Suporte a **parênteses** e **precedência de operadores**  
✅ Funções matemáticas:
```
sin, cos, tan, sqrt, log (base 10), ln, abs, floor, ceil, round, max(a,b), min(a,b)
```
✅ Constantes matemáticas:
```
pi, e
```
✅ Variável especial:
```
ans → guarda o último resultado
```
✅ Comandos interativos:
```
:help   → mostra ajuda
:const  → lista constantes
:func   → lista funções
:quit   → sai da calculadora
```

---

## 🧩 Exemplo de utilização

```bash
$ go run calculadora.go
Calculadora em Go — REPL (:help para ajuda)
> 2+2*3
= 8
> (1+2)^3/9
= 3
> sin(pi/2)
= 1
> sqrt(2)
= 1.4142135623731
> max(3, 9)
= 9
> 1+ans
= 10
```

---

## 🛠️ Como compilar e executar

```bash
# Clonar o repositório
git clone https://github.com/teu-usuario/calculadora-go.git
cd calculadora-go

# Executar diretamente
go run calculadora.go

# Ou compilar e executar
go build -o calc calculadora.go
./calc
```

---

## 📂 Estrutura do projeto

```
calculadora-go/
├── calculadora.go   # Código principal da calculadora
└── README.md        # Este ficheiro
```

---

## 🧠 Licença

Este projeto é distribuído sob a licença **MIT**.  
Sinta-se à vontade para modificar e partilhar!
