package network

import (
	"Peerster/entities"
	"Peerster/fail"
	"Peerster/files"
	"Peerster/messages"
	"crypto/sha256"
	"net"
	"time"

	"github.com/dedis/protobuf"
)

// DataRequestRepeatIntervalSec represents the amount of time after which a DataRequest is resent if it wasn't answered
const DataRequestRepeatIntervalSec = 5

// OnSendDataRequest - Sends a data request
func OnSendDataRequest(g *entities.Gossiper, request *messages.DataRequest, target *net.UDPAddr) error {

	// Create the packet
	pkt := messages.GossipPacket{DataRequest: request}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return &fail.CustomError{Fun: "OnSendDataRequest", Desc: "failed to encode DataRequest"}
	}

	// Send the packet
	if _, err := g.GossipChannel.WriteToUDP(buf, target); err != nil {
		return &fail.CustomError{Fun: "OnSendDataRequest", Desc: "failed to send DataRequest"}
	}

	return nil
}

// OnSendTimedDataRequest - Sends a data request with timeout
func OnSendTimedDataRequest(g *entities.Gossiper, request *messages.DataRequest,
	ref *files.HashRef, target *net.UDPAddr) {

	// Attempt to add the DataRequest to the index of pending requests
	if !g.TODataRequest.AddDataRequest(request, ref) {
		return
	}

	for {
		// Send the request
		if err := OnSendDataRequest(g, request, target); err != nil {
			return
		}

		// Wait for some time
		time.Sleep(DataRequestRepeatIntervalSec * time.Second)

		// Check if the response was received
		if g.TODataRequest.CheckResponseAndDelete(request.HashValue) {
			return
		}
	}

}

// OnSendDataReply - Sends a data reply
func OnSendDataReply(g *entities.Gossiper, reply *messages.DataReply, target *net.UDPAddr) {

	// Create the packet
	pkt := messages.GossipPacket{DataReply: reply}
	buf, err := protobuf.Encode(&pkt)
	if err != nil {
		return
	}

	// Send the packet
	g.GossipChannel.WriteToUDP(buf, target)
}

// OnReceiveDataRequest - Called when a data request is received
func OnReceiveDataRequest(g *entities.Gossiper, request *messages.DataRequest, sender *net.UDPAddr) {

	if g.Args.Name == request.Destination { // Message is for me
		// Add the contact to our routing table
		g.Router.AddContactIfAbsent(request.Origin, sender)

		if request.HashValue != nil {

			// Request data for this hash
			data := g.FileIndex.GetDataFromHash(request.HashValue)

			// Craft DataReply
			reply := &messages.DataReply{Origin: g.Args.Name,
				Destination: request.Origin,
				HopLimit:    16,
				HashValue:   request.HashValue,
				Data:        data,
			}

			// Pick the target (should exist) and send
			if target := g.Router.GetTarget(request.Origin); target != nil {
				OnSendDataReply(g, reply, target)
			}

		}

	} else { // Message is for someone else
		// Decrement hop limit
		request.HopLimit--

		// Send/Relay private message if hop-limit not exhausted
		if request.HopLimit != 0 {

			// Pick the target (should exist) and send
			target := g.Router.GetTarget(request.Destination)
			if target != nil {
				OnSendDataRequest(g, request, target)
			}
		}
	}

}

// OnReceiveDataReply - Called when a data reply is received
func OnReceiveDataReply(g *entities.Gossiper, reply *messages.DataReply, sender *net.UDPAddr) {

	if g.Args.Name == reply.Destination { // Message is for me

		// Check that the data contained in the message corresponds to the hash
		receivedDataHash := sha256.Sum256(reply.Data[:len(reply.Data)])
		if files.ToHex(reply.HashValue[:]) != files.ToHex(receivedDataHash[:]) {
			fail.LeveledPrint(1, "OnReceiveDataReply", `Received data doesn't correspond
				to hash: %s != %s `, files.ToHex(reply.HashValue[:]), files.ToHex(receivedDataHash[:]))
			return
		}

		// Look for the corresponding data request
		if ref := g.TODataRequest.SearchHashAndAcknowledge(reply); ref != nil {
			// Handle the reply and request next chunk if there is one
			if nextChunk, target := g.FileIndex.HandleDataReply(ref, reply); nextChunk != 0 {
				OnRemoteChunkRequest(g, ref.File, nextChunk, target)
			} else {
				fail.LeveledPrint(0, "", "RECONSTRUCTED file %s", ref.File.Filename)
			}
		}

	} else { // Message is for someone else
		// Decrement hop limit
		reply.HopLimit--

		// Send/Relay private message if hop-limit not exhausted
		if reply.HopLimit != 0 {

			// Pick the target (should exist) and send
			target := g.Router.GetTarget(reply.Destination)
			if target != nil {
				OnSendDataReply(g, reply, target)
			}
		}
	}

}

// OnRemoteChunkRequest - Request the chunks of a remote file
func OnRemoteChunkRequest(g *entities.Gossiper, file *files.SharedFile, chunkIndex uint64, remotePeer string) {

	// Check that the remote peer exists
	target := g.Router.GetTarget(remotePeer)
	if target == nil {
		return
	}

	// Get hash
	index := (chunkIndex - 1) * files.HashSizeBytes
	hash := file.Metafile[index : index+files.HashSizeBytes]

	// Create chunk request
	request := &messages.DataRequest{Origin: g.Args.Name,
		Destination: remotePeer,
		HopLimit:    16,
		HashValue:   hash,
	}

	// Send with timeout
	ref := files.NewHashRef(file, chunkIndex)
	fail.LeveledPrint(0, "", "DOWNLOADING %s chunk %d from %s\n", file.Filename, chunkIndex, remotePeer)
	OnSendTimedDataRequest(g, request, ref, target)
}

// OnRemoteMetaFileRequest - Request the metafile of a remote file
func OnRemoteMetaFileRequest(g *entities.Gossiper, metahash []byte, localFilename, remotePeer string) {

	// Check that the remote peer exists
	target := g.Router.GetTarget(remotePeer)
	if target == nil {
		return
	}

	// Create a shared file
	sharedFile := g.FileIndex.AddMonoSourceFile(localFilename, remotePeer, metahash)
	if sharedFile == nil {
		// Error: filename already exists
		return
	}

	// Create metafile request
	request := &messages.DataRequest{Origin: g.Args.Name,
		Destination: remotePeer,
		HopLimit:    16,
		HashValue:   metahash,
	}

	// Send with timeout
	ref := files.NewHashRef(sharedFile, 0)
	fail.LeveledPrint(0, "", "DOWNLOADING metafile of %s from %s\n", localFilename, remotePeer)
	OnSendTimedDataRequest(g, request, ref, target)
}
