package main

import (
	"fmt"
)

type AST interface {
}

type PackageAST struct {
	Name     string
	Imports  []string
	Services []*ServiceAST
}

type ServiceAST struct {
	Name    string
	Methods []*MethodAST
}

type MethodAST struct {
	Name    string
	Args    []string
	Returns []string
	IsGo    bool
}

func (s *ServiceAST) String() string {
	result := fmt.Sprintf("ServiceAST(name=%s, methods=[", s.Name)
	for _, method := range s.Methods {
		result += method.String() + ", "
	}
	return result[:len(result)-2] + "]"
}

func (m *MethodAST) String() string {
	return fmt.Sprintf("MethodAST(name=%s, args=%v, returns=%v, isgo=%v)", m.Name, m.Args, m.Returns, m.IsGo)
}

// Grammar rule:
// package: package ID (import PATH)* serviceList
// serviceList: service | service serviceList
// service: SERVICE ID LCURLY serviceMethodList RCURLY
// serviceMethodList: serviceMethodDecl serviceMethodList | serviceMethodDecl
// serviceMethodDecl: (GO)+ ID LPAREN formalParametersList RPAREN returnValue
// parametersList: parameter
//    			 | parameter COMMA parametersList
//               | empty
// parameter: ID
// PATH: ".+"
// ID: (*)?([])?[a-zA-Z_][a-zA-Z0-9_]*
// returnValue: empty | LPAREN parameter COMMA parameter RPAREN

type Parser struct {
	lexer        *Lexer
	currentToken *Token
}

func NewPackageAST(name string, imports []string, services []*ServiceAST) *PackageAST {
	return &PackageAST{
		Name:     name,
		Imports:  imports,
		Services: services,
	}
}

func NewServiceAST(name string, methods []*MethodAST) *ServiceAST {
	return &ServiceAST{
		Name:    name,
		Methods: methods,
	}
}

func NewMethodAST(name string, args, returns []string, isGo bool) *MethodAST {
	return &MethodAST{
		Name:    name,
		Args:    args,
		Returns: returns,
		IsGo:    isGo,
	}
}

func NewParser(lexer *Lexer) *Parser {
	parser := &Parser{
		lexer:        lexer,
		currentToken: lexer.get_next_token(),
	}
	return parser
}

func (p *Parser) eat(tokenType string) error {
	if p.currentToken.tp != tokenType {
		return fmt.Errorf("unexpected token -> %s", p.currentToken.String())
	}

	p.currentToken = p.lexer.get_next_token()
	return nil
}

func (p *Parser) parse() (*PackageAST, error) {
	module, err := p._package()
	if err != nil {
		return nil, err
	}

	if p.currentToken.tp != EOF {
		return nil, fmt.Errorf("unexpected token -> %s", p.currentToken.String())
	}

	return module, nil
}

func (p *Parser) _package() (*PackageAST, error) {
	err := p.eat(PACKAGE)
	if err != nil {
		return nil, err
	}

	packageName := p.currentToken.value
	err = p.eat(ID)
	if err != nil {
		return nil, err
	}

	imports := make([]string, 0)
	for p.currentToken.tp == IMPORT {
		err = p.eat(IMPORT)
		if err != nil {
			return nil, err
		}

		imports = append(imports, p.currentToken.value)
		err = p.eat(PATH)
		if err != nil {
			return nil, err
		}
	}

	services, err := p.serviceList()
	if err != nil {
		return nil, err
	}

	return NewPackageAST(packageName, imports, services), nil
}

func (p *Parser) serviceList() ([]*ServiceAST, error) {
	services := make([]*ServiceAST, 0)

	service, err := p.service()
	if err != nil {
		return nil, err
	}
	services = append(services, service)

	for p.currentToken.tp == SERVICE {
		service, err := p.service()
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, nil
}

func (p *Parser) service() (*ServiceAST, error) {
	err := p.eat(SERVICE)
	if err != nil {
		return nil, err
	}

	serviceName := p.currentToken.value
	err = p.eat(ID)
	if err != nil {
		return nil, err
	}

	err = p.eat(LCURLY)
	if err != nil {
		return nil, err
	}

	methods, err := p.serviceMethodList()
	if err != nil {
		return nil, err
	}

	err = p.eat(RCURLY)
	if err != nil {
		return nil, err
	}

	return NewServiceAST(serviceName, methods), nil
}

func (p *Parser) serviceMethodList() ([]*MethodAST, error) {
	methods := make([]*MethodAST, 0)

	method, err := p.serviceMethodDecl()
	if err != nil {
		return nil, err
	}
	methods = append(methods, method)

	for p.currentToken.tp == ID || p.currentToken.tp == GO {
		method, err := p.serviceMethodDecl()
		if err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}

	return methods, nil
}

func (p *Parser) serviceMethodDecl() (*MethodAST, error) {
	isGo := false
	if p.currentToken.tp == GO {
		err := p.eat(GO)
		if err != nil {
			return nil, err
		}
		isGo = true
	}

	methodName := p.currentToken.value
	err := p.eat(ID)
	if err != nil {
		return nil, err
	}

	err = p.eat(LPAREN)
	if err != nil {
		return nil, err
	}

	args, err := p.formalParametersList()
	if err != nil {
		return nil, err
	}

	err = p.eat(RPAREN)
	if err != nil {
		return nil, err
	}

	returns, err := p.returnValue()
	if err != nil {
		return nil, err
	}

	return NewMethodAST(methodName, args, returns, isGo), nil
}

func (p *Parser) formalParametersList() ([]string, error) {
	args := make([]string, 0)

	if p.currentToken.tp == RPAREN {
		return args, nil
	}

	arg, err := p.parameter()
	if err != nil {
		return nil, err
	}
	args = append(args, arg)

	for p.currentToken.tp == COMMA {
		err = p.eat(COMMA)
		if err != nil {
			return nil, err
		}

		arg, err := p.parameter()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	return args, nil
}

func (p *Parser) parameter() (string, error) {
	arg := p.currentToken.value
	err := p.eat(ID)
	if err != nil {
		return "", err
	}
	return arg, nil
}

func (p *Parser) returnValue() ([]string, error) {
	// no returns
	if p.currentToken.tp != LPAREN {
		return nil, nil
	}

	err := p.eat(LPAREN)
	if err != nil {
		return nil, err
	}

	returns := make([]string, 0)
	param, err := p.parameter()
	if err != nil {
		return nil, err
	}
	returns = append(returns, param)
	p.eat(COMMA)
	param, err = p.parameter()
	if err != nil {
		return nil, err
	}
	returns = append(returns, param)

	err = p.eat(RPAREN)
	if err != nil {
		return nil, err
	}

	return returns, nil
}
