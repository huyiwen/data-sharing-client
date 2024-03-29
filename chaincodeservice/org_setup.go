package chaincodeservice

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
}

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

func (setup *OrgSetup) SubmitAsync(chainCodeName, channelID, function string, args []string) error {
	fmt.Printf("SubmitAsync channel: %s, chaincode: %s, function: %s, args: %s\n", channelID, chainCodeName, function, args)

	network := setup.Gateway.GetNetwork(channelID)
	contract := network.GetContract(chainCodeName)

	submitResult, commit, err := contract.SubmitAsync(function, client.WithArguments(args...))
	if err != nil {
		return fmt.Errorf("failed to submit transaction asynchronously: %w", err)
	}

	fmt.Printf("\n*** Successfully submitted transaction to transfer ownership from %s to Mark. \n", string(submitResult))
	fmt.Println("*** Waiting for transaction commit.")

	if commitStatus, err := commit.Status(); err != nil {
		return fmt.Errorf("failed to get commit status: %w", err)
	} else if !commitStatus.Successful {
		return fmt.Errorf("transaction %s failed to commit with status: %d", commitStatus.TransactionID, int32(commitStatus.Code))
	}

	return nil
}

// newGrpcConnection creates a gRPC connection to the Gateway server.
func (setup OrgSetup) newGrpcConnection() *grpc.ClientConn {
	certificate, err := loadCertificate(setup.TLSCertPath)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, setup.GatewayPeer)

	connection, err := grpc.Dial(setup.PeerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func (setup OrgSetup) newIdentity() (*identity.X509Identity, *ecdsa.PublicKey) {
	certificate, err := loadCertificate(setup.CertPath)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(setup.MSPID, certificate)
	if err != nil {
		panic(err)
	}

	pubKey, ok := certificate.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, nil
	}
	fmt.Println("newIdentity generates public key:", pubKey)

	return id, pubKey
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func (setup OrgSetup) newSign() (identity.Sign, crypto.Signer, *ecdsa.PrivateKey) {
	files, err := os.ReadDir(setup.KeyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key directory: %w", err))
	}
	privateKeyPEM, err := os.ReadFile(path.Join(setup.KeyPath, files[0].Name()))

	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	privateKeySigner, ok := privateKey.(crypto.Signer)
	if !ok {
		panic(fmt.Errorf("failed to generate signer"))
	}

	return sign, privateKeySigner, privateKey.(*ecdsa.PrivateKey)
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}
