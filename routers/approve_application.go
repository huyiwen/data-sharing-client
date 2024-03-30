package routers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handle application approval in the front-end. Args: ServiceID, InitiatorIdentity
func (r *Routers) IApproveApplication() func(c *gin.Context) {
	return func(c *gin.Context) {
		var httpData map[string]interface{}
		if err := c.ShouldBindJSON(&httpData); err != nil {
			fmt.Printf("error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		serviceID := httpData["ServiceID"].(string)
		recipientIdentity := httpData["InitiatorID"].(string)
		recipientIdentity = strings.ReplaceAll(recipientIdentity, " ", "")

		tokenId, err := r.ServiceContract.ApproveServiceFor(serviceID, recipientIdentity)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 4. 删除applicationToMe里的对应项，先别删了
		// for i := 0; i < len(r.ApplicationToMe); i++ {
		// 	if r.ApplicationToMe[i].InitiatorID == httpData["InitiatorID"].(string) && r.ApplicationToMe[i].ServiceID == httpData["ServiceID"].(string) {
		// 		r.ApplicationToMe = append(r.ApplicationToMe[:i], r.ApplicationToMe[i+1:]...)
		// 		break
		// 	}
		// }

		c.JSON(http.StatusOK, gin.H{"tokenId": tokenId})
	}
}
