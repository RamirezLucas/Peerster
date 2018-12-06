package network

import (
	"Peerster/entities"
	"Peerster/messages"
	"net"
)

/* ================ TRANSACTIONS ================ */

/*OnSendTransaction @TODO*/
func OnSendTransaction(channel *net.UDPConn, tx *messages.TxPublish, target *net.UDPAddr) {

}

/*OnReceiveTransaction @TODO*/
func OnReceiveTransaction(gossiper *entities.Gossiper, tx *messages.TxPublish, sender *net.UDPAddr) {

}

/* ================ BLOCKS ================ */

/*OnSendBlock @TODO*/
func OnSendBlock(channel *net.UDPConn, tx *messages.BlockPublish, target *net.UDPAddr) {

}

/*OnReceiveBlock @TODO*/
func OnReceiveBlock(gossiper *entities.Gossiper, tx *messages.BlockPublish, sender *net.UDPAddr) {

}
