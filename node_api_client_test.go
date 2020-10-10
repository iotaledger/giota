package iota_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

const nodeAPIUrl = "http://127.0.0.1:14265"

func TestNodeAPI_Info(t *testing.T) {
	defer gock.Off()

	originInfo := &iota.NodeInfoResponse{
		Name:                     "HORNET",
		Version:                  "1.0.0",
		IsHealthy:                true,
		CoordinatorPublicKey:     "733ed2810f2333e9d6cd702c7d5c8264cd9f1ae454b61e75cf702c451f68611d",
		LatestMilestoneMessageID: "5e4a89c549456dbec74ce3a21bde719e9cd84e655f3b1c5a09058d0fbf9417fe",
		LatestMilestoneIndex:     1337,
		SolidMilestoneMessageID:  "598f7a3186bf7291b8199a3147bb2a81d19b89ac545788b4e5d8adbee7db0f13",
		SolidMilestoneIndex:      666,
		PruningIndex:             142857,
		Features:                 []string{"Lazers"},
	}

	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRouteInfo).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originInfo})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	info, err := nodeAPI.Info()
	assert.NoError(t, err)
	assert.EqualValues(t, originInfo, info)
}

func TestNodeAPI_Tips(t *testing.T) {
	defer gock.Off()

	originRes := &iota.NodeTipsResponse{
		Tip1: "733ed2810f2333e9d6cd702c7d5c8264cd9f1ae454b61e75cf702c451f68611d",
		Tip2: "5e4a89c549456dbec74ce3a21bde719e9cd84e655f3b1c5a09058d0fbf9417fe",
	}

	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRouteTips).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	tips, err := nodeAPI.Tips()
	assert.NoError(t, err)
	assert.EqualValues(t, originRes, tips)
}

func TestNodeAPI_MessageMetadataByMessageID(t *testing.T) {
	defer gock.Off()

	identifier := rand32ByteHash()
	parent1 := rand32ByteHash()
	parent2 := rand32ByteHash()

	queryHash := hex.EncodeToString(identifier[:])
	parent1MessageID := hex.EncodeToString(parent1[:])
	parent2MessageID := hex.EncodeToString(parent2[:])

	originRes := &iota.MessageMetadataResponse{
		MessageID:                  queryHash,
		Parent1:                    parent1MessageID,
		Parent2:                    parent2MessageID,
		Solid:                      true,
		ReferencedByMilestoneIndex: nil,
		LedgerInclusionState:       nil,
		ShouldPromote:              nil,
		ShouldReattach:             nil,
	}

	gock.New(nodeAPIUrl).
		Get(strings.Replace(iota.NodeAPIRouteMessageMetadata, iota.ParameterMessageID, queryHash, 1)).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	meta, err := nodeAPI.MessageMetadataByMessageID(identifier)
	assert.NoError(t, err)

	metaJson, err := json.Marshal(meta)
	assert.NoError(t, err)

	originJson, err := json.Marshal(originRes)
	assert.NoError(t, err)

	assert.EqualValues(t, originJson, metaJson)
}

func TestNodeAPI_MessageByMessageID(t *testing.T) {
	defer gock.Off()

	identifier := rand32ByteHash()
	queryHash := hex.EncodeToString(identifier[:])

	originMsg := &iota.Message{
		Version: 1,
		Parent1: rand32ByteHash(),
		Parent2: rand32ByteHash(),
		Payload: nil,
		Nonce:   16345984576234,
	}

	responseBuf := &bytes.Buffer{}

	data, err := originMsg.Serialize(iota.DeSeriModePerformValidation)
	assert.NoError(t, err)
	responseBuf.Write(data)

	gock.New(nodeAPIUrl).
		Get(strings.Replace(iota.NodeAPIRouteMessageBytes, iota.ParameterMessageID, queryHash, 1)).
		Reply(200).
		Body(responseBuf)

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	responseMsg, err := nodeAPI.MessageByMessageID(identifier)
	assert.NoError(t, err)

	responseMsgJson, err := json.Marshal(responseMsg)
	assert.NoError(t, err)

	originMsgJson, err := originMsg.MarshalJSON()
	assert.NoError(t, err)

	assert.EqualValues(t, originMsgJson, responseMsgJson)
}

func TestNodeAPI_ChildrenByMessageID(t *testing.T) {
}

func TestNodeAPI_MessageIDsByIndex(t *testing.T) {
	/*
		defer gock.Off()

		tag := "बेकार पाठ"

		msg := &iota.Message{
			Version: 1,
			Parent1: rand32ByteHash(),
			Parent2: rand32ByteHash(),
			Payload: nil,
			Nonce:   16345984576234,
		}

		gock.New(nodeAPIUrl).
			Get(iota.NodeAPIRouteMessagesByTag).
			MatchParam("tags", tag).
			Reply(200).
			JSON(&iota.HTTPOkResponseEnvelope{Data: []*iota.Message{msg}})

		nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
		msgs, err := nodeAPI.MessagesByTag(tag)
		assert.NoError(t, err)
		assert.Len(t, msgs, 1)

		msgJson, err := json.Marshal(msgs[0])
		assert.NoError(t, err)

		originMsgJson, err := msg.MarshalJSON()
		assert.NoError(t, err)

		assert.EqualValues(t, originMsgJson, msgJson)
	*/
}

func TestNodeAPI_SubmitMessage(t *testing.T) {
	/*
		defer gock.Off()

		msgHash := rand32ByteHash()
		msgHashStr := hex.EncodeToString(msgHash[:])

		incompleteMsg := &iota.Message{Version: 1}
		completeMsg := &iota.Message{
			Version: 1,
			Parent1: rand32ByteHash(),
			Parent2: rand32ByteHash(),
			Payload: nil,
			Nonce:   3495721389537486,
		}

		gock.New(nodeAPIUrl).
			Post(iota.NodeAPIRouteMessageSubmit).
			MatchType("json").
			JSON(incompleteMsg).
			Reply(200).AddHeader("Location", msgHashStr)

		gock.New(nodeAPIUrl).
			Get(iota.NodeAPIRouteMessagesByID).
			MatchParam("hashes", msgHashStr).
			Reply(200).
			JSON(&iota.HTTPOkResponseEnvelope{Data: []*iota.Message{completeMsg}})

		nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
		resp, err := nodeAPI.SubmitMessage(incompleteMsg)
		assert.NoError(t, err)

		assert.EqualValues(t, completeMsg, resp)
	*/
}

func TestNodeAPI_MilestoneByIndex(t *testing.T) {
}

func TestNodeAPI_OutputByID(t *testing.T) {
	/*
		originOutput, _ := randSigLockedSingleOutput(iota.AddressEd25519)
		sigDepJson, err := originOutput.MarshalJSON()
		assert.NoError(t, err)
		rawMsgSigDepJson := json.RawMessage(sigDepJson)

		txID := rand32ByteHash()
		hexTxID := hex.EncodeToString(txID[:])
		originRes := []iota.NodeOutputResponse{
			{
				HexTransactionID: hexTxID,
				OutputIndex:      3,
				Spent:            true,
				RawOutput:        &rawMsgSigDepJson,
			},
		}

		utxoInput := &iota.UTXOInput{TransactionID: txID, TransactionOutputIndex: 3}
		utxoInputId := utxoInput.ID()

		gock.New(nodeAPIUrl).
			Get(iota.NodeAPIRouteOutputsByID).
			MatchParam("ids", utxoInputId.ToHex()).
			Reply(200).
			JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

		nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
		resp, err := nodeAPI.OutputsByID(iota.UTXOInputIDs{utxoInputId})
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.EqualValues(t, originRes, resp)

		respOutput, err := resp[0].Output()
		assert.NoError(t, err)
		assert.EqualValues(t, originOutput, respOutput)

		sigTxPayloadHash, err := resp[0].TransactionID()
		assert.NoError(t, err)
		assert.EqualValues(t, txID, *sigTxPayloadHash)
	*/
}

func TestNodeAPI_BalanceByAddress(t *testing.T) {
}

func TestNodeAPI_OutputIDsByAddress(t *testing.T) {
	/*
		originOutput, _ := randSigLockedSingleOutput(iota.AddressEd25519)
		sigDepJson, err := originOutput.MarshalJSON()
		assert.NoError(t, err)
		rawMsgSigDepJson := json.RawMessage(sigDepJson)

		addr, _ := randEd25519Addr()
		addrHex := addr.String()
		originRes := map[string][]iota.NodeOutputResponse{
			addrHex: {{RawOutput: &rawMsgSigDepJson, Spent: true}},
		}

		gock.New(nodeAPIUrl).
			Get(iota.NodeAPIRouteOutputsByAddress).
			MatchParam("addresses", addrHex).
			Reply(200).
			JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

		nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
		resp, err := nodeAPI.OutputsByAddress(addrHex)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.EqualValues(t, originRes, resp)

		respOutput, err := resp[addrHex][0].Output()
		assert.NoError(t, err)
		assert.EqualValues(t, originOutput, respOutput)
	*/
}
