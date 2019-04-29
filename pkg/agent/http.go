package agent

import (
	"github.com/farmer-hutao/k6s/pkg/agent/exec"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"net/http"
)

func Install(c *gin.Context) {
	fileAddress := c.PostForm("install")
	if len(fileAddress) == 0 {
		glog.Errorf("%s field of the post form is null", "install")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "install field of the post form is null",
		})
		return
	}

	user := c.DefaultPostForm("user", "root")
	glog.V(3).Infoln("User is" + user)

	err := exec.Install(user, fileAddress)
	if err != nil {
		glog.Errorf("%s: Install failed", fileAddress)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "install failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func Start(c *gin.Context) {
	fileAddress := c.PostForm("start")
	if len(fileAddress) == 0 {
		glog.Errorf("%s field of the post form is null", "start")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "start field of the post form is null",
		})
		return
	}

	user := c.DefaultPostForm("user", "root")
	glog.V(3).Infoln("User is" + user)

	err := exec.Start(user, fileAddress)
	if err != nil {
		glog.Errorf("%s: start failed", fileAddress)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "start failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func Restart(c *gin.Context) {
	fileAddress := c.PostForm("restart")
	if len(fileAddress) == 0 {
		glog.Errorf("%s field of the post form is null", "restart")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "restart field of the post form is null",
		})
		return
	}

	user := c.DefaultPostForm("user", "root")
	glog.V(3).Infoln("User is" + user)

	err := exec.Restart(user, fileAddress)
	if err != nil {
		glog.Errorf("%s: restart failed", fileAddress)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "restart failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func Stop(c *gin.Context) {
	fileAddress := c.PostForm("stop")
	if len(fileAddress) == 0 {
		glog.Errorf("%s field of the post form is null", "stop")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "stop field of the post form is null",
		})
		return
	}

	user := c.DefaultPostForm("user", "root")
	glog.V(3).Infoln("User is" + user)

	err := exec.Stop(user, fileAddress)
	if err != nil {
		glog.Errorf("%s: stop failed", fileAddress)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "stop failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func Delete(c *gin.Context) {
	fileAddress := c.PostForm("delete")
	if len(fileAddress) == 0 {
		glog.Errorf("%s field of the post form is null", "delete")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "delete field of the post form is null",
		})
		return
	}

	user := c.DefaultPostForm("user", "root")
	glog.V(3).Infoln("User is" + user)

	err := exec.Delete(user, fileAddress)
	if err != nil {
		glog.Errorf("%s: delete failed", fileAddress)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "delete failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func NewGinEngine() *gin.Engine {
	router := gin.Default()
	router.POST("/install", Install)
	router.POST("/start", Start)
	router.POST("/restart", Restart)
	router.POST("/stop", Stop)
	router.POST("/delete", Delete)

	return router
}
