package iota_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	legacy "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
	"github.com/iotaledger/iota.go/v2"
	"github.com/iotaledger/iota.go/v2/ed25519"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// returns length amount random bytes
func randBytes(length int) []byte {
	var b []byte
	for i := 0; i < length; i++ {
		b = append(b, byte(rand.Intn(256)))
	}
	return b
}

func randTrytes(length int) trinary.Trytes {
	var trytes strings.Builder
	for i := 0; i < length; i++ {
		trytes.WriteByte(legacy.TryteAlphabet[rand.Intn(len(legacy.TryteAlphabet))])
	}
	return trytes.String()
}

func rand32ByteHash() [32]byte {
	var h [32]byte
	b := randBytes(32)
	copy(h[:], b)
	return h
}

func rand49ByteHash() [49]byte {
	var h [49]byte
	b := randBytes(49)
	copy(h[:], b)
	return h
}

func sortedRand32ByteHashes(count int) [][32]byte {
	hashes := make(iota.LexicalOrdered32ByteArrays, count)
	for i := 0; i < count; i++ {
		hashes[i] = rand32ByteHash()
	}
	sort.Sort(hashes)
	return hashes
}

func rand64ByteHash() [64]byte {
	var h [64]byte
	b := randBytes(64)
	copy(h[:], b)
	return h
}

func randEd25519Addr() (*iota.Ed25519Address, []byte) {
	// type
	edAddr := &iota.Ed25519Address{}
	addr := randBytes(iota.Ed25519AddressBytesLength)
	copy(edAddr[:], addr)
	// serialized
	var b [iota.Ed25519AddressSerializedBytesSize]byte
	b[0] = iota.AddressEd25519
	copy(b[iota.SmallTypeDenotationByteSize:], addr)
	return edAddr, b[:]
}

func randEd25519Signature() (*iota.Ed25519Signature, []byte) {
	// type
	edSig := &iota.Ed25519Signature{}
	pub := randBytes(ed25519.PublicKeySize)
	sig := randBytes(ed25519.SignatureSize)
	copy(edSig.PublicKey[:], pub)
	copy(edSig.Signature[:], sig)
	// serialized
	var b [iota.Ed25519SignatureSerializedBytesSize]byte
	b[0] = iota.SignatureEd25519
	copy(b[iota.SmallTypeDenotationByteSize:], pub)
	copy(b[iota.SmallTypeDenotationByteSize+ed25519.PublicKeySize:], sig)
	return edSig, b[:]
}

func randEd25519SignatureUnlockBlock() (*iota.SignatureUnlockBlock, []byte) {
	edSig, edSigData := randEd25519Signature()
	block := &iota.SignatureUnlockBlock{Signature: edSig}
	return block, append([]byte{iota.UnlockBlockSignature}, edSigData...)
}

func randReferenceUnlockBlock() (*iota.ReferenceUnlockBlock, []byte) {
	return referenceUnlockBlock(uint16(rand.Intn(1000)))
}

func referenceUnlockBlock(index uint16) (*iota.ReferenceUnlockBlock, []byte) {
	var b [iota.ReferenceUnlockBlockSize]byte
	b[0] = iota.UnlockBlockReference
	binary.LittleEndian.PutUint16(b[iota.SmallTypeDenotationByteSize:], index)
	return &iota.ReferenceUnlockBlock{Reference: index}, b[:]
}

func randTransactionEssence() (*iota.TransactionEssence, []byte) {
	var buf bytes.Buffer

	tx := &iota.TransactionEssence{}
	must(buf.WriteByte(iota.TransactionEssenceNormal))

	inputsBytes := iota.LexicalOrderedByteSlices{}
	inputCount := rand.Intn(10) + 1
	must(binary.Write(&buf, binary.LittleEndian, uint16(inputCount)))
	for i := inputCount; i > 0; i-- {
		_, inputData := randUTXOInput()
		inputsBytes = append(inputsBytes, inputData)
	}

	sort.Sort(inputsBytes)

	for _, inputData := range inputsBytes {
		_, err := buf.Write(inputData)
		must(err)
		input := &iota.UTXOInput{}
		if _, err := input.Deserialize(inputData, iota.DeSeriModePerformValidation); err != nil {
			panic(err)
		}
		tx.Inputs = append(tx.Inputs, input)
	}

	outputsBytes := iota.LexicalOrderedByteSlices{}
	outputCount := rand.Intn(10) + 1
	must(binary.Write(&buf, binary.LittleEndian, uint16(outputCount)))
	for i := outputCount; i > 0; i-- {
		_, depData := randSigLockedSingleOutput(iota.AddressEd25519)
		outputsBytes = append(outputsBytes, depData)
	}

	sort.Sort(outputsBytes)
	for _, outputData := range outputsBytes {
		_, err := buf.Write(outputData)
		must(err)
		output := &iota.SigLockedSingleOutput{}
		if _, err := output.Deserialize(outputData, iota.DeSeriModePerformValidation); err != nil {
			panic(err)
		}
		tx.Outputs = append(tx.Outputs, output)
	}

	// empty payload
	must(binary.Write(&buf, binary.LittleEndian, uint32(0)))

	return tx, buf.Bytes()
}

func randMigratedFundsEntry() (*iota.MigratedFundsEntry, []byte) {
	tailTxHash := rand49ByteHash()
	addr, addrBytes := randEd25519Addr()
	deposit := rand.Uint64()

	var b bytes.Buffer
	_, err := b.Write(tailTxHash[:])
	must(err)
	_, err = b.Write(addrBytes)
	must(err)
	must(binary.Write(&b, binary.LittleEndian, deposit))

	return &iota.MigratedFundsEntry{
		TailTransactionHash: tailTxHash,
		Address:             addr,
		Deposit:             deposit,
	}, b.Bytes()
}

func randReceipt() (*iota.Receipt, []byte) {
	receipt := &iota.Receipt{MigratedAt: 1000, Final: true}

	var b bytes.Buffer

	must(binary.Write(&b, binary.LittleEndian, iota.ReceiptPayloadTypeID))
	must(binary.Write(&b, binary.LittleEndian, receipt.MigratedAt))
	must(b.WriteByte(1))

	migFundsEntriesBytes := iota.LexicalOrderedByteSlices{}
	migFundsEntriesCount := rand.Intn(10) + 1
	must(binary.Write(&b, binary.LittleEndian, uint16(migFundsEntriesCount)))
	for i := migFundsEntriesCount; i > 0; i-- {
		_, migFundsEntryBytes := randMigratedFundsEntry()
		migFundsEntriesBytes = append(migFundsEntriesBytes, migFundsEntryBytes)
	}

	sort.Sort(migFundsEntriesBytes)

	for _, migFundEntryBytes := range migFundsEntriesBytes {
		_, err := b.Write(migFundEntryBytes)
		must(err)
		migFundsEntry := &iota.MigratedFundsEntry{}
		if _, err := migFundsEntry.Deserialize(migFundEntryBytes, iota.DeSeriModePerformValidation); err != nil {
			panic(err)
		}
		receipt.Funds = append(receipt.Funds, migFundsEntry)
	}

	randTreasuryTx, randTreasuryTxBytes := randTreasuryTransaction()
	receipt.Transaction = randTreasuryTx

	must(binary.Write(&b, binary.LittleEndian, uint32(len(randTreasuryTxBytes))))
	if _, err := b.Write(randTreasuryTxBytes); err != nil {
		must(err)
	}

	return receipt, b.Bytes()
}

func randMilestone(parents iota.MessageIDs) (*iota.Milestone, []byte) {
	inclusionMerkleProof := randBytes(iota.MilestoneInclusionMerkleProofLength)
	const sigsCount = 3

	if parents == nil {
		parents = sortedRand32ByteHashes(1 + rand.Intn(7))
	}

	msPayload := &iota.Milestone{
		Index:     uint32(rand.Intn(1000)),
		Timestamp: uint64(time.Now().Unix()),
		Parents:   parents,
		InclusionMerkleProof: func() [iota.MilestoneInclusionMerkleProofLength]byte {
			b := [iota.MilestoneInclusionMerkleProofLength]byte{}
			copy(b[:], inclusionMerkleProof)
			return b
		}(),
		PublicKeys: func() [][iota.MilestonePublicKeyLength]byte {
			msPubKeys := make([][iota.MilestonePublicKeyLength]byte, sigsCount)
			for i := 0; i < sigsCount; i++ {
				msPubKeys[i] = rand32ByteHash()
				// ensure lexical ordering
				msPubKeys[i][0] = byte(i)
			}
			return msPubKeys
		}(),
		Signatures: func() [][iota.MilestoneSignatureLength]byte {
			msSigs := make([][iota.MilestoneSignatureLength]byte, sigsCount)
			for i := 0; i < sigsCount; i++ {
				msSigs[i] = randMilestoneSig()
			}
			return msSigs
		}(),
	}

	var b bytes.Buffer
	must(binary.Write(&b, binary.LittleEndian, iota.MilestonePayloadTypeID))
	must(binary.Write(&b, binary.LittleEndian, msPayload.Index))
	must(binary.Write(&b, binary.LittleEndian, msPayload.Timestamp))
	must(binary.Write(&b, binary.LittleEndian, byte(len(msPayload.Parents))))
	for _, parent := range msPayload.Parents {
		if _, err := b.Write(parent[:]); err != nil {
			panic(err)
		}
	}
	if _, err := b.Write(msPayload.InclusionMerkleProof[:]); err != nil {
		panic(err)
	}
	must(b.WriteByte(sigsCount))
	for _, pubKey := range msPayload.PublicKeys {
		if _, err := b.Write(pubKey[:]); err != nil {
			panic(err)
		}
	}
	must(binary.Write(&b, binary.LittleEndian, uint32(0)))
	must(b.WriteByte(sigsCount))
	for _, sig := range msPayload.Signatures {
		if _, err := b.Write(sig[:]); err != nil {
			panic(err)
		}
	}

	return msPayload, b.Bytes()
}

func randMilestoneSig() [iota.MilestoneSignatureLength]byte {
	var sig [iota.MilestoneSignatureLength]byte
	copy(sig[:], randBytes(iota.MilestoneSignatureLength))
	return sig
}

func randIndexation(dataLength ...int) (*iota.Indexation, []byte) {
	const index = "寿司を作って"

	var data []byte
	switch {
	case len(dataLength) > 0:
		data = randBytes(dataLength[0])
	default:
		data = randBytes(rand.Intn(200) + 1)
	}

	indexationPayload := &iota.Indexation{Index: []byte(index), Data: data}

	var b bytes.Buffer
	must(binary.Write(&b, binary.LittleEndian, iota.IndexationPayloadTypeID))

	strLen := uint16(len(index))
	must(binary.Write(&b, binary.LittleEndian, strLen))

	if _, err := b.Write([]byte(index)); err != nil {
		panic(err)
	}

	must(binary.Write(&b, binary.LittleEndian, uint32(len(indexationPayload.Data))))
	if _, err := b.Write(indexationPayload.Data); err != nil {
		panic(err)
	}

	return indexationPayload, b.Bytes()
}

func randMessage(withPayloadType uint32) (*iota.Message, []byte) {
	var payload iota.Serializable
	var payloadData []byte

	parents := sortedRand32ByteHashes(1 + rand.Intn(7))

	switch withPayloadType {
	case iota.TransactionPayloadTypeID:
		payload, payloadData = randTransaction()
	case iota.IndexationPayloadTypeID:
		payload, payloadData = randIndexation()
	case iota.MilestonePayloadTypeID:
		payload, payloadData = randMilestone(parents)
	}

	m := &iota.Message{}
	m.NetworkID = 1
	m.Payload = payload
	m.Nonce = uint64(rand.Intn(1000))
	m.Parents = parents

	var b bytes.Buffer
	must(binary.Write(&b, binary.LittleEndian, m.NetworkID))
	must(binary.Write(&b, binary.LittleEndian, byte(len(m.Parents))))

	for _, parent := range m.Parents {
		if _, err := b.Write(parent[:]); err != nil {
			panic(err)
		}
	}

	switch {
	case payload == nil:
		// zero length payload
		must(binary.Write(&b, binary.LittleEndian, uint32(0)))
	default:
		must(binary.Write(&b, binary.LittleEndian, uint32(len(payloadData))))
		if _, err := b.Write(payloadData); err != nil {
			panic(err)
		}
	}

	must(binary.Write(&b, binary.LittleEndian, m.Nonce))

	return m, b.Bytes()
}

func randTransaction() (*iota.Transaction, []byte) {
	var buf bytes.Buffer
	must(binary.Write(&buf, binary.LittleEndian, iota.TransactionPayloadTypeID))

	sigTxPayload := &iota.Transaction{}
	unTx, unTxData := randTransactionEssence()
	_, err := buf.Write(unTxData)
	must(err)
	sigTxPayload.Essence = unTx

	unlockBlocksCount := len(unTx.Inputs)
	must(binary.Write(&buf, binary.LittleEndian, uint16(unlockBlocksCount)))
	for i := unlockBlocksCount; i > 0; i-- {
		unlockBlock, unlockBlockData := randEd25519SignatureUnlockBlock()
		_, err := buf.Write(unlockBlockData)
		must(err)
		sigTxPayload.UnlockBlocks = append(sigTxPayload.UnlockBlocks, unlockBlock)
	}

	return sigTxPayload, buf.Bytes()
}

func randTreasuryInput() (*iota.TreasuryInput, []byte) {
	treasuryInput := &iota.TreasuryInput{}
	input := randBytes(iota.TreasuryInputBytesLength)
	copy(treasuryInput[:], input)
	var b [iota.TreasuryInputSerializedBytesSize]byte
	b[0] = iota.InputTreasury
	copy(b[iota.SmallTypeDenotationByteSize:], input)
	return treasuryInput, b[:]
}

func randUTXOInput() (*iota.UTXOInput, []byte) {
	utxoInput := &iota.UTXOInput{}
	var b [iota.UTXOInputSize]byte
	b[0] = iota.InputUTXO

	txID := randBytes(iota.TransactionIDLength)
	copy(b[iota.SmallTypeDenotationByteSize:], txID)
	copy(utxoInput.TransactionID[:], txID)

	index := uint16(rand.Intn(iota.RefUTXOIndexMax))
	binary.LittleEndian.PutUint16(b[len(b)-iota.UInt16ByteSize:], index)
	utxoInput.TransactionOutputIndex = index
	return utxoInput, b[:]
}

func randTreasuryOutput() (*iota.TreasuryOutput, []byte) {
	var b bytes.Buffer

	deposit := rand.Uint64()
	must(binary.Write(&b, binary.LittleEndian, iota.OutputTreasuryOutput))
	must(binary.Write(&b, binary.LittleEndian, deposit))

	return &iota.TreasuryOutput{Amount: deposit}, b.Bytes()
}

func randTreasuryTransaction() (*iota.TreasuryTransaction, []byte) {
	var b bytes.Buffer

	treasuryInput, treasuryInputBytes := randTreasuryInput()
	treasuryOutput, treasuryOutputBytes := randTreasuryOutput()
	must(binary.Write(&b, binary.LittleEndian, iota.TreasuryTransactionPayloadTypeID))
	_, err := b.Write(treasuryInputBytes)
	must(err)
	_, err = b.Write(treasuryOutputBytes)
	must(err)
	return &iota.TreasuryTransaction{
		Input:  treasuryInput,
		Output: treasuryOutput,
	}, b.Bytes()
}

func randSigLockedSingleOutput(addrType iota.AddressType) (*iota.SigLockedSingleOutput, []byte) {
	var buf bytes.Buffer
	must(buf.WriteByte(iota.OutputSigLockedSingleOutput))

	dep := &iota.SigLockedSingleOutput{}

	var addrData []byte
	switch addrType {
	case iota.AddressEd25519:
		dep.Address, addrData = randEd25519Addr()
	default:
		panic(fmt.Sprintf("invalid addr type: %d", addrType))
	}

	_, err := buf.Write(addrData)
	must(err)

	amount := uint64(rand.Intn(10000))
	must(binary.Write(&buf, binary.LittleEndian, amount))
	dep.Amount = amount

	return dep, buf.Bytes()
}

func oneInputOutputTransaction() *iota.Transaction {
	return &iota.Transaction{
		Essence: &iota.TransactionEssence{
			Inputs: []iota.Serializable{
				&iota.UTXOInput{
					TransactionID: func() [iota.TransactionIDLength]byte {
						var b [iota.TransactionIDLength]byte
						copy(b[:], randBytes(iota.TransactionIDLength))
						return b
					}(),
					TransactionOutputIndex: 0,
				},
			},
			Outputs: []iota.Serializable{
				&iota.SigLockedSingleOutput{
					Address: func() iota.Serializable {
						edAddr, _ := randEd25519Addr()
						return edAddr
					}(),
					Amount: 1337,
				},
			},
			Payload: nil,
		},
		UnlockBlocks: []iota.Serializable{
			&iota.SignatureUnlockBlock{
				Signature: func() iota.Serializable {
					edSig, _ := randEd25519Signature()
					return edSig
				}(),
			},
		},
	}
}

func randEd25519PrivateKey() ed25519.PrivateKey {
	seed := randEd25519Seed()
	return ed25519.NewKeyFromSeed(seed[:])
}

func randEd25519Seed() [ed25519.SeedSize]byte {
	var b [ed25519.SeedSize]byte
	read, err := rand.Read(b[:])
	if read != ed25519.SeedSize {
		panic(fmt.Sprintf("could not read %d required bytes from secure RNG", ed25519.SeedSize))
	}
	if err != nil {
		panic(err)
	}
	return b
}
