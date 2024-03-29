package routers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Routers) IDebugQuery() func(*gin.Context) {
	return func(c *gin.Context) {

		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var onChainResults string
		var err error

		function := data["Function"].(string)

		if function == "QueryExists" {
			onChainResults, err = r.QueryContract.QueryExists(data["QueryID"].(string))

		} else if function == "ReadQuery" {
			queryID := data["QueryID"].(string)
			fmt.Printf("type of queryID (%s): %T\n", queryID, queryID)
			onChainResults, err = r.QueryContract.ReadQuery(queryID)

		} else if function == "GetAllQuerys" {
			queries, err := r.QueryContract.GetAllQuerys()
			if err != nil {
				err = fmt.Errorf("failed to %s: %s", function, err)
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			for _, query := range queries {
				fmt.Printf("query: %+v\n", query)
			}
			results, _ := json.Marshal(queries)
			onChainResults = string(results)

		} else if function == "ClientAccountID" {
			onChainResults, err = r.ServiceContract.ClientAccountID()
		}

		if err != nil {
			err = fmt.Errorf("failed to %s: %s", function, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": onChainResults})
	}
}
