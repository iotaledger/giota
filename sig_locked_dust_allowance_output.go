package iotago

import (
	"encoding/json"
	"fmt"
)

const (
	// SigLockedDustAllowanceOutputEd25519AddrBytesSize is the size of a SigLockedDustAllowanceOutput containing an Ed25519Address as its deposit address.
	SigLockedDustAllowanceOutputEd25519AddrBytesSize = SmallTypeDenotationByteSize + Ed25519AddressSerializedBytesSize + UInt64ByteSize

	// SigLockedDustAllowanceOutputBytesMinSize defines the minimum size of a SigLockedDustAllowanceOutput.
	SigLockedDustAllowanceOutputBytesMinSize = SigLockedDustAllowanceOutputEd25519AddrBytesSize
	// SigLockedDustAllowanceOutputAddressOffset defines the offset at which the address portion within a SigLockedDustAllowanceOutput begins.
	SigLockedDustAllowanceOutputAddressOffset = SmallTypeDenotationByteSize
)

// SigLockedDustAllowanceOutput functions like a SigLockedSingleOutput but as a special property
// it is used to increase the allowance/amount of dust outputs on a given address.
type SigLockedDustAllowanceOutput struct {
	// The actual address.
	Address Serializable `json:"address"`
	// The amount to deposit.
	Amount uint64 `json:"amount"`
}

func (s *SigLockedDustAllowanceOutput) Type() OutputType {
	return OutputSigLockedDustAllowanceOutput
}

func (s *SigLockedDustAllowanceOutput) Target() (Serializable, error) {
	return s.Address, nil
}

func (s *SigLockedDustAllowanceOutput) Deposit() (uint64, error) {
	return s.Amount, nil
}

func (s *SigLockedDustAllowanceOutput) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	return NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := checkMinByteLength(SigLockedDustAllowanceOutputBytesMinSize, len(data)); err != nil {
					return fmt.Errorf("invalid signature locked dust allowance output bytes: %w", err)
				}
				if err := checkTypeByte(data, OutputSigLockedDustAllowanceOutput); err != nil {
					return fmt.Errorf("unable to deserialize signature locked dust allowance output: %w", err)
				}
			}
			return nil
		}).
		Skip(SmallTypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip signature locked dust allowance output type during deserialization: %w", err)
		}).
		ReadObject(func(seri Serializable) { s.Address = seri }, deSeriMode, TypeDenotationByte, AddressSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize address for signature locked dust allowance output: %w", err)
		}).
		ReadNum(&s.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for signature locked dust allowance output: %w", err)
		}).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, s); err != nil {
					return fmt.Errorf("%w: unable to deserialize signature locked dust allowance output", err)
				}
			}
			return nil
		}).
		Done()
}

func (s *SigLockedDustAllowanceOutput) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	return NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, s); err != nil {
					return fmt.Errorf("%w: unable to serialize signature locked dust allowance output", err)
				}

				switch s.Address.(type) {
				case *Ed25519Address:
				default:
					return fmt.Errorf("%w: signature locked dust allowance output defines unknown address", ErrUnknownAddrType)
				}
			}
			return nil
		}).
		WriteNum(OutputSigLockedDustAllowanceOutput, func(err error) error {
			return fmt.Errorf("unable to serialize signature locked dust allowance output type ID: %w", err)
		}).
		WriteObject(s.Address, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize signature locked dust allowance output address: %w", err)
		}).
		WriteNum(s.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize signature locked dust allowance output amount: %w", err)
		}).Serialize()
}

func (s *SigLockedDustAllowanceOutput) MarshalJSON() ([]byte, error) {
	jSigLockedDustAllowanceOutput := &jsonSigLockedDustAllowanceOutput{}

	addrJsonBytes, err := s.Address.MarshalJSON()
	if err != nil {
		return nil, err
	}
	jsonRawMsgAddr := json.RawMessage(addrJsonBytes)

	jSigLockedDustAllowanceOutput.Type = int(OutputSigLockedDustAllowanceOutput)
	jSigLockedDustAllowanceOutput.Address = &jsonRawMsgAddr
	jSigLockedDustAllowanceOutput.Amount = int(s.Amount)
	return json.Marshal(jSigLockedDustAllowanceOutput)
}

func (s *SigLockedDustAllowanceOutput) UnmarshalJSON(bytes []byte) error {
	jSigLockedDustAllowanceOutput := &jsonSigLockedDustAllowanceOutput{}
	if err := json.Unmarshal(bytes, jSigLockedDustAllowanceOutput); err != nil {
		return err
	}
	seri, err := jSigLockedDustAllowanceOutput.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*SigLockedDustAllowanceOutput)
	return nil
}

// jsonSigLockedDustAllowanceOutput defines the json representation of a SigLockedDustAllowanceOutput.
type jsonSigLockedDustAllowanceOutput struct {
	Type    int              `json:"type"`
	Address *json.RawMessage `json:"address"`
	Amount  int              `json:"amount"`
}

func (j *jsonSigLockedDustAllowanceOutput) ToSerializable() (Serializable, error) {
	dep := &SigLockedDustAllowanceOutput{Amount: uint64(j.Amount)}

	jsonAddr, err := DeserializeObjectFromJSON(j.Address, jsonAddressSelector)
	if err != nil {
		return nil, fmt.Errorf("can't decode address type from JSON: %w", err)
	}

	dep.Address, err = jsonAddr.ToSerializable()
	if err != nil {
		return nil, err
	}
	return dep, nil
}
