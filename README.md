# 1C BSL Parser

🇬🇧 [English](#english) | 🇷🇺 [Русский](#русский)

---

## English

A fast and accurate parser for **1C:Enterprise (BSL)** programming language, written in Go. Generates Abstract Syntax Tree (AST) for code analysis, transformation, and tooling.

> **Based on** [LazarenkoA/1c-language-parser](https://github.com/LazarenkoA/1c-language-parser) — extended with preprocessor support, PostgreSQL indexer, and web interface.

### Features

- **Full BSL syntax support** — procedures, functions, control flow, expressions
- **Preprocessor directives** — `#If`, `#ElseIf`, `#Else`, `#EndIf`, `#Region`
- **Compiler directives** — `&AtServer`, `&AtClient`, etc.
- **JSON output** — AST serialization for tooling integration
- **Pretty printer** — regenerate formatted code from AST

### Installation

#### As a library

```bash
go get github.com/vitaly-andr/1c-bsl-parser
```

#### CLI tools

```bash
# Parser CLI
go install github.com/vitaly-andr/1c-bsl-parser/cmd/bsl-ast@latest

# PostgreSQL indexer
go install github.com/vitaly-andr/1c-bsl-parser/cmd/bsl-index@latest
```

### Quick Start

#### Library Usage

```go
package main

import (
    "fmt"
    "github.com/vitaly-andr/1c-bsl-parser/ast"
)

func main() {
    code := `
    Procedure ProcessOrder(Order) Export
        If Order.Status = "New" Then
            Order.Process();
        EndIf;
    EndProcedure
    `

    parser := ast.NewAST(code)
    if err := parser.Parse(); err != nil {
        panic(err)
    }

    // Get JSON AST
    jsonData, _ := parser.JSON()
    fmt.Println(string(jsonData))

    // Regenerate code
    fmt.Println(parser.Print())
}
```

#### CLI Usage

```bash
# Parse from stdin
echo 'Function Test() Return 1; EndFunction' | bsl-ast

# Parse file
bsl-ast < module.bsl

# Index to PostgreSQL
bsl-index --source /path/to/1c/config --config myconfig
```

### AST Example

Input:
```bsl
#Region Public

Procedure Hello() Export
    Message("Hello, World!");
EndProcedure

#EndRegion
```

Output (simplified):
```json
{
  "Body": [
    {
      "Type": "RegionStatement",
      "Name": "Public",
      "Body": [
        {
          "Type": "FunctionOrProcedure",
          "Name": "Hello",
          "ProcType": 1,
          "Export": true,
          "Body": [
            {
              "Type": "MethodStatement",
              "Name": "Message",
              "Params": ["Hello, World!"]
            }
          ]
        }
      ]
    }
  ]
}
```

### Project Structure

```
├── ast/                    # Parser package
│   ├── grammar.y           # Yacc grammar definition
│   ├── tokens.go           # Lexer implementation
│   ├── ast.go              # Main parser API
│   ├── ast_struct.go       # AST node types
│   └── ast_print.go        # Code generator
└── examples/               # Usage examples
```

### Development

```bash
# Clone
git clone https://github.com/vitaly-andr/1c-bsl-parser.git
cd 1c-bsl-parser

# Install dependencies
go mod download

# Regenerate parser (after grammar.y changes)
go generate ./ast/...

# Run tests
go test -v ./...

# Build CLI tools
go build -o bin/bsl-ast ./cmd/bsl-ast
go build -o bin/bsl-index ./cmd/bsl-index
```

### Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

### Credits

- Original parser by [LazarenkoA](https://github.com/LazarenkoA/1c-language-parser)
- Grammar based on [1C:Enterprise documentation](https://its.1c.ru/)

---

## Русский

Быстрый и точный парсер языка **1С:Предприятие (BSL)**, написанный на Go. Генерирует абстрактное синтаксическое дерево (AST) для анализа, трансформации и создания инструментов.

> **Основан на** [LazarenkoA/1c-language-parser](https://github.com/LazarenkoA/1c-language-parser) — расширен поддержкой препроцессора, индексатором PostgreSQL и веб-интерфейсом.

### Возможности

- **Полная поддержка синтаксиса BSL** — процедуры, функции, управляющие конструкции, выражения
- **Директивы препроцессора** — `#Если`, `#ИначеЕсли`, `#Иначе`, `#КонецЕсли`, `#Область`
- **Директивы компиляции** — `&НаСервере`, `&НаКлиенте` и др.
- **JSON вывод** — сериализация AST для интеграции
- **Форматирование кода** — регенерация отформатированного кода из AST

### Установка

#### Как библиотека

```bash
go get github.com/vitaly-andr/1c-bsl-parser
```

#### CLI инструменты

```bash
# Парсер
go install github.com/vitaly-andr/1c-bsl-parser/cmd/bsl-ast@latest

# Индексатор PostgreSQL
go install github.com/vitaly-andr/1c-bsl-parser/cmd/bsl-index@latest
```

### Быстрый старт

#### Использование как библиотеки

```go
package main

import (
    "fmt"
    "github.com/vitaly-andr/1c-bsl-parser/ast"
)

func main() {
    code := `
    Процедура ОбработатьЗаказ(Заказ) Экспорт
        Если Заказ.Статус = "Новый" Тогда
            Заказ.Обработать();
        КонецЕсли;
    КонецПроцедуры
    `

    parser := ast.NewAST(code)
    if err := parser.Parse(); err != nil {
        panic(err)
    }

    // Получить JSON AST
    jsonData, _ := parser.JSON()
    fmt.Println(string(jsonData))

    // Сгенерировать код обратно
    fmt.Println(parser.Print())
}
```

#### Использование CLI

```bash
# Парсинг из stdin
echo 'Функция Тест() Возврат 1; КонецФункции' | bsl-ast

# Парсинг файла
bsl-ast < module.bsl

# Индексация в PostgreSQL
bsl-index --source /path/to/1c/config --config myconfig
```

### Пример AST

Входной код:
```bsl
#Область ПрограммныйИнтерфейс

Процедура Привет() Экспорт
    Сообщить("Привет, мир!");
КонецПроцедуры

#КонецОбласти
```

Результат (упрощённо):
```json
{
  "Body": [
    {
      "Type": "RegionStatement",
      "Name": "ПрограммныйИнтерфейс",
      "Body": [
        {
          "Type": "FunctionOrProcedure",
          "Name": "Привет",
          "ProcType": 1,
          "Export": true,
          "Body": [
            {
              "Type": "MethodStatement",
              "Name": "Сообщить",
              "Params": ["Привет, мир!"]
            }
          ]
        }
      ]
    }
  ]
}
```

### Структура проекта

```
├── ast/                    # Пакет парсера
│   ├── grammar.y           # Yacc грамматика
│   ├── tokens.go           # Лексер
│   ├── ast.go              # Основной API
│   ├── ast_struct.go       # Типы узлов AST
│   └── ast_print.go        # Генератор кода
└── examples/               # Примеры использования
```

### Разработка

```bash
# Клонирование
git clone https://github.com/vitaly-andr/1c-bsl-parser.git
cd 1c-bsl-parser

# Установка зависимостей
go mod download

# Перегенерация парсера (после изменения grammar.y)
go generate ./ast/...

# Запуск тестов
go test -v ./...

# Сборка CLI
go build -o bin/bsl-ast ./cmd/bsl-ast
go build -o bin/bsl-index ./cmd/bsl-index
```

### Участие в разработке

Приветствуются pull request'ы!

1. Сделайте форк репозитория
2. Создайте ветку для фичи (`git checkout -b feature/amazing-feature`)
3. Закоммитьте изменения (`git commit -m 'feat: add amazing feature'`)
4. Запушьте ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

### Лицензия

Проект распространяется под лицензией MIT — см. файл [LICENSE](LICENSE).

### Благодарности

- Оригинальный парсер: [LazarenkoA](https://github.com/LazarenkoA/1c-language-parser)
- Грамматика основана на [документации 1С:Предприятие](https://its.1c.ru/)

