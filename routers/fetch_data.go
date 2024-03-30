package routers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Routers) FetchData() func(c *gin.Context) {
	return func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			fmt.Printf("error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		data["InitiatorID"] = r.MyIdentity
		data["InitiatorURL"] = r.MyURL
		MyPubKey := r.QueryContract.OrgSetup.PublicKey
		data["InitiatorPublicKeyX"] = MyPubKey.X.Text(10)
		data["InitiatorPublicKeyY"] = MyPubKey.Y.Text(10)
		data["InitiatorIdentity"] = r.MyIdentity
		PublisherURL := data["PublisherURL"].(string)

		sendData, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		req, err := http.NewRequest("POST", PublisherURL+"/request_data", bytes.NewBuffer(sendData))
		if err != nil {
			fmt.Printf("error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		refererURL := r.MyURL
		req.Header.Set("Referer", refererURL)

		httpClient := &http.Client{}
		res, err := httpClient.Do(req)
		if err != nil {
			fmt.Println("forward_application err on httpClient.Do()", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if res.StatusCode != http.StatusOK {
			respBody, err := io.ReadAll(res.Body)
			if err != nil {
				fmt.Println("fetch_data err on ioutil.ReadAll()", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			var respData map[string]interface{}
			err = json.Unmarshal(respBody, &respData)
			if err != nil {
				fmt.Println("fetch_data err on json.Unmarshal()", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			fmt.Printf("error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": respData["error"]})
			return
		}
		defer res.Body.Close()

		fmt.Println("======================== fetch_data receive response ========================")

		respBody, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println("fetch_data err on ioutil.ReadAll()", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var respData map[string]interface{}
		err = json.Unmarshal(respBody, &respData)
		if err != nil {
			fmt.Println("fetch_data err on json.Unmarshal()", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		retdata := respData["data"].(string)
		// TODO: decrypt data
		responsdata, err := r.DeCryptByEcies(retdata, r.QueryContract.OrgSetup.PrivateKey)
		if err != nil {
			fmt.Println("fetch_data failed in decrypt the data.")
			panic(err)
		}

		c.JSON(200, gin.H{"data": responsdata})

	}
}
