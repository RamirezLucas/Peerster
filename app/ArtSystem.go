package app

import (
	"Peerster/fail"
	"Peerster/messages"
	"sync"
)

/*SigLenBytes is the length (in bytes) of an artist's signature. */
const SigLenBytes = 256

/*ArtSystem holds all the information about the art system.*/
type ArtSystem struct {
	artists       map[string]*Artist
	artworks      map[string]*Artwork
	subscriptions map[string]*Artist

	mux sync.Mutex
}

/*NewArtSystem instantiates a new `ArtSystem` and returns a pointer to
the newly created struct. */
func NewArtSystem() *ArtSystem {
	return &ArtSystem{
		artists:       make(map[string]*Artist),
		artworks:      make(map[string]*Artwork),
		subscriptions: make(map[string]*Artist),
	}
}

/*AddArtist adds an artist to the database.*/
func (art *ArtSystem) AddArtist(artist *messages.ArtistInfo) bool {
	// Grab the mutex
	art.mux.Lock()
	defer art.mux.Unlock()

	// Check that we don't already know the artist
	if _, ok := art.artists[artist.Signature]; ok {
		return false
	}

	// We don't know the artist
	art.artists[artist.Signature] = NewArtist(artist)
	return true
}

/*AddArtwork adds an artwork to the database.*/
func (art *ArtSystem) AddArtwork(artwork *messages.ArtworkInfo) *Artwork {
	// Grab the mutex
	art.mux.Lock()
	defer art.mux.Unlock()

	// Check that the artist exists
	if artist, ok := art.artists[artwork.AuthorSignature]; !ok {
		fail.CustomPanic("ArtSystem.AddArtwork",
			"Attempting to add artwork from inexistent artist.\n"+
				"\tArtwork: %s\n", artwork.ToString())
	} else {
		// Add the artwork to the database
		if newArtwork := artist.addArtwork(artwork); newArtwork != nil {
			if _, ok := art.subscriptions[artwork.AuthorSignature]; ok { // We are subscribed to the artist
				// Return the artwork to download
				return newArtwork
			}
		}
	}

	return nil
}

/*InvalidateArtwork invalidates an artwork in the database.*/
func (art *ArtSystem) InvalidateArtwork(artwork *messages.ArtworkInfo) bool {
	// Grab the mutex
	art.mux.Lock()
	defer art.mux.Unlock()

	// Check that the artwork exists
	if knownArtwork, ok := art.artworks[artwork.AuthorSignature]; !ok {
		fail.CustomPanic("ArtSystem.AddArtwork",
			"Attempting to invalidate inexistent artwork.\n"+
				"\tArtwork: %s\n", artwork.ToString())
	} else {
		return knownArtwork.Invalidate()
	}
	return false
}

/*Subscribe subscribes the user to a new artist in the database.
The receiver must have been instantiated with a call to `NewArtSystem()`.

`signature` The signature of the artists to subscribe to.*/
func (art *ArtSystem) Subscribe(signature string) ([]*Artwork, *Artist) {
	// Grab the mutex
	art.mux.Lock()
	defer art.mux.Unlock()

	// Check that the artist exists
	if artist, ok := art.artists[signature]; !ok {
		fail.CustomPanic("ArtSystem.AddArtwork",
			"Attempting to subscribe to inexistent artist %s.\n", signature)
	} else {
		// Check that we are not already subscribed to the artist
		if _, ok := art.subscriptions[signature]; !ok {
			// Subscribe
			art.subscriptions[signature] = artist

			// Return the list of all artworks to download
			var artworksToDownload []*Artwork
			for _, artwork := range artist.artworks {
				artworksToDownload = append(artworksToDownload, artwork)
			}
			return artworksToDownload, artist
		}
	}

	return nil, nil
}
