package routers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Args: DisplayName, Description, IP, Port, User, Password, Database, Table
func (r *Routers) IPutService() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 1. 接受前端传来的httpData
		var httpData map[string]interface{}
		if err := c.ShouldBindJSON(&httpData); err != nil {
			fmt.Printf("error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// 2. 生成唯一的ServiceID
		serviceID, err := r.ServiceContract.NewService(r.MyURL)
		if err != nil {
			err = fmt.Errorf("failed to generate new service ID: %v", err)
			fmt.Printf("error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		information := ServiceInformation{
			DisplayName: httpData["serviceName"].(string),
			Description: httpData["comment"].(string),
		}
		var credentials ServiceCredentials
		if _, ok := httpData["Password"]; ok {
			credentials = ServiceCredentials{
				DatabaseIP:       httpData["IP"].(string),
				DatabasePort:     httpData["Port"].(string),
				DatabaseUser:     httpData["User"].(string),
				DatabasePassword: httpData["Password"].(string),
				DatabaseName:     httpData["Database"].(string),
				DatabaseTable:    httpData["Table"].(string),
			}
		}

		// 3. 将httpData存入数据库
		r.Config.Services[serviceID] = ServiceType{Information: information, Credentials: credentials}
		err = r.updateConfig()
		if err != nil {
			err = fmt.Errorf("failed to update config: %v", err)
			fmt.Printf("error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"serviceID": serviceID,
			"error_msg": "None",
		})
	}
}
