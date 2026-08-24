package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/edwdch/web-tty/internal/session"
)

type SessionLister interface {
	List() []session.Info
}

type SessionCloser interface {
	Close(id string) error
}

func ListSessions(hub SessionLister) gin.HandlerFunc {
	return func(c *gin.Context) {
		list := hub.List()
		if list == nil {
			list = []session.Info{}
		}
		c.JSON(http.StatusOK, gin.H{"sessions": list})
	}
}

func DeleteSession(hub SessionCloser) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := hub.Close(c.Param("id")); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
