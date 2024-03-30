package routers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ViewService struct {
	ServiceName        string `json:"ServiceName"`
	ServiceID          string `json:"ServiceID"`
	PublisherURL       string `json:"PublisherURL"`
	PublisherPublicKey string `json:"PublisherPublicKey"`
	Comment            string `json:"Comment"`
	Table              string `json:"Table"`
	Approved           bool   `json:"Approved"`
	NoAccess           bool   `json:"NoAccess"`
}

func (r *Routers) IGetServices() func(c *gin.Context) {
	return func(c *gin.Context) {
		serviceURIs, err := r.ServiceContract.GetServices()
		if err != nil {
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

			s := ViewService{
				ServiceName:        r.Config.Services[serviceID].Information.DisplayName,
				ServiceID:          serviceID,
				Comment:            r.Config.Services[serviceID].Information.Description,
				PublisherURL:       serviceURL,
				PublisherPublicKey: "Not available",
				Table:              r.Config.Services[serviceID].Credentials.DatabaseTable,
				Approved:           access,
				NoAccess:           !access,
			}
			services = append(services, s)
		}

		c.JSON(http.StatusOK, gin.H{"services": services})
	}
}
