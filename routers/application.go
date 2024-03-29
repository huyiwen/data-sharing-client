package routers

import (
	"crypto/ecdsa"
	"net/http"

	"github.com/gin-gonic/gin"
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
