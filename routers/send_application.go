package routers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (r *Routers) SendApplication() func(c *gin.Context) {
	return func(c *gin.Context) {
		fmt.Println("================= send_application start DEBUG ================")
		// fmt.Println("send_application receive referer: ", referer)
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println("send_application read body err:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var application map[string]interface{}
		err = json.Unmarshal(body, &application)
		if err != nil {
			fmt.Println("send_application unmarshal response err: ", err)
			return
		}
		fmt.Println("send_application receive application:", application)
		X := application["InitiatorPublicKeyX"].(string)
		Y := application["InitiatorPublicKeyY"].(string)
		InitiatorPublicKey := GetPublicKey(X, Y)
		fmt.Println("initiatorpublickeyX: ", InitiatorPublicKey.X, "y:  ", InitiatorPublicKey.Y)
		// InitiatorPublicKey, err := decodePublicKey(encodedPubKey.([]byte))
		fmt.Println("send_application decode public Key", InitiatorPublicKey)
		// InitiatorPublicKey := application["InitiatorPublicKey"]
		referer := c.Request.Referer()
		newUrl := referer
		// 验签
		verified, err := r.execVerify(referer, InitiatorPublicKey)
		if err != nil {
			fmt.Println("send_application verify signature err: ", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if verified {
			application["InitiatorURL"] = newUrl
			newApplication := Application{
				InitiatorPublicKey: InitiatorPublicKey,
				ApplicationTime:    time.Now().Format("2006-01-02 15:04:05"),
				InitiatorURL:       newUrl,
				InitiatorID:        application["InitiatorID"].(string),
				ServiceID:          application["ServiceID"].(string),
				ServiceName:        application["ServiceName"].(string),
			}
			fmt.Println("send_application successfully verify a signature.")
			r.ApplicationToMe = append(r.ApplicationToMe, newApplication)
			c.JSON(http.StatusOK, gin.H{"new_application": application})
		} else {
			fmt.Println("send_application failed in verifying a signature.")
			c.JSON(http.StatusBadRequest, nil)
		}
	}
}
