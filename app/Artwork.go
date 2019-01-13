package app

import (
	"Peerster/files"
	"Peerster/messages"
	"fmt"
	"sync"
)

/*Artwork represents an artwork.*/
type Artwork struct {
	file                *files.SharedFile
	isValidInBlockchain bool
	isDownloaded        bool

	Info *messages.ArtworkInfo
	mux  sync.Mutex
}

/*NewArtwork instantiates a new `Artwork` from a `ArtowrkInfo` and
returns a pointer to the newly created struct. */
func NewArtwork(artworkInfo *messages.ArtworkInfo) *Artwork {
	return &Artwork{
		file:                nil,
		isValidInBlockchain: true,
		isDownloaded:        false,
		Info:                artworkInfo,
	}
}

/*Validate re-validates an artwork (happens on fork). */
func (artwork *Artwork) Validate() bool {
	// Grab the mutex
	artwork.mux.Lock()
	defer artwork.mux.Unlock()

	oldVal := artwork.isValidInBlockchain
	artwork.isValidInBlockchain = true
	return oldVal
}

/*Invalidate invalidates an artwork (happens on fork). */
func (artwork *Artwork) Invalidate() bool {
	// Grab the mutex
	artwork.mux.Lock()
	defer artwork.mux.Unlock()

	oldVal := artwork.isValidInBlockchain
	artwork.isValidInBlockchain = false
	return oldVal
}

/*CreateSharedFile creates the `SharedFile` structure corresponding to the artwork.*/
func (artwork *Artwork) CreateSharedFile(fileIndex *files.FileIndex, artTx *messages.ArtTx) *files.SharedFile {
	// Grab the mutex
	artwork.mux.Lock()
	defer artwork.mux.Unlock()

	return fileIndex.AddMonoSourceFile(artwork.Info.Filename, artwork.Info.Metahash[:], true, artTx)
}

/*ToString returns a textual representation of an artwork. */
func (artwork *Artwork) ToString() string {
	return fmt.Sprintf("%s buy %s. Is valid ? -> %t",
		artwork.Info.Name, artwork.Info.AuthorSignature, artwork.isValidInBlockchain)
}
