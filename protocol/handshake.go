package protocol

type HandShakeFunc func(Peer) error

func DumbHandShakeFunc(Peer) error { return nil }
