package service

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestDummyHashIsRealBcryptAtProductionCost proves the absent-user compare does
// the same bcrypt work a present user's would: the dummy hash is a well-formed
// bcrypt hash at the production cost, so a missing username cannot be told from
// a present one by timing.
func TestDummyHashIsRealBcryptAtProductionCost(t *testing.T) {
	cost, err := bcrypt.Cost([]byte(dummyHash))

	if nil != err {
		t.Fatalf("the dummy hash must be a valid bcrypt hash, got %v", err)
	}

	if bcrypt.DefaultCost != cost {
		t.Fatalf("the dummy hash must match the production cost %d, got %d", bcrypt.DefaultCost, cost)
	}
}
