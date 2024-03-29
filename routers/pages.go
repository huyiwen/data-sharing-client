package routers

import (
	"github.com/gin-gonic/gin"
)

type viewPortData struct {
	Port string
}

func (r *Routers) ILogin() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.HTML(200, "login.html", nil)
	}
}

func (r *Routers) IIndex() func(c *gin.Context) {
	return func(c *gin.Context) {
		data := viewPortData{Port: r.Port}
		c.HTML(200, "index.html", data)
	}
}

func (r *Routers) IApplicationToMe() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.HTML(200, "applicationToMe.html", gin.H{
			"Port":         r.Port,
			"applications": r.ApplicationToMe,
		})
	}
}

func (r *Routers) IMyApplication() func(c *gin.Context) {
	return func(c *gin.Context) {
		// data := ViewData{Port: port}
		c.HTML(200, "myapplication.html", gin.H{
			"Port":         r.Port,
			"applications": r.MyApplication,
		})
	}
}
