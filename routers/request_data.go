package routers

import (
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Routers) CRequestData() func(*gin.Context) {
	return func(c *gin.Context) {
		// 接收从Initiator发来的数据申请，验证其权限后返回
		// 1. 解析申请数据包
		// 2. 调用链码 TODO 或许需要一个由InitiatorID+ServiceID共同索引的GetQuery
		// 3. 根据链码结果决定向Initiator的fetch_data发什么内容
		// 3.1 从数据库获取data

		var httpData map[string]interface{}
		if err := c.ShouldBindJSON(&httpData); err != nil {
			err = fmt.Errorf("failed to parse request data: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// data: ServiceID, InitiatorURL, InitiatorPublicKeyX, InitiatorPublicKeyY, InitiatorIdentity

		// serviceID format: Service-123
		serviceID := httpData["ServiceID"].(string)
		X := httpData["InitiatorPublicKeyX"].(string)
		Y := httpData["InitiatorPublicKeyY"].(string)
		identity := httpData["InitiatorIdentity"].(string)
		initiatorURL := httpData["InitiatorURL"].(string)
		publicKey := GetPublicKey(X, Y)
		certificate := fmt.Sprint(publicKey)
		service, valid := r.Config.Services[serviceID]
		if !valid {
			err := fmt.Errorf("service not found in config: %s", serviceID)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		createQuery := func(hashStr string, legitimacy string) string {
			queryID, err := r.QueryContract.CreateQuery(certificate, hashStr, 1, identity, "initiatorMSPID", legitimacy, service.Credentials.DatabaseTable, "SELECT * FROM crfm", serviceID)
			if err != nil {
				fmt.Printf("failed to create query: %s", err)
				return ""
			}
			return queryID
		}

		// verify signature
		fmt.Println("request_data get publicKey ", publicKey)
		verified, err := r.execVerify(initiatorURL, publicKey)
		if err != nil {
			fmt.Println("send_application verify signature err: ", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !verified {
			err := fmt.Errorf("failed to verify signature")
			queryID := createQuery("", "unkown user")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "queryID": queryID})
			return
		}

		access, err := r.ServiceContract.HasAccessToService(identity, serviceID)
		if err != nil {
			err = fmt.Errorf("failed to get balance: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !access {
			err = fmt.Errorf("insufficient balance")
			queryID := createQuery("", "no access")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "queryID": queryID})
			return
		}

		// 获取数据
		// data := "[{'apple': 10}, {'apple': 20}, {'apple': 30}]"
		data := r.dataBase(service.Credentials.DatabaseUser, service.Credentials.DatabasePassword, service.Credentials.DatabaseIP, service.Credentials.DatabasePort, service.Credentials.DatabaseName, service.Credentials.DatabaseTable)
		cryData, err := r.EnCryptByEcies(data, publicKey)
		if err != nil {
			panic(err)
		}

		hash := sha256.Sum256([]byte(data))
		hashStr := fmt.Sprintf("%x", hash)

		queryID := createQuery(hashStr, "true")

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("data requested: %s", hashStr), "queryID": queryID, "data": cryData})
	}
}
