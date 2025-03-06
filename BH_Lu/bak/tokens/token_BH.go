package tokens

import (
	"fmt"
	"regexp"
	"strings"
)

// создаём константы
const (
	TOKEN_NUMBER       = "NUMBER"
	TOKEN_OPERATOR     = "OPERATOR"
	TOKEN_PARENT_OPEN  = "OPEN"
	TOKEN_PARENT_CLOSE = "CLOSE"
	TOKEN_MULT         = "MULT"
	TOKEN_UNARY_MINUS  = "UNARY_MINUS"
	TOKEN_UNKNOWN      = "UNKNOWN"
)

// структура к которой обращаеться токенизатор
type Token struct {
	Type  string
	Value string
}

func Tokenize_BH(input string) ([]Token, error) {
	input = strings.ReplaceAll(input, " ", "")

	//регулярные выражения для всех типов токенов
	numberRegex := regexp.MustCompile(`^(\d+(\.\d*)?|\.\d+)`)
	operatorRegex := regexp.MustCompile(`^(\+|-|\*|/|=)`)
	openRegex := regexp.MustCompile(`^\(`)
	closeRegex := regexp.MustCompile(`^\)`)
	unknowRegex := regexp.MustCompile(`.`) //все значения которые не поддерживаються

	token_BH := []Token{}
	unariMinusNext := false //указатель на наличие унарного минуса

	for len(input) > 0 {
		// токенизация чисел
		if numberMatch := numberRegex.FindString(input); numberMatch != "" {
			numberValue := numberMatch
			//проверяем наличие унарного минуса
			if unariMinusNext {
				numberValue = "-" + numberValue
				unariMinusNext = false //сбрасываем указатель
			}
			token_BH = append(token_BH, Token{TOKEN_NUMBER, numberValue})
			input = input[len(numberMatch):]
			//если после числа следует открывающая скобка добовим токен MULT
			input = strings.TrimSpace(input)
			if len(input) > 0 {
				if openRegex.MatchString(string(input[0])) {
					token_BH = append(token_BH, Token{TOKEN_MULT, "*"})
				}
			}
			continue
		}
		// токенизация операторов
		if operatorMatch := operatorRegex.FindString(input); operatorMatch != "" {
			operatorValue := operatorMatch
			input = input[len(operatorMatch):]

			if operatorValue == "-" {
				// проверяем унарный ли минус
				if len(token_BH) == 0 || //начало строки
					(len(token_BH) > 0 && token_BH[len(token_BH)-1].Type == TOKEN_PARENT_OPEN) || //после открывающей скобки
					(len(token_BH) > 0 && token_BH[len(token_BH)-1].Type == TOKEN_OPERATOR) || // после другово оператора
					(len(token_BH) > 0 && token_BH[len(token_BH)-1].Type == TOKEN_UNARY_MINUS) {
					unariMinusNext = true
					continue
				} else {
					token_BH = append(token_BH, Token{TOKEN_OPERATOR, operatorValue})
					continue
				}
			} else {
				//это другой оператор
				token_BH = append(token_BH, Token{TOKEN_OPERATOR, operatorValue})
				continue
			}

		}
		//токенизация открывающей скобки
		if openMatch := openRegex.FindString(input); openMatch != "" {
			openValue := openMatch
			if unariMinusNext {
				openValue = "-" + openValue
				unariMinusNext = false
			}
			// проверяем, есть ли коэффициент перед скобкой
			if len(token_BH) > 0 {
				lastToken := token_BH[len(token_BH)-1]
				if lastToken.Type == TOKEN_NUMBER {
					// eсли перед скобкой число, значит коэффициент есть. Ничего не добавляем.
				} else {
					if lastToken.Type == TOKEN_OPERATOR {
						// eсли перед скобкой оператор добавляем "1*"
						token_BH = append(token_BH, Token{TOKEN_NUMBER, "1"})
						token_BH = append(token_BH, Token{TOKEN_MULT, "*"})
					} else if lastToken.Type == TOKEN_PARENT_CLOSE {
						//если перед скобкой закрывающая скобка то добавляем "*1*"
						token_BH = append(token_BH, Token{TOKEN_MULT, "*"})
						token_BH = append(token_BH, Token{TOKEN_NUMBER, "1"})
						token_BH = append(token_BH, Token{TOKEN_MULT, "*"})
					}
				}
			} else {
				// eсли это первая скобка в выражении, добавляем "1*"
				token_BH = append(token_BH, Token{TOKEN_NUMBER, "1"})
				token_BH = append(token_BH, Token{TOKEN_MULT, "*"})
			}
			token_BH = append(token_BH, Token{TOKEN_PARENT_OPEN, openValue})
			input = input[len(openMatch):]
			continue
		}
		// токенизиция закрывающей скобки
		if closeMatch := closeRegex.FindString(input); closeMatch != "" {
			token_BH = append(token_BH, Token{TOKEN_PARENT_CLOSE, closeMatch})
			input = input[len(closeMatch):]
			continue
		}
		//если токен не поддерживаеться
		if unknouMatch := unknowRegex.FindString(input); unknouMatch != "" {
			token_BH = append(token_BH, Token{TOKEN_PARENT_CLOSE, unknouMatch})
			input = input[1:]
			if len(token_BH) != 0 {
				return nil, fmt.Errorf("токен %s не распозднан", unknouMatch)
			}
		}
	}
	//fmt.Println("распашенное уравнение", token_BH)
	return token_BH, nil
}
