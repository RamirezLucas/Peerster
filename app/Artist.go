package app

import (
	"Peerster/messages"
	"fmt"
)

/*Artist represents an artist.*/
type Artist struct {
	artworks map[string]*Artwork // List of artworks produced by this artist

	Info *messages.ArtistInfo
}

/*NewArtist instantiates a new `Artist` from a `ArtistInfo` and
returns a pointer to the newly created struct. */
func NewArtist(artistInfo *messages.ArtistInfo) *Artist {
	return &Artist{
		artworks: make(map[string]*Artwork),
		Info:     artistInfo,
	}
}

/*addArtwork returns a textual representation of an artist. */
func (artist *Artist) addArtwork(artwork *messages.ArtworkInfo) *Artwork {

	// Check if the artwork is already registered
	if existingArtwork, ok := artist.artworks[artwork.Name]; ok {

		// If it is, validate the transaction again (happens on fork)
		existingArtwork.Validate()
		// @TODO tell frontend

		return nil
	}

	newArtwork := NewArtwork(artwork)
	artist.artworks[artwork.Name] = newArtwork
	return newArtwork
}

/*ToString returns a textual representation of an artist. */
func (artist *Artist) ToString() string {
	return fmt.Sprintf("%s (%s)", artist.Info.Name, artist.Info.Signature)
}
