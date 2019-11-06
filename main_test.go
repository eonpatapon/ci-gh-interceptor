package main

import "testing"
import "regexp"

func TestSanitizeBranchName(t *testing.T) {
	inputs := []string{
		"foo/bar",
		"foo_bar/foo.com",
		"foo-héhé",
		"FOO/BAR",
	}
	validate := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`
	for _, input := range inputs {
		got := sanitizeBranchName(input)
		_, err := regexp.MatchString(validate, got)
		if err != nil {
			t.Errorf("with %s got %s, doesn't match %s", input, got, validate)
		}
	}
}

func TestBranchNameFromRef(t *testing.T) {
	inputs := []string{
		"refs/heads/foo",
		"refs/heads/foo/bar",
		"refs/heads/foo/bar/foo",
	}
	expects := []string{
		"foo",
		"foo/bar",
		"foo/bar/foo",
	}
	for idx, input := range inputs {
		got := branchNameFromRef(input)
		if got != expects[idx] {
			t.Errorf("with %s got %s expected %s", input, got, expects[idx])
		}
	}
}
