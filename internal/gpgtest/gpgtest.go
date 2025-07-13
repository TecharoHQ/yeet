package gpgtest

import (
	"context"
	"os"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/TecharoHQ/yeet/internal/yeet"
)

// lastNRunes gets the last n runes of a string.
//
// This code was written by ChatGPT.
func lastNRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[len(runes)-n:])
}

func MakeKey(ctx context.Context, fname string) (string, error) {
	keyString, err := yeet.Output(ctx, "go", "tool", "gosop", "generate-key", "Yeet CI <social+yeet-ci@techaro.lol>")
	if err != nil {
		return "", err
	}

	key, err := crypto.NewKeyFromArmored(keyString)
	if err != nil {
		return "", err
	}

	return lastNRunes(key.GetFingerprint(), 16), os.WriteFile(fname, []byte(keyString), 0600)
}
