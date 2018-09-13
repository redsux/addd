package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redsux/addd/core"
)

func registerRoutes(srv *gin.Engine) {
	records := srv.Group("/records")
	{
		forAll(records)
		record := records.Group("/:name")
		{
			record.Use(parseParams)
			forOne(record)
			wtype := record.Group("/:type")
			{
				forOne(wtype)
			}
		}
	}
	members := srv.Group("/members")
	{
		members.GET("", getMembers)
		members.GET("/", getMembers)
	}
}

func getMembers(c *gin.Context) {
	lst, err := addd.IPs()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		addd.Log.DebugF("[API] %v", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"members": lst,
	})
}

func forAll(router *gin.RouterGroup) {
	router.GET("", allRecords)
	router.GET("/", allRecords)

	router.POST("", newRecord)
	router.POST("/", newRecord)
}

func forOne(router *gin.RouterGroup) {
	router.GET("", getRecord)
	router.GET("/", getRecord)

	router.PUT("", updRecord)
	router.PUT("/", updRecord)

	router.DELETE("", delRecord)
	router.DELETE("/", delRecord)
}

func allRecords(c *gin.Context) {
	lst, err := addd.ListRecords()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		addd.Log.DebugF("[API] %v", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"records": lst,
	})
}

func newRecord(c *gin.Context) {
	var err error
	newRec := addd.DefaultRecord()

	// Bind body
	if err = c.BindJSON(&newRec); err != nil {
		return
	}

	// Not existing
	if _, err = addd.GetRecord(newRec.Name, newRec.Type); err != nil {
		if err = addd.StoreRecord(newRec); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"status": "created",
				"record": newRec,
			})
			return
		}
	} else {
		err = fmt.Errorf("Record already exist")
	}
	c.AbortWithError(http.StatusInternalServerError, err)
	addd.Log.DebugF("[API] %v", err.Error())
}

func getRecord(c *gin.Context) {
	rec := c.MustGet("record").(*addd.Record)
	c.JSON(http.StatusOK, rec)
}

func updRecord(c *gin.Context) {
	var err error
	rec := c.MustGet("record").(*addd.Record)
	newRec := &addd.Record{
		Type:  rec.Type,
		Class: rec.Class,
		TTL:   rec.TTL,
	}

	// Bind body
	if err = c.BindJSON(newRec); err != nil {
		return
	}

	if rec.Name == newRec.Name && rec.Type == newRec.Type {
		// Delete existing
		if err = addd.DeleteRecord(rec); err == nil {
			// Save new
			if err = addd.StoreRecord(newRec); err == nil {
				c.JSON(http.StatusOK, gin.H{
					"status":     "updated",
					"old-record": rec,
					"new-record": newRec,
				})
				return
			}
		}
	} else {
		err = fmt.Errorf("Body doesn't suit URI path")
	}
	c.AbortWithError(http.StatusInternalServerError, err)
	addd.Log.DebugF("[API] %v", err.Error())
}

func delRecord(c *gin.Context) {
	rec := c.MustGet("record").(*addd.Record)

	if err := addd.DeleteRecord(rec); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "deleted",
		"record": rec,
	})
}

func parseParams(c *gin.Context) {
	name := c.Param("name")
	rtype := c.Param("type")
	if rtype == "" {
		rtype = "A"
	}

	rec, err := addd.GetRecord(name, rtype)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		addd.Log.DebugF("[API] %v", err.Error())
		return
	}

	c.Set("record", rec)
	c.Next()
}
