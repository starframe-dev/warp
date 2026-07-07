package warp

import (
	"reflect"
	"testing"
)

func TestWordWrap(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected []string
	}{
		{
			name:     "width zero",
			text:     "hello world",
			width:    0,
			expected: nil,
		},
		{
			name:     "width negative",
			text:     "hello world",
			width:    -5,
			expected: nil,
		},
		{
			name:     "empty string",
			text:     "",
			width:    10,
			expected: []string{""},
		},
		{
			name:     "short text no wrap",
			text:     "hello",
			width:    10,
			expected: []string{"hello"},
		},
		{
			name:     "wrap at word boundary",
			text:     "hello world",
			width:    5,
			expected: []string{"hello", "world"},
		},
		{
			name:     "long word broken mid-word",
			text:     "abcdef",
			width:    3,
			expected: []string{"abc", "def"},
		},
		{
			name:     "multiple newlines preserved",
			text:     "line1\nline2\nline3",
			width:    10,
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "extra spaces trimmed at breaks",
			text:     "hello   world",
			width:    5,
			expected: []string{"hello", "world"},
		},
		{
			name:     "wide runes",
			text:     "こんにちは世界",
			width:    10,
			expected: []string{"こんにちは", "世界"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WordWrap(tt.text, tt.width)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("WordWrap(%q, %d) = %v, want %v", tt.text, tt.width, got, tt.expected)
			}
		})
	}
}

func TestSpaceWrap(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected []string
	}{
		{
			name:     "width zero",
			text:     "hello world",
			width:    0,
			expected: nil,
		},
		{
			name:     "width negative",
			text:     "hello world",
			width:    -1,
			expected: nil,
		},
		{
			name:     "empty string",
			text:     "",
			width:    10,
			expected: []string{""},
		},
		{
			name:     "short text no wrap",
			text:     "hello world",
			width:    20,
			expected: []string{"hello world"},
		},
		{
			name:     "wrap at space",
			text:     "hello world",
			width:    5,
			expected: []string{"hello", "world"},
		},
		{
			name:     "long word overflows",
			text:     "abcdef",
			width:    3,
			expected: []string{"abcdef"},
		},
		{
			name:     "multiple newlines preserved",
			text:     "one two\nthree four",
			width:    5,
			expected: []string{"one", "two", "three", "four"},
		},
		{
			name:     "multiple spaces preserved when fits",
			text:     "hello   world",
			width:    20,
			expected: []string{"hello   world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SpaceWrap(tt.text, tt.width)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("SpaceWrap(%q, %d) = %v, want %v", tt.text, tt.width, got, tt.expected)
			}
		})
	}
}

func TestWrapToString(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		width        int
		useSpaceWrap bool
		expected     string
	}{
		{
			name:         "word wrap",
			text:         "hello world",
			width:        5,
			useSpaceWrap: false,
			expected:     "hello\nworld",
		},
		{
			name:         "space wrap",
			text:         "hello world",
			width:        5,
			useSpaceWrap: true,
			expected:     "hello\nworld",
		},
		{
			name:         "invalid width",
			text:         "hello",
			width:        0,
			useSpaceWrap: false,
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapToString(tt.text, tt.width, tt.useSpaceWrap)
			if got != tt.expected {
				t.Errorf("WrapToString(%q, %d, %v) = %q, want %q", tt.text, tt.width, tt.useSpaceWrap, got, tt.expected)
			}
		})
	}
}

func TestIsWordBreak(t *testing.T) {
	if !isWordBreak(' ') {
		t.Error("expected space to be a word break")
	}
	if !isWordBreak('\t') {
		t.Error("expected tab to be a word break")
	}
	if isWordBreak('a') {
		t.Error("expected 'a' not to be a word break")
	}
}
