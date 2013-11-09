package main

import (
	"testing"
	"regexp"
)

func Test_GenId(t *testing.T) {
	id := GenId()
	r := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
	if ! r.MatchString(id) {
		t.Error("invalid ID", id)
	}
}