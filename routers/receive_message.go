package routers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Routers) ReceiveMessage() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 1. receive raw message
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			fmt.Printf("error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Println("============= receive_message start debug =============")
		fmt.Println("receive_message receive message", data)
		// 2. generate the signature
		message := data["message"]
		signature, err := SignMessage(message.(string), r.PrivateKeySigner)
		if err != nil {
			fmt.Println("receive_message SignMessage() err", err)
		}
		fmt.Println("receive_message generate signature", signature)
		fmt.Println("receive_message 's signer ", r.PrivateKeySigner)
		// content := map[string]interface{}{
		// 	"message": message,
		// 	"sign":    signature,
		// }
		// 3. send in response
		// sendData, _ := json.Marshal(content)
		// res, err := http.Post(service["sellerurl"]+"/receive_sign",
		// "application/json", bytes.NewBuffer([]byte(sendData)))
		//返回值
		c.JSON(http.StatusOK, gin.H{"message": message, "signature": signature})
	}
}
