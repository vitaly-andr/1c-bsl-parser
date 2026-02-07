package ast

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/mock.go
type Iast interface {
	SrsCode() string
}

type Position struct {
	Line   int
	Column int
}

type Token struct {
	ast           Iast
	value         interface{}
	literal       string
	position      Position
	offset        int
	prevDot       bool
	prevTokenType int  // для определения начала statement (lookahead)
	inProcedure   bool // true после Procedure/Function, false после EndProcedure/EndFunction
	regionDepth   int  // tracks #Region nesting to skip orphan #EndRegion
}

const (
	EOL      = '\n' // end of line.
	emptyLit = ""

	// Buffer size hints for scanner operations
	identifierBufSize = 10 // typical identifier length for pre-allocation
)

var (
	tokens = map[string]int{
		"процедура":         Procedure,
		"перем":             Var,
		"перейти":           GoTo,
		"конецпроцедуры":    EndProcedure,
		"знач":              ValueParam,
		"если":              If,
		"тогда":             Then,
		"иначеесли":         ElseIf,
		"иначе":             Else,
		"конецесли":         EndIf,
		"для":               For,
		"каждого":           Each,
		"из":                In,
		"по":                To,
		"цикл":              Loop,
		"конеццикла":        EndLoop,
		"прервать":          Break,
		"продолжить":        Continue,
		"попытка":           Try,
		"новый":             New,
		"исключение":        Catch,
		"пока":              While,
		"конецпопытки":      EndTry,
		"функция":           Function,
		"конецфункции":      EndFunction,
		"возврат":           Return,
		"вызватьисключение": Throw,
		"и":                 And,
		"или":               OR,
		"истина":            True,
		"ложь":              False,
		"неопределено":      Undefined,
		"null":              Null,
		"не":                Not,
		"экспорт":           Export,
		"выполнить":         Execute,
		"while":             While,
		"do":                Loop,
		"enddo":             EndLoop,
		"to":                To,
		"асинх":             Async,
		"async":             Async,
		"ждать":             Await,
		"await":             Await,
		//"вычислить":         Eval,
		// "массив":            Array,
		// "структура":         Struct,
		// "соответствие":      Dictionary,
	}

	// общие директивы
	directives = map[string]int{
		"&наклиенте":                      Directive,
		"&насервере":                      Directive,
		"&насерверебезконтекста":          Directive,
		"&наклиентенасерверебезконтекста": Directive,
		"&наклиентенасервере":             Directive,
	}

	// директивы расширений
	extDirectives = map[string]int{
		"&перед":              ExtDirective,
		"&после":              ExtDirective,
		"&вместо":             ExtDirective,
		"&изменениеиконтроль": ExtDirective,
	}
)

// DebugLexer enables token logging when set to true.
// Use for debugging parser issues.
var DebugLexer = false

func (t *Token) Next(ast Iast) (token int, err error) {
	t.ast = ast
	token, t.literal, err = t.next()

	// Сохраняем тип токена для определения начала statement
	defer func() { t.prevTokenType = token }()

	// Track procedure context for body-level preprocessor tokens
	switch token {
	case Procedure, Function:
		t.inProcedure = true
	case EndProcedure, EndFunction:
		t.inProcedure = false
	}

	if DebugLexer {
		pos := t.GetPosition()
		fmt.Printf("[LEXER] line=%d col=%d token=%d lit=%q inProc=%v\n",
			pos.Line, pos.Column, token, t.literal, t.inProcedure)
	}

	// Emit different tokens for preprocessor and Var inside procedure bodies
	// This avoids reduce/reduce conflict between main and body rules
	if t.inProcedure {
		switch token {
		case PreprocIf:
			token = PreprocIfBody
		case PreprocElseIf:
			token = PreprocElseIfBody
		case PreprocElse:
			token = PreprocElseBody
		case PreprocEndIf:
			token = PreprocEndIfBody
		case PreprocRegion:
			token = PreprocRegionBody
		case PreprocEndRegion:
			token = PreprocEndRegionBody
		case Var:
			token = VarBody
		}
	}

	switch token {
	case Number:
		t.value, err = strconv.ParseFloat(t.literal, 64)
	case String:
		t.value = t.literal
	case Date:
		formats := []string{"20060102", "200601021504", "20060102150405"} // Допускается опускать либо время целиком, либо только секунды.
		for _, f := range formats {
			// если все 0 это равносильно пустой дате
			if strings.Count(t.literal, "0") == len(t.literal) {
				t.value = time.Time{}
				return
			}

			if t.value, err = time.Parse(f, t.literal); err == nil {
				break
			}
		}
	case Undefined:
		t.value = nil
	case True:
		t.value = true
	case False:
		t.value = false
	}

	return
}

func (t *Token) next() (int, string, error) {
	t.skipSpace()
	t.skipComment()

	// Handle preprocessor directives (#Если, #Область, #Использовать)
	if t.currentLet() == '#' {
		return t.handlePreprocessor()
	}

	if t.prevDot {
		defer func() { t.prevDot = false }()
	}

	switch let := t.currentLet(); {
	case isLetter(let):
		literal := t.scanIdentifier()
		lowLit := fastToLower(literal)

		if tName, ok := tokens[lowLit]; ok && !t.prevDot {
			return tName, literal, nil
		}

		// Для идентификаторов: определяем тип по контексту
		// Если в начале statement — используем lookahead для различения lvalue/call
		if t.atStatementStart() && !t.prevDot {
			if t.hasEqualBeforeSemicolon() {
				return LVALUE_IDENT, literal, nil
			}
			return CALL_IDENT, literal, nil
		}

		// Обычный идентификатор (в выражениях, после точки, и т.д.)
		return token_identifier, literal, nil
	case let == '.':
		// если после точки у нас следует идентификатор то нам нужно читать его обычным идентификатором
		// Могут быть таие случаи стр.Истина = 1 или стр.Функция = 2 (стр в данном случае какой-то объект, например структура)
		// нам нужно что бы то что следует после точки считалось Identifier, а не определенным зарезервированным токеном
		t.prevDot = true

		t.nextPos()
		return int(let), string(let), nil
	case isDigit(let):
		if literal, err := t.scanNumber(); err != nil {
			return EOF, emptyLit, err
		} else {
			return Number, literal, nil
		}

	case let == 0x27:
		literal, err := t.scanString(let)
		if err != nil {
			return EOF, emptyLit, err
		}

		// В литерале даты игнорируются все значения, отличные от цифр.
		if literal = extractDigits(literal); literal == "" {
			return EOF, emptyLit, fmt.Errorf("incorrect Date type constant")
		}

		return Date, literal, nil
	case let == '/' || let == ';' || let == '(' || let == ')' || let == ',' || let == '-' || let == '+' || let == '*' || let == '?' || let == '[' || let == ']' || let == ':' || let == '%':
		t.nextPos()
		return int(let), string(let), nil
	case let == '=':
		t.nextPos()
		return EQUAL, string(let), nil
	case let == '"':
		literal, err := t.scanString(let)
		if err != nil {
			return EOF, emptyLit, err
		}

		return String, literal, nil
	case let == '<':
		if t.nextLet() == '>' {
			t.nextPos()
			t.nextPos()
			return NeEQ, "<>", nil
		} else if t.nextLet() == '=' {
			t.nextPos()
			t.nextPos()
			return LE, "<=", nil
		} else {
			t.nextPos()
			return int(let), string(let), nil
		}
	case let == '>':
		if t.nextLet() == '=' {
			t.nextPos()
			t.nextPos()
			return GE, ">=", nil
		} else {
			t.nextPos()
			return int(let), string(let), nil
		}

	// case let == '#':
	// literal, err := t.scanIdentifier()
	// if err != nil {
	// 	return EOF, emptyLit, err
	// }

	case let == '&':
		t.nextPos()
		pos := t.offset

		literal := t.scanIdentifier()
		lowLit := fastToLower("&" + literal)

		if tName, ok := directives[lowLit]; ok {
			return tName, "&" + literal, nil
		} else if tName, ok = extDirectives[lowLit]; ok {
			return tName, "&" + literal, nil
		} else {
			t.offset = pos
			return int(let), string(let), fmt.Errorf(`syntax error %q`, string(let))
		}
	case let == '~':
		t.nextPos()
		return GoToLabel, t.scanIdentifier(), nil
	default:
		switch let {
		case EOF:
			t.nextPos()
			return EOF, emptyLit, nil
		// case '\n':
		// 	t.nextPos()
		// 	return int(let), string(let), nil
		default:
			t.nextPos()
			return int(let), string(let), fmt.Errorf(`syntax error %q`, string(let))
		}
	}
}

func (t *Token) scanIdentifier() string {
	ret := make([]rune, 0, identifierBufSize) // pre-allocate for typical short identifiers

	for {
		let := t.currentLet()
		if !isLetter(let) && !isDigit(let) {
			break
		}

		ret = append(ret, let)
		t.nextPos()
	}

	return string(ret)
}

func (t *Token) scanString(end rune) (string, error) {
	var ret []rune

eos:
	for {
		t.nextPos()

		switch cl := t.currentLet(); {
		case cl == EOL:
			t.nextPos()

			t.skipSpace()
			t.skipComment() // комментарии могут быть в тексте, в тексте запроса например
			if cl = t.currentLet(); cl != '|' && !isSpace(cl) {
				return "", fmt.Errorf("unexpected EOL")
			}

			ret = append(append(ret, EOL), cl)
		case cl == EOF:
			return "", fmt.Errorf("unexpected EOF")
		case cl == end:
			// пропускаем двойные "
			if t.nextLet() == '"' {
				t.nextPos()
				ret = append(ret, '"', '"')
				continue
			}
			t.nextPos()
			t.skipSpace()

			// для случаев (в 1с такое допускается, считается одна строка)
			// d = "ввава"
			// "fffff"
			if t.currentLet() == '"' {
				continue
			}

			break eos
		default:
			ret = append(ret, cl)
		}
	}

	return string(ret), nil
}

func (t *Token) skipSpace() {
	for isSpace(t.currentLet()) {
		t.nextPos()
	}
}

func (t *Token) skipComment() {
	if t.currentLet() == '/' && t.nextLet() == '/' {
		for ch := t.currentLet(); ch != EOL && ch != EOF; ch = t.currentLet() {
			t.nextPos()
		}
		t.skipSpace()
	} else {
		return
	}

	// проверяем что на новой строке нет комментария, если есть, рекурсия
	// Примечание: #-директивы НЕ пропускаем — они обрабатываются в handlePreprocessor()
	if cl := t.currentLet(); cl == '/' {
		t.skipComment()
	}
}

func (t *Token) handlePreprocessor() (int, string, error) {
	if t.currentLet() != '#' {
		return 0, "", nil
	}

	t.nextPos() // skip #
	directive := t.scanIdentifier()
	lowDirective := fastToLower(directive)

	switch lowDirective {
	case "если", "if":
		condition := t.scanUntilThen()
		return PreprocIf, condition, nil
	case "иначеесли", "elseif":
		condition := t.scanUntilThen()
		return PreprocElseIf, condition, nil
	case "иначе", "else":
		t.skipToEOL()
		return PreprocElse, "", nil
	case "конецесли", "endif":
		t.skipToEOL()
		return PreprocEndIf, "", nil
	case "область", "region":
		t.regionDepth++
		name := t.scanRegionName()
		return PreprocRegion, name, nil
	case "конецобласти", "endregion":
		t.skipToEOL()
		if t.regionDepth > 0 {
			t.regionDepth--
			return PreprocEndRegion, "", nil
		}
		// Orphan #EndRegion (no matching #Region) — skip silently, like 1C platform
		return t.next()
	case "использовать", "use":
		path := t.scanUsePath()
		return PreprocUse, path, nil
	default:
		t.skipToEOL()
		return t.next()
	}
}

// scanUntilThen сканирует условие препроцессора до ключевого слова Тогда/Then.
func (t *Token) scanUntilThen() string {
	var buf strings.Builder
	for {
		t.skipSpaceOnly()
		ch := t.currentLet()
		if ch == EOL || ch == EOF {
			break
		}

		word := t.scanIdentifier()
		if word == "" {
			if ch == '(' || ch == ')' {
				buf.WriteRune(ch)
				t.nextPos()
				continue
			}
			break
		}

		lowWord := fastToLower(word)
		if lowWord == "тогда" || lowWord == "then" {
			break
		}

		if buf.Len() > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteString(word)
	}
	return buf.String()
}

// scanRegionName сканирует имя области после #Область.
func (t *Token) scanRegionName() string {
	t.skipSpaceOnly()
	return t.scanIdentifier()
}

// scanUsePath сканирует путь после #Использовать/#Use.
func (t *Token) scanUsePath() string {
	t.skipSpaceOnly()
	if t.currentLet() == '"' {
		path, _ := t.scanString('"')
		return path
	}
	return t.scanIdentifier()
}

// skipToEOL пропускает все символы до конца строки.
func (t *Token) skipToEOL() {
	for ch := t.currentLet(); ch != EOL && ch != EOF; ch = t.currentLet() {
		t.nextPos()
	}
}

// skipSpaceOnly пропускает только пробелы и табы, но НЕ переводы строк.
func (t *Token) skipSpaceOnly() {
	for {
		ch := t.currentLet()
		if ch == ' ' || ch == '\t' || ch == '\u00A0' {
			t.nextPos()
		} else {
			break
		}
	}
}

// atStatementStart определяет, находимся ли мы в начале statement.
// Используется для различения lvalue (присваивание) от call_stmt (вызов метода).
func (t *Token) atStatementStart() bool {
	switch t.prevTokenType {
	// Токены, которые ПРОДОЛЖАЮТ выражение
	case '.':
		return false // после точки идёт свойство/метод
	case '(':
		return false // внутри скобок (аргументы, группировка)
	case '[':
		return false // внутри индекса
	case ',':
		return false // аргументы функции
	case '+', '-', '*', '/', '%':
		return false // арифметика
	case '<', '>':
		return false // сравнение
	case EQUAL:
		return false // присваивание или сравнение в выражении
	case NeEQ, LE, GE:
		return false // сравнение
	case And, OR:
		return false // логические операторы
	case Not, Await:
		return false // унарные операторы
	case ValueParam:
		return false // после Знач идёт имя параметра

	// Токены, после которых идёт НЕ statement, а часть конструкции
	case Procedure, Function, Async:
		return false // после них идёт имя функции
	case Var, VarBody:
		return false // после Var идёт имя переменной
	case For:
		return false // после For идёт переменная цикла или Each
	case Each:
		return false // после Each идёт переменная
	case In, To:
		return false // после In/To идёт выражение
	case New:
		return false // после New идёт тип
	case GoTo:
		return false // после GoTo идёт метка
	case Return, Throw, Execute:
		return false // после них идёт выражение или параметр
	case If, ElseIf, While:
		return false // после них идёт условие
	}
	// Для всех остальных (;, ), ], Then, Loop, Else, Catch, Try, identifiers, literals, etc.)
	// предполагаем что это начало statement
	return true
}

// hasEqualBeforeSemicolon сканирует вперёд и ищет одиночный "=" до ";".
// Используется для различения lvalue (присваивание) от call_stmt (вызов метода).
// Учитывает вложенность скобок — "=" внутри скобок игнорируется.
// ВАЖНО: Также останавливается на ключевых словах, заканчивающих statement
// (КонецЕсли, КонецЦикла, КонецПроцедуры, etc.)
func (t *Token) hasEqualBeforeSemicolon() bool {
	srsCode := t.ast.SrsCode()
	depth := 0
	pos := t.offset
	prevWasDot := false

	for pos < len(srsCode) {
		ch, size := utf8.DecodeRuneInString(srsCode[pos:])
		switch ch {
		case '"':
			// Пропускаем строковые литералы — внутри могут быть =, ; и др.
			pos += size
			for pos < len(srsCode) {
				sch, ssize := utf8.DecodeRuneInString(srsCode[pos:])
				pos += ssize
				if sch == '"' {
					// В 1С "" внутри строки = экранирование, проверяем следующий символ
					if pos < len(srsCode) {
						next, _ := utf8.DecodeRuneInString(srsCode[pos:])
						if next == '"' {
							pos += ssize // skip escaped ""
							continue
						}
					}
					break
				}
			}
			prevWasDot = false
			continue
		case '.':
			prevWasDot = true
		case '(', '[':
			depth++
			prevWasDot = false
		case ')', ']':
			depth--
			prevWasDot = false
		case '=':
			// "=" на верхнем уровне вложенности = присваивание
			// В 1С нет "==", поэтому любой "=" вне скобок — это присваивание
			if depth == 0 {
				return true
			}
			prevWasDot = false
		case ';':
			if depth == 0 {
				return false // конец statement без "="
			}
			// ';' внутри скобок — может быть в строке-параметре, продолжаем
			prevWasDot = false
		case EOF:
			return false
		// '\n' НЕ останавливает поиск — в 1С statement может занимать несколько строк
		default:
			// Сканируем идентификатор целиком для корректного продвижения позиции
			if isLetter(ch) {
				wordPos := pos
				for wordPos < len(srsCode) {
					wch, wsize := utf8.DecodeRuneInString(srsCode[wordPos:])
					if !isLetter(wch) && !isDigit(wch) {
						break
					}
					wordPos += wsize
				}
				// Проверяем на ключевые слова, заканчивающие блок
				// НО только на верхнем уровне и НЕ после точки (свойства — не ключевые слова)
				if depth == 0 && !prevWasDot {
					word := srsCode[pos:wordPos]
					lowWord := fastToLower(word)
					switch lowWord {
					case "конецесли", "endif",
						"конеццикла", "enddo",
						"конецпроцедуры", "endprocedure",
						"конецфункции", "endfunction",
						"конецпопытки", "endtry",
						"иначе", "else",
						"иначеесли", "elseif",
						"исключение", "except":
						return false // конец блока — это не присваивание
					}
				}
				pos = wordPos
				prevWasDot = false
				continue
			}
			prevWasDot = false
		}
		pos += size
	}
	return false
}

func (t *Token) nextLet() rune {
	srsCode := t.ast.SrsCode()
	_, size := utf8.DecodeRuneInString(srsCode[t.offset:])
	t.offset += size
	defer func() { t.offset -= size }()

	return t.currentLet()
}

func (t *Token) currentLet() rune {
	srsCode := t.ast.SrsCode()

	if t.offset >= len(srsCode) {
		return EOF
	}

	char, _ := utf8.DecodeRuneInString(srsCode[t.offset:])
	if char == utf8.RuneError {
		fmt.Println(fmt.Errorf("error decoding the character"))
		return char
	}

	return char
}

func (t *Token) GetPosition() Position {
	srsCode := t.ast.SrsCode()
	eol := strings.LastIndex(srsCode[:t.offset], "\n") + 1
	lineBegin := IF[int](eol < 0, 0, eol)

	return Position{
		Line:   strings.Count(srsCode[:t.offset], "\n") + 1,
		Column: len([]rune(srsCode[lineBegin:t.offset])) + 1,
	}
}

func (t *Token) nextPos() {
	srsCode := t.ast.SrsCode()
	_, size := utf8.DecodeRuneInString(srsCode[t.offset:])
	t.offset += size
}

func (t *Token) scanNumber() (string, error) {
	var ret []rune

	let := t.currentLet()
	for ; isDigit(let) || let == '.'; let = t.currentLet() {
		ret = append(ret, let)
		t.nextPos()
	}

	if isLetter(let) {
		return "", fmt.Errorf("identifier immediately follow the number")
	}

	return string(ret), nil
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isSpace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' || ch == '\u00A0'
}

func IF[T any](condition bool, a, b T) T {
	if condition {
		return a
	} else {
		return b
	}
}

func IsDigit(str string) bool {
	for _, c := range str {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func extractDigits(str string) string {
	result := make([]rune, 0, len(str))
	for _, c := range str {
		if c >= '0' && c <= '9' {
			result = append(result, c)
		}
	}
	return string(result)
}

func fastToLower_old(s string) string {
	rs := bytes.NewBuffer(make([]byte, 0, len(s)))
	for _, rn := range s {
		switch {
		case (rn >= 'А' && rn <= 'Я') || (rn >= 'A' && rn <= 'Z'):
			rs.WriteRune(rn + 0x20)
		case rn == 'Ё':
			rs.WriteRune('ё')
		default:
			rs.WriteRune(rn)
		}
	}
	return rs.String()
}

func fastToLower(s string) string {
	b := []byte(s)
	for i, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			b[i] = s[i] + ('a' - 'A')
		case r >= 'А' && r <= 'Я':
			if s[i] == 208 && r > 'П' { // от "П" и дальше
				b[i], b[i+1] = b[i]+1, s[i+1]-('а'-'А')
			} else {
				b[i+1] = s[i+1] + ('а' - 'А')
			}
		case r == 'Ё':
			b[i], b[i+1] = 209, 145
		}
	}

	return unsafe.String(unsafe.SliceData(b), len(b))
}
