package iotago

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	// Defines the indexation payload's ID.
	IndexationPayloadTypeID uint32 = 2
	// type bytes + index prefix + one char + data length
	IndexationBinSerializedMinSize = TypeDenotationByteSize + UInt16ByteSize + OneByte + UInt32ByteSize
	// Defines the max length of the index within an Indexation.
	IndexationIndexMaxLength = 64
	// Defines the min length of the index within an Indexation.
	IndexationIndexMinLength = 1
)

var (
	// Returned when an Indexation's index exceeds IndexationIndexMaxLength.
	ErrIndexationIndexExceedsMaxSize = errors.New("index exceeds max size")
	// Returned when an Indexation's index is under IndexationIndexMinLength.
	ErrIndexationIndexUnderMinSize = errors.New("index is below min size")
)

// Indexation is a payload which holds an index and associated data.
type Indexation struct {
	// The index to use to index the enclosing message and data.
	Index []byte `json:"index"`
	// The data within the payload.
	Data []byte `json:"data"`
}

func (u *Indexation) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	return NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := checkMinByteLength(IndexationBinSerializedMinSize, len(data)); err != nil {
					return fmt.Errorf("invalid indexation bytes: %w", err)
				}
				if err := checkType(data, IndexationPayloadTypeID); err != nil {
					return fmt.Errorf("unable to deserialize indexation: %w", err)
				}
			}
			return nil
		}).
		Skip(TypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip indexation payload ID during deserialization: %w", err)
		}).
		ReadVariableByteSlice(&u.Index, SeriSliceLengthAsUint16, func(err error) error {
			return fmt.Errorf("unable to deserialize indexation index: %w", err)
		}, IndexationIndexMaxLength).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				switch {
				case len(u.Index) < IndexationIndexMinLength:
					return fmt.Errorf("unable to deserialize indexation index: %w", ErrIndexationIndexUnderMinSize)
				}
			}
			return nil
		}).
		ReadVariableByteSlice(&u.Data, SeriSliceLengthAsUint32, func(err error) error {
			return fmt.Errorf("unable to deserialize indexation data: %w", err)
		}, MessageBinSerializedMaxSize). // obviously can never be that size
		Done()
}

func (u *Indexation) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	return NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				switch {
				case len(u.Index) > IndexationIndexMaxLength:
					return fmt.Errorf("unable to serialize indexation index: %w", ErrIndexationIndexExceedsMaxSize)
				case len(u.Index) < IndexationIndexMinLength:
					return fmt.Errorf("unable to serialize indexation index: %w", ErrIndexationIndexUnderMinSize)
				}
				// we do not check the length of the data field as in any circumstance
				// the max size it can take up is dependent on how big the enclosing
				// parent object is
			}
			return nil
		}).
		WriteNum(IndexationPayloadTypeID, func(err error) error {
			return fmt.Errorf("unable to serialize indexation payload ID: %w", err)
		}).
		WriteVariableByteSlice(u.Index, SeriSliceLengthAsUint16, func(err error) error {
			return fmt.Errorf("unable to serialize indexation index: %w", err)
		}).
		WriteVariableByteSlice(u.Data, SeriSliceLengthAsUint32, func(err error) error {
			return fmt.Errorf("unable to serialize indexation data: %w", err)
		}).
		Serialize()
}

func (u *Indexation) MarshalJSON() ([]byte, error) {
	jsonIndexPayload := &jsonIndexation{}
	jsonIndexPayload.Type = int(IndexationPayloadTypeID)
	jsonIndexPayload.Index = hex.EncodeToString(u.Index)
	jsonIndexPayload.Data = hex.EncodeToString(u.Data)
	return json.Marshal(jsonIndexPayload)
}

func (u *Indexation) UnmarshalJSON(bytes []byte) error {
	jsonIndexPayload := &jsonIndexation{}
	if err := json.Unmarshal(bytes, jsonIndexPayload); err != nil {
		return err
	}
	seri, err := jsonIndexPayload.ToSerializable()
	if err != nil {
		return err
	}
	*u = *seri.(*Indexation)
	return nil
}

// jsonIndexation defines the json representation of an Indexation.
type jsonIndexation struct {
	Type  int    `json:"type"`
	Index string `json:"index"`
	Data  string `json:"data"`
}

func (j *jsonIndexation) ToSerializable() (Serializable, error) {
	indexBytes, err := hex.DecodeString(j.Index)
	if err != nil {
		return nil, fmt.Errorf("unable to decode index from JSON for indexation: %w", err)
	}

	dataBytes, err := hex.DecodeString(j.Data)
	if err != nil {
		return nil, fmt.Errorf("unable to decode data from JSON for indexation: %w", err)
	}

	return &Indexation{Index: indexBytes, Data: dataBytes}, nil
}
