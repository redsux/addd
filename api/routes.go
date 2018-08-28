package api

import (
	"fmt"
	"net/http"
	"github.com/redsux/addd/core"
	"github.com/gin-gonic/gin"
)

func registerRoutes(srv *gin.Engine) {
	records := srv.Group("/records")
	{
		records.GET("/", allRecords )
		records.POST("/", newRecord )

		record := records.Group("/:name/:type")
		{
			record.Use( parseParams )
			record.GET("/", getRecord )
			record.PUT("/", updRecord )
			record.DELETE("/", delRecord )
		}
	}
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
		err = fmt.Errorf("Record already exist.")
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
		Type: rec.Type,
		Class: rec.Class,
		Ttl: rec.Ttl,
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
					"status": "updated",
					"old-record": rec,
					"new-record": newRec,
				})
				return
			}
		}
	} else {
		err = fmt.Errorf("Body doesn't suit URI path.")
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

	rec, err := addd.GetRecord(name, rtype)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		addd.Log.DebugF("[API] %v", err.Error())
		return
	}
	
	c.Set("record", rec)
	c.Next()
}