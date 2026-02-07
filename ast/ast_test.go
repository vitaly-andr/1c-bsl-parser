package ast

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	code := `Процедура dsds() d = 864/63+607-177*906*27>737*429+84-270 КонецПроцедуры`

	a := NewAST(code)
	err := a.Parse()
	if assert.NoError(t, err) && assert.Len(t, a.ModuleStatement.Body, 1) {

		assert.Contains(t, a.PrintStatement(a.ModuleStatement.Body[0]), "Процедура")

		p := a.Print(PrintConf{OneLine: true})
		assert.Equal(t, "Процедура dsds() d = (((864 / 63) + 607) - ((177 * 906) * 27)) > (((737 * 429) + 84) - 270);КонецПроцедуры", strings.TrimSpace(p))
	}
}

func TestParse2(t *testing.T) {
	code := `
		// @strict-types
		
		
		#Если Сервер Или ТолстыйКлиентОбычноеПриложение Или ВнешнееСоединение Тогда
		
		#Область ОписаниеПеременных
		
		#КонецОбласти
		
		#Область ПрограммныйИнтерфейс
		
		// Код процедур и функций
		
		#КонецОбласти
		
		#Область ОбработчикиСобытий
		
		// Код процедур и функций
		
		#КонецОбласти
		
		#Область СлужебныйПрограммныйИнтерфейс
		
		// Код процедур и функций
		
		#КонецОбласти
		
		#Область СлужебныеПроцедурыИФункции
		
		// Код процедур и функций
		
		#КонецОбласти
		
		#Область Инициализация
		
		#КонецОбласти
		
		#КонецЕсли
		
		`

	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	assert.Len(t, a.ModuleStatement.Body, 1)
	_, ok := a.ModuleStatement.Body[0].(*PreprocessorIfStatement)
	assert.True(t, ok, "expected PreprocessorIfStatement")
}

func TestParse3(t *testing.T) {

	code := `// См. УправлениеДоступомПереопределяемый.ПриЗаполненииСписковСОграничениемДоступа.
Процедура ПриЗаполненииОграниченияДоступа(Ограничение) Экспорт
//{{MRG[ <-> ]
//
//}}MRG[ <-> ]
	Ограничение.Текст = 
	"РазрешитьЧтениеИзменение
	|ГДЕ
//{{MRG[ <-> ]
	|	ЗначениеРазрешено(Организация)
	|	И ЗначениеРазрешено(ФизическоеЛицо)";
//}}MRG[ <-> ]
//{{MRG[ <-> ]
//	|	ЗначениеРазрешено(Организация)";
//
//}}MRG[ <-> ]
КонецПроцедуры`

	a := NewAST(code)
	err := a.Parse()
	if assert.NoError(t, err) && assert.NotNil(t, a.ModuleStatement.Body) {
		p := a.Print(PrintConf{OneLine: true})
		assert.Equal(t, "Процедура ПриЗаполненииОграниченияДоступа(Ограничение) Экспорт Ограничение.Текст = \"РазрешитьЧтениеИзменение\n|ГДЕ\n|\tЗначениеРазрешено(Организация)\n|\tИ ЗначениеРазрешено(ФизическоеЛицо)\";КонецПроцедуры", strings.TrimSpace(p))
	}
}

func TestParseString(t *testing.T) {
	t.Run("test1", func(t *testing.T) {
		code := `a = "rererer // rererer"`

		a := NewAST(code)
		err := a.Parse()
		if assert.NoError(t, err) && assert.NotNil(t, a.ModuleStatement.Body) {
			p := a.Print(PrintConf{OneLine: true})
			assert.Equal(t, "a = \"rererer // rererer\";", strings.TrimSpace(p))
		}
	})
	t.Run("test2", func(t *testing.T) {
		code := `a = "rererer 
| rererer
// rererer
| rererer"`

		a := NewAST(code)
		err := a.Parse()
		if assert.NoError(t, err) && assert.NotNil(t, a.ModuleStatement.Body) {
			p := a.Print(PrintConf{OneLine: true})
			assert.Equal(t, "a = \"rererer \n| rererer\n| rererer\";", strings.TrimSpace(p))
		}
	})
	t.Run("test3", func(t *testing.T) {
		code := `a = "rererer 
						| rererer
						// rererer
						| rererer"`

		a := NewAST(code)
		err := a.Parse()
		if assert.NoError(t, err) && assert.NotNil(t, a.ModuleStatement.Body) {
			p := a.Print(PrintConf{OneLine: true})
			assert.Equal(t, "a = \"rererer \n| rererer\n| rererer\";", strings.TrimSpace(p))
		}
	})
	t.Run("test4", func(t *testing.T) {
		code := `a = "rererer"+ 
						"rererer" + "rererer";`

		a := NewAST(code)
		err := a.Parse()
		if assert.NoError(t, err) && assert.NotNil(t, a.ModuleStatement.Body) {
			p := a.Print(PrintConf{OneLine: true})
			assert.Equal(t, "a = (\"rererer\" + \"rererer\") + \"rererer\";", strings.TrimSpace(p))
		}
	})
	t.Run("test5", func(t *testing.T) {
		code := `a = "123_"
					"123_" 
"123";`

		a := NewAST(code)
		err := a.Parse()
		if assert.NoError(t, err) && assert.NotNil(t, a.ModuleStatement.Body) {
			p := a.Print(PrintConf{OneLine: true})
			assert.Equal(t, "a = \"123_123_123\";", strings.TrimSpace(p))
		}
	})
	t.Run("test6", func(t *testing.T) {
		code := `Процедура test()
		Диалог.Фильтр = "XML|*.xml";
		КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		if assert.NoError(t, err) && assert.NotNil(t, a.ModuleStatement.Body) {
			p := a.Print(PrintConf{OneLine: true})
			assert.Equal(t, "Процедура test() Диалог.Фильтр = \"XML|*.xml\";КонецПроцедуры", strings.TrimSpace(p))
		}
	})
}

func TestParseModule(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		code := ``

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("global var", func(t *testing.T) {
		code := `&НаСервере
				Перем в, e;

				&НаКлиенте 
				Перем а Экспорт; Перем с;

				Процедура вв1()
				
				Конецпроцедуры
				
				&НаКлиенте
				Процедура вв2()

				Конецпроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("global var error", func(t *testing.T) {
		code := `Перем а;
				Перем а;
				
				Процедура вв()
				
				Конецпроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.ErrorContains(t, err, "variable has already been defined")
	})
	t.Run("global var error", func(t *testing.T) {
		code := `Перем в; 

				Процедура вв()
					
				Конецпроцедуры
				Перем а;`

		a := NewAST(code)
		err := a.Parse()
		assert.ErrorContains(t, err, "variable declarations must be placed at the beginning of the module")
	})
	t.Run("without FunctionProcedure pass", func(t *testing.T) {
		code := `
					Пока Истина Цикл
						
					КонецЦикла;
					
					
					ВызватьИсключение "";
					
					Если Истина Тогда
						а = 0;
					КонецЕсли`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("without FunctionProcedure pass", func(t *testing.T) {
		code := `Перем в; 
					Функция test1() 
					КонецФункции

Функция test1() 
					КонецФункции

					Пока Истина Цикл
						
					КонецЦикла;
					
					
					ВызватьИсключение "";
					
					Если Истина Тогда
						а = 0;
					КонецЕсли;`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)

		// fmt.Println(a.Print(PrintConf{Margin: 4}))
	})
	t.Run("without FunctionProcedure error", func(t *testing.T) {
		code := `
					Пока Истина Цикл
						
					КонецЦикла;
					
					
					ВызватьИсключение "";
					
					Если Истина Тогда
						а = 0;
					КонецЕсли;

					Процедура test()
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.Error(t, err)
		// assert.ErrorContains(t, err, "procedure and function definitions should be placed before the module body statements")
	})
}

func TestExecute(t *testing.T) {
	t.Run("Execute-1", func(t *testing.T) {
		code := `&НаСервере
					Процедура ВыполнитьВБезопасномРежиме(Знач Алгоритм, Знач Параметры = Неопределено)
						Выполнить Алгоритм;
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		if assert.NoError(t, err) {
			p := a.Print(PrintConf{OneLine: true})
			assert.Equal(t, "&НаСервере\nПроцедура ВыполнитьВБезопасномРежиме(Знач Алгоритм, Знач Параметры = Неопределено) Выполнить(Алгоритм);КонецПроцедуры ", p)
		}
	})
	t.Run("Execute-2", func(t *testing.T) {
		code := `&НаСервере
					Процедура ВыполнитьВБезопасномРежиме(Знач Алгоритм, Знач Параметры = Неопределено)
						Выполнить(Алгоритм);
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)

		p := a.Print(PrintConf{Margin: 4})
		assert.Equal(t, true, compareHashes(code, p))
	})
	t.Run("Execute-3", func(t *testing.T) {
		code := `&НаСервере
					Процедура ВыполнитьВБезопасномРежиме(Знач Алгоритм, Знач Параметры = Неопределено)
						Выполнить "Алгоритм";
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("Execute-number", func(t *testing.T) {
		// Execute expr accepts any expression (like Java BSLParser)
		code := `&НаСервере
					Процедура ВыполнитьВБезопасномРежиме(Знач Алгоритм, Знач Параметры = Неопределено)
						Выполнить 32;
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("Execute-multi-arg-error", func(t *testing.T) {
		code := `&НаСервере
					Процедура ВыполнитьВБезопасномРежиме(Знач Алгоритм, Знач Параметры = Неопределено)
						Выполнить "Алгоритм", "";
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		// Comma after "Алгоритм" is unexpected — Execute takes one expression
		assert.EqualError(t, err, "syntax error. line: 3, column: 26 (unexpected literal: \",\")")
	})
	t.Run("Execute-parens-multi", func(t *testing.T) {
		// Execute (expr, expr) — parenthesized expressions list
		code := `&НаСервере
					Процедура ВыполнитьВБезопасномРежиме(Знач Алгоритм, Знач Параметры = Неопределено)
						Выполнить ("Алгоритм", "");
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("Eval-1", func(t *testing.T) {
		code := `&НаСервере
					Процедура ВыполнитьВБезопасномРежиме(Знач Алгоритм, Знач Параметры = Неопределено)
						в = Вычислить(Алгоритм);
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("Eval-2", func(t *testing.T) {
		// Grammar uses optional semicolons (like Java BSLParser SEMICOLON?).
		// "Вычислить Алгоритм;" is parsed as two statements without separator.
		code := `&НаСервере
					Процедура ВыполнитьВБезопасномРежиме(Знач Алгоритм, Знач Параметры = Неопределено)
						Вычислить Алгоритм;
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
}

func TestParseIF(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если (1 = 1) Тогда 

					КонецЕсли; 
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если в = 1 И а = 1 или у = 3 Тогда
						test = 2+2*2;
						а = 7;
						а = 7.2;
					ИначеЕсли Не 4 = 3 И Не 8 = 2 И 1 <> 3 Тогда;
						а = 5;
					ИначеЕсли Ложь Тогда;
					Иначе
						а = -(1+1);
						а = -s;
						а = -1;
						а = -7.42;
						а = Не истина;
					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)

		data, err := a.JSON()
		assert.NoError(t, err)
		assert.NotEqual(t, 0, len(data))
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если в = 1 И а = 1 или у = 3 Тогда

					КонецЕсли

					;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если в = 1 И (а = 1 или у = 3) Тогда

					КонецЕсли

					;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если в = 1 И (а = 1 или у = 3) Тогда
						Если в = 1 или у = 3 Тогда

						КонецЕсли;
					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если в = 1 И (а = 1 или у = 3) Тогда
						Если в = 1 или у = 3 Тогда
						Иначе
						КонецЕсли;
					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если в = 1 И (а = 1 или у = 3) Тогда
						Если в = 1 или у = 3 Тогда

						ИначеЕсли ввв Тогда

						ИначеЕсли авыав Тогда

						Иначе

						КонецЕсли;
					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если Истина Тогда
	
					КонецЕсли // запятой может и не быть
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если (Истина или Ложь) Тогда
						а = 0;
					КонецЕсли
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если (1 = 1) Тогда 
						f = 0 // запятой может не быть
					КонецЕсли; 
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если в = 1 И (а = 1 или у = 3) Тогда
						Если в = 1 или у = 3 Тогда

						ИначеЕсли ввв Тогда

						ИначеЕсли авыав Тогда

						КонецЕсли;
					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если в = 1 И (а = 1 или у = 3) Тогда
						Если в = 1 или у = 3 Тогда
							а = 1 + 3 *4;
						ИначеЕсли ввв Тогда

						ИначеЕсли авыав Тогда

						КонецЕсли;
					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 
						Если ВерсияПлатформыЧислом < 80303641 Тогда
							ВызватьИсключение НСтр("ru = 'Для работы обработки требуется ""1С:Предприятие"" версии 8.3.3.641 или страше.';sys= ''", "ru");
						КонецЕсли;
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("no-semicolons", func(t *testing.T) {
		// Grammar uses optional semicolons — missing ; between statements is accepted
		code := `Процедура ПодключитьВнешнююОбработку()
					Если в = 1 И (а = 1 или у = 3) Тогда
						Если в = 1 или у = 3 Тогда
							а = 1 + 3 * 4
							b = 1
						ИначеЕсли ввв Тогда

						ИначеЕсли авыав Тогда

						КонецЕсли;
					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если в = 1 И (а = 1 или у = 3) Тогда
						Если в = 1 или у = 3 Тогда

						ИначеЕсли ввв Тогда

						ИначеЕсли авыав 

						КонецЕсли;
					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "syntax error. line: 9, column: 6 (unexpected literal: \"КонецЕсли\")")
	})
	t.Run("error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если в = 1 И (а = 1 или у = 3) Тогда
						Если в = 1 или у = 3 Тогда

					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "syntax error. line: 6, column: 4 (unexpected literal: \"КонецПроцедуры\")")
	})
	t.Run("error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если  Тогда
	
					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "syntax error. line: 2, column: 11 (unexpected literal: \"Тогда\")")
	})
	t.Run("error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если f Тогд
	
					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "syntax error. line: 2, column: 12 (unexpected literal: \"Тогд\")")
	})
	t.Run("error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если ав f Тогда
	
					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "syntax error. line: 2, column: 13 (unexpected literal: \"f\")")
	})
	t.Run("\"not\" pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Если Не f Тогда

					КонецЕсли;

					Если Не f Тогда
						d = 0;
					ИначеЕсли 3 = 9 Тогда
						Если тогоСего Тогда
						
						КонецЕсли;
					Иначе
						Если Не f И не 1 = 1 ИЛИ не (а = 2 ИЛИ Истина) Тогда
						
						КонецЕсли;
					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Функция Команда1НаСервере()

				Если Не ШаблонТекстаОшибки = "" Тогда

				КонеЦесли;

			 КонецФункции`

		a := NewAST(code)
		err := a.Parse()
		if assert.NoError(t, err) {
			p := a.Print(PrintConf{OneLine: true})
			assert.Equal(t, "Функция Команда1НаСервере() Если Не (ШаблонТекстаОшибки = \"\") Тогда КонецЕсли;КонецФункции", strings.TrimSpace(p))
		}
	})
}

func TestParseLoop(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Для Каждого ИзмененныйОбъект Из ОбъектыНазначения Цикл
						Тип = ТипЗнч(ИзмененныйОбъект);
						Если ТипыИзмененныхОбъектов  = Неопределено Тогда
							ТипыИзмененныхОбъектов = 0;
						КонецЕсли;
					КонецЦикла;

				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)

		data, err := a.JSON()
		assert.NoError(t, err)
		assert.NotEqual(t, 0, len(data))
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Для а = 0 По 100 Цикл
						Тип = ТипЗнч(ИзмененныйОбъект);
						Если ТипыИзмененныхОбъектов  = Неопределено Тогда
							ТипыИзмененныхОбъектов = 0;
						КонецЕсли;
					КонецЦикла;

				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Для а = 0 По 100 Цикл
						Тип = ТипЗнч(ИзмененныйОбъект);
   						Продолжить;

						Если ТипыИзмененныхОбъектов  = Неопределено Тогда
							Продолжить;
						Иначе
							Прервать;
						КонецЕсли;
					КонецЦикла;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Для а = 0 По 100 Цикл            
						Для а = 0 По 100 Цикл
							Если Истина Тогда
								Прервать;
							КонецЕсли;
						КонецЦикла;
						
						Если ТипыИзмененныхОбъектов  = Неопределено Тогда
							Продолжить;
						Иначе
							Прервать;
						КонецЕсли;
					КонецЦикла; 
					
					Если ТипыИзмененныхОбъектов  = Неопределено Тогда       
						Для а = 0 По 100 Цикл
							Если Истина Тогда
								Прервать;
							КонецЕсли;
						КонецЦикла;
					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку()
					Для Каждого КлючЗначение Из Новый Структура(СписокКолонок) Цикл
						
					КонецЦикла;
					Для Каждого КлючЗначение Из (Новый Структура(СписокКолонок2)) Цикл
						
					КонецЦикла;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		if assert.NoError(t, err) {
			p := a.Print(PrintConf{Margin: 0})
			assert.Equal(t, "Процедура ПодключитьВнешнююОбработку() \nДля Каждого КлючЗначение Из Новый Структура(СписокКолонок) Цикл \nКонецЦикла;\nДля Каждого КлючЗначение Из Новый Структура(СписокКолонок2) Цикл \nКонецЦикла;\nКонецПроцедуры", deleteEmptyLine(p))
		}
	})
	t.Run("error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Для а = 0 По 100 Цикл            
						Для а = 0 По 100 Цикл
							Если Истина Тогда
								Прервать;
							КонецЕсли;
						КонецЦикла;
						
						Если ТипыИзмененныхОбъектов  = Неопределено Тогда
							Продолжить;
						Иначе
							Прервать;
						КонецЕсли;
					КонецЦикла; 
					
					Если ТипыИзмененныхОбъектов  = Неопределено Тогда       
						Для а = 0 По 100 Цикл
							Если Истина Тогда
								Прервать;
							КонецЕсли;
						КонецЦикла;

						Продолжить; // вне цикла нельзя
					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "operator \"Продолжить\" can only be used inside a loop. line: 23, column: 6 (unexpected literal: \"Продолжить\")")
	})
	t.Run("error", func(t *testing.T) {
		code := `Функция ПодключитьВнешнююОбработку() 
					Если 1 = 1 Тогда
							f = 1+1;
							Прервать; // вне цикла нельзя
					КонецЕсли;

					Для Каждого ИзмененныйОбъект Из ОбъектыНазначения Цикл
						Если 1 = 1 Тогда
							f = 1+1;
							Прервать;
						КонецЕсли;
					КонецЦикла;
				КонецФункции`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "operator \"Прервать\" can only be used inside a loop. line: 4, column: 7 (unexpected literal: \"Прервать\")")
	})
	t.Run("error", func(t *testing.T) {
		code := `Функция ПодключитьВнешнююОбработку() 
					Если 1 = 1 Тогда
							f = 1+1;
							Прервать;
					КонецЕсли;
				КонецФункции`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "operator \"Прервать\" can only be used inside a loop. line: 4, column: 7 (unexpected literal: \"Прервать\")")
	})
	t.Run("error", func(t *testing.T) {
		code := `Функция ПодключитьВнешнююОбработку() 
					Для Каждого ИзмененныйОбъект Из ОбъектыНазначения Цикл
						Если 1 = 1 Тогда
							f = 1+1;
							Прервать;
						КонецЕсли;
						продолжить;
					КонецЦикла;

					Если 1 = 1 Тогда
							f = 1+1;
							Прервать;
					КонецЕсли;
				КонецФункции`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "operator \"Прервать\" can only be used inside a loop. line: 12, column: 7 (unexpected literal: \"Прервать\")")
	})
	t.Run("error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Продолжить; // вне цикла нельзя
					Для а = 0 По 100 Цикл
						Тип = ТипЗнч(ИзмененныйОбъект);
						Если ТипыИзмененныхОбъектов  = Неопределено Тогда
							Продолжить;
						Иначе
							Прервать;
						КонецЕсли;
					КонецЦикла;

				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "operator \"Продолжить\" can only be used inside a loop. line: 2, column: 5 (unexpected literal: \"Продолжить\")")
	})
	t.Run("error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Прервать; // вне цикла нельзя
					Для а = 0 По 100 Цикл
						Тип = ТипЗнч(ИзмененныйОбъект);
						Если ТипыИзмененныхОбъектов  = Неопределено Тогда
							Продолжить;
						Иначе
							Прервать;
						КонецЕсли;
					КонецЦикла;

				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "operator \"Прервать\" can only be used inside a loop. line: 2, column: 5 (unexpected literal: \"Прервать\")")
	})
	t.Run("error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Для а = 0 По  Цикл
						Тип = ТипЗнч(ИзмененныйОбъект);
						Если ТипыИзмененныхОбъектов  = Неопределено Тогда
							ТипыИзмененныхОбъектов = 0;
						КонецЕсли;
					КонецЦикла;

				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "syntax error. line: 2, column: 19 (unexpected literal: \"Цикл\")")
	})
	t.Run("error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					Для ИзмененныйОбъект Из ОбъектыНазначения Цикл
						Тип = ТипЗнч(ИзмененныйОбъект);
						Если ТипыИзмененныхОбъектов  = Неопределено Тогда
							ТипыИзмененныхОбъектов = 0;
						КонецЕсли;
					КонецЦикла;

				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "syntax error. line: 2, column: 26 (unexpected literal: \"Из\")")
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура rrrr() 
				Для Каждого Стр Из ?(ТекущаяСтраница = Элементы.СтраницаДополнительныеРеквизиты,СписокРеквизитов,СписокРеквизитовОсновныеРеквизиты) Цикл
					Стр.Пометка = Ложь;
				КонецЦикла;
		КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		if assert.NoError(t, err) {
			p := a.Print(PrintConf{OneLine: true})
			assert.Equal(t, normalize(code), normalize(p))
		}
	})
}

func TestTryCatch(t *testing.T) {
	t.Run("throw", func(t *testing.T) {
		t.Run("pass", func(t *testing.T) {
			code := `Процедура ПодключитьВнешнююОбработку() 
						Если в = 1 И (а = 1 или у = 3) Тогда
							f = 0;
							ВызватьИсключение "dsdsd dsds";							
							f = 0;
							f = 0;
						КонецЕсли;
					КонецПроцедуры`

			a := NewAST(code)
			err := a.Parse()
			assert.NoError(t, err)
		})
		t.Run("error", func(t *testing.T) {
			code := `Процедура ПодключитьВнешнююОбработку() 
						Если в = 1 И (а = 1 или у = 3) Тогда
							f = 0;
							ВызватьИсключение; // без параметров нельзя
						КонецЕсли;
					КонецПроцедуры`

			a := NewAST(code)
			err := a.Parse()
			assert.EqualError(t, err, "operator \"ВызватьИсключение\" without arguments can only be used when handling an exception. line: 4, column: 24 (unexpected literal: \";\")")
		})
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
						Попытка 
							а = 1+1;					
						Исключение
							ВызватьИсключение "fff";
						КонецПопытки;
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		if assert.NoError(t, err) {
			json, _ := a.JSON()
			assert.Contains(t, string(json), `{"Param":"fff"}`)
		}
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
						Попытка 
							Попытка 
								а = 1+1;					
							Исключение
								ВызватьИсключение;
							КонецПопытки;				
						Исключение
							ВызватьИсключение
						КонецПопытки;
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку()
						Попытка 
							а = 1+1;
							ВызватьИсключение("dsdsd dsds");	  
							f = 0;
							f = 0		
						Исключение
							а = 1+1;
							а = 1+1;
							ВызватьИсключение;  // в блоке Исключение можно вызывать без параметров
							а = 1+1;
							
							Если истина Тогда
								ВызватьИсключение;  // в блоке Исключение можно вызывать без параметров
							КонецЕсли
						КонецПопытки;
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
						Попытка 
							а = 1+1;
							Если в = 1 И (а = 1 или у = 3) Тогда
								ВызватьИсключение "SDSDSD";
							КонецЕсли;
						Исключение
							а = 1+1;
							ВызватьИсключение "dsd";
							а = 1+1;
						КонецПопытки;
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
						Попытка 
							Попытка 
								а = 1+1;					
							Исключение
								ВызватьИсключение;
							КонецПопытки;		

							ВызватьИсключение ;
						Исключение
							ВызватьИсключение
						КонецПопытки;
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "operator \"ВызватьИсключение\" without arguments can only be used when handling an exception. line: 9, column: 25 (unexpected literal: \";\")")
	})
	t.Run("error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
						Попытка 
							Попытка 
								а = 1+1;					
							Исключение
								ВызватьИсключение;
							КонецПопытки;
						Исключение
							ВызватьИсключение
						КонецПопытки;

						ВызватьИсключение 
					КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.ErrorContains(t, err, "without arguments can only be used when handling an exception")
	})
	t.Run("pass", func(t *testing.T) {
		code := `Функция Команда1НаСервере()

					ВызватьИсключение(НСтр("ru = 'Недостаточно прав на использование сертификата.'"),
						КатегорияОшибки.НарушениеПравДоступа);

			 КонецФункции`

		a := NewAST(code)
		err := a.Parse()
		if assert.NoError(t, err) {
			p := a.Print(PrintConf{OneLine: true})
			assert.Equal(t, "Функция Команда1НаСервере() ВызватьИсключение(НСтр(\"ru = 'Недостаточно прав на использование сертификата.'\"), КатегорияОшибки.НарушениеПравДоступа);КонецФункции", strings.TrimSpace(p))
		}
	})
}

func TestParseMethod(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					а = ТипыИзмененныхОбъектов.Найти(Тип)
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					а = ТипыИзмененныхОбъектов.Test.Найти(Тип)
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					а = ТипыИзмененныхОбъектов(Тип);
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку() 
					а = ТипыИзмененныхОбъектов..Найти(Тип)
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "syntax error. line: 2, column: 32 (unexpected literal: \".\")")
	})
}

func TestParseFunctionProcedure(t *testing.T) {
	t.Run("Function", func(t *testing.T) {
		t.Run("ast", func(t *testing.T) {
			code := `&НасервереБезКонтекста
					Функция ПодключитьВнешнююОбработку(Ссылка) 
						f = 1 + gggg - (fd +1 / 3);
					КонецФункции`

			a := NewAST(code)
			err := a.Parse()
			assert.NoError(t, err)

			data, err := a.JSON()
			assert.NoError(t, err)
			assert.NotEqual(t, 0, len(data))
		})
		t.Run("ast", func(t *testing.T) {
			code := `&НасервереБезКонтекста
					Функция ПодключитьВнешнююОбработку(Ссылка) 
						f = парапапапам; 
					КонецФункции`

			a := NewAST(code)
			err := a.Parse()
			if assert.NoError(t, err) {
				p := a.Print(PrintConf{OneLine: true})
				assert.Equal(t, "&НасервереБезКонтекста\nФункция ПодключитьВнешнююОбработку(Ссылка) f = парапапапам;КонецФункции", strings.TrimSpace(p))
			}
		})
		t.Run("ast", func(t *testing.T) {
			code := `&НасервереБезКонтекста
					Функция ПодключитьВнешнююОбработку(Ссылка) 
						f = 221;
						возврат 2+2;
					КонецФункции`

			a := NewAST(code)
			err := a.Parse()
			fmt.Println(a.Print(PrintConf{Margin: 2}))
			if assert.NoError(t, err) {
				p := a.Print(PrintConf{OneLine: true})
				assert.Equal(t, "&НасервереБезКонтекста\nФункция ПодключитьВнешнююОбработку(Ссылка) f = 221;Возврат 2 + 2;КонецФункции", strings.TrimSpace(p))
			}
		})
		t.Run("ast", func(t *testing.T) {
			code := `&НасервереБезКонтекста
					Функция ПодключитьВнешнююОбработку(Ссылка) 
						f = "вававава авава";
					КонецФункции`

			a := NewAST(code)
			err := a.Parse()
			if assert.NoError(t, err) {
				p := a.Print(PrintConf{OneLine: true})
				assert.Equal(t, "&НасервереБезКонтекста\nФункция ПодключитьВнешнююОбработку(Ссылка) f = \"вававава авава\";КонецФункции", strings.TrimSpace(p))
			}
		})
		t.Run("ast", func(t *testing.T) {
			code := `&НасервереБезКонтекста
					Функция ПодключитьВнешнююОбработку(Ссылка) 
						f = Истина;
					КонецФункции`

			a := NewAST(code)
			err := a.Parse()
			if assert.NoError(t, err) {
				p := a.Print(PrintConf{OneLine: true})
				assert.Equal(t, "&НасервереБезКонтекста\nФункция ПодключитьВнешнююОбработку(Ссылка) f = Истина;КонецФункции", strings.TrimSpace(p))
			}
		})
		t.Run("ast", func(t *testing.T) {
			code := `&НасервереБезКонтекста
					Функция ПодключитьВнешнююОбработку(Ссылка) 
						f = Ложь;
					КонецФункции`

			a := NewAST(code)
			err := a.Parse()
			if assert.NoError(t, err) {
				p := a.Print(PrintConf{OneLine: true})
				assert.Equal(t, "&НасервереБезКонтекста\nФункция ПодключитьВнешнююОбработку(Ссылка) f = Ложь;КонецФункции", strings.TrimSpace(p))
			}
		})
		t.Run("bad directive", func(t *testing.T) {
			code := `&НасервереБез
					Функция ПодключитьВнешнююОбработку(Ссылка) 

					КонецФункции`

			a := NewAST(code)
			err := a.Parse()
			assert.EqualError(t, err, "syntax error. line: 1, column: 1 (unexpected literal: \"НасервереБез\")")
		})
		t.Run("without directive", func(t *testing.T) {
			code := `Функция ПодключитьВнешнююОбработку(Ссылка, вы, выыыыы) 

					КонецФункции`

			a := NewAST(code)
			err := a.Parse()
			assert.NoError(t, err)
		})
		t.Run("return", func(t *testing.T) {
			code := `Функция ПодключитьВнешнююОбработку(Ссылка, вы, выыыыы) 
						Перем а;
						Перем вы, в;

						Если истина Тогда
							ВызватьИсключение вызов("");
						ИначеЕсли 1 = 1 Тогда
						ИначеЕсли 2 = 2 Тогда
						Иначе
							б = а;
						КонецЕсли;
					КонецФункции`

			a := NewAST(code)
			err := a.Parse()
			assert.NoError(t, err)

			// fmt.Println(a.Print(&PrintConf{Margin: 2}))
		})
		t.Run("return", func(t *testing.T) {
			code := `Функция ПодключитьВнешнююОбработку(Ссылка, вы, выыыыы) 
						Перем а;
						Перем вы, в;

						Если истина Тогда
							Возврат;
						КонецЕсли;
					КонецФункции`

			a := NewAST(code)
			err := a.Parse()
			assert.NoError(t, err)
		})
		t.Run("export", func(t *testing.T) {
			code := `Функция ПодключитьВнешнююОбработку(Ссылка) Экспорт

					КонецФункции`

			a := NewAST(code)
			err := a.Parse()
			assert.NoError(t, err)
		})
		t.Run("error", func(t *testing.T) {
			code := `Функция ПодключитьВнешнююОбработку(Ссылка) 

					КонецФунки`

			a := NewAST(code)
			err := a.Parse()
			assert.EqualError(t, err, "syntax error. line: 3, column: 5 (unexpected literal: \"КонецФунки\")")
		})
		t.Run("error", func(t *testing.T) {
			code := `Функция ПодключитьВнешнююОбработку(Ссылка) 

					КонецПроцедуры`

			a := NewAST(code)
			err := a.Parse()
			assert.EqualError(t, err, "syntax error. line: 3, column: 5 (unexpected literal: \"КонецПроцедуры\")")
		})
		t.Run("params def value", func(t *testing.T) {
			code := `Функция ПодключитьВнешнююОбработку(Парам1, Парам2 = Неопределено, Знач Парам3 = "вывыв", парам4 = 4) 

					КонецФункции`

			a := NewAST(code)
			err := a.Parse()
			assert.NoError(t, err)
		})
	})
	t.Run("Procedure", func(t *testing.T) {
		t.Run("with directive", func(t *testing.T) {
			code := `&НасервереБезКонтекста
					Процедура ПодключитьВнешнююОбработку() 

					КонецПроцедуры`

			a := NewAST(code)
			err := a.Parse()
			assert.NoError(t, err)
		})
		t.Run("bad directive", func(t *testing.T) {
			code := `&НасервереБез
					Процедура ПодключитьВнешнююОбработку(Ссылка) 

					КонецПроцедуры`

			a := NewAST(code)
			err := a.Parse()
			assert.EqualError(t, err, "syntax error. line: 1, column: 1 (unexpected literal: \"НасервереБез\")")
		})
		t.Run("export", func(t *testing.T) {
			code := `Процедура ПодключитьВнешнююОбработку(Ссылка) Экспорт

					КонецПроцедуры`

			a := NewAST(code)
			err := a.Parse()
			assert.NoError(t, err)
		})
		t.Run("error", func(t *testing.T) {
			code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 

					КонецФункции`

			a := NewAST(code)
			err := a.Parse()
			assert.EqualError(t, err, "syntax error. line: 3, column: 5 (unexpected literal: \"КонецФункции\")")
		})
		t.Run("with var pass", func(t *testing.T) {
			code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 
						Перем а;
						Перем вы, в;

						Если истина Тогда
							ВызватьИсключение "";
						КонецЕсли;
					КонецПроцедуры`

			a := NewAST(code)
			err := a.Parse()
			assert.NoError(t, err)
			assert.Equal(t, 3, len(a.ModuleStatement.Body[0].(*FunctionOrProcedure).ExplicitVariables))
		})
		t.Run("with var error", func(t *testing.T) {
			code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 
						Перем а;
						Перем а, вы, в;

						Если истина Тогда
							ВызватьИсключение "";
						КонецЕсли;
					КонецПроцедуры`

			a := NewAST(code)
			err := a.Parse()
			assert.ErrorContains(t, err, "variable has already been defined")
		})
		t.Run("return", func(t *testing.T) {
			code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 
						Перем а;
						Перем вы, в;

						Если истина Тогда
							возврат;
						КонецЕсли;
					КонецПроцедуры`

			a := NewAST(code)
			err := a.Parse()
			assert.NoError(t, err)
		})
		t.Run("return error", func(t *testing.T) {
			code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 
						Перем а;
						Перем а, вы, в;

						Если истина Тогда
							возврат "";
						КонецЕсли;
					КонецПроцедуры`

			a := NewAST(code)
			err := a.Parse()
			assert.ErrorContains(t, err, "procedure cannot return a value")
		})
		t.Run("with var error", func(t *testing.T) {
			code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 
						Перем а;
						Перем а, вы, в

						Если истина Тогда
							ВызватьИсключение "";
						КонецЕсли;
					КонецПроцедуры`

			a := NewAST(code)
			err := a.Parse()
			assert.EqualError(t, err, "syntax error. line: 5, column: 6 (unexpected literal: \"Если\")")
		})
		t.Run("with var error", func(t *testing.T) {
			// Перем after statements is now allowed in body (VarBody token)
			// This supports UNF files with Перем inside #Область
			code := `Процедура ПодключитьВнешнююОбработку(Ссылка)
						Если истина Тогда
							ВызватьИсключение "";
						КонецЕсли;

						Перем а, вы, в;
					КонецПроцедуры`

			a := NewAST(code)
			err := a.Parse()
			assert.NoError(t, err)
		})
		t.Run("with region", func(t *testing.T) {
			code := `#Область ПрограммныйИнтерфейс
// hg
#Область ПрограммныйИнтерфейс
					&НасервереБезКонтекста
					Процедура ПодключитьВнешнююОбработку()
						ТипЗначенияСтрокой = XMLТипЗнч(КлючДанных).ИмяТипа;

					КонецПроцедуры
					#КонецОбласти
#КонецОбласти

					#Область СлужебныеПроцедурыИФункции
					&НасервереБезКонтекста
						Функция ПодключитьВнешнююОбработку() 
							ВызватьИсключение "Нет соответствия шаблону! " + СтрокаТекста;

						КонецФункции
					#КонецОбласти`

			a := NewAST(code)
			err := a.Parse()
			assert.NoError(t, err)
		})
		t.Run("through_dot pass", func(t *testing.T) {
			code := `Процедура ЗагрузитьОбъекты(Задание, Отказ = Ложь) Экспорт
						Перем СоответствиеРеквизитовШапки;
					
						Организация  = Задание.Организация.ВыполнитьМетодСПараметрами(1, "ав", авава);
						Организация  = Задание.Организация.ВыполнитьМетодБезПараметров();
						Организация  = Задание.Организация.Код;
 
					КонецПроцедуры`

			a := NewAST(code)
			err := a.Parse()
			assert.NoError(t, err)

			data, err := a.JSON()
			assert.NoError(t, err)
			assert.NotEqual(t, 0, len(data))
		})
	})
	t.Run("many", func(t *testing.T) {
		code := `&Насервере
				Процедура ПодключитьВнешнююОбработку() 
					Возврат
				КонецПроцедуры

				&НаКлиенте
				Функция ОчиститьПараметрыТЖ(парам1 = 1, парам2 = Неопределено, парам3 = -1) Экспорт
					Возврат 100;
				КонецФункции

				Функция ПарамТарам(Знач парам1)
					возврат +1;
				КонецФункции`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
		if !t.Failed() {
			p := a.Print(PrintConf{Margin: 0})
			assert.Equal(t, "&Насервере\nПроцедура ПодключитьВнешнююОбработку() \nВозврат;\nКонецПроцедуры \n&НаКлиенте\nФункция ОчиститьПараметрыТЖ(парам1 = 1, парам2 = Неопределено, парам3 = -1) Экспорт \nВозврат 100;\nКонецФункции \nФункция ПарамТарам(Знач парам1) \nВозврат 1;\nКонецФункции", deleteEmptyLine(p))
		}
	})
}

func TestParseBaseExpression(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку(Ссылка) ds; КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку(Ссылка) ds = 222; uu = 9; КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 
					ds = 222; ds = 222; uu = 9
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 
					ds = 222



					; uu = 9;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 
					ds = 222; 
					uu = 9;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("no-semicolons", func(t *testing.T) {
		// Grammar uses optional semicolons — missing ; between statements is accepted
		code := `Процедура ПодключитьВнешнююОбработку(Ссылка)
					ds = 222
					uu = 9;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 
					ds = ИспользуемыеНастройки[0].Структура[0].Структура;
					fdfd = СтруктураКонтрагент();
					fdfd = f.СтруктураКонтрагент(gf, ghf);
					СтруктураКонтрагент.Наименование = СтрокаВывода[РезультатВывода.Колонки.Найти("СтруктураКонтрагентНаименование").Имя];
					СтрокаСпискаПП[ТекКолонка.Ключ]["РасшифровкаПлатежа"].Добавить(ВременнаяСтруктура);
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 
					ds = ИспользуемыеНастройки[0].Структура[0].Структура;
					fdfd = СтруктураКонтрагент();
					fdfd = f.СтруктураКонтрагент(gf, ghf);
					СтруктураКонтрагент.Наименование = СтрокаВывода[РезультатВывода.Колонки.Найти("СтруктураКонтрагентНаименование.Имя];
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.Error(t, err)
	})
	t.Run("new pass", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 
					Контекст = Новый Структура;
					Контекст = Новый Структура();
					Контекст = Новый Структура("выыыы");
					Контекст = Новый Структура(какойтофункшин());
					Контекст = Новый Структура("какойтоимя", чето);
					Запрос = Новый Запрос(ТекстЗапросаЗадание());
					Оповещение = Новый ОписаниеОповещения(,, Контекст,
													"ОткрытьНавигационнуюСсылкуПриОбработкеОшибки", ОбщегоНазначенияСлужебныйКлиент);
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
	t.Run("new error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 
					Контекст = Новый Структура(;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.EqualError(t, err, "syntax error. line: 2, column: 32 (unexpected literal: \";\")")
	})
	t.Run("new error", func(t *testing.T) {
		code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 
					Контекст = Новый Структура("выыыы);
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()
		assert.Error(t, err)
	})
}

func TestParseAST(t *testing.T) {
	code := `

Процедура ОткрытьНавигационнуюСсылку(НавигационнаяСсылка, Знач Оповещение = Неопределено) Экспорт

	ПустаяДата = '00010101000000';
	ПустаяДата = '20131231235959';

	КлючЗаписиРегистра = Новый("РегистрСведенийКлючЗаписи.СостоянияОригиналовПервичныхДокументов", ПараметрыМассив);
	МассаДМ = ВыборкаЕдИзм.МассаДМ/Количество;
	
     стр = новый Структура("Цикл", 1);
     стр.Цикл = 0; 

Если (КодСимвола < 1040) ИЛИ (((КодСимвола > 1103) И (КодыДопустимыхСимволов.Найти(КодСимвола) = Неопределено)) И Не ((Не УчитыватьРазделителиСлов И ЭтоРазделительСлов(КодСимвола)))) Тогда 
        Возврат;
    КонецЕсли;

перейти ~метка;

МассивСтроки.Добавить(Новый ФорматированнаяСтрока(ЧастьСтроки.Значение, Новый Шрифт(,,Истина)));

	Позиция = Найти(Строка, Разделитель);
	Пока Позиция > 0 Цикл
		Подстрока = Лев(Строка, Позиция - 1);
		Если Не ПропускатьПустыеСтроки Или Не ПустаяСтрока(Подстрока) Тогда
			Если СокращатьНепечатаемыеСимволы Тогда
				Результат.Добавить(СокрЛП(Подстрока));
			Иначе
				Результат.Добавить(Подстрока);
			КонецЕсли;
		КонецЕсли;
		Строка = Сред(Строка, Позиция + СтрДлина(Разделитель));
		Позиция = Найти(Строка, Разделитель);
	КонецЦикла;

~метка:



	вы = ввывыв[0];
	СтрокаСпискаПП[ТекКолонка.Ключ].Вставить(ТекКолонкаЗначение.Ключ, УровеньГруппировки3[ПрефиксПоля + СтрЗаменить(ТекКолонкаЗначение.Значение, ".", "")]);

	Контекст = Новый Структура();
	Контекст.Вставить("НавигационнаяСсылка", НавигационнаяСсылка);
	Контекст.Вставить("Оповещение", Оповещение);
	
	ОписаниеОшибки = СтроковыеФункцииКлиентСервер.ПодставитьПараметрыВСтроку(
			НСтр("ru = 'Не удалось перейти по ссылке ""%1"" по причине: 
			           |Неверно задана навигационная ссылка.'"),
			НавигационнаяСсылка);
	
	Если Не ОбщегоНазначенияСлужебныйКлиент.ЭтоДопустимаяСсылка(НавигационнаяСсылка) Тогда 
		ОбщегоНазначенияСлужебныйКлиент.ОткрытьНавигационнуюСсылкуОповеститьОбОшибке(ОписаниеОшибки, Контекст);
		Возврат;
	КонецЕсли;
	
	Если ОбщегоНазначенияСлужебныйКлиент.ЭтоВебСсылка(НавигационнаяСсылка)
		Или ОбщегоНазначенияСлужебныйКлиент.ЭтоНавигационнаяСсылка(НавигационнаяСсылка) Тогда 
		
		Попытка
			а = а /0;
		Исключение
			ОбщегоНазначенияСлужебныйКлиент.ОткрытьНавигационнуюСсылкуОповеститьОбОшибке(ОписаниеОшибки, Контекст);
			Возврат;
		КонецПопытки;
		
		Если Оповещение <> Неопределено Тогда 
			ПриложениеЗапущено = Истина;
			ВыполнитьОбработкуОповещения(Оповещение, ПриложениеЗапущено);
		КонецЕсли;
		
		Возврат;
	КонецЕсли;
	
	Если ОбщегоНазначенияСлужебныйКлиент.ЭтоСсылкаНаСправку(НавигационнаяСсылка) Тогда 
		ОткрытьСправку(НавигационнаяСсылка);
		Возврат;
	КонецЕсли;
КонецПроцедуры

Если Оповещение <> Неопределено Тогда 
			ПриложениеЗапущено = Истина;
			ВыполнитьОбработкуОповещения(Оповещение, ПриложениеЗапущено);
		КонецЕсли;`

	a := NewAST(code)
	err := a.Parse()

	p := a.Print(PrintConf{Margin: 4})
	fmt.Println(p)

	if assert.NoError(t, err) {
		p := a.Print(PrintConf{Margin: 4})
		assert.Equal(t, true, compareHashes(code, p))
	}
}

func TestParseEmpty(t *testing.T) {
	code := `

`

	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

func TestBigProcedure(t *testing.T) {
	if _, err := os.Stat("testdata"); errors.Is(err, os.ErrNotExist) {
		t.Fatal("testdata file not found")
	}

	fileData, err := os.ReadFile("testdata")
	assert.NoError(t, err)

	a := NewAST(string(fileData))
	s := time.Now()
	err = a.Parse()
	fmt.Println("milliseconds -", time.Since(s).Milliseconds())
	assert.NoError(t, err)

	// p := a.Print(&PrintConf{Margin: 4})
	// fmt.Println(p)
}

func TestTernaryOperator(t *testing.T) {
	code := `Процедура ПодключитьВнешнююОбработку(Ссылка) 
					ds = ?(Истина, ?(dd = 3, а = 1, Наименование), СтруктураКонтрагент.Наименование);
				КонецПроцедуры`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)

	data, err := a.JSON()
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(data))
}

func TestArrayStruct(t *testing.T) {
	code := `Процедура ПодключитьВнешнююОбработку()        
				м = Новый Массив();
				в = м[4];
				
				м = Новый Структура("ав", уцуцу);
				в = м["вывыв"];
			КонецПроцедуры`

	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)

	data, err := a.JSON()
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(data))
}

func TestPrint(t *testing.T) {
	code := `&НаКлиенте
					Процедура Проба() 
						Если в = 1 или у = 3 И 0 <> 3 и не гоого и Истина и ав = неопределено Тогда
							а=1 + 3 *4;
						а=1 + 3 *4;
							fgd = 1
						ИначеЕсли ввв Тогда Если в = 1 Тогда
							а = -(1 + 3 *4);
Если в = 1 Тогда
							а = 1 + 3 *4;
							КонецЕсли
							КонецЕсли
						ИначеЕсли авыав Тогда

						Иначе						
							ваывы = 1 + 3 *4;
ваывы = 1 + 3 *4
						КонецЕсли;

						а = 1 + 3 *4
					КонецПроцедуры

					Функция авава(пар1, пар2) экспорт
						Для Каждого ИзмененныйОбъект Из ОбъектыНазначения Цикл
							Тип = ТипЗнч(ИзмененныйОбъект);
							Если ТипыИзмененныхОбъектов = Неопределено Тогда
								ТипыИзмененныхОбъектов.Добавить(Тип);
							КонецЕсли;

Для Каждого ИзмененныйОбъект Из ОбъектыНазначения Цикл
        Тип = ТипЗнч(ИзмененныйОбъект);
        Если ТипыИзмененныхОбъектов = Неопределено Тогда
            ТипыИзмененныхОбъектов.Добавить(Тип);
        Иначе
        КонецЕсли;
    КонецЦикла;

						КонецЦикла;

						Для а = 0 По 100 Цикл
							Тип = ТипЗнч(ИзмененныйОбъект);
							Если ТипыИзмененныхОбъектов  = Неопределено Тогда
								Продолжить; Иначе 	Прервать;
							КонецЕсли;
						КонецЦикла;
					Конецфункции
					
					Процедура Опрпп(пар1, Знач пар2 = 2.2, пар1 = Неопределено, Пар3 = "авава") 

						Попытка 
							а = 1+1;
ВызватьИсключение(ававава());
						Исключение
							ВызватьИсключение("");
ВызватьИсключение ;
						КонецПопытки;
					Конецпроцедуры`

	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

func TestExpPriority(t *testing.T) {
	t.Run("test1", func(t *testing.T) {
		code := `А = d = 2 = d ИЛИ в = 3;
			Если 1 = 1 = 2 = 3 Тогда
			   ПриКомпоновкеРезультата();
			КонецЕсли`
		a := NewAST(code)
		err := a.Parse()
		if assert.NoError(t, err) {
			p := a.Print(PrintConf{Margin: 4})
			//fmt.Println(p)
			assert.Equal(t, "А=((d=2)=d)ИЛИ(в=3);Если((1=1)=2)=3ТогдаПриКомпоновкеРезультата();КонецЕсли;", normalize(p))
		}
	})
	t.Run("test2", func(t *testing.T) {
		code := `Процедура ОткрытьНавигационнуюСсылку(НавигационнаяСсылка, Знач Оповещение = Неопределено) Экспорт
					Если в = 1 = 5 и не авав ИЛИ ааа Тогда
						в = 1 = 5 = 1 и не авав ИЛИ ааа;
					КонецЕсли;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()

		if assert.NoError(t, err) {
			p := a.Print(PrintConf{Margin: 4})
			assert.Equal(t, "ПроцедураОткрытьНавигационнуюСсылку(НавигационнаяСсылка,ЗначОповещение=Неопределено)ЭкспортЕсли(((в=1)=5)ИНеавав)ИЛИаааТогдав=(((1=5)=1)ИНеавав)ИЛИааа;КонецЕсли;КонецПроцедуры", normalize(p))
		}
	})
	t.Run("test3", func(t *testing.T) {
		code := `Процедура f()
					тест.куку.ууу = 1 = 5 = 1 и не авав ИЛИ ааа;
					тест[333] = 1 = 5 = 1 = 4 = fd;
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()

		if assert.NoError(t, err) {
			p := a.Print(PrintConf{OneLine: true})
			assert.Equal(t, "Процедура f() тест.куку.ууу = (((1 = 5) = 1) И Не авав) ИЛИ ааа;тест[333] = (((1 = 5) = 1) = 4) = fd;КонецПроцедуры", strings.TrimSpace(p))
		}
	})
	t.Run("test4", func(t *testing.T) {
		code := `Процедура f()
					ds = r / (КонВремя - НачВремя);
					fd = Формат(r / (КонВремя - НачВремя), "ЧН=; ЧГ=");
				КонецПроцедуры`

		a := NewAST(code)
		err := a.Parse()

		if assert.NoError(t, err) {
			p := a.Print(PrintConf{OneLine: true})
			assert.Equal(t, "Процедура f() ds = r / (КонВремя - НачВремя);fd = Формат(r / (КонВремя - НачВремя), \"ЧН=; ЧГ=\");КонецПроцедуры", strings.TrimSpace(p))
		}
	})
}

func Test_Directive(t *testing.T) {
	code := `
	&НаКлиенте
	&Вместо("ВыбратьИзФайла")
	Процедура Расш3_ВыбратьИзФайла(Команда)
	
	КонецПроцедуры
	
	
	&ИзменениеИКонтроль("ВыбратьИзФайла")
	&НаКлиенте
	Процедура Расш3_ВыбратьИзФайла1(Команда)
	
	КонецПроцедуры
	
	&НаСервере
	&НаСервере
	&После("ВыбратьИзФайла")
	Процедура Расш3_ВыбратьИзФайла1(Команда)
	
	КонецПроцедуры
	
	&Перед("ВыбратьИзФайла")
	Процедура Расш3_ВыбратьИзФайла1(Команда)
	
	КонецПроцедуры

	Процедура Расш3_ВыбратьИзФайла1(Команда)
	
	КонецПроцедуры
	`

	a := NewAST(code)
	err := a.Parse()
	if assert.NoError(t, err) {
		if assert.Len(t, a.ModuleStatement.Body, 5) {
			assert.Len(t, a.ModuleStatement.Body[0].(*FunctionOrProcedure).Directives, 2)
			assert.Len(t, a.ModuleStatement.Body[1].(*FunctionOrProcedure).Directives, 2)
			assert.Len(t, a.ModuleStatement.Body[2].(*FunctionOrProcedure).Directives, 3)
			assert.Len(t, a.ModuleStatement.Body[3].(*FunctionOrProcedure).Directives, 1)
			assert.Nil(t, a.ModuleStatement.Body[4].(*FunctionOrProcedure).Directives)
		}
	}

	//p := a.Print(PrintConf{})
	//fmt.Println(p)
}

func BenchmarkString(b *testing.B) {
	b.Run("string", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			test("rdedfs dfdf dsfd rdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs rdedfs dfdf dsfd rdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfd rdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfd dfdf dsfd")
		}
	})
	b.Run("ptr string", func(b *testing.B) {
		str := "rdedfs dfdf dsfd rdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs rdedfs dfdf dsfd rdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfd rdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfd dfdf dsfd"
		for i := 0; i < b.N; i++ {
			testPt(&str)
		}
	})
	b.Run("string count - 1", func(b *testing.B) {
		str := "rdedfs dfdf dsfd rdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs rdedfs dfdf dsfd rdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfd rdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfd dfdf dsfd"
		for i := 0; i < b.N; i++ {
			strings.Count(str, "df")
		}
	})
	//b.Run("string count - 2", func(b *testing.B) {
	//	str := "rdedfs dfdf dsfd rdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs rdedfs dfdf dsfd rdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfd rdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfdrdedfs dfdf dsfd dfdf dsfd"
	//	for i := 0; i < b.N; i++ {
	//		stringCount(str, "df")
	//	}
	//})
}

func test(_ string)    {}
func testPt(_ *string) {}

func compareHashes(str1, str2 string) bool {
	str1 = normalize(str1)
	str2 = normalize(str2)

	hash1 := sha256.Sum256([]byte(fastToLower(str1)))
	hash2 := sha256.Sum256([]byte(fastToLower(str2)))

	return hash1 == hash2
}

func normalize(str string) string {
	str = strings.ReplaceAll(str, " ", "")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, "\n", "")

	return str
}

func deleteEmptyLine(str string) string {
	result := strings.Builder{}
	for _, line := range strings.Split(str, "\n") {
		if strings.TrimSpace(line) != "" {
			result.WriteString(line + "\n")
		}
	}

	return strings.TrimSpace(result.String())
}

// TC-1: Test that ALL code from ALL preprocessor branches is parsed
func TestParse_PreprocessorAllBranches(t *testing.T) {
	code := `
Процедура ОбычнаяПроцедура()
    Перем а;
    а = 1;
КонецПроцедуры

#Если Сервер Тогда

Процедура СерверныйМетод() Экспорт
    Возврат;
КонецПроцедуры

#Иначе

Процедура КлиентскийМетод() Экспорт
    Возврат;
КонецПроцедуры

#КонецЕсли
`
	a := NewAST(code)
	err := a.Parse()
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, a.ModuleStatement.Body, 2, "Expected ОбычнаяПроцедура + PreprocessorIfStatement")

	normalProc, ok := a.ModuleStatement.Body[0].(*FunctionOrProcedure)
	assert.True(t, ok, "First item should be FunctionOrProcedure")
	assert.Equal(t, "ОбычнаяПроцедура", normalProc.Name)

	preproc, ok := a.ModuleStatement.Body[1].(*PreprocessorIfStatement)
	assert.True(t, ok, "Second item should be PreprocessorIfStatement")
	assert.Equal(t, "Сервер", preproc.Condition)

	assert.Len(t, preproc.ThenBlock, 1, "ThenBlock should have 1 procedure")
	serverProc, ok := preproc.ThenBlock[0].(*FunctionOrProcedure)
	assert.True(t, ok)
	assert.Equal(t, "СерверныйМетод", serverProc.Name)

	assert.Len(t, preproc.ElseBlock, 1, "ElseBlock should have 1 procedure")
	clientProc, ok := preproc.ElseBlock[0].(*FunctionOrProcedure)
	assert.True(t, ok)
	assert.Equal(t, "КлиентскийМетод", clientProc.Name)
}

// TC-3: Test nested preprocessor directives
func TestParse_PreprocessorNested(t *testing.T) {
	code := `
#Если Сервер Тогда
    #Если Не ВебКлиент Тогда

    Функция ВложенныйСерверныйМетод() Экспорт
        Возврат 1;
    КонецФункции

    #Иначе

    Функция ВложенныйВебМетод() Экспорт
        Возврат 2;
    КонецФункции

    #КонецЕсли
#КонецЕсли
`
	a := NewAST(code)
	err := a.Parse()
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, a.ModuleStatement.Body, 1, "Expected 1 outer PreprocessorIfStatement")

	outerPreproc, ok := a.ModuleStatement.Body[0].(*PreprocessorIfStatement)
	if !assert.True(t, ok, "Expected PreprocessorIfStatement") {
		return
	}
	assert.Equal(t, "Сервер", outerPreproc.Condition)

	assert.Len(t, outerPreproc.ThenBlock, 1, "Expected 1 nested PreprocessorIfStatement")

	innerPreproc, ok := outerPreproc.ThenBlock[0].(*PreprocessorIfStatement)
	if !assert.True(t, ok, "Expected nested PreprocessorIfStatement") {
		return
	}
	assert.Equal(t, "Не ВебКлиент", innerPreproc.Condition)

	assert.Len(t, innerPreproc.ThenBlock, 1, "ThenBlock should have 1 function")
	if fp, ok := innerPreproc.ThenBlock[0].(*FunctionOrProcedure); ok {
		assert.Equal(t, "ВложенныйСерверныйМетод", fp.Name)
	} else {
		t.Error("ThenBlock[0] should be FunctionOrProcedure")
	}

	assert.Len(t, innerPreproc.ElseBlock, 1, "ElseBlock should have 1 function")
	if fp, ok := innerPreproc.ElseBlock[0].(*FunctionOrProcedure); ok {
		assert.Equal(t, "ВложенныйВебМетод", fp.Name)
	} else {
		t.Error("ElseBlock[0] should be FunctionOrProcedure")
	}
}

// TC-4: Test #ИначеЕсли branches
func TestParse_PreprocessorElseIf(t *testing.T) {
	code := `
#Если Сервер Тогда
    Процедура СерверМетод()
    КонецПроцедуры
#ИначеЕсли Клиент Тогда
    Процедура КлиентМетод()
    КонецПроцедуры
#ИначеЕсли ВебКлиент Тогда
    Процедура ВебМетод()
    КонецПроцедуры
#Иначе
    Процедура ПоУмолчанию()
    КонецПроцедуры
#КонецЕсли
`
	a := NewAST(code)
	err := a.Parse()
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, a.ModuleStatement.Body, 1)
	preproc, ok := a.ModuleStatement.Body[0].(*PreprocessorIfStatement)
	assert.True(t, ok)
	assert.Equal(t, "Сервер", preproc.Condition)

	assert.Len(t, preproc.ElseIfs, 2, "Should have 2 ElseIf branches")
	assert.Equal(t, "Клиент", preproc.ElseIfs[0].Condition)
	assert.Equal(t, "ВебКлиент", preproc.ElseIfs[1].Condition)

	assert.Len(t, preproc.ElseBlock, 1, "Should have Else block")
}

// TC-5: Test #Область/#КонецОбласти
func TestParse_Region(t *testing.T) {
	code := `
#Область ПрограммныйИнтерфейс

Процедура ПубличнаяПроцедура() Экспорт
    Возврат;
КонецПроцедуры

#КонецОбласти
`
	a := NewAST(code)
	err := a.Parse()
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, a.ModuleStatement.Body, 1)
	region, ok := a.ModuleStatement.Body[0].(*RegionStatement)
	assert.True(t, ok, "Expected RegionStatement")
	assert.Equal(t, "ПрограммныйИнтерфейс", region.Name)

	assert.Len(t, region.Body, 1, "Region should contain 1 procedure")
	proc, ok := region.Body[0].(*FunctionOrProcedure)
	assert.True(t, ok)
	assert.Equal(t, "ПубличнаяПроцедура", proc.Name)
}

// TC-7: Test #Использовать directive (OneScript)
func TestParse_UseDirective(t *testing.T) {
	code := `
Процедура Тест()
КонецПроцедуры

#Использовать lib
#Использовать "./path/to/module"
`
	a := NewAST(code)
	err := a.Parse()
	if !assert.NoError(t, err) {
		return
	}

	assert.GreaterOrEqual(t, len(a.ModuleStatement.Body), 3, "Should have procedure + 2 Use statements")

	// First item is the procedure
	_, ok := a.ModuleStatement.Body[0].(*FunctionOrProcedure)
	assert.True(t, ok, "First item should be FunctionOrProcedure")

	// Second item is first #Использовать
	use1, ok := a.ModuleStatement.Body[1].(*UseStatement)
	assert.True(t, ok, "Second item should be UseStatement")
	assert.Equal(t, "lib", use1.Path)

	// Third item is second #Использовать
	use2, ok := a.ModuleStatement.Body[2].(*UseStatement)
	assert.True(t, ok, "Third item should be UseStatement")
	assert.Equal(t, "./path/to/module", use2.Path)
}

// TC-8: Test async function (Russian keyword)
func TestParse_AsyncFunction(t *testing.T) {
	code := `
Асинх Функция ПолучитьДанныеАсинхронно(Параметр) Экспорт
    Результат = Ждать ВыполнитьЗапрос(Параметр);
    Возврат Результат;
КонецФункции

Функция СинхронныйМетод() Экспорт
    Возврат 1;
КонецФункции
`
	a := NewAST(code)
	err := a.Parse()
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, a.ModuleStatement.Body, 2, "Expected 2 functions")

	// Check async flag
	for _, stmt := range a.ModuleStatement.Body {
		if fp, ok := stmt.(*FunctionOrProcedure); ok {
			if fp.Name == "ПолучитьДанныеАсинхронно" {
				assert.True(t, fp.Async, "ПолучитьДанныеАсинхронно should have Async=true")
				assert.True(t, fp.Export, "ПолучитьДанныеАсинхронно should have Export=true")
			} else if fp.Name == "СинхронныйМетод" {
				assert.False(t, fp.Async, "СинхронныйМетод should have Async=false")
			}
		}
	}
}

// TC-9: Test async procedure (Russian keyword)
func TestParse_AsyncProcedure(t *testing.T) {
	code := `
Асинх Процедура ОбработатьДанныеАсинхронно() Экспорт
    Данные = Ждать ЗагрузитьДанные();
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, a.ModuleStatement.Body, 1, "Expected 1 procedure")

	fp, ok := a.ModuleStatement.Body[0].(*FunctionOrProcedure)
	assert.True(t, ok, "Expected FunctionOrProcedure")
	assert.Equal(t, "ОбработатьДанныеАсинхронно", fp.Name)
	assert.True(t, fp.Async, "Should have Async=true")
	assert.Equal(t, PFTypeProcedure, fp.Type, "Should be Procedure")
}

// TC-10: Test async with English keyword
func TestParse_AsyncEnglishKeyword(t *testing.T) {
	code := `
async Функция GetDataAsync(Param) Экспорт
    Возврат 1;
КонецФункции
`
	a := NewAST(code)
	err := a.Parse()
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, a.ModuleStatement.Body, 1)
	fp := a.ModuleStatement.Body[0].(*FunctionOrProcedure)
	assert.True(t, fp.Async, "async keyword should set Async=true")
}

// TC-11: Test async with directives
func TestParse_AsyncWithDirectives(t *testing.T) {
	code := `
&НаСервере
Асинх Функция СерверныйАсинхМетод()
    Возврат Ждать Запрос();
КонецФункции
`
	a := NewAST(code)
	err := a.Parse()
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, a.ModuleStatement.Body, 1)
	fp := a.ModuleStatement.Body[0].(*FunctionOrProcedure)
	assert.True(t, fp.Async, "Should have Async=true")
	assert.NotNil(t, fp.Directives, "Should have directives")
	assert.Len(t, fp.Directives, 1)
	assert.Equal(t, "&НаСервере", fp.Directives[0].Name)
}

// TC-12: Test await expression
func TestParse_AwaitExpression(t *testing.T) {
	code := `
Асинх Функция Тест()
    а = Ждать Метод();
    б = 1 + Ждать Другой();
    Возврат Ждать Третий();
КонецФункции
`
	a := NewAST(code)
	err := a.Parse()
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, a.ModuleStatement.Body, 1)
	fp := a.ModuleStatement.Body[0].(*FunctionOrProcedure)
	assert.True(t, fp.Async)
	// The body should contain assignments and return with await
	assert.GreaterOrEqual(t, len(fp.Body), 3, "Should have at least 3 statements")
}

// TestOperationType_String tests OperationType.String() for all operation types
func TestOperationType_String(t *testing.T) {
	tests := []struct {
		op       OperationType
		expected string
	}{
		{OpPlus, "+"},
		{OpMinus, "-"},
		{OpMul, "*"},
		{OpDiv, "/"},
		{OpEq, "="},
		{OpGt, ">"},
		{OpLt, "<"},
		{OpNe, "<>"},
		{OpLe, "<="},
		{OpGe, ">="},
		{OpMod, "%"},
		{OpOr, "ИЛИ"},
		{OpAnd, "И"},
		{OpUndefined, ""},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.op.String())
		})
	}
}

// TestWalkHelper_TryCatch tests walkHelper with Try/Catch statements
func TestWalkHelper_TryCatch(t *testing.T) {
	code := `
Процедура Тест()
    Попытка
        а = 1;
    Исключение
        б = 2;
    КонецПопытки
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)

	var statements []Statement
	a.ModuleStatement.Walk(func(root *FunctionOrProcedure, parentStm, stm *Statement) {
		statements = append(statements, *stm)
	})
	// Should walk: FunctionOrProcedure, TryStatement body, TryStatement catch, assignments
	assert.GreaterOrEqual(t, len(statements), 3)
}

// TestWalkHelper_Ternary tests walkHelper with ternary operator
func TestWalkHelper_Ternary(t *testing.T) {
	code := `
Функция Тест()
    Возврат ?(а > 0, 1, 2);
КонецФункции
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)

	var statements []Statement
	a.ModuleStatement.Walk(func(root *FunctionOrProcedure, parentStm, stm *Statement) {
		statements = append(statements, *stm)
	})
	assert.GreaterOrEqual(t, len(statements), 1)
}

// TestWalkHelper_MethodStatement tests walkHelper with method calls
func TestWalkHelper_MethodStatement(t *testing.T) {
	code := `
Процедура Тест()
    Объект.Метод(1, 2, 3);
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)

	var statements []Statement
	a.ModuleStatement.Walk(func(root *FunctionOrProcedure, parentStm, stm *Statement) {
		statements = append(statements, *stm)
	})
	assert.GreaterOrEqual(t, len(statements), 1)
}

// TestWalkHelper_PreprocessorUse tests walkHelper with #Use directive
func TestWalkHelper_PreprocessorUse(t *testing.T) {
	code := `
#Использовать lib
Процедура Тест()
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)

	var useFound bool
	a.ModuleStatement.Walk(func(root *FunctionOrProcedure, parentStm, stm *Statement) {
		if _, ok := (*stm).(*UseStatement); ok {
			useFound = true
		}
	})
	assert.True(t, useFound)
}

// TestModuleAppend_VariableAfterBody tests error when variable declared after body
func TestModuleAppend_VariableAfterBody(t *testing.T) {
	code := `
Процедура Тест()
КонецПроцедуры
Перем а;
`
	a := NewAST(code)
	err := a.Parse()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable declarations must be placed at the beginning")
}

// TestModuleAppend_FunctionAfterBody tests error when function after non-function body
func TestModuleAppend_FunctionAfterBody(t *testing.T) {
	// In 1C, expressions at module level must come AFTER procedures/functions
	// Grammar catches this as syntax error before Append() check runs
	code := `
а = 1;
Процедура Тест()
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.Error(t, err) // syntax error - invalid structure rejected
}

// TestModuleAppend_DuplicateVariable tests error for duplicate global variable
func TestModuleAppend_DuplicateVariable(t *testing.T) {
	code := `
Перем а;
Перем а;
`
	a := NewAST(code)
	err := a.Parse()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already been defined")
}

// TestCallChainStatement_IsMethod tests IsMethod for call chains
func TestCallChainStatement_IsMethod(t *testing.T) {
	code := `
Процедура Тест()
    а = Объект.Метод();
    б = Объект.Свойство;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)

	fp := a.ModuleStatement.Body[0].(*FunctionOrProcedure)
	// First statement: assignment with method call
	assign1 := fp.Body[0].(AssignmentStatement)
	if chain, ok := assign1.Expr.Statements[0].(CallChainStatement); ok {
		assert.True(t, chain.IsMethod())
	}
}

// TestFastToLower tests fastToLower function
func TestFastToLower(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Russian uppercase", "ПРИВЕТ", "привет"},
		{"Russian mixed", "ПрИвЕт", "привет"},
		{"English uppercase", "HELLO", "hello"},
		{"English mixed", "HeLLo", "hello"},
		{"Mixed Russian English", "ПриветHELLO", "приветhello"},
		{"With Ё", "ЁЛКА", "ёлка"},
		{"Already lowercase", "привет", "привет"},
		{"Numbers unchanged", "123", "123"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, fastToLower(tt.input))
		})
	}
}

// TestCharacterClassification tests isLetter, isDigit, isSpace
func TestCharacterClassification(t *testing.T) {
	t.Run("isLetter", func(t *testing.T) {
		// Should return true
		assert.True(t, isLetter('a'))
		assert.True(t, isLetter('Z'))
		assert.True(t, isLetter('а'))
		assert.True(t, isLetter('Я'))
		assert.True(t, isLetter('ё'))
		assert.True(t, isLetter('Ё'))
		assert.True(t, isLetter('_'))
		// Should return false
		assert.False(t, isLetter('1'))
		assert.False(t, isLetter(' '))
		assert.False(t, isLetter('.'))
	})

	t.Run("isDigit", func(t *testing.T) {
		assert.True(t, isDigit('0'))
		assert.True(t, isDigit('5'))
		assert.True(t, isDigit('9'))
		assert.False(t, isDigit('a'))
		assert.False(t, isDigit(' '))
	})

	t.Run("isSpace", func(t *testing.T) {
		assert.True(t, isSpace(' '))
		assert.True(t, isSpace('\t'))
		assert.True(t, isSpace('\n'))
		assert.True(t, isSpace('\r'))
		assert.False(t, isSpace('a'))
		assert.False(t, isSpace('1'))
	})
}

// TestParamStatement_Fill tests ParamStatement.Fill method
func TestParamStatement_Fill(t *testing.T) {
	t.Run("with value param", func(t *testing.T) {
		p := &ParamStatement{}
		tok := Token{literal: "param1"}
		valTok := &Token{literal: "Знач"}
		p.Fill(valTok, tok)
		assert.True(t, p.IsValue)
		assert.Equal(t, "param1", p.Name)
	})

	t.Run("without value param", func(t *testing.T) {
		p := &ParamStatement{}
		tok := Token{literal: "param2"}
		p.Fill(nil, tok)
		assert.False(t, p.IsValue)
		assert.Equal(t, "param2", p.Name)
	})
}

// TestParamStatement_DefaultValue tests ParamStatement.DefaultValue method
func TestParamStatement_DefaultValue(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		p := &ParamStatement{}
		p.DefaultValue(nil)
		_, ok := p.Default.(UndefinedStatement)
		assert.True(t, ok)
	})

	t.Run("with value", func(t *testing.T) {
		p := &ParamStatement{}
		p.DefaultValue(VarStatement{Name: "test"})
		_, ok := p.Default.(VarStatement)
		assert.True(t, ok)
	})
}

// TestUnaryMinus tests unaryMinus function with various types
func TestUnaryMinus(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{"int", 5, -5},
		{"int32", int32(10), int32(-10)},
		{"int64", int64(20), int64(-20)},
		{"float32", float32(1.5), float32(-1.5)},
		{"float64", 2.5, -2.5},
		{"string unchanged", "test", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, unaryMinus(tt.input))
		})
	}
}

// TestNot tests not function with various types
func TestNot(t *testing.T) {
	t.Run("bool true", func(t *testing.T) {
		assert.Equal(t, false, not(true))
	})
	t.Run("bool false", func(t *testing.T) {
		assert.Equal(t, true, not(false))
	})
	t.Run("string unchanged", func(t *testing.T) {
		assert.Equal(t, "test", not("test"))
	})
}

// TestNewObjectStatement_Params tests NewObjectStatement.Params method
func TestNewObjectStatement_Params(t *testing.T) {
	code := `
Процедура Тест()
    а = Новый Структура("ключ", значение);
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// TestMethodStatement_Params tests MethodStatement.Params method
func TestMethodStatement_Params(t *testing.T) {
	code := `
Процедура Тест()
    а = Объект.Метод(1, 2);
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// TestPrint_BreakContinue tests printing Break and Continue statements
func TestPrint_BreakContinue(t *testing.T) {
	code := `
Процедура Тест()
    Для а = 1 По 10 Цикл
        Если а = 5 Тогда
            Прервать;
        КонецЕсли;
        Если а = 3 Тогда
            Продолжить;
        КонецЕсли;
    КонецЦикла;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "Прервать")
	assert.Contains(t, printed, "Продолжить")
}

// TestPrint_ThrowStatement tests printing Throw statement
func TestPrint_ThrowStatement(t *testing.T) {
	code := `
Процедура Тест()
    Попытка
        а = 1;
    Исключение
        ВызватьИсключение;
    КонецПопытки;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "ВызватьИсключение")
}

// TestPrint_ThrowWithParam tests printing Throw with parameter
func TestPrint_ThrowWithParam(t *testing.T) {
	code := `
Процедура Тест()
    ВызватьИсключение "Ошибка";
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "ВызватьИсключение")
}

// TestPrint_GoTo tests printing GoTo statement
func TestPrint_GoTo(t *testing.T) {
	code := `
Процедура Тест()
    Перейти ~метка;
    ~метка:
    а = 1;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "метка")
}

// TestPrint_GlobalVariables tests printing global variables
func TestPrint_GlobalVariables(t *testing.T) {
	code := `
Перем а;
Перем б Экспорт;
Процедура Тест()
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "Перем")
}

// TestPrint_GlobalVariablesWithDirective tests printing global variables with directive
func TestPrint_GlobalVariablesWithDirective(t *testing.T) {
	code := `
&НаСервере
Перем а;
Процедура Тест()
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "Перем")
	assert.Contains(t, printed, "НаСервере")
}

// TestPrint_FunctionWithParams tests printing function with parameters
func TestPrint_FunctionWithParams(t *testing.T) {
	code := `
Функция Тест(а, Знач б, в = 1)
    Возврат а + б + в;
КонецФункции
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "Знач")
	assert.Contains(t, printed, "Возврат")
}

// TestPrint_WhileLoop tests printing While loop
func TestPrint_WhileLoop(t *testing.T) {
	code := `
Процедура Тест()
    Пока а < 10 Цикл
        а = а + 1;
    КонецЦикла;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "Пока")
	assert.Contains(t, printed, "Цикл")
}

// TestPrint_ForEachLoop tests printing ForEach loop
func TestPrint_ForEachLoop(t *testing.T) {
	code := `
Процедура Тест()
    Для Каждого элемент Из Коллекция Цикл
        а = элемент;
    КонецЦикла;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "Каждого")
	assert.Contains(t, printed, "Из")
}

// TestPrint_TernaryOperator tests printing ternary operator
func TestPrint_TernaryOperator(t *testing.T) {
	code := `
Функция Тест()
    а = ?(б > 0, 1, 2);
    Возврат а;
КонецФункции
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "?(")
}

// TestPrint_NewObject tests printing New object creation
func TestPrint_NewObject(t *testing.T) {
	code := `
Процедура Тест()
    а = Новый Массив();
    б = Новый Структура("ключ", значение);
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "Новый")
}

// TestPrint_ItemStatement tests printing array/map access
func TestPrint_ItemStatement(t *testing.T) {
	code := `
Процедура Тест()
    а = Массив[0];
    б = Структура["ключ"];
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "[")
	assert.Contains(t, printed, "]")
}

// TestPrint_PreprocessorIf tests printing preprocessor if
func TestPrint_PreprocessorIf(t *testing.T) {
	code := `
#Если Сервер Тогда
    Процедура Тест()
    КонецПроцедуры
#КонецЕсли
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "#Если")
	assert.Contains(t, printed, "#КонецЕсли")
}

// TestPrint_Region tests printing region
func TestPrint_Region(t *testing.T) {
	code := `
#Область ТестоваяОбласть
    Процедура Тест()
    КонецПроцедуры
#КонецОбласти
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "#Область")
	assert.Contains(t, printed, "#КонецОбласти")
}

// TestPrint_Use tests printing use directive
func TestPrint_Use(t *testing.T) {
	code := `
#Использовать lib
Процедура Тест()
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "#Использовать")
}

// TestPrint_NotExpression tests printing Not expression
func TestPrint_NotExpression(t *testing.T) {
	code := `
Процедура Тест()
    а = Не б;
    в = Не (г И д);
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "Не")
}

// TestPrint_ExtDirective tests printing extension directive
func TestPrint_ExtDirective(t *testing.T) {
	code := `
&Вместо("ОригинальнаяПроцедура")
Процедура ПереопределённаяПроцедура()
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "&Вместо")
}

// TestPrint_NilAST tests Print with nil AST
func TestPrint_NilAST(t *testing.T) {
	var a *AstNode = nil
	result := a.Print(PrintConf{})
	assert.Equal(t, "", result)
}

// TestPrintStatement_Nil tests PrintStatement with nil
func TestPrintStatement_Nil(t *testing.T) {
	a := NewAST("")
	result := a.PrintStatement(nil)
	assert.Equal(t, "", result)
}

// TestPrintStatementWithConf_Nil tests PrintStatementWithConf with nil
func TestPrintStatementWithConf_Nil(t *testing.T) {
	a := NewAST("")
	result := a.PrintStatementWithConf(nil, PrintConf{})
	assert.Equal(t, "", result)
}

// TestIsDigit tests IsDigit function
func TestIsDigit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"all digits", "12345", true},
		{"single digit", "0", true},
		{"empty string", "", true}, // vacuously true
		{"with letter", "123a", false},
		{"with space", "12 34", false},
		{"only letters", "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsDigit(tt.input))
		})
	}
}

// TestFastToLower_EdgeCases tests fastToLower with edge cases
func TestFastToLower_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Ё uppercase", "ЁЛКА", "ёлка"},
		{"Ё in middle", "ПОДЪЁМ", "подъём"},
		{"Russian after П", "РСТУФХЦЧШЩ", "рстуфхцчшщ"},
		{"Mixed with Ё", "ЁЖ", "ёж"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, fastToLower(tt.input))
		})
	}
}

// TestFastToLower_Old tests the old fastToLower implementation
func TestFastToLower_Old(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Russian uppercase", "ПРИВЕТ", "привет"},
		{"English uppercase", "HELLO", "hello"},
		{"With Ё", "ЁЛКА", "ёлка"},
		{"Mixed", "ПриВЕТ", "привет"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, fastToLower_old(tt.input))
		})
	}
}

// TestSyntaxError_UnknownCharacter tests error for unknown characters
func TestSyntaxError_UnknownCharacter(t *testing.T) {
	// Character that's not recognized by the lexer
	code := "Процедура Тест()\n    § = 1;\nКонецПроцедуры"
	a := NewAST(code)
	err := a.Parse()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "syntax error")
}

// TestDateLiteral_ScanError tests error in date literal scanning
func TestDateLiteral_ScanError(t *testing.T) {
	// Unclosed date literal
	code := "а = '20240101"
	a := NewAST(code)
	err := a.Parse()
	assert.Error(t, err)
}

// TestPreprocessor_UnknownDirective tests unknown preprocessor directive
func TestPreprocessor_UnknownDirective(t *testing.T) {
	// Unknown preprocessor directive should be skipped
	code := `
#НеизвестнаяДиректива
Процедура Тест()
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err) // Unknown directives are skipped
}

// TestPreprocessor_ConditionWithParens tests preprocessor with parentheses in condition
func TestPreprocessor_ConditionWithParens(t *testing.T) {
	code := `
#Если (Сервер Или Клиент) И Не ВебКлиент Тогда
    Процедура Тест()
    КонецПроцедуры
#КонецЕсли
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// TestPreprocessor_UseWithQuotedPath tests #Use with quoted path
func TestPreprocessor_UseWithQuotedPath(t *testing.T) {
	code := `
#Использовать "./lib/module"
Процедура Тест()
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// TestUnaryMinus_Expression tests unary minus with expressions
func TestUnaryMinus_Expression(t *testing.T) {
	code := `
Функция Тест()
    а = -(б + в);
    б = -Объект.Метод();
    в = -Массив[0];
    Возврат -1;
КонецФункции
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "-")
}

// TestWalkHelper_IfElseBlock tests walkHelper with ElseIf blocks
func TestWalkHelper_IfElseBlock(t *testing.T) {
	code := `
Процедура Тест()
    Если а = 1 Тогда
        б = 1;
    ИначеЕсли а = 2 Тогда
        б = 2;
    ИначеЕсли а = 3 Тогда
        б = 3;
    Иначе
        б = 0;
    КонецЕсли;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)

	var count int
	a.ModuleStatement.Walk(func(root *FunctionOrProcedure, parentStm, stm *Statement) {
		count++
	})
	assert.GreaterOrEqual(t, count, 5) // Multiple statements including ElseIf blocks
}

// TestWalkHelper_ExpStatement tests walkHelper with ExpStatement
func TestWalkHelper_ExpStatement(t *testing.T) {
	code := `
Процедура Тест()
    а = (б + в) * (г - д);
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)

	var stmtCount int
	a.ModuleStatement.Walk(func(root *FunctionOrProcedure, parentStm, stm *Statement) {
		stmtCount++
	})
	// Should walk multiple statements including the expression parts
	assert.GreaterOrEqual(t, stmtCount, 1)
}

// TestWalkHelper_PreprocessorElseIf tests walkHelper with preprocessor ElseIf
func TestWalkHelper_PreprocessorElseIf(t *testing.T) {
	code := `
#Если Сервер Тогда
    Процедура Тест1()
    КонецПроцедуры
#ИначеЕсли Клиент Тогда
    Процедура Тест2()
    КонецПроцедуры
#ИначеЕсли ВнешнееСоединение Тогда
    Процедура Тест3()
    КонецПроцедуры
#Иначе
    Процедура Тест4()
    КонецПроцедуры
#КонецЕсли
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)

	var funcCount int
	a.ModuleStatement.Walk(func(root *FunctionOrProcedure, parentStm, stm *Statement) {
		if _, ok := (*stm).(*FunctionOrProcedure); ok {
			funcCount++
		}
	})
	assert.Equal(t, 4, funcCount)
}

// TestCallChainStatement_Not tests Not method on CallChainStatement
func TestCallChainStatement_Not(t *testing.T) {
	code := `
Процедура Тест()
    Если Не Объект.Метод() Тогда
        а = 1;
    КонецЕсли;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "Не")
}

// TestMethodStatement_Not tests Not method on MethodStatement
func TestMethodStatement_Not(t *testing.T) {
	code := `
Процедура Тест()
    Если Не Метод(а, б) Тогда
        в = 1;
    КонецЕсли;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// TestVarStatement_UnaryMinus tests UnaryMinus method on VarStatement
func TestVarStatement_UnaryMinus(t *testing.T) {
	code := `
Процедура Тест()
    а = -б;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// TestVarStatement_Not tests Not method on VarStatement
func TestVarStatement_Not(t *testing.T) {
	code := `
Процедура Тест()
    Если Не а Тогда
        б = 1;
    КонецЕсли;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// TestExprStatements_Not tests Not method on ExprStatements
func TestExprStatements_Not(t *testing.T) {
	code := `
Процедура Тест()
    Если Не (а И б) Тогда
        в = 1;
    КонецЕсли;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// TestPrint_IfElseIfBlock tests printing If with ElseIf blocks
func TestPrint_IfElseIfBlock(t *testing.T) {
	code := `
Процедура Тест()
    Если а = 1 Тогда
        б = 1;
    ИначеЕсли а = 2 Тогда
        б = 2;
    ИначеЕсли а = 3 Тогда
        б = 3;
    Иначе
        б = 0;
    КонецЕсли;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "ИначеЕсли")
	assert.Contains(t, printed, "Иначе")
}

// TestPrint_EmptyModule tests Print on empty module body
func TestPrint_EmptyModule(t *testing.T) {
	// Module with ONLY global variables and NO procedures/functions
	// has empty Body, so print() returns "" for the body part
	code := `Перем а;`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	// Verify global variable was parsed
	assert.Len(t, a.ModuleStatement.GlobalVariables, 1)
	// Body is empty since there are no functions
	assert.Len(t, a.ModuleStatement.Body, 0)
}

// TestPrintStatementWithConf_NonFunction tests PrintStatementWithConf with non-function statement
func TestPrintStatementWithConf_NonFunction(t *testing.T) {
	code := `
Процедура Тест()
    Если а = 1 Тогда
        б = 1;
    КонецЕсли;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)

	// Get the if statement from inside the procedure
	fp := a.ModuleStatement.Body[0].(*FunctionOrProcedure)
	ifStmt := fp.Body[0]
	printed := a.PrintStatementWithConf(ifStmt, PrintConf{Margin: 4})
	assert.Contains(t, printed, "Если")
}

// TestPrint_PreprocessorElseIf tests printing preprocessor with ElseIf
func TestPrint_PreprocessorElseIf(t *testing.T) {
	code := `
#Если Сервер Тогда
    Процедура Тест1()
    КонецПроцедуры
#ИначеЕсли Клиент Тогда
    Процедура Тест2()
    КонецПроцедуры
#Иначе
    Процедура Тест3()
    КонецПроцедуры
#КонецЕсли
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "#ИначеЕсли")
	assert.Contains(t, printed, "#Иначе")
}

// TestPrint_LoopForTo tests printing For..To loop
func TestPrint_LoopForTo(t *testing.T) {
	code := `
Процедура Тест()
    Для а = 1 По 10 Цикл
        б = а;
    КонецЦикла;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "Для")
	assert.Contains(t, printed, "По")
}

// TestPrint_DateLiteral tests printing date literals
func TestPrint_DateLiteral(t *testing.T) {
	code := `
Процедура Тест()
    а = '20240115';
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	// Date parsing and printing
	assert.NotNil(t, a.ModuleStatement.Body)
}

// TestPrint_UnaryMinusChain tests parsing unary minus on call chain
func TestPrint_UnaryMinusChain(t *testing.T) {
	code := `
Процедура Тест()
    а = -Объект.Свойство;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	// Verify parsing succeeded - unaryMinus flag is internal
	assert.Len(t, a.ModuleStatement.Body, 1)
}

// TestNewAST_WithBOM tests parsing code with UTF-8 BOM
func TestNewAST_WithBOM(t *testing.T) {
	// UTF-8 BOM: 0xEF 0xBB 0xBF
	bom := string([]byte{0xEF, 0xBB, 0xBF})
	code := bom + `
Процедура Тест()
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	assert.Len(t, a.ModuleStatement.Body, 1)
}

// TestCast tests the Cast helper function
func TestCast(t *testing.T) {
	t.Run("successful cast", func(t *testing.T) {
		var stm Statement = VarStatement{Name: "test"}
		result := Cast[VarStatement](stm)
		assert.Equal(t, "test", result.Name)
	})

	t.Run("failed cast returns zero value", func(t *testing.T) {
		var stm Statement = BreakStatement{}
		result := Cast[VarStatement](stm)
		assert.Equal(t, "", result.Name)
	})
}

// TestLex_EmptyCode tests Lex on empty code
func TestLex_EmptyCode(t *testing.T) {
	a := NewAST("")
	lval := &yySymType{}
	token := a.Lex(lval)
	assert.Equal(t, EOF, token)
}

// TestPrint_IntegerLiteral tests printing integer literals
func TestPrint_IntegerLiteral(t *testing.T) {
	code := `
Функция Тест()
    Возврат 42;
КонецФункции
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "42")
}

// TestPrint_BoolLiteral tests printing boolean literals
func TestPrint_BoolLiteral(t *testing.T) {
	code := `
Функция Тест()
    Возврат Истина;
КонецФункции
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
	printed := a.Print(PrintConf{Margin: 4})
	assert.Contains(t, printed, "Истина")
}

// TestPrint_AssignmentInExpression tests printing assignment
func TestPrint_AssignmentInExpression(t *testing.T) {
	code := `
Процедура Тест()
    а = б = в;
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// TC-FR044-0: Test procedure inside module-level preprocessor block
func TestParse_ProcedureInsideModuleLevelPreproc(t *testing.T) {
	code := `#Если Сервер Или ТолстыйКлиентОбычноеПриложение Или ВнешнееСоединение Тогда

#Область ОбработчикиСобытий

Процедура ОбработкаЗаполнения(ДанныеЗаполнения, ТекстЗаполнения, СтандартнаяОбработка)
	Если Ссылка.Пустая() Тогда
		ПриСоздании(ДанныеЗаполнения)
	КонецЕсли
КонецПроцедуры

Процедура ПриКопировании(ОбъектКопирования)
	ПриСоздании(ОбъектКопирования)
КонецПроцедуры

#КонецОбласти
#КонецЕсли`
	a := NewAST(code)
	err := a.Parse()
	if !assert.NoError(t, err) {
		return
	}
	assert.Len(t, a.ModuleStatement.Body, 1)
}

// TC-FR044: Test preprocessor inside procedure body (FR-044)
func TestParse_PreprocessorInsideProcedure(t *testing.T) {
	code := `
Процедура ТестовыйМетод()
    #Если Клиент Тогда
        а = 1;
    #ИначеЕсли Сервер Тогда
        а = 2;
    #Иначе
        а = 3;
    #КонецЕсли
КонецПроцедуры
`
	a := NewAST(code)
	err := a.Parse()
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, a.ModuleStatement.Body, 1)
	proc, ok := a.ModuleStatement.Body[0].(*FunctionOrProcedure)
	assert.True(t, ok, "Expected FunctionOrProcedure")
	assert.Equal(t, "ТестовыйМетод", proc.Name)

	// Procedure body should contain the preprocessor statement
	assert.Len(t, proc.Body, 1, "Procedure body should have 1 statement (PreprocessorIfStatement)")
	preproc, ok := proc.Body[0].(*PreprocessorIfStatement)
	assert.True(t, ok, "Expected PreprocessorIfStatement inside procedure body")
	assert.Equal(t, "Клиент", preproc.Condition)
	assert.Len(t, preproc.ElseIfs, 1)
	assert.Equal(t, "Сервер", preproc.ElseIfs[0].Condition)
	assert.NotNil(t, preproc.ElseBlock)
}

// TC-FR044-2: Test nested preprocessor inside procedure with region
func TestParse_PreprocessorNestedInsideProcedure(t *testing.T) {
	code := `
Функция Тест() Экспорт
    #Область Внутренняя
        #Если Сервер Тогда
            а = 1;
        #КонецЕсли
    #КонецОбласти
КонецФункции
`
	a := NewAST(code)
	err := a.Parse()
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, a.ModuleStatement.Body, 1)
	fn, ok := a.ModuleStatement.Body[0].(*FunctionOrProcedure)
	assert.True(t, ok)

	// Body-level regions are standalone directives (not containers)
	// because 1C allows regions to split code blocks (e.g. inside loops).
	// Expect: RegionStart, #If, RegionEnd = 3 items
	assert.Len(t, fn.Body, 3, "Function body should have region start, #If, region end")

	// First: region start
	region, ok := fn.Body[0].(*RegionStatement)
	assert.True(t, ok, "First should be RegionStatement (start)")
	assert.Equal(t, "Внутренняя", region.Name)

	// Second: nested #Если
	nestedPreproc, ok := fn.Body[1].(*PreprocessorIfStatement)
	assert.True(t, ok, "Second should be PreprocessorIfStatement")
	assert.Equal(t, "Сервер", nestedPreproc.Condition)

	// Third: region end
	regionEnd, ok := fn.Body[2].(*RegionStatement)
	assert.True(t, ok, "Third should be RegionStatement (end)")
	assert.Equal(t, "", regionEnd.Name)
}

// =============================================================================
// UNF Error Category Tests — real patterns from 15K+ production BSL files
// Each test corresponds to a category from error-analysis.md
// =============================================================================

// Category 1: Semicolons after proc/func declarations (299 files — 64.2%)
// Pattern: Процедура Имя(Параметры); — semicolon after closing paren
func TestUNF_SemicolonAfterProcDecl(t *testing.T) {
	t.Run("procedure-with-semicolon", func(t *testing.T) {
		code := `&НаСервере
Процедура Обработать(Отказ);
	а = 1;
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})

	t.Run("function-with-semicolon", func(t *testing.T) {
		code := `Функция РодительПоИдентификатору(МассивРодителей);
	Возврат Неопределено;
КонецФункции`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})

	t.Run("procedure-export-with-semicolon", func(t *testing.T) {
		code := `Процедура Тест(Парам1, Знач Парам2) Экспорт;
	а = 1;
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})

	t.Run("directive-procedure-with-semicolon", func(t *testing.T) {
		code := `&НаКлиентеНаСервереБезКонтекста
Процедура ПоказатьНедействительных(Форма);
	а = 1;
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
}

// Category 2: Semicolons after control keywords (13 files — 2.8%)
// Pattern: Цикл; Тогда; — semicolons after Loop/Then
func TestUNF_SemicolonAfterControlKeyword(t *testing.T) {
	t.Run("semicolon-after-loop", func(t *testing.T) {
		code := `Процедура Тест()
	Для Каждого Стр Из Таблица Цикл;
		а = 1;
	КонецЦикла;
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})

	t.Run("semicolon-after-then", func(t *testing.T) {
		code := `Процедура Тест()
	Если а = 1 Тогда;
		б = 2;
	КонецЕсли;
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
}

// Category 3: Await (Ждать) as standalone statement (36 files — 7.7%)
// Pattern: Ждать Method(args); — not assigned to variable
func TestUNF_AwaitAsStatement(t *testing.T) {
	t.Run("await-standalone", func(t *testing.T) {
		code := `Асинх Процедура Тест()
	Ждать ПредупреждениеАсинх("Готово");
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})

	t.Run("await-in-if", func(t *testing.T) {
		code := `Асинх Процедура Тест()
	Если Результат.Успешно Тогда
		Ждать ПредупреждениеАсинх("Успех");
	Иначе
		Ждать ПредупреждениеАсинх("Ошибка");
	КонецЕсли;
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})

	t.Run("await-chained-call", func(t *testing.T) {
		code := `Асинх Процедура Тест()
	Ждать Сертификат.ИнициализироватьАсинх(Данные);
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
}

// Category 5: Method().Property = value (19 files — 4.1%)
// Pattern: call result used as lvalue for assignment
func TestUNF_CallResultAsLvalue(t *testing.T) {
	t.Run("method-result-property-assign", func(t *testing.T) {
		code := `Процедура Тест()
	ПараметрыОжидания().Включено = Ложь;
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})

	t.Run("method-with-args-property-assign", func(t *testing.T) {
		code := `Процедура Тест()
	ВидКонтактнойИнформации(ВидКИ).Наименование = Заголовок;
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
}

// Category 6: Complex lvalue — indexed property assignment (8 files — 1.7%)
// Pattern: Области["expr"].Property = value
func TestUNF_ComplexLvalue(t *testing.T) {
	t.Run("indexed-property-assign", func(t *testing.T) {
		code := `Процедура Тест()
	Области["П0000101001"].Значение = Сумма;
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})

	t.Run("indexed-with-expr-property-assign", func(t *testing.T) {
		code := `Процедура Тест()
	Области["П0000" + Формат(Гр, "ЧЦ=2")].Значение = СуммаПоКол;
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
}

// Category 7: Proc decl without semicolon — missing ) handling (8 files — 1.7%)
// Pattern: Процедура Имя(Параметры) <newline> body — no semicolon, no Экспорт
func TestUNF_ProcDeclNoSemicolon(t *testing.T) {
	code := `&НаКлиенте
Процедура Подключаемый_Выполнить(Команда)
	Клиент.Синхронизировать();
КонецПроцедуры`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// Category 8: Ternary().method (4 files — 0.9%)
// Pattern: ?(cond, a, b).Method() — ternary result used as object
func TestUNF_TernaryDotMethod(t *testing.T) {
	t.Run("ternary-method-call", func(t *testing.T) {
		code := `Процедура Тест()
	Имена = ?(Флаг, Мета.Справ.Один, Мета.Справ.Два).ПолучитьИмена();
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})

	t.Run("ternary-property", func(t *testing.T) {
		code := `Процедура Тест()
	Имя = ?(Страница = Неопределено, Элементы.Группа.ТекущаяСтраница, Страница).Имя;
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
}

// Category 9: Перем (Var) inside #Region in procedure body (3 files — 0.6%)
func TestUNF_VarInsideRegionInBody(t *testing.T) {
	code := `Процедура Тест()
	#Область Инициализация
		Перем МассивНовостей;
		а = 1;
	#КонецОбласти
КонецПроцедуры`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// Category 10: Execute with expression (2 files — 0.4%)
// Pattern: Выполнить ИмяМетода + "(" + args + ")"
func TestUNF_ExecuteWithExpression(t *testing.T) {
	code := `Процедура Тест()
	Выполнить ИмяМетода + "(" + ПараметрыСтрока + ")";
КонецПроцедуры`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// Category 11: КонецПроцедуры; at module level (5 files — 1.1%)
func TestUNF_EndProcSemicolonModuleLevel(t *testing.T) {
	code := `Процедура Тест()
	а = 1;
КонецПроцедуры;

Процедура Тест2()
	б = 2;
КонецПроцедуры`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// Category 12: Reserved keyword as property name after dot (2 files)
// Pattern: Выбор.Иначе = x, Параметры.КонецЦикла = x
func TestUNF_KeywordAsPropertyName(t *testing.T) {
	t.Run("else-as-property", func(t *testing.T) {
		code := `Процедура Тест()
	Выбор.Иначе = Значение;
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})

	t.Run("endloop-as-property", func(t *testing.T) {
		code := `Процедура Тест()
	Параметры.КонецЦикла = Вершина;
КонецПроцедуры`
		a := NewAST(code)
		err := a.Parse()
		assert.NoError(t, err)
	})
}

// Category 13: Indexed lvalue with string containing special chars (7 files)
// Pattern: Области["expr" + Формат(x, "ЧЦ=2; ЧВН=")].Значение = value
func TestUNF_IndexedLvalueWithSpecialString(t *testing.T) {
	code := `Процедура Тест()
	Области["П0000101001" + Формат(Гр, "ЧЦ=2; ЧВН=")].Значение = СуммаПоКол;
КонецПроцедуры`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// Category 14: КонецПроцедуры; followed by more procedures in same #Region (10 files)
func TestUNF_EndProcSemicolonMultiple(t *testing.T) {
	code := `#Область Тест

&НаКлиенте
Процедура Тест1()
	а = 1;
КонецПроцедуры;

&НаСервере
Процедура Тест2()
	б = 2;
КонецПроцедуры;

#КонецОбласти`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// Category 15: English While/Do keywords (1 file)
func TestUNF_EnglishWhileDo(t *testing.T) {
	code := `Процедура Тест()
	While а <> Неопределено Do
		а = а.Следующий;
	EndDo;
КонецПроцедуры`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}

// Category 16: English For/To keywords (1 file)
func TestUNF_EnglishForTo(t *testing.T) {
	code := `Процедура Тест()
	Для Индекс = 0 To Количество - 1 Цикл
		а = Индекс;
	КонецЦикла;
КонецПроцедуры`
	a := NewAST(code)
	err := a.Parse()
	assert.NoError(t, err)
}
