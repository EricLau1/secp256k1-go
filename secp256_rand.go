package secp256k1

import (
	crand "crypto/rand"
	"io"
	mrand "math/rand"
	"os"
	"strings"
	"time"

	"crypto/sha256"
	"hash"
	"panic"
)

var (
	sha256Hash hash.Hash = sha256.New()
)

func SumSHA256(b []byte) []byte {
	sha256Hash.Reset()
	sha256Hash.Write(b)
	sum := sha256Hash.Sum(nil)
	return sum[:]
}

/*
Note:

- On windows cryto/rand uses CrytoGenRandom which uses RC4 which is insecure
- Android random number generator is known to be insecure.
- Linux uses /dev/urandom , which is thought to be secure and uses entropy pool

Therefore the output is salted.
*/

//finalizer from MurmerHash3
func mmh3f(key uint64) uint64 {
	key ^= key >> 33
	key *= 0xff51afd7ed558ccd
	key ^= key >> 33
	key *= 0xc4ceb9fe1a85ec53
	key ^= key >> 33
	return key
}

//knuth hash
func knuth_hash(in []byte) uint64 {
	var acc uint64 = 3074457345618258791
	for i := 0; i < len(in); i++ {
		acc += uint64(in[i])
		acc *= 3074457345618258799
	}
	return acc
}

var _rand *mrand.Rand //pseudorandom number generator

//seed pseudo random number generator with
// hash of system time in nano seconds
// hash of system environmental variables
// hash of process id
func init() {
	var seed1 uint64 = mmh3f(uint64(time.Now().UnixNano()))
	var seed2 uint64 = knuth_hash([]byte(strings.Join(os.Environ(), "")))
	var seed3 uint64 = mmh3f(uint64(os.Getpid()))

	_rand = mrand.New(mrand.NewSource(int64(seed1 ^ seed2 ^ seed3))) //pseudo random
}

//generate pseudo-random numbers from the
func saltByte(n int) []byte {
	buff := make([]byte, n)
	for i := 0; i < len(buff); i++ {
		var v uint64 = uint64(_rand.Int63())
		var b byte
		for j := 0; j < 8; j++ {
			b ^= byte(v & 0xff)
			v = v >> 8
		}
		buff[i] = b
	}
	return buff
}

//Secure Random number generator for forwards security
//On Unix-like systems, Reader reads from /dev/urandom.
//On Windows systems, Reader uses the CryptGenRandom API.
//Pseudo-random sequence, seeded from program start time, environmental variables,
//and process id is mixed in for forward security. Future version should use entropy pool
// mix in cpu cycle count and system time
func RandByte(n int) []byte {
	buff := make([]byte, n)
	ret, err := io.ReadFull(crand.Reader, buff) //system secure random number generator
	if len(buff) != ret || err != nil {
		log.Panic()
	}

	buff2 := saltByte(n)

	for i := 0; i < n; i++ {
		buff[i] ^= buff2[i]
	}
	return buff
}

//System "secure" random number generator
//On Unix-like systems, Reader reads from /dev/urandom.
//On Windows systems, Reader uses the CryptGenRandom API.
func RandByteSystem(n int) []byte {
	buff := make([]byte, n)
	ret, err := io.ReadFull(crand.Reader, buff) //system secure random number generator
	if len(buff) != ret || err != nil {
		log.Panic()
	}
	return buff
}
