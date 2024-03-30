package routers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ViewService struct {
	ServiceName    string `json:"ServiceName"`
	ServiceID      string `json:"ServiceID"`
	PublisherURL   string `json:"PublisherURL"`
	PublisherMSPID string `json:"PublisherMSPID"`
	Comment        string `json:"Comment"`
	Table          string `json:"Table"`
	Approved       bool   `json:"Approved"`
	NoAccess       bool   `json:"NoAccess"`
}

func (r *Routers) IGetServices() func(c *gin.Context) {
	return func(c *gin.Context) {
		serviceURIs, err := r.ServiceContract.GetServices()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"services": nil, "error": err.Error()})
			return
		}

		var services []ViewService
		for _, serviceURI := range serviceURIs {
			splitted := strings.Split(serviceURI, "|")
			serviceID, serviceURL := splitted[0], splitted[1]

			access, err := r.ServiceContract.HasAccessToService(r.MyIdentity, serviceID)
			if err != nil {
				access = false
			}

			publisher, err := r.ServiceContract.Owner()
			if err != nil {
				publisher = "Not available"
			}

			s := ViewService{
				ServiceName:    r.Config.Services[serviceID].Information.DisplayName,
				ServiceID:      serviceID,
				Comment:        r.Config.Services[serviceID].Information.Description,
				PublisherURL:   serviceURL,
				PublisherMSPID: publisher,
				Table:          r.Config.Services[serviceID].Credentials.DatabaseTable,
				Approved:       access,
				NoAccess:       !access,
			}
			services = append(services, s)
		}

		c.JSON(http.StatusOK, gin.H{"services": services})
	}
}
