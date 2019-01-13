package network

import (
	"Peerster/app"
	"Peerster/entities"
	"Peerster/fail"
	"Peerster/files"
	"Peerster/messages"
	"net"
)

/*OnReceiveArtTx handles a new transaction containing an artist/artwork pair.*/
func OnReceiveArtTx(gossiper *entities.Gossiper, artTx *messages.ArtTx, sender *net.UDPAddr) {

	// Add the contact to our routing table
	if gossiper.Args.Name != artTx.Artist.Name {
		gossiper.Router.AddContactIfAbsent(artTx.Artist.Name, sender)
	}

	// Attempt to add the artist to our database
	gossiper.ArtSystem.AddArtist(artTx.Artist)

	// Attempt to add the artwork to our database
	if toDownload := gossiper.ArtSystem.AddArtwork(artTx.Artwork); toDownload != nil {
		OnDownloadArtwork(gossiper, toDownload, artTx)
	}

}

/*OnInvalidateArtTx allows to invalidate transactions when a fork happens on the blockchain.*/
func OnInvalidateArtTx(gossiper *entities.Gossiper, artTx *messages.ArtTx) {
	gossiper.ArtSystem.InvalidateArtwork(artTx.Artwork)
}

/*OnSubscribe subscribes the user to an artist.*/
func OnSubscribe(gossiper *entities.Gossiper, signature string) {

	if toDownload, artist := gossiper.ArtSystem.Subscribe(signature); toDownload != nil {
		for _, artwork := range toDownload {
			go OnDownloadArtwork(gossiper, artwork, &messages.ArtTx{
				Artist:  artist.Info,
				Artwork: artwork.Info,
			})
		}
	}

}

/*OnDownloadArtwork downloads an artwork from the network.*/
func OnDownloadArtwork(gossiper *entities.Gossiper, artwork *app.Artwork, artTx *messages.ArtTx) {

	// Check that the remote peer exists
	target := gossiper.Router.GetTarget(artTx.Artist.Name)
	if target == nil {
		// @TODO: broadcast
		return
	}

	// Create a shared file
	shared := artwork.CreateSharedFile(gossiper.FileIndex, artTx)
	if shared == nil {
		// Error: filename already exists
		return
	}

	// Create metafile request
	request := &messages.DataRequest{Origin: gossiper.Args.Name,
		Destination: artTx.Artist.Name,
		HopLimit:    16,
		HashValue:   artwork.Info.Metahash[:],
	}

	// Send with timeout
	ref := files.NewHashRef(shared, 0)
	fail.LeveledPrint(0, "", "DOWNLOADING metafile of %s from %s", artwork.Info.Name, artTx.Artist.Name)
	OnSendTimedDataRequest(gossiper, request, ref, target)

}
