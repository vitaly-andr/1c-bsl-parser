%{
package ast
%}

%type<body> body
%type<opt_body> opt_body
%type<stmt> stmt
%type<stmt_loop> stmt_loop
%type<funcProc> funcProc
%type<stmt_if> stmt_if
%type<opt_elseif_list> opt_elseif_list
%type<opt_else> opt_else
%type<stmt> opt_stmt
%type<exprs> exprs 
%type<stmt> expr
%type<opt_export> opt_export
%type<stmt> simple_expr
%type<declarations_method_params> declarations_method_params
%type<declarations_method_param> declarations_method_param
%type<stmt> opt_expr
%type<stmt> execute_param
%type<stmt> through_dot
%type<stmt> loopExp
%type<stmt> new_object
%type<stmt> ternary
%type<opt_explicit_variables> opt_explicit_variables
%type<explicit_variables> explicit_variables
%type<identifiers> identifiers
%type<stmt> stmt_tryCatch
%type<stmt> identifier
%type<goToLabel> goToLabel
%type<token> separator
%type<token> semicolon
%type<token> colon
%type<token> ':'
%type<token> ';'
%type<token> ','
%type<global_variables> global_variables
%type<token> comma
%type<directive> one_directive
%type<directive> opt_directive
%type<directives> directives
%type<directives> opt_directives
%type<stmt> lvalue_rest
%type<stmt> call_rest
%type<stmt> call_chain
%type<stmt> call_part

%union {
    token Token
    stmt_if *IfStatement
    opt_elseif_list Statements
    opt_else Statements
    stmt    Statement
    stmt_loop *LoopStatement
    funcProc *FunctionOrProcedure
    body Statements
    opt_body Statements
    declarations_method_params []ParamStatement
    declarations_method_param ParamStatement
    exprs ExprStatements
    opt_export *Token
    explicit_variables map[string]VarStatement
    global_variables []GlobalVariables
    opt_explicit_variables map[string]VarStatement
    identifiers []Token
    goToLabel *GoToLabelStatement
    opt_goToLabel *GoToLabelStatement
    directive *DirectiveStatement
    directives []*DirectiveStatement
}

%token<token> Directive ExtDirective token_identifier Procedure Var EndProcedure If Then ElseIf Else EndIf For Each In To Loop EndLoop Break Not ValueParam While GoToLabel
%token<token> Continue Try Catch EndTry Number String New Function EndFunction Return Throw NeEQ EQUAL LE GE OR And True False Undefined Export Date GoTo Execute
%token<token> LVALUE_IDENT CALL_IDENT

%nonassoc LOW_PREC /* самый низкий приоритет */
%left OR
%left And
%left NeEQ
%left LE
%left GE
%right Not
%left EQUAL
%left '>' '<'
%left '+' '-'
%left '*' '/' '%'
%right UNARMinus UNARYPlus /* самый высокий приоритет */

%%


module: /* empty */ {  }
    |body {
         if ast, ok := yylex.(*AstNode); ok {
            ast.ModuleStatement.Append($1, yylex)
        }
    }
    | main_items opt_body {
         if ast, ok := yylex.(*AstNode); ok {
            ast.ModuleStatement.Append($2, yylex)
        }
    };

main_items: main
    | main_items main
;

main: global_variables {  
        if ast, ok := yylex.(*AstNode); ok {
            ast.ModuleStatement.Append($1, yylex)
        }
    }
    | funcProc {
        if ast, ok := yylex.(*AstNode); ok {
            ast.ModuleStatement.Append($1, yylex)
        }
    }
;


/* Директивы - refactored to eliminate shift/reduce conflicts */
one_directive: Directive { $$ = &DirectiveStatement{ Name: $1.literal }}
        | ExtDirective '(' String ')' { $$ = &DirectiveStatement{ Name: $1.literal, Src: $3.literal }}
;

opt_directive: { $$ = nil }
        | one_directive { $$ = $1 }
;

directives: one_directive { $$ = []*DirectiveStatement{$1} }
         | directives one_directive { $$ = append($$, $2) }
;

opt_directives: { $$ = nil }
        | directives { $$ = $1 }
;

opt_export: { $$ = nil}
        | Export { $$ = &$1}
;

global_variables: opt_directive Var identifiers opt_export semicolon {
        $$ = make([]GlobalVariables,  len($3), len($3))
        for i, v := range $3 {
            if $1 != nil {
                $$[i].Directive = $1
            }

            $$[i].Export = $4 != nil 
            $$[i].Var = VarStatement { Name: v.literal }
        }
};


funcProc: opt_directives Function token_identifier '(' declarations_method_params ')' opt_export { isFunction(true, yylex) } opt_explicit_variables opt_body EndFunction
        {
            $$ = createFunctionOrProcedure(PFTypeFunction, $1, $3.literal, $5, $7, $9, $10)
            isFunction(false, yylex)
        }
        | opt_directives Procedure token_identifier '(' declarations_method_params ')' opt_export opt_explicit_variables opt_body EndProcedure
        {
            $$ = createFunctionOrProcedure(PFTypeProcedure, $1, $3.literal, $5, $7, $8, $9)
        }
;

opt_body: { $$ = nil }
	| body { $$ = $1 }
;
    

body: stmt { $$ = Statements{$1}}
    | opt_body separator opt_stmt { 
        if $2.literal == ":" && len($1) > 0 {
            if _, ok := $1[len($1)-1].(*GoToLabelStatement); !ok {
                yylex.Error("semicolon (;) is expected")
            }
        }
        if $3 != nil {
            $$ = append($$, $3) 
        }
    }
    
;

opt_stmt: { $$ = nil }
        | stmt { $$ = $1 }
;

separator: semicolon { $$ = $1} | colon { $$ = $1};


/* переменные */ 
opt_explicit_variables: { $$ = map[string]VarStatement{} }
            | explicit_variables { $$ = $1 }
;

explicit_variables: Var identifiers semicolon { 
                    if vars, err := appendVarStatements(map[string]VarStatement{}, $2); err != nil {
                        yylex.Error(err.Error()) 
                    } else {
                        $$ = vars
                    }
                }
            | explicit_variables Var identifiers semicolon {
                    if vars, err := appendVarStatements($1, $3); err != nil {
                        yylex.Error(err.Error()) 
                    } else {
                        $$ = vars
                    }
                }
;


/* Если Конецесли */
stmt_if : If expr Then opt_body opt_elseif_list opt_else EndIf {
    $$ = &IfStatement {
        Expression: $2,
        TrueBlock:  $4,
        IfElseBlock: $5,
        ElseBlock: $6,
    }
};

/* ИначеЕсли */
opt_elseif_list : { $$ = Statements{} }
        | ElseIf expr Then opt_body opt_elseif_list {
             $$ = append($5, &IfStatement{
                Expression: $2,
                TrueBlock:  $4,
            })
        };

/* Иначе */
opt_else : { $$ = nil }
        | Else opt_body { $$ = $2 };

/* тернарный оператор */
ternary: '?' '(' expr comma expr comma expr ')' {
    $$ = TernaryStatement{
            Expression: $3,
            TrueBlock: $5,
            ElseBlock: $7,
        } 
};

/* циклы */
stmt_loop: For Each token_identifier In loopExp Loop { setLoopFlag(true, yylex) } opt_body EndLoop {
        $$ = &LoopStatement{
            For: $3.literal,
            In: $5,
            Body: $8,
        }
        setLoopFlag(false, yylex) 
    } 
    | For expr To expr Loop { setLoopFlag(true, yylex) } opt_body EndLoop {
        $$ = &LoopStatement{
            For: $2,
            To: $4,
            Body: $7,
        }
        setLoopFlag(false, yylex)
    }
    | While expr Loop { setLoopFlag(true, yylex) } opt_body EndLoop {
        $$ = &LoopStatement{
            WhileExpr: $2,
            Body: $5,
        }
};


/* описыввает выражения которые можно использовать в циккле Для Каждого */
loopExp: through_dot { $$ = $1 }
        | new_object { $$ = $1 }
        |'(' new_object ')' { $$ = $2 }
        | ternary { $$ = $1 }
;


/* ═══════════════════════════════════════════════════════════════════
   lvalue_rest — продолжение цепочки для левой части присваивания
   Может содержать свойства, индексы, и даже методы посередине:
   а.б, а[0], а.Метод().в
   ═══════════════════════════════════════════════════════════════════ */
lvalue_rest: /* пусто — просто идентификатор */ { $$ = nil }
           | lvalue_rest '.' token_identifier {
               prop := VarStatement{ Name: $3.literal }
               if $1 != nil {
                   $$ = CallChainStatement{ Unit: prop, Call: $1 }
               } else {
                   $$ = prop
               }
           }
           | lvalue_rest '.' token_identifier '(' exprs ')' {
               method := MethodStatement{ Name: $3.literal, Param: $5 }
               if $1 != nil {
                   $$ = CallChainStatement{ Unit: method, Call: $1 }
               } else {
                   $$ = method
               }
           }
           | lvalue_rest '[' expr ']' {
               if $1 != nil {
                   $$ = ItemStatement{ Object: $1, Item: $3 }
               } else {
                   $$ = ItemStatement{ Object: nil, Item: $3 }
               }
           }
;

/* ═══════════════════════════════════════════════════════════════════
   call_rest — продолжение вызова (после CALL_IDENT)
   Примеры: (), .Метод(), .свойство, [индекс]
   ═══════════════════════════════════════════════════════════════════ */
call_rest: /* пусто — голый идентификатор "а;" */ { $$ = nil }
         | '(' exprs ')' {
             // Foo() — вызов метода
             $$ = MethodStatement{ Name: "", Param: $2 }
         }
         | '(' exprs ')' call_chain {
             // Foo().Bar... — вызов с продолжением
             method := MethodStatement{ Name: "", Param: $2 }
             $$ = CallChainStatement{ Unit: $4, Call: method }
         }
         | call_chain {
             // .свойство, .Метод(), [индекс] или их цепочка
             $$ = $1
         }
;

/* ═══════════════════════════════════════════════════════════════════
   call_chain — цепочка доступа к свойствам/методам/индексам
   ═══════════════════════════════════════════════════════════════════ */
call_chain: call_part { $$ = $1 }
          | call_chain call_part {
              $$ = CallChainStatement{ Unit: $2, Call: $1 }
          }
;

call_part: '.' token_identifier {
             $$ = VarStatement{ Name: $2.literal }
         }
         | '.' token_identifier '(' exprs ')' {
             $$ = MethodStatement{ Name: $2.literal, Param: $4 }
         }
         | '[' expr ']' {
             $$ = ItemStatement{ Object: nil, Item: $2 }
         }
;


/* ═══════════════════════════════════════════════════════════════════
   stmt — оператор (statement)
   Лексер уже определил тип по lookahead: LVALUE_IDENT или CALL_IDENT
   ═══════════════════════════════════════════════════════════════════ */
stmt: LVALUE_IDENT lvalue_rest EQUAL expr {
        // Присваивание: а = 1, а.б = 1, а.Метод().в = 1
        lhs := VarStatement{ Name: $1.literal }
        if $2 != nil {
            $$ = AssignmentStatement{ Var: CallChainStatement{ Unit: $2, Call: lhs }, Expr: ExprStatements{ Statements: Statements{$4}} }
        } else {
            $$ = AssignmentStatement{ Var: lhs, Expr: ExprStatements{ Statements: Statements{$4}} }
        }
    }
    | CALL_IDENT call_rest {
        // Вызов как statement: Foo(), а.Метод(), а.б
        if $2 == nil {
            // Голый идентификатор: а;
            $$ = VarStatement{ Name: $1.literal }
        } else if method, ok := $2.(MethodStatement); ok && method.Name == "" {
            // Простой вызов Foo() — НЕ оборачиваем в CallChainStatement
            method.Name = $1.literal
            $$ = method
        } else {
            // Цепочка: а.Метод(), а.б, а[0]
            base := VarStatement{ Name: $1.literal }
            $$ = CallChainStatement{ Unit: $2, Call: base }
        }
    }
    | stmt_if { $$ = $1 }
    | stmt_loop { $$ = $1 }
    | stmt_tryCatch { $$ = $1 }
    | Continue { $$ = ContinueStatement{}; checkLoopOperator($1, yylex) }
    | Break { $$ = BreakStatement{}; checkLoopOperator($1, yylex) }
    | Throw opt_expr { $$ = ThrowStatement{ Param: $2 }; checkThrowParam($1, $2, yylex) }
    | Return opt_expr { $$ = &ReturnStatement{ Param: $2 }; checkReturnParam($2, yylex) }
    | Execute execute_param { $$ = MethodStatement{ Name: $1.literal, Param: ExprStatements{ Statements: Statements{$2}} } }
    | Execute '(' expr ')' { $$ = MethodStatement{ Name: $1.literal, Param: ExprStatements{ Statements: Statements{$3}} } }
    | GoTo goToLabel { $$ = GoToStatement{ Label: $2 } }
    | goToLabel { $$ = $1 }  /* метка ~метка: как statement */
;


/* вызовы через точку */
through_dot: identifier { $$ = $1 }
        | through_dot dot identifier { $$ = CallChainStatement{ Unit: $3, Call:  $1 } }
;

/* вызовы процедур, функций */
/* вызовы выполнить */
/* выполнить может вызываться так выполнить("что-то") или так выполнить "что-то" */
identifier: token_identifier { $$ = VarStatement{ Name: $1.literal } }
        | token_identifier '(' exprs ')' { $$ = MethodStatement{ Name: $1.literal, Param: $3 } }
        | identifier '[' expr ']' { $$ = ItemStatement{ Object: $1, Item: $3 } }
        | Execute execute_param { $$ = MethodStatement{ Name: $1.literal, Param:   ExprStatements{ Statements: Statements{$2}} } }
        | Execute '(' expr ')' { $$ = MethodStatement{ Name: $1.literal, Param:   ExprStatements{ Statements: Statements{$3}} } }
;

execute_param: String { $$ = $1.value  }
             | token_identifier { $$ = VarStatement{ Name: $1.literal }};

/* попытка */
stmt_tryCatch: Try opt_body Catch { setTryFlag(true, yylex) } opt_body EndTry { 
    $$ = TryStatement{ Body: $2, Catch: $5 }
    setTryFlag(false, yylex)
};

/* все что может учавствовать в выражениях */
expr : simple_expr { $$ = $1 }
    | '(' exprs ')' { $$ = $2 }
    | expr '+' expr { $$ = &ExpStatement{Operation: OpPlus, Left: $1, Right: $3} }
    | expr '-' expr { $$ = &ExpStatement{Operation: OpMinus, Left: $1, Right: $3} }
    | expr '*' expr { $$ = &ExpStatement{Operation: OpMul, Left: $1, Right: $3} }
    | expr '/' expr { $$ = &ExpStatement{Operation: OpDiv, Left: $1, Right: $3} }
    | expr '%' expr { $$ = &ExpStatement{Operation: OpMod, Left: $1, Right: $3} }
    | expr '>' expr { $$ = &ExpStatement{Operation: OpGt, Left: $1, Right: $3} }
    | expr '<' expr { $$ = &ExpStatement{Operation: OpLt, Left: $1, Right: $3} }
    | expr EQUAL expr {$$ = &ExpStatement{Operation: OpEq, Left: $1, Right: $3 }}
    | expr OR expr {  $$ = &ExpStatement{Operation: OpOr, Left: $1, Right: $3 } }
    | expr And expr { $$ = &ExpStatement{Operation: OpAnd, Left: $1, Right: $3 } }
    | expr NeEQ expr { $$ = &ExpStatement{Operation: OpNe, Left: $1, Right: $3 } }
    | expr LE expr { $$ = &ExpStatement{Operation: OpLe, Left: $1, Right: $3 } }
    | expr GE expr { $$ = &ExpStatement{Operation: OpGe, Left: $1, Right: $3 } }
    | Not expr { $$ = not($2) }
    | new_object { $$ = $1 }
    | GoTo goToLabel { $$ = GoToStatement{ Label: $2 } }
    | ternary { $$ =  $1  } /* тернарный оператор */
    | through_dot {
	    if tok, ok := $1.(Token); ok {
		    $$ = tok.literal
	    } else {
		    $$ =  $1
	    }
	}
;

opt_expr: { $$ = nil } | expr { $$ = $1 };

exprs : opt_expr {$$ = ExprStatements{ Statements: Statements{$1}} }
	| exprs comma opt_expr { $$.Statements = append($$.Statements, $3) }
;

simple_expr: String { $$ = $1.value  }
            | Number { $$ =  $1.value }
            | '-' expr %prec UNARMinus { $$ = unaryMinus($2) }
            | '+' expr %prec UNARYPlus { $$ = $2 }
            | True { $$ =  $1.value  }
            | False { $$ =  $1.value  }
            | Date { $$ =  $1.value  }
            | Undefined { $$ = UndefinedStatement{} }
            | goToLabel { $$ = $1}
;

// опиасываются правила по которым можно объявлять параметры в функции или процедуре
declarations_method_param: token_identifier {  $$ = *(&ParamStatement{}).Fill(nil, $1) } // обычный параметр
            | ValueParam token_identifier { $$ = *(&ParamStatement{}).Fill(&$1, $2) } // знач
            | declarations_method_param EQUAL simple_expr { $$ = *($$.DefaultValue($3)) } // необязательный параметр
;

declarations_method_params : { $$ = []ParamStatement{} }
                | declarations_method_param  { $$ = []ParamStatement{$1} }
                | declarations_method_params comma declarations_method_param { $$ = append($1, $3) }
;

// для ключевого слова Новый
// 1С допускает такие конструкции
// новый Структура(), новый Массив() ...
// но так же и такие
// Новый("РегистрСведенийКлючЗаписи.СостоянияОригиналовПервичныхДокументов", ПараметрыМассив);
new_object:  New token_identifier { $$ = NewObjectStatement{ Constructor: $2.literal } }
            | New token_identifier '(' exprs ')' { $$ = NewObjectStatement{ Constructor: $2.literal, Param: $4 } }
            | New '(' exprs ')' { $$ = NewObjectStatement{ Param: $3 } }
;



goToLabel: GoToLabel { $$ = &GoToLabelStatement{ Name: $1.literal } }

identifiers: token_identifier %prec LOW_PREC  { $$ = []Token{$1} }
        | identifiers comma token_identifier %prec LOW_PREC {$$ = append($$, $3) }
;

semicolon: ';' {$$ = $1};
colon: ':'{$$ = $1};
comma: ',' {$$ = $1};
dot: '.';

%%