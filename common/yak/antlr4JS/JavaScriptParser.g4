/*
 * The MIT License (MIT)
 *
 * Copyright (c) 2014 by Bart Kiers (original author) and Alexandre Vitorelli (contributor -> ported to CSharp)
 * Copyright (c) 2017-2020 by Ivan Kochurkin (Positive Technologies):
    added ECMAScript 6 support, cleared and transformed to the universal grammar.
 * Copyright (c) 2018 by Juan Alvarez (contributor -> ported to Go)
 * Copyright (c) 2019 by Student Main (contributor -> ES2020)
 *
 * Permission is hereby granted, free of charge, to any person
 * obtaining a copy of this software and associated documentation
 * files (the "Software"), to deal in the Software without
 * restriction, including without limitation the rights to use,
 * copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following
 * conditions:
 *
 * The above copyright notice and this permission notice shall be
 * included in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 * EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
 * OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
 * NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
 * HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 * FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */
parser grammar JavaScriptParser;

// Insert here @header for C++ parser.

options {
    tokenVocab=JavaScriptLexer;
    superClass=JavaScriptParserBase;
}

program
    : HashBangLine? sourceElements? EOF
    ;

sourceElement
    : statement
    ;

statement
    : block
    | variableStatement
    | importStatement
    | exportStatement
    | emptyStatement_
    | classDeclaration
    | functionDeclaration
    | expressionStatement
    | ifStatement
    | iterationStatement
    | continueStatement
    | breakStatement
    | returnStatement
    | yieldStatement
    | withStatement
    | labelledStatement
    | switchStatement
    | throwStatement
    | tryStatement
    | debuggerStatement
    ;

block
    : '{' statementList? '}'
    ;

statementList
    : statement+
    ;

importStatement
    : Import importFromBlock
    ;

importFromBlock
    : importDefault? (importNamespace | importModuleItems) importFrom eos
    | StringLiteral eos
    ;

importModuleItems
    : '{' (importAliasName ',')* (importAliasName ','?)? '}'
    ;

importAliasName
    : moduleExportName (As importedBinding)?
    ;

moduleExportName
    : identifierName
    | StringLiteral
    ;

// yield and await are permitted as BindingIdentifier in the grammar
importedBinding
    : Identifier
    | Yield
    | Await
    ;

importDefault
    : aliasName ','
    ;

importNamespace
    : ('*' | identifierName) (As identifierName)?
    ;

importFrom
    : From StringLiteral
    ;

aliasName
    : identifierName (As identifierName)?
    ;

exportStatement
    : Export Default? (exportFromBlock | declaration) eos    # ExportDeclaration
    | Export Default singleExpression eos                    # ExportDefaultDeclaration
    ;

exportFromBlock
    : importNamespace importFrom eos
    | exportModuleItems importFrom? eos
    ;

exportModuleItems
    : '{' (exportAliasName ',')* (exportAliasName ','?)? '}'
    ;

exportAliasName
    : moduleExportName (As moduleExportName)?
    ;

declaration
    : variableStatement
    | classDeclaration
    | functionDeclaration
    ;

variableStatement
    : variableDeclarationList eos
    ;

variableDeclarationList
    : varModifier variableDeclaration (',' variableDeclaration)*
    ;

variableDeclaration
    : assignable ('=' singleExpression)? // ECMAScript 6: Array & Object Matching
    ;

emptyStatement_
    : SemiColon
    ;

expressionStatement
    : {p.notOpenBraceAndNotFunction()}? expressionSequence eos
    ;

ifStatement
    : If '(' singleExpression ')' statement (Else statement)?
    ;

forFirst
    : singleExpression | variableDeclarationList
    ;

forSecond
    : singleExpression
    ;

forThird
    : singleExpression
    ;

iterationStatement
    : Do statement While '(' singleExpression ')' eos                                                                       # DoStatement
    | While '(' singleExpression ')' statement                                                                              # WhileStatement
    | For '(' forFirst? ';' forSecond? ';' forThird? ')' statement   # ForStatement
    | For '(' (singleExpression | variableDeclaration) In singleExpression ')' statement                                # ForInStatement
    // strange, 'of' is an identifier. and p.p("of") not work in sometime.
    | For Await? '(' (singleExpression | variableDeclaration) identifier{p.p("of")}? singleExpression ')' statement  # ForOfStatement
    ;

varModifier  // let, const - ECMAScript 6
    : Var
    | let_
    | Const
    ;

continueStatement
    : Continue ({p.notLineTerminator()}? identifier)? eos
    ;

breakStatement
    : Break ({p.notLineTerminator()}? identifier)? eos
    ;

returnStatement
    : Return ({p.notLineTerminator()}? expressionSequence)? eos
    ;

yieldStatement
    : Yield ({p.notLineTerminator()}? expressionSequence)? eos
    ;

withStatement
    : With '(' expressionSequence ')' statement
    ;

switchStatement
    : Switch '(' singleExpression ')' caseBlock
    ;

caseBlock
    : '{' caseClauses? (defaultClause caseClauses?)? '}'
    ;

caseClauses
    : caseClause+
    ;

caseClause
    : Case singleExpression ':' statementList?
    ;

defaultClause
    : Default ':' statementList?
    ;

labelledStatement
    : identifier ':' statement
    ;

throwStatement
    : Throw {p.notLineTerminator()}? expressionSequence eos
    ;

tryStatement
    : Try block (catchProduction finallyProduction? | finallyProduction)
    ;

catchProduction
    : Catch ('(' assignable? ')')? block
    ;

finallyProduction
    : Finally block
    ;

debuggerStatement
    : Debugger eos
    ;

functionDeclaration
    : Async? Function_ '*'? identifier '(' formalParameterList? ')' functionBody
    ;

classDeclaration
    : Class identifier classTail
    ;

classTail
    : (Extends singleExpression)? '{' classElement* '}'
    ;

classElement
    : (Static | {p.n("static")}? identifier)? methodDefinition
    | (Static | {p.n("static")}? identifier)? fieldDefinition
    | (Static | {p.n("static")}? identifier) block
    | emptyStatement_
    ;

methodDefinition
    : (Async {p.notLineTerminator()}?)? '*'? classElementName '(' formalParameterList? ')' functionBody
    | '*'? getter '(' ')' functionBody
    | '*'? setter '(' formalParameterList? ')' functionBody
    ;

fieldDefinition
    : classElementName initializer?
    ;

classElementName
    : propertyName
    | privateIdentifier
    ;

privateIdentifier
    : '#' identifierName
    ;

formalParameterList
    : formalParameterArg (',' formalParameterArg)* (',' lastFormalParameterArg)?
    | lastFormalParameterArg
    ;

formalParameterArg
    : assignable ('=' singleExpression)?      // ECMAScript 6: Initialization
    ;

lastFormalParameterArg                        // ECMAScript 6: Rest Parameter
    : Ellipsis singleExpression
    ;

functionBody
    : '{' sourceElements? '}'
    ;

sourceElements
    : sourceElement+
    ;

arrayLiteral
    : ('[' elementList ']')
    ;

elementList
    : ','* arrayElement? (','+ arrayElement)* ','* // Yes, everything is optional
    ;

arrayElement
    : Ellipsis? singleExpression
    ;

propertyAssignment
    : propertyName ':' singleExpression                                             # PropertyExpressionAssignment
    | '[' singleExpression ']' ':' singleExpression                                 # ComputedPropertyExpressionAssignment
    | Async? '*'? propertyName '(' formalParameterList?  ')'  functionBody  # FunctionProperty
    | getter '(' ')' functionBody                                           # PropertyGetter
    | setter '(' formalParameterArg ')' functionBody                        # PropertySetter
    | Ellipsis? singleExpression                                                    # PropertyShorthand
    ;

propertyName
    : identifierName
    | StringLiteral
    | numericLiteral
    | '[' singleExpression ']'
    ;

arguments
    : '('(argument (',' argument)* ','?)?')'
    ;

argument
    : Ellipsis? (singleExpression | identifier)
    ;

expressionSequence
    : singleExpression (',' singleExpression)*
    ;

singleExpression
    : anonymousFunction                                                     # FunctionExpression
    | Class identifier? classTail                                           # ClassExpression
    | singleExpression '?.' singleExpression                                # OptionalChainExpression
    | singleExpression '?.'? '[' singleExpression ']'                     # MemberIndexExpression
    | singleExpression '?'? '.' '#'? identifierName                         # MemberDotExpression
    // Split to try `new Date()` first, then `new Date`.
    | New singleExpression arguments                                        # NewExpression
    | New singleExpression                                                  # NewExpression
    | singleExpression arguments                                            # ArgumentsExpression
    | New '.' identifier                                                    # MetaExpression // new.target
    | singleExpression {p.notLineTerminator()}? '++'                     # PostIncrementExpression
    | singleExpression {p.notLineTerminator()}? '--'                     # PostDecreaseExpression
    | Delete singleExpression                                               # DeleteExpression
    | Void singleExpression                                                 # VoidExpression
    | Typeof singleExpression                                               # TypeofExpression
    | '++' singleExpression                                                 # PreIncrementExpression
    | '--' singleExpression                                                 # PreDecreaseExpression
    | '+' singleExpression                                                  # UnaryPlusExpression
    | '-' singleExpression                                                  # UnaryMinusExpression
    | '~' singleExpression                                                  # BitNotExpression
    | '!' singleExpression                                                  # NotExpression
    | Await singleExpression                                                # AwaitExpression
    | <assoc=right> singleExpression '**' singleExpression                  # PowerExpression
    | singleExpression ('*' | '/' | '%') singleExpression                   # MultiplicativeExpression
    | singleExpression ('+' | '-') singleExpression                         # AdditiveExpression
    | singleExpression '??' singleExpression                                # CoalesceExpression
    | singleExpression ('<<' | '>>' | '>>>') singleExpression               # BitShiftExpression
    | singleExpression ('<' | '>' | '<=' | '>=') singleExpression           # RelationalExpression
    | singleExpression Instanceof singleExpression                          # InstanceofExpression
    | singleExpression In singleExpression                                  # InExpression
    | singleExpression ('==' | '!=' | '===' | '!==') singleExpression       # EqualityExpression
    | singleExpression '&' singleExpression                                 # BitAndExpression
    | singleExpression '^' singleExpression                                 # BitXOrExpression
    | singleExpression '|' singleExpression                                 # BitOrExpression
    | singleExpression '&&' singleExpression                                # LogicalAndExpression
    | singleExpression '||' singleExpression                                # LogicalOrExpression
    | singleExpression '?' singleExpression ':' singleExpression            # TernaryExpression
    | <assoc=right> singleExpression '=' singleExpression                   # AssignmentExpression
    | <assoc=right> singleExpression assignmentOperator singleExpression    # AssignmentOperatorExpression
    | Import '(' singleExpression ')'                                       # ImportExpression
    | singleExpression templateStringLiteral                                # TemplateStringExpression  // ECMAScript 6
    | yieldStatement                                                        # YieldExpression // ECMAScript 6
    | This                                                                  # ThisExpression
    | identifier                                                            # IdentifierExpression
    | Super                                                                 # SuperExpression
    | literal                                                               # LiteralExpression
    | arrayLiteral                                                          # ArrayLiteralExpression
    | objectLiteral                                                         # ObjectLiteralExpression
    | '(' expressionSequence ')'                                            # ParenthesizedExpression
    ;

initializer
// TODO: must be `= AssignmentExpression` and we have such label alredy but it doesn't respect the specification.
//  See https://tc39.es/ecma262/multipage/ecmascript-language-expressions.html#prod-Initializer
    : '=' singleExpression
    ;

assignable
    : identifier
    | arrayLiteral
    | objectLiteral
    ;

objectLiteral
    : '{' (propertyAssignment (',' propertyAssignment)* ','?)? '}'
    ;

anonymousFunction
    : Async? Function_ '*'? '(' formalParameterList? ')' functionBody    # AnonymousFunctionDecl
    | Async? arrowFunctionParameters '=>' arrowFunctionBody                    # ArrowFunction
    ;

arrowFunctionParameters
    : identifier
    | '(' formalParameterList? ')'
    ;

arrowFunctionBody
    : functionBody
    | singleExpression
    ;

assignmentOperator
    : '*='
    | '/='
    | '%='
    | '+='
    | '-='
    | '<<='
    | '>>='
    | '>>>='
    | '&='
    | '^='
    | '|='
    | '**='
    ;

literal
    : NullLiteral
    | BooleanLiteral
    | StringLiteral
    | templateStringLiteral
    | RegularExpressionLiteral
    | numericLiteral
    | bigintLiteral
    ;

templateStringLiteral
    : BackTick templateStringAtom* BackTick
    ;

templateStringAtom
    : TemplateStringAtom
    | TemplateStringStartExpression singleExpression TemplateCloseBrace
    ;

numericLiteral
    : DecimalLiteral
    | HexIntegerLiteral
    | OctalIntegerLiteral
    | OctalIntegerLiteral2
    | BinaryIntegerLiteral
    ;

bigintLiteral
    : BigDecimalIntegerLiteral
    | BigHexIntegerLiteral
    | BigOctalIntegerLiteral
    | BigBinaryIntegerLiteral
    ;

getter
    : {p.n("get")}? identifier classElementName
    ;

setter
    : {p.n("set")}? identifier classElementName
    ;

identifierName
    : identifier
    | reservedWord
    ;

identifier
    : Identifier
    | NonStrictLet
    | Async
    | As
    | From
    ;

reservedWord
    : keyword
    | NullLiteral
    | BooleanLiteral
    ;

keyword
    : Break
    | Do
    | Instanceof
    | Typeof
    | Case
    | Else
    | New
    | Var
    | Catch
    | Finally
    | Return
    | Void
    | Continue
    | For
    | Switch
    | While
    | Debugger
    | Function_
    | This
    | With
    | Default
    | If
    | Throw
    | Delete
    | In
    | Try

    | Class
    | Enum
    | Extends
    | Super
    | Const
    | Export
    | Import
    | Implements
    | let_
    | Private
    | Public
    | Interface
    | Package
    | Protected
    | Static
    | Yield
    | Async
    | Await
    | From
    | As
    ;

let_
    : NonStrictLet
    | StrictLet
    ;

eos
    : SemiColon
    | EOF
    | {p.lineTerminatorAhead()}?
    | {p.closeBrace()}?
    ;
