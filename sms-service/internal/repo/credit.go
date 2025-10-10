package repo

import (
	"fmt"
	"math/rand"
)

type Credit interface {
	ReserveCredit(org string, amount uint64) bool
}

type InternalCreditHandler struct{}

func (c *InternalCreditHandler) ReserveCredit(org string, amount uint64) bool {
	n := rand.Intn(3)
	fmt.Println("n is", n)
	return !(n == 1)
}

func NewCreditRepo() Credit {
	return &InternalCreditHandler{}
}
