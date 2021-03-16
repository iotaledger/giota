package iotago

import (
	"encoding/json"
	"fmt"
)

const (
	// Defines the binary serialized size of a TreasuryOutput.
	TreasuryOutputBytesSize = SmallTypeDenotationByteSize + UInt64ByteSize
)

// TreasuryOutput is an output which holds the treasury of a network.
type TreasuryOutput struct {
	// The currently residing funds in the treasury.
	Amount uint64 `json:"deposit"`
}

func (t *TreasuryOutput) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	return NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := checkMinByteLength(TreasuryOutputBytesSize, len(data)); err != nil {
					return fmt.Errorf("invalid treasury output bytes: %w", err)
				}
				if err := checkTypeByte(data, OutputTreasuryOutput); err != nil {
					return fmt.Errorf("unable to deserialize treasury output: %w", err)
				}
			}
			return nil
		}).
		Skip(SmallTypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip treasury output type during deserialization: %w", err)
		}).
		ReadNum(&t.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for treasury output: %w", err)
		}).
		Done()
}

func (t *TreasuryOutput) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	return NewSerializer().
		WriteNum(OutputTreasuryOutput, func(err error) error {
			return fmt.Errorf("unable to serialize treasury output type ID: %w", err)
		}).
		WriteNum(t.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize treasury output amount: %w", err)
		}).Serialize()
}

func (t *TreasuryOutput) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonTreasuryOutput{
		Type:   int(OutputTreasuryOutput),
		Amount: int(t.Amount),
	})
}

func (t *TreasuryOutput) UnmarshalJSON(bytes []byte) error {
	jTreasuryOutput := &jsonTreasuryOutput{}
	if err := json.Unmarshal(bytes, jTreasuryOutput); err != nil {
		return err
	}
	seri, err := jTreasuryOutput.ToSerializable()
	if err != nil {
		return err
	}
	*t = *seri.(*TreasuryOutput)
	return nil
}

// jsonTreasuryOutput defines the json representation of a TreasuryOutput.
type jsonTreasuryOutput struct {
	Type   int `json:"type"`
	Amount int `json:"amount"`
}

func (j *jsonTreasuryOutput) ToSerializable() (Serializable, error) {
	return &TreasuryOutput{Amount: uint64(j.Amount)}, nil
}
