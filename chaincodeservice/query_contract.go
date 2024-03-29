package chaincodeservice

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type Query struct {
	Certificate    string `json:"Certificate"`
	DataDigest     string `json:"DataDigest"`
	DataRows       int    `json:"DatatRows"`
	InitiatorID    string `json:"InitiatorID"`
	InitiatorMSPID string `json:"InitiatorMSPID"`
	Legitimacy     string `json:"Legitimacy"`
	QueriedTable   string `json:"QueriedTable"`
	QueryDigest    string `json:"QueryDigest"`
	QueryID        string `json:"QueryID"`
	ServiceID      string `json:"ServiceID"`
	Timestamp      int    `json:"Timestamp"`
}

type QueryContract struct {
	OrgSetup      *OrgSetup
	ChaincodeName string
	ChannelID     string
}

func (cc *QueryContract) CreateQuery(certificate, dataDigest string, dataRows int, initiatorID, initiatorMSPID, legitimacy, queriedTable, queryDigest, serviceID string) (string, error) {
	queryID, err := cc.getNextQueryID()
	if err != nil {
		return "", fmt.Errorf("error getting next QueryID: %s", err)
	}
	timestamp := time.Now().Unix()

	args := []string{certificate, dataDigest, strconv.Itoa(dataRows), initiatorID, initiatorMSPID, legitimacy, queriedTable, queryDigest, queryID, serviceID, strconv.FormatInt(timestamp, 10)}

	_, err = cc.OrgSetup.Invoke(cc.ChaincodeName, cc.ChannelID, "CreateQuery", args)
	if err != nil {
		return "", fmt.Errorf("error invoking CreateQuery: %s", err)
	}

	return queryID, err
}

func (cc *QueryContract) ReadQuery(queryID string) (string, error) {
	return cc.OrgSetup.Query(cc.ChaincodeName, cc.ChannelID, "ReadQuery", []string{queryID})
}

func (cc *QueryContract) QueryExists(queryID string) (string, error) {
	return cc.OrgSetup.Query(cc.ChaincodeName, cc.ChannelID, "QueryExists", []string{queryID})
}

func (cc *QueryContract) GetAllQuerys() ([]Query, error) {
	jsonQueries, err := cc.OrgSetup.Query(cc.ChaincodeName, cc.ChannelID, "GetAllQuerys", []string{})
	if err != nil {
		return nil, fmt.Errorf("error invoking GetAllQuerys: %s", err)
	}
	var queries []Query
	json.Unmarshal([]byte(jsonQueries), &queries)
	return queries, nil
}

func (cc *QueryContract) getNextQueryID() (string, error) {
	queries, err := cc.GetAllQuerys()
	if err != nil {
		return "", fmt.Errorf("error getting all queries: %s", err)
	}
	return strconv.Itoa(len(queries) + 1), nil
}
