package routers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Args: PublisherURL
func (r *Routers) ForwardApplication() func(c *gin.Context) {
	return func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			fmt.Printf("error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		newApplication := ApplicationAnswer{
			InitiatorID:     r.MyIdentity,
			InitiatorURL:    r.MyURL,
			ServiceID:       data["ServiceID"].(string),
			ServiceName:     data["ServiceName"].(string),
			PublisherURL:    data["PublisherURL"].(string),
			Status:          0,
			ApplicationTime: time.Now().Format("2006-01-02 15:04:05")}
		r.MyApplication = append(r.MyApplication, newApplication)
		PublisherURL := data["PublisherURL"].(string)
		data["InitiatorID"] = r.MyIdentity
		data["InitiatorURL"] = r.MyURL
		// encode pubKey
		// encodedPubKey, err := encodePublicKey(MyPubKey)
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 	return
		// }
		//data["InitiatorPublicKeyCurve"] = MyPubKey.Curve.Params()
		data["InitiatorPublicKeyX"] = r.QueryContract.OrgSetup.PublicKey.X.Text(10)
		data["InitiatorPublicKeyY"] = r.QueryContract.OrgSetup.PublicKey.Y.Text(10)

		sendData, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		fmt.Println("======================== forward_application sending data ========================")
		fmt.Println("data: ", data)
		fmt.Println("objURL:", PublisherURL+"/send_application")

		req, err := http.NewRequest("POST", PublisherURL+"/send_application", bytes.NewBuffer(sendData))
		if err != nil {
			fmt.Printf("error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		refererURL := r.MyURL
		// refererURL, err := url.Parse((MyIP + ":" + port))
		// if err != nil {
		// 	fmt.Println("forward_application parsing url error: ", err)
		// 	return
		// }
		req.Header.Set("Referer", refererURL)
		fmt.Println("forward_application set header:", refererURL)

		httpClient := &http.Client{}
		res, err := httpClient.Do(req)
		if err != nil {
			fmt.Println("forward_application err on httpClient.Do()", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		fmt.Println("======================== forward_application receive response ========================")
		fmt.Println("res content: ", res)
		c.JSON(200, gin.H{"success": "success"})
		defer res.Body.Close()

	}
}
