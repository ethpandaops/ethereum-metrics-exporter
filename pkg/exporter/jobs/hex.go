package jobs

import (
	"bytes"
	"encoding/hex"
	"math/big"
)

func hexStringToFloat64(hexStr string) float64 {
	f := new(big.Float)
	f.SetString(hexStr)
	balance, _ := f.Float64()

	return balance
}

func hexStringToString(hexStr string) (string, error) {
	bs, err := hex.DecodeString(hexStr[2:])
	if err != nil {
		return "", err
	}

	// split on EOT
	split := bytes.Split(bs, []byte{4})
	last := split[len(split)-1]

	// trim null bytes and spaces
	last = bytes.Trim(last, "\x00")
	last = bytes.TrimSpace(last)

	return string(last), nil
}
