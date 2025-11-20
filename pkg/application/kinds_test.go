package application

import "testing"

func TestKindFromStringSuccess(t *testing.T) {
	tests := []struct {
		input    string
		expected Kind
	}{
		{input: "concert", expected: KindConcert},
		{input: "theater", expected: KindTheater},
		{input: "movie", expected: KindMovie},
		{input: "festival", expected: KindFestival},
		{input: "party", expected: KindParty},
		{input: "dance", expected: KindParty},
		{input: "live-music", expected: KindParty},
		{input: "karaoke", expected: KindKaraoke},
		{input: "meetups", expected: KindBusiness},
		{input: "workshops", expected: KindBusiness},
		{input: "unknown", expected: KindUnknown},
		{input: "foo bar", expected: KindUnknown},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			actual := KindFromString(test.input)
			if actual != test.expected {
				t.Errorf("expected %s, got %s", test.expected, actual)
			}
		})
	}
}

func TestFirstKindMatchSuccess(t *testing.T) {
	tests := map[string]struct {
		input    []string
		expected Kind
	}{
		"when mathing the first string": {
			input:    []string{"concert", "foo", "bar"},
			expected: KindConcert,
		},
		"when mathing the second string": {
			input:    []string{"foo", "theater", "bar"},
			expected: KindTheater,
		},
		"when mathing the second string and after": {
			input:    []string{"foo", "theater", "concert"},
			expected: KindTheater,
		},
		"when no match": {
			input:    []string{"foo", "bar"},
			expected: KindUnknown,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := FirstKindMatch(test.input)
			if actual != test.expected {
				t.Errorf("expected %s, got %s", test.expected, actual)
			}
		})
	}
}
