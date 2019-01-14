package messages

import (
	"fmt"
)

// HashSizeBytes is the size of a hash in bytes.
const HashSizeBytes = 32

/*ArtworkInfo represents an artwork as transfered over the blockchain*/
type ArtworkInfo struct {
	Name            string
	Description     string
	AuthorSignature string

	Filename string
	Metahash string
}

/*ToString returns a textual representation of an artwork. */
func (artwork *ArtworkInfo) ToString() string {
	return fmt.Sprintf("%s by %s", artwork.Name, artwork.AuthorSignature)
}
