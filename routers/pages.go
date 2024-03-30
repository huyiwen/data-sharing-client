package routers

import (
	"github.com/gin-gonic/gin"
)

func (r *Routers) ILogin() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.HTML(200, "login.html", nil)
	}
}

func (r *Routers) IIndex() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"MyIdentity": r.OrgSetup.Identity,
			"MyMSPID":    r.OrgSetup.MSPID,
		})
	}
}

func (r *Routers) IApplicationToMe() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.HTML(200, "applicationToMe.html", gin.H{
			"applications": r.ApplicationToMe,
			"MyIdentity":   r.OrgSetup.Identity,
			"MyMSPID":      r.OrgSetup.MSPID,
		})
	}
}

func (r *Routers) IMyApplication() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.HTML(200, "myapplication.html", gin.H{
			"applications": r.MyApplication,
			"MyIdentity":   r.OrgSetup.Identity,
			"MyMSPID":      r.OrgSetup.MSPID,
		})
	}
}
