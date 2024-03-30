package routers

import (
	_ "crypto/ecdsa"
	"encoding/json"
	"fmt"
	"service-client/chaincodeservice"
	"time"
)

type Routers struct {
	Port            string
	QueryContract   chaincodeservice.QueryContract
	ServiceContract chaincodeservice.ServiceContract
	ApplicationToMe []Application
	MyApplication   []ApplicationAnswer
	Config          Config
	configFile      string
	OrgSetup        *chaincodeservice.OrgSetup
	MyURL           string
}

func Default(configFile string, getOrgSetup func(string) chaincodeservice.OrgSetup) *Routers {
	port, err := loadPort()
	if err != nil {
		panic(fmt.Errorf("error loading port: %s", err))
	}
	config, err := loadConfig(configFile)
	if err != nil {
		panic(fmt.Errorf("error loading config: %s", err))
	}

	myIP, err := getOuterIP()
	if err != nil {
		panic(fmt.Errorf("error getting outer IP: %s", err))
	}
	myURL := "http://" + myIP + ":" + port

	// setup org
	orgSetup, err := chaincodeservice.Initialize(getOrgSetup(port))
	if err != nil {
		panic(fmt.Errorf("error initializing OrgSetup: %s", err))
	}
	bytes, _ := json.Marshal(orgSetup)
	fmt.Printf("Initializing OrgSetup - OrgSetup %s\n", string(bytes))

	// get contracts
	queryContract := chaincodeservice.QueryContract{OrgSetup: orgSetup, ChaincodeName: config.QueryContract.ChaincodeName, ChannelID: config.QueryContract.ChannelID}
	serviceContract := chaincodeservice.ServiceContract{OrgSetup: orgSetup, ChaincodeName: config.ServiceContract.ChaincodeName, ChannelID: config.ServiceContract.ChannelID}

	// test contract
	queries, err := queryContract.GetAllQuerys()
	if err != nil {
		panic(fmt.Errorf("error initializing QueryContract: %s, check network status", err))
	}
	fmt.Printf("Initializing QueryContract - Queries: %d\n", len(queries))

	// initialize service contract if needed
	_ = serviceContract.Initialize("DataAccessCard", "dac", orgSetup.MSPID)
	time.Sleep(5 * time.Second)
	myIdentity, err := serviceContract.ClientAccountID()
	if err != nil {
		panic(fmt.Errorf("error initializing ServiceContract: %s, check service contract", err))
	}
	services, err := serviceContract.GetServices()
	if err != nil {
		panic(fmt.Errorf("error initializing ServiceContract: %s, check service contract initialization", err))
	}
	orgSetup.Identity = myIdentity
	fmt.Printf("Initializing ServiceContract - My Identity: %s\n", myIdentity)
	fmt.Printf("Initializing ServiceContract - Services: %d\n", len(services))

	r := Routers{
		Port:            port,
		QueryContract:   queryContract,
		ServiceContract: serviceContract,
		Config:          config,
		configFile:      configFile,
		OrgSetup:        orgSetup,
		MyURL:           myURL,
	}

	queryContract.StartListen(nil)
	serviceContract.StartListen([]chaincodeservice.EventListener{r.ListenTransfer})

	return &r
}
