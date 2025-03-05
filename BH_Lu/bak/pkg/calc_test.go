package calculator

import (
	"errors"
	"math"
	"testing"
)

func TestCalc(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected float64
		err      error
	}{
		{
			name:     "Simple addition",
			input:    "2 + 3",
			expected: 5,
			err:      nil,
		},
		{
			name:     "Simple subtraction",
			input:    "5 - 2",
			expected: 3,
			err:      nil,
		},
		{
			name:     "Simple multiplication",
			input:    "4 * 6",
			expected: 24,
			err:      nil,
		},
		{
			name:     "Simple division",
			input:    "8 / 2",
			expected: 4,
			err:      nil,
		},
		{
			name:     "Expression with parentheses",
			input:    "(2 + 3) * 4",
			expected: 20,
			err:      nil,
		},
		{
			name:     "Expression with multiple operations",
			input:    "10 + 2 * 3 - 4 / 2",
			expected: 14,
			err:      nil,
		},
		{
			name:     "Complex expression with parentheses",
			input:    "((10 / 2) + 10) * 2",
			expected: 30,
			err:      nil,
		},
		{
			name:     "Division by zero",
			input:    "5 / 0",
			expected: 0,
			err:      errors.New("division by zero"),
		},
		{
			name:     "Mismatched parentheses - opening",
			input:    "(2 + 3",
			expected: 0,
			err:      errors.New("mismatched parentheses"),
		},
		{
			name:     "Mismatched parentheses - closing",
			input:    "2 + 3)",
			expected: 0,
			err:      errors.New("mismatched parentheses"),
		},
		{
			name:     "Invalid character",
			input:    "2 + a",
			expected: 0,
			err:      errors.New("invalid expression"),
		},
		{
			name:     "Negative numbers",
			input:    "-2 + 5",
			expected: 3,
			err:      nil,
		},
		{
			name:     "Float numbers",
			input:    "2.5 + 2.5",
			expected: 5,
			err:      nil,
		},
		{
			name:     "Float number as result",
			input:    "5 / 2",
			expected: 2.5,
			err:      nil,
		},
		{
			name:     "Negative result",
			input:    "2-5",
			expected: -3,
			err:      nil,
		},
		{
			name:     "Very big numbers",
			input:    "1000000000000 + 2000000000000",
			expected: 3000000000000,
			err:      nil,
		},
		{
			name:     "Very big float number",
			input:    "1000000000000.5 + 2000000000000.5",
			expected: 3000000000001,
			err:      nil,
		},
		{
			name:     "Floating point imprecision",
			input:    "0.1+0.2",
			expected: 0.3,
			err:      nil,
		},
		{
			name:     "Empty expression",
			input:    "",
			expected: 0,
			err:      errors.New("invalid character"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := Calc(tc.input)
			if tc.err != nil {
				if err == nil || err.Error() != tc.err.Error() {
					t.Errorf("expected error '%v', got '%v'", tc.err, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(actual-tc.expected) > 1e-10 {
				t.Errorf("expected %f, got %f", tc.expected, actual)
			}
		})
	}
}
