package utils

import (
	"math/rand"
	"testing"
	"time"
)

func TestRandString(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	str := GetRandomString(10, nil)
	if len(str) != 10 {
		t.Error("got: ", str)
		t.FailNow()
	}
	t.Log("got: ", str)
}
