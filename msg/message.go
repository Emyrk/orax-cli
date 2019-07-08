package msg

import (
	"encoding/binary"
	"errors"

	"gitlab.com/pbernier3/orax-cli/common"
)

const CurrentVersion = 0

var log = common.GetLog()


type MessageType = uint8

const (
	// Signals are sent from the orchetrator to the miners
	MineSignal   MessageType = iota
	SubmitSignal MessageType = iota
	// Miner messages to the orchestrator
	MinerSubmission MessageType = iota
)

type Message struct {
	Version uint8
	Type    MessageType
}

type MineSignalMessage struct {
	Message
	OprHash []byte
}

func NewMineSignalMessage() *MineSignalMessage {
	ssm := new(MineSignalMessage)
	ssm.Version = CurrentVersion
	ssm.Type = MineSignal
	return ssm
}

func (mm *MineSignalMessage) Marshal() []byte {
	bytes := []byte{mm.Version, mm.Type}

	bytes = append(bytes, mm.OprHash...)

	return bytes
}

type SubmitSignalMessage struct {
	Message
	WindowDurationSec uint8
}

func NewSubmitSignalMessage() *SubmitSignalMessage {
	ssm := new(SubmitSignalMessage)
	ssm.Version = CurrentVersion
	ssm.Type = SubmitSignal
	return ssm
}

func (mm *SubmitSignalMessage) Marshal() []byte {
	bytes := make([]byte, 3)

	bytes[0] = mm.Version
	bytes[1] = mm.Type
	bytes[2] = mm.WindowDurationSec

	return bytes
}

type MinerSubmissionMessage struct {
	Message
	OprHash    []byte
	Nonce      []byte
	Difficulty uint64
	HashRate   uint64
}

func NewMinerSubmissionMessage() *MinerSubmissionMessage {
	ssm := new(MinerSubmissionMessage)
	ssm.Version = CurrentVersion
	ssm.Type = MinerSubmission
	return ssm
}

func (mm *MinerSubmissionMessage) Marshal() []byte {
	bytes := []byte{mm.Version, mm.Type}

	bytes = append(bytes, mm.OprHash...)
	bytes = append(bytes, mm.Nonce...)

	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, mm.Difficulty)
	bytes = append(bytes, bs...)

	binary.LittleEndian.PutUint64(bs, mm.HashRate)
	bytes = append(bytes, bs...)

	return bytes
}

////////////////////////////////////////
// Unmarshall
////////////////////////////////////////

func UnmarshalMessage(bytes []byte) (interface{}, error) {
	if len(bytes) <= 2 {
		return nil, errors.New("Message too short")
	}

	switch bytes[1] {
	case MineSignal:
		return unmarshalMineSignalMessage(bytes)
	case SubmitSignal:
		return unmarshalSubmitSignalMessage(bytes)
	case MinerSubmission:
		return unmarshalMinerSubmissionMessage(bytes)
	default:
		return nil, errors.New("Unknown message type")
	}
}

func unmarshalMineSignalMessage(bytes []byte) (*MineSignalMessage, error) {
	if len(bytes) != 34 {
		return nil, errors.New("Wrong message size")
	}

	m := new(MineSignalMessage)

	m.Version = bytes[0]
	m.Type = bytes[1]
	m.OprHash = bytes[2:]

	return m, nil
}

func unmarshalSubmitSignalMessage(bytes []byte) (*SubmitSignalMessage, error) {
	if len(bytes) != 3 {
		return nil, errors.New("Wrong message size")
	}

	m := new(SubmitSignalMessage)

	m.Version = bytes[0]
	m.Type = bytes[1]
	m.WindowDurationSec = bytes[2]

	return m, nil
}

func unmarshalMinerSubmissionMessage(bytes []byte) (*MinerSubmissionMessage, error) {
	if len(bytes) != 82 {
		return nil, errors.New("Wrong message size")
	}

	m := new(MinerSubmissionMessage)

	m.Version = bytes[0]
	m.Type = bytes[1]
	m.OprHash = bytes[2:34]
	m.Nonce = bytes[34:66]

	m.Difficulty = binary.LittleEndian.Uint64(bytes[66:74])
	m.HashRate = binary.LittleEndian.Uint64(bytes[74:82])

	return m, nil
}
