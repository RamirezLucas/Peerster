package messages

import "fmt"

/*ArtistInfo represents an artist as transfered over the blockchain*/
type ArtistInfo struct {
	Name      string // Assumed to be unique
	Signature string // Unique signature
}

/*ToString returns a textual representation of an artist. */
func (artist *ArtistInfo) ToString() string {
	return fmt.Sprintf("%s (%s)", artist.Name, artist.Signature)
}
