package msg

import (
	"encoding/json"
	"errors"

	"gitlab.com/pbernier3/orax-cli/common"
)

var log = common.GetLog()

type MessageType = uint8

const (
	// Signals are sent from the orchetrator to the miners
	MineSignal   MessageType = iota
	SubmitSignal MessageType = iota
	// Miner messages to the orchestrator
	MinerSubmission MessageType = iota
)

type MineSignalMessage struct {
	MaxNonces int    `json:"maxNonces"`
	OprHash   []byte `json:"oprHash"`
}

func NewMineSignalMessage(oprHash []byte, maxNonces int) *MineSignalMessage {
	m := new(MineSignalMessage)
	m.OprHash = oprHash
	m.MaxNonces = maxNonces
	return m
}

type SubmitSignalMessage struct {
	WindowDurationSec int `json:"oprHash"`
}

type MinerSubmissionMessage struct {
	OprHash  []byte  `json:"oprHash"`
	Nonces   []Nonce `json:"nonces"`
	Duration int64   `json:"duration"`
	OpCount  int64   `json:"opCount"`
}

type Nonce struct {
	Nonce      []byte `json:"nonce"`
	Difficulty uint64 `json:"diff"`
}

func (msm *MineSignalMessage) Marshal() ([]byte, error) {
	return marshal(MineSignal, msm)
}

func (ssm *SubmitSignalMessage) Marshal() ([]byte, error) {
	return marshal(SubmitSignal, ssm)
}

func (msm *MinerSubmissionMessage) Marshal() ([]byte, error) {
	return marshal(MinerSubmission, msm)
}

func marshal(t MessageType, i interface{}) ([]byte, error) {
	bytes := []byte{t}

	j, err := json.Marshal(i)
	if err != nil {
		return []byte{}, err
	}
	bytes = append(bytes, j...)

	return bytes, nil
}

////////////////////////////////////////
// Unmarshall
////////////////////////////////////////

func UnmarshalMessage(bytes []byte) (interface{}, error) {
	if len(bytes) <= 1 {
		return nil, errors.New("Message too short")
	}

	switch bytes[0] {
	case MineSignal:
		return unmarshalMineSignalMessage(bytes[1:])
	case SubmitSignal:
		return unmarshalSubmitSignalMessage(bytes[1:])
	case MinerSubmission:
		return unmarshalMinerSubmissionMessage(bytes[1:])
	default:
		return nil, errors.New("Unknown message type")
	}
}

func unmarshalMineSignalMessage(bytes []byte) (*MineSignalMessage, error) {
	var msm MineSignalMessage
	err := json.Unmarshal(bytes, &msm)
	return &msm, err
}

func unmarshalSubmitSignalMessage(bytes []byte) (*SubmitSignalMessage, error) {
	var ssm SubmitSignalMessage
	err := json.Unmarshal(bytes, &ssm)
	return &ssm, err
}

func unmarshalMinerSubmissionMessage(bytes []byte) (*MinerSubmissionMessage, error) {
	var msm MinerSubmissionMessage
	err := json.Unmarshal(bytes, &msm)
	return &msm, err
}
