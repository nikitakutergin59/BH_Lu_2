package tokens

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestTokenize_BH(t *testing.T) {
	testCase := []struct {
		name   string
		input  string
		result string
		err    error
	}{
		{
			name:   "defelt_0",
			input:  "52+52",
			result: "{NUMBER 52} {OPERATOR +} {NUMBER 52}",
			err:    nil,
		},
		{
			name:   "unaryMinus",
			input:  "-52+52",
			result: "{NUMBER -52} {OPERATOR +} {NUMBER 52}",
			err:    nil,
		},
		{
			name:   "floatValue",
			input:  "0.25+0.25",
			result: "{NUMBER 0.25} {OPERATOR +} {NUMBER 0.25}",
			err:    nil,
		},
		{
			name:   "parents",
			input:  "15*52+(14-66)",
			result: "{NUMBER 15} {OPERATOR *} {NUMBER 52} {OPERATOR +} {NUMBER 1} {MULT *} {OPEN (} {NUMBER 14} {OPERATOR -} {NUMBER 66} {CLOSE )}",
			err:    nil,
		},
		{
			name:   "unaryMinusAfterParent",
			input:  "-(52-44)*93-(400/20)",
			result: "{NUMBER 1} {MULT *} {OPEN -(} {NUMBER 52} {OPERATOR -} {NUMBER 44} {CLOSE )} {OPERATOR *} {NUMBER 93} {OPERATOR -} {NUMBER 1} {MULT *} {OPEN (} {NUMBER 400} {OPERATOR /} {NUMBER 20} {CLOSE )}",
			err:    nil,
		},
		{
			name:   "parentParent",
			input:  "(52+52)(44-43)",
			result: "{NUMBER 1} {MULT *} {OPEN (} {NUMBER 52} {OPERATOR +} {NUMBER 52} {CLOSE )} {MULT *} {NUMBER 1} {MULT *} {OPEN (} {NUMBER 44} {OPERATOR -} {NUMBER 43} {CLOSE )}",
			err:    nil,
		},
		{
			name:   "cfParent",
			input:  "52*2(45-44)",
			result: "{NUMBER 52} {OPERATOR *} {NUMBER 2} {MULT *} {OPEN (} {NUMBER 45} {OPERATOR -} {NUMBER 44} {CLOSE )}",
			err:    nil,
		},
		{
			name:   "ERROR",
			input:  "52+52|",
			result: "Ошибка при токенизации: токен не распозднан",
			err:    errors.New("токен | не распозднан"),
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			tokens, err := Tokenize_BH(tc.input)

			// Проверяем ошибку
			if tc.err != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tc.err.Error()) {
					t.Fatalf("expected error '%v', got '%v'", tc.err, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Проверяем результат
			if tc.err == nil {
				var resultString string
				for _, token := range tokens {
					resultString += fmt.Sprintf("%v ", token)
				}
				resultString = strings.TrimSpace(resultString) // Удаляем лишний пробел в конце

				if resultString != tc.result {
					t.Fatalf("expected '%s', got '%s'", tc.result, resultString)
				}
			}
		})
	}
}
