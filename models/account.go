package models

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
)

type Account struct {
	Address common.Address
	PKey    *ecdsa.PrivateKey
}

func NewAccount(address common.Address, pkey *ecdsa.PrivateKey) *Account {
	return &Account{Address: address, PKey: pkey}
}

type TwitterData struct {
	Ct0       string
	AuthToken string
}

func NewTwitterData(ct0, authToken string) *TwitterData {
	return &TwitterData{Ct0: ct0, AuthToken: authToken}
}
