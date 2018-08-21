package crypt

import (
	"math/big"
	"math/rand"
)

var P, G *big.Int

// a
func Randomkey() *big.Int {
	n := new(big.Int)
	tmp := make([]byte, 8)
	for i := 0; i < 8; i++ {
		tmp[i] = byte(rand.Intn(256))
	}
	return n.SetBytes(tmp)
}

// G**a mod p
func DHExchange(key *big.Int) *big.Int {
	n := new(big.Int)
	return n.Exp(G, key, P)
}

// exchange**a mod p
func DHSecret(key, exchange *big.Int) *big.Int {
	n := new(big.Int)
	return n.Exp(exchange, key, P)
}

func init() {
	P = new(big.Int)
	P.SetString("FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7EDEE386BFB5A899FA5AE9F24117C4B1FE649286651ECE65381FFFFFFFFFFFFFFFF", 16)
	G = new(big.Int)
	G.SetInt64(2)
}
