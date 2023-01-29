package manager

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

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
	serviceName, ok := c.Params.Get("name")
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}

	err := s.Manager.Start(serviceName)
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
}

func (s *App) stop(c *gin.Context) {
	serviceName, ok := c.Params.Get("name")
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}

	err := s.Manager.Stop(serviceName)
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
}

func (s *App) delete(c *gin.Context) {
	serviceName, ok := c.Params.Get("name")
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}

	err := s.Manager.Delete(serviceName)
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
}

func (s *App) add(c *gin.Context) {
	serviceName, ok := c.Params.Get("name")
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}

	fileName := strings.Join([]string{serviceName, ".yaml"}, "")
	path := path.Join(ServiceDefinitionDir, fileName)
	file, err := os.Open(path)
	if err != nil {
		c.String(http.StatusInternalServerError, "could not open file at %s. %s", path, err.Error())
		return
	}

	opts, err := ServiceReader(file, serviceName)
	if err != nil {
		c.String(http.StatusInternalServerError, "could not load service %s. %s", serviceName, err.Error())
		return
	}

	err = s.Manager.Add(opts)
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
}

func (s *App) list(c *gin.Context) {
	services := s.Manager.List()
	c.JSON(http.StatusOK, services)
}
