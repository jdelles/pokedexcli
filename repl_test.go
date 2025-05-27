package main

import (
	"fmt"
	"testing"
)

func TestCleanInput(t *testing.T) {
    cases := []struct {
	input    string
	expected []string
}{
	{
		input:    "  hello  world  ",
		expected: []string{"hello", "world"},
	},
	{
		input: "testing......    ",
		expected: []string{"testing......",},
	},
}

	for _, c := range cases {
		actual := cleanInput(c.input)
		fmt.Println(actual)
		fmt.Println(c.expected)
		if len(actual) != len(c.expected) {
			t.Errorf("Length mismatch: got %d words, expected %d words", len(actual), len(c.expected))
			continue
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("Word mismatch at position %d: got '%s', expected '%s'", i, word, expectedWord)
			}
		}	
	}
}