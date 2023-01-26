package manager

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	Address = "127.0.0.1"
	Port    = 8080
)

func (s *App) startHTTPServer() error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(
		gin.Recovery(),
	)

	router.POST("/services/:name", s.add)
	router.DELETE("/services/:name", s.delete)
	router.PUT("/services/:name/start", s.start)
	router.PUT("/services/:name/stop", s.stop)
	router.GET("/services", s.list)

	err := router.Run(fmt.Sprintf("%s:%d", Address, Port))
	return err
}

func (s *App) start(c *gin.Context) {
	wrapper(c, s.Manager.Start)
}

func (s *App) stop(c *gin.Context) {
	wrapper(c, s.Manager.Stop)
}

func (s *App) delete(c *gin.Context) {
	wrapper(c, s.Manager.Delete)
}

func (s *App) add(c *gin.Context) {
	wrapper(c, s.Manager.Add)
}

func (s *App) list(c *gin.Context) {
	services := s.Manager.List()
	c.JSON(http.StatusOK, services)
}

func wrapper(c *gin.Context, action func(serviceName string) error) {
	serviceName, ok := c.Params.Get("name")
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}

	err := action(serviceName)
	if err != nil {
		switch {
		case errors.Is(err, ErrBadRequest):
			c.String(http.StatusBadRequest, err.Error())

		case errors.Is(err, ErrSminitInternalError):
			c.String(http.StatusInternalServerError, err.Error())
		default:
			c.String(http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.Status(http.StatusOK)
	return

}
