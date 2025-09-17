package gravatar

import (
	"crypto/md5"
	"fmt"
	"strings"
)

// URL returns the Gravatar image URL for the provided email.
// If size is non-positive, the default Gravatar size is used.
func URL(email string, size int) string {
	normalized := strings.TrimSpace(strings.ToLower(email))
	hash := md5.Sum([]byte(normalized))

	if size > 0 {
		return fmt.Sprintf("https://www.gravatar.com/avatar/%x?d=identicon&s=%d", hash, size)
	}

	return fmt.Sprintf("https://www.gravatar.com/avatar/%x?d=identicon", hash)
}
