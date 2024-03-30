package routers

import (
	"crypto/ecdsa"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

type ApplicationAnswer struct {
	InitiatorID  string `json:"InitiatorID"`
	InitiatorURL string `json:"InitiatorURL"`
	ServiceID    string `json:"ServiceID"`
	ServiceName  string `json:"ServiceName"`
	PublisherURL string `json:"PublisherURL"`
	// TODO
	ApplicationTime string `json:"ApplicationTime"`
	ProcessTime     string `json:"ProcessTime"`
	Status          int    `json:"Status"` // 0-pending 1-approved 2-rejected
}

type Application struct {
	InitiatorURL       string           `json:"InitiatorURL"`
	InitiatorPublicKey *ecdsa.PublicKey `json:"InitiatorPublicKey"`
	InitiatorID        string           `json:"InitiatorID"`
	ServiceID          string           `json:"ServiceID"`
	ServiceName        string           `json:"ServiceName"`
	ApplicationTime    string           `json:"ApplicationTime"`
	Status             int              `json:"Status"`
}

func (r *Routers) GetToMe() func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"applications": r.ApplicationToMe})
	}
}

func (r *Routers) GetSendOut() func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"applications": r.MyApplication})
	}
}

func (r *Routers) ListenTransfer(e *client.ChaincodeEvent) {
	fmt.Printf("Transfer event received: %s\n", e.Payload)
	if e.EventName == "Transfer" {
		fmt.Printf("Transfer event received: %s\n", e.Payload)
	}
}
