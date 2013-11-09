package main

import (
	"testing"
	"regexp"
)

func Test_GenId(t *testing.T) {
	id, err := GenId()
	if err != nil {
		t.Error("err returned: ", err)
	}
	r := regexp.MustCompile(`^[A-Z0-9]{16}$`)
	if ! r.MatchString(id) {
		t.Error("invalid ID: ", id)
	}
}