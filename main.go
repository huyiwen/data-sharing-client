package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"service-client/chaincodeservice"
	"service-client/routers"
)

func getOrgSetup(port string) chaincodeservice.OrgSetup {
	peers := "/home/ubuntu/hyperledger/fabric-samples/test-network/organizations/peerOrganizations"
	if port == "3999" {
		cryptoPath := peers + "/org1.example.com"
		return chaincodeservice.OrgSetup{
			OrgName:      "Org1",
			MSPID:        "Org1MSP",
			CryptoPath:   cryptoPath,
			CertPath:     cryptoPath + "/users/User1@org1.example.com/msp/signcerts/cert.pem",
			KeyPath:      cryptoPath + "/users/User1@org1.example.com/msp/keystore/",
			TLSCertPath:  cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt",
			PeerEndpoint: "localhost:7051",
			GatewayPeer:  "peer0.org1.example.com",
		}
	} else {
		cryptoPath := peers + "/org2.example.com"
		return chaincodeservice.OrgSetup{
			OrgName:      "Org2",
			MSPID:        "Org2MSP",
			CryptoPath:   cryptoPath,
			CertPath:     cryptoPath + "/users/User1@org2.example.com/msp/signcerts/cert.pem",
			KeyPath:      cryptoPath + "/users/User1@org2.example.com/msp/keystore/",
			TLSCertPath:  cryptoPath + "/peers/peer0.org2.example.com/tls/ca.crt",
			PeerEndpoint: "localhost:9051",
			GatewayPeer:  "peer0.org2.example.com",
		}
	}
}

func isInternalIP(ip string) bool {
	return ip == "127.0.0.1" || ip == "localhost" || ip == "62.234.49.75"
}

func internalOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if !isInternalIP(clientIP) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.Next()
	}
}

func listenConfig(r *routers.Routers) {
	r.ListenConfig()
}

func runApp(app *gin.Engine, port string) {
	app.Run(":" + port)
}

func main() {

	r := routers.Default("configs", getOrgSetup)
	app := gin.Default()
	app.LoadHTMLGlob("/home/ubuntu/webApp/data-sharing-webui/templates/*")

	// pages
	app.Static("/static", "/home/ubuntu/webApp/data-sharing-webui/static")
	app.GET("/", r.IIndex())
	app.GET("/login", r.ILogin())
	app.GET("/applicationToMe", r.IApplicationToMe())
	app.GET("/myApplication", r.IMyApplication())

	// apis
	app.Any("/send_application", r.SendApplication())
	app.POST("/request_data", r.CRequestData())
	app.POST("/put_service", r.IPutService())
	app.POST("/forward_application", r.ForwardApplication())
	app.POST("/receive_message", r.ReceiveMessage())
	app.POST("/fetch_data", r.FetchData())
	app.POST("/approve_application", r.IApproveApplication())
	app.POST("/debug_query", r.IDebugQuery())
	app.GET("/get_toMe", r.GetToMe())
	app.GET("/get_sendOut", r.GetSendOut())
	app.GET("/get_services", r.IGetServices())

	listenConfig(r)
	runApp(app, r.Port)
}
