package expiringlink

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ExpiringLink struct {
	// Epoch defines the oldest possible date the system
	// could use. It can be any date in the past, but once
	// selected must NEVER be changed, as doing so could
	// cause previously expired tokens to be valid again.
	// Dates closer to the present will result in shorter
	// hashes
	Epoch time.Time
	// Expire is how long before a generated token expires
	// (in seconds) This value can be changed at any time
	// and will affect any tokens generate after the change.
	Expire time.Duration
	// Rounds controls the complexity of the hash. Larger
	// values result in more secure hashes but use more
	// CPU. Can be changed at any time and will only
	// apply to newly generated hashes.
	Rounds int
}

func (e *ExpiringLink) Generate(secret string) string {
	expire := time.Since(e.Epoch) + e.Expire
	expTime := uint64(expire / time.Second)
	hash := hashRounds(e.Rounds, formatHash(expTime, e.Rounds, secret))
	return formatHash(expTime, e.Rounds, hash)
}

func (e *ExpiringLink) Check(hash, secret string) bool {
	part := strings.Split(hash, "g")
	ts, err := strconv.ParseInt(part[0], 16, 64)
	if err != nil {
		return false
	}
	if e.Epoch.Add(time.Second * time.Duration(ts)).Before(time.Now()) {
		return false
	}
	rounds, err := strconv.ParseInt(part[1], 16, 64)
	if err != nil {
		return false
	}
	genHash := hashRounds(int(rounds), formatHash(uint64(ts), int(rounds), secret))
	genFormatted := formatHash(uint64(ts), int(rounds), genHash)
	return genFormatted == hash
}

func formatHash(age uint64, rounds int, hash string) string {
	return fmt.Sprintf("%xg%xg%s", age, rounds, hash)
}

func hashRounds(rounds int, v string) string {
	hash := sha1.New()
	chain := []byte(v)
	for x := 0; x < rounds; x++ {
		hash.Write(chain)
		chain = hash.Sum(nil)
		hash.Reset()
	}
	return hex.EncodeToString(chain)
}
