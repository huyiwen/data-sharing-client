package chaincodeservice

import (
	"fmt"
	"strconv"
	"strings"
)

var servicePrefix = "Service-"
var mintPrefix = "Mint|"

type ServiceContract struct {
	OrgSetup      *OrgSetup
	ChaincodeName string
	ChannelID     string
}

func (cc *ServiceContract) StartListen(callbacks []EventListener) {

	cc.OrgSetup.StartListen(cc.ChaincodeName, cc.ChannelID, callbacks)
}

// Put a new service on chain. Return the new service ID
func (cc *ServiceContract) NewService(servicePostfix string) (string, error) {

	services, err := cc.GetServices()
	if err != nil {
		err = fmt.Errorf("failed to get balance: %w", err)
		return "", err
	}
	clientCount := cc.TotalSupply()

	tokenID := strconv.Itoa(clientCount)
	newServiceID := servicePrefix + strconv.Itoa(len(services))
	err = cc.MintWithTokenURI(tokenID, mintPrefix+newServiceID+"|"+servicePostfix)
	if err != nil {
		err = fmt.Errorf("failed to mint %s new service %s: %w", tokenID, newServiceID, err)
		return "", err
	}

	return newServiceID, nil
}

// Get all services
func (cc *ServiceContract) GetServices() ([]string, error) {

	var services []string
	maxTokenID := cc.TotalSupply()
	for i := 0; i < maxTokenID; i++ {
		tokenID := strconv.Itoa(i)

		tokenURI, err := cc.TokenURI(tokenID)
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(tokenURI, mintPrefix) {
			continue
		}
		services = append(services, tokenURI[len(mintPrefix):])
	}
	return services, nil
}

// Approve a service for a user. Returns the token ID
func (cc *ServiceContract) ApproveServiceFor(serviceID string, recipientIdentity string) (string, error) {

	totalSupply := cc.TotalSupply()
	tokenID := strconv.Itoa(totalSupply)
	err := cc.MintWithTokenURI(tokenID, serviceID)
	if err != nil {
		err = fmt.Errorf("failed to mint %s new service %s for %s: %w", tokenID, serviceID, recipientIdentity, err)
		return "", err
	}
	minter, err := cc.ClientAccountID()
	if err != nil {
		err = fmt.Errorf("failed to get operator: %w", err)
		return "", err
	}
	err = cc.TransferFrom(minter, recipientIdentity, tokenID)
	if err != nil {
		err = fmt.Errorf("failed to transfer %s new service %s for %s: %w", tokenID, serviceID, recipientIdentity, err)
		return "", err
	}
	return tokenID, nil
}

func (cc *ServiceContract) HasAccessToService(serviceID string) (bool, error) {

	operatorMSPID, err := cc.OwnerMSPID()
	if err != nil {
		return false, err
	}
	if operatorMSPID == cc.OrgSetup.MSPID {
		return true, nil
	}

	accessBalance, err := cc.BalanceOfByURIPrefix(cc.OrgSetup.Identity, serviceID)
	if err != nil {
		return false, err
	}
	return accessBalance > 0, nil
}

// ======= Original Contract Interfaces =======

func (cc *ServiceContract) Initialize(name string, symbol string, ownerMSPID string) error {
	_, err := cc.OrgSetup.Invoke(cc.ChaincodeName, cc.ChannelID, "Initialize", []string{name, symbol, ownerMSPID})
	return err
}

func (cc *ServiceContract) OwnerMSPID() (string, error) {
	return cc.OrgSetup.Query(cc.ChaincodeName, cc.ChannelID, "OwnerMSPID", []string{})
}

func (cc *ServiceContract) Burn(tokenId string) error {
	_, err := cc.OrgSetup.Invoke(cc.ChaincodeName, cc.ChannelID, "Burn", []string{tokenId})
	return err
}

func (cc *ServiceContract) OwnerOf(tokenId string) (string, error) {
	return cc.OrgSetup.Query(cc.ChaincodeName, cc.ChannelID, "OwnerOf", []string{tokenId})
}

func (cc *ServiceContract) TransferFrom(from string, to string, tokenId string) error {
	return cc.OrgSetup.Submit(cc.ChaincodeName, cc.ChannelID, "TransferFrom", []string{from, to, tokenId})

	// cmdString := fmt.Sprintf("CORE_PEER_TLS_ENABLED=true CORE_PEER_LOCALMSPID=\"Org1MSP\" CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt CORE_PEER_ADDRESS=localhost:7051 ../bin/peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile \"${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem\" --peerAddresses localhost:7051 --tlsRootCertFiles \"${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt\" --peerAddresses localhost:9051 --tlsRootCertFiles \"${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt\" -C %s -n %s -c '{\"function\":\"TransferFrom\",\"Args\":[\"%s\", \"%s\", \"%s\"]}'", cc.ChannelID, cc.ChaincodeName, from, to, tokenId)
	// err := os.WriteFile("/home/ubuntu/hyperledger/fabric-samples/test-network/chaincode-invoke.sh", []byte(cmdString), 0644)
	// if err != nil {
	// 	return err
	// }
	// cmd := exec.Command("/home/ubuntu/hyperledger/fabric-samples/test-network/chaincode-invoke.sh")
	// output, err := cmd.Output()
	// fmt.Println("output: ", string(output))
	// return nil
}

func (cc *ServiceContract) TokenURI(tokenId string) (string, error) {
	return cc.OrgSetup.Query(cc.ChaincodeName, cc.ChannelID, "TokenURI", []string{tokenId})
}

func (cc *ServiceContract) TotalSupply() int {
	supply, _ := cc.OrgSetup.Query(cc.ChaincodeName, cc.ChannelID, "TotalSupply", []string{})
	supplyInt, _ := strconv.Atoi(supply)
	return supplyInt
}

func (cc *ServiceContract) ClientAccountID() (string, error) {
	clientAccountID, err := cc.OrgSetup.Query(cc.ChaincodeName, cc.ChannelID, "ClientAccountID", []string{})
	if err != nil {
		return "", err
	}
	return clientAccountID, nil
}

func (cc *ServiceContract) MintWithTokenURI(tokenId, tokenURI string) error {
	_, err := cc.OrgSetup.Invoke(cc.ChaincodeName, cc.ChannelID, "MintWithTokenURI", []string{tokenId, tokenURI})
	if err != nil {
		return err
	}
	return nil
}

func (cc *ServiceContract) BalanceOf(owner string) (int, error) {

	balance, err := cc.OrgSetup.Query(cc.ChaincodeName, cc.ChannelID, "BalanceOf", []string{owner})
	if err != nil {
		return 0, err
	}

	balanceInt, err := strconv.Atoi(balance)
	if err != nil {
		return 0, err
	}

	return balanceInt, nil
}

func (cc *ServiceContract) BalanceOfByURIPrefix(owner string, tokenURI string) (int, error) {

	balance, err := cc.OrgSetup.Query(cc.ChaincodeName, cc.ChannelID, "BalanceOfByURIPrefix", []string{owner, tokenURI})
	if err != nil {
		return 0, err
	}

	balanceInt, err := strconv.Atoi(balance)
	if err != nil {
		return 0, err
	}

	return balanceInt, nil
}

func (cc *ServiceContract) BalanceOfByURI(owner string, tokenURI string) (int, error) {

	balance, err := cc.OrgSetup.Query(cc.ChaincodeName, cc.ChannelID, "BalanceOfByURI", []string{strings.Replace(owner, " ", "", -1), tokenURI})
	if err != nil {
		return 0, err
	}
	fmt.Println("balance: ", balance)

	balanceInt, err := strconv.Atoi(balance)
	if err != nil {
		return 0, err
	}

	return balanceInt, nil
}
