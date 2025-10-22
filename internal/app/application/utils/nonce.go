package utils

import (
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

func GetPeerIDFromPubkey(pubkey []byte) (string, error) {
	pubKey, err := crypto.UnmarshalPublicKey(pubkey)
	if err != nil {
		return "", err
	}

	peerID, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return "", err
	}

	return peerID.String(), nil
}
