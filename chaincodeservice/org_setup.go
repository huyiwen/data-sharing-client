package chaincodeservice

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

type OrgSetup struct {
	OrgName          string
	MSPID            string
	CryptoPath       string
	CertPath         string
	KeyPath          string
	TLSCertPath      string
	PeerEndpoint     string
	GatewayPeer      string
	Gateway          client.Gateway
	PublicKey        *ecdsa.PublicKey
	PrivateKeySigner crypto.Signer
	PrivateKey       *ecdsa.PrivateKey
	Identity         string
}

type EventListener func(*client.ChaincodeEvent)

func Initialize(setup OrgSetup) (*OrgSetup, error) {
	log.Printf("Initializing connection for %s...\n", setup.OrgName)
	clientConnection := setup.newGrpcConnection()
	id, publicKey := setup.newIdentity()
	sign, privateKeySigner, privateKey := setup.newSign()

	// Connect to a Fabric Gateway using a client identity, gRPC connection and signing implementation.
	// func Connect(id identity.Identity, options ...ConnectOption) (*Gateway, error)
	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	setup.Gateway = *gateway
	setup.PublicKey = publicKey
	setup.PrivateKeySigner = privateKeySigner
	setup.PrivateKey = privateKey
	log.Println("Initialization complete")
	return &setup, nil
}

// query state from the ledger
func (setup OrgSetup) Query(chainCodeName, channelID, function string, args []string) (string, error) {

	fmt.Printf("Query channel: %s, chaincode: %s, function: %s, args: %s\n", channelID, chainCodeName, function, args)

	network := setup.Gateway.GetNetwork(channelID)
	contract := network.GetContract(chainCodeName)
	evaluateResponse, err := contract.EvaluateTransaction(function, args...)
	if err != nil {
		return "", err
	}

	return string(evaluateResponse), nil
}

// store state to the ledger
func (setup *OrgSetup) Invoke(chainCodeName, channelID, function string, args []string) (string, error) {

	fmt.Printf("Invoke channel: %s, chaincode: %s, function: %s, args: %s\n", channelID, chainCodeName, function, args)

	network := setup.Gateway.GetNetwork(channelID)
	contract := network.GetContract(chainCodeName)
	txn_proposal, err := contract.NewProposal(function, client.WithArguments(args...))
	if err != nil {
		return "", err
	}
	txn_endorsed, err := txn_proposal.Endorse()
	if err != nil {
		return "", err
	}
	txn_committed, err := txn_endorsed.Submit()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Transaction ID : %s Response: %s", txn_committed.TransactionID(), txn_endorsed.Result()), nil
}

func (setup *OrgSetup) Submit(chainCodeName, channelID, function string, args []string) error {

	fmt.Printf("Submit channel: %s, chaincode: %s, function: %s, args: %s\n", channelID, chainCodeName, function, args)

	network := setup.Gateway.GetNetwork(channelID)
	contract := network.GetContract(chainCodeName)

	_, err := contract.SubmitTransaction(function, args...)
	if err != nil {
		return fmt.Errorf("failed to submit transaction: %w", err)
	}
	return nil
}

func (setup *OrgSetup) StartListen(chainCodeName, channelID string, callbacks []EventListener) {

	network := setup.Gateway.GetNetwork(channelID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := network.ChaincodeEvents(ctx, chainCodeName)
	if err != nil {
		panic(fmt.Errorf("failed to start chaincode event listening: %w", err))
	}
	fmt.Printf("Listening for chaincode events on channel %s for chaincode %s\n", channelID, chainCodeName)

	go func() {
		for event := range events {
			fmt.Printf("Received chaincode event: %s %s %s\n", event.EventName, event.TransactionID, event.Payload)
		}
	}()
}
