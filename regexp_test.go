package main

import (
	"regexp"
	"testing"
)

func TestReg(t *testing.T) {
	pattern := regexp.MustCompile(`(.*[a-z].*)(.*[A-Z].*)(.*\d.*)(.*\..*)`)
	if pattern.MatchString("abc") {
		t.Fail()
	}
	if pattern.MatchString("abdA1") {
		t.Fail()
	}
	if pattern.MatchString(".asdA1") {
		t.Fail()
	}
	if pattern.MatchString("asd.1") {
		t.Fail()
	}
	if !pattern.MatchString("ds.23") {
		t.Fail()
	}
}
