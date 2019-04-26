package server

import (
	"net/http"

	"github.com/devshorts/incidentql/pkg/graph/db"
	_ "github.com/devshorts/incidentql/pkg/server/docs"
	"github.com/devshorts/incidentql/pkg/server/model"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

type Server struct {
	engine *gin.Engine
	db     *db.DB
}

func NewServer(db *db.DB) *Server {
	engine := gin.Default()

	server := &Server{engine, db}

	server.configureRoutes()

	return server
}

func (x Server) configureRoutes() {
	x.engine.Use(func(c *gin.Context) {
		c.Next()

		if c.Errors != nil {
			c.JSON(http.StatusBadRequest, c.Errors.JSON())
		}
	})

	x.engine.Use(gin.Logger())

	x.engine.Use(gin.Recovery())

	x.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := x.engine.Group("/api/v1")
	{
		api.POST("/incidents", x.createIncident)
		api.GET("/incidents/:id/shared/:shared_id", x.sharedInfra)
		api.GET("/incidents/:id/related", x.relatedIncidents)
		api.GET("/infra/hotspots", x.infraHotSpots)
		api.GET("/infra/communities", x.infraCommunities)
		api.GET("/ping", x.ping)
		api.POST("/links", x.createLinks)
	}
}

// @Summary Infra that has a high level of betweenness. I.e. Infra that is used everywhere
// @Tags infra
// @Accept  json
// @Produce  json
// @Success 200 {object} model.InfraHotSpotsResponse
// @Router /infra/hotspots [get]
func (x Server) infraHotSpots(c *gin.Context) {

	hotspots, err := x.db.Hotspots()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, model.InfraHotSpotsResponse{
			Hotspots: hotspots,
		})
	}
}

// @Summary Infra that is likley to be grouped together. These are related communities
// @Tags infra
// @Accept  json
// @Produce  json
// @Success 200 {object} model.InfraCommunitiesResponse
// @Router /infra/communities [get]
func (x Server) infraCommunities(c *gin.Context) {

	communities, err := x.db.Communities()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, model.InfraCommunitiesResponse{
			Communities: communities,
		})
	}

}

// @Summary Shared Infra Between Incidents
// @Tags incidents
// @Accept  json
// @Produce  json
// @Param id path string true "Incident Name"
// @Param shared_id path string true "Related Incident Name"
// @Success 200 {object} model.SharedInfraResponse
// @Router /incidents/{id}/shared/{shared_id} [get]
func (x Server) sharedInfra(c *gin.Context) {
	source := db.IncidentName(c.Param("id"))
	target := db.IncidentName(c.Param("shared_id"))

	paths, err := x.db.SharedInfra(source, target)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, model.SharedInfraResponse{Paths: paths})
}

// @Summary Related Incidents
// @Tags incidents
// @Accept  json
// @Produce  json
// @Param id path string true "Incident Name"
// @Success 200 {object} model.RelatedIncidentsResponse
// @Router /incidents/{id}/related [get]
func (x Server) relatedIncidents(c *gin.Context) {
	incidentName := db.IncidentName(c.Param("id"))

	relatedIncidents, err := x.db.RelatedIncidents(incidentName)

	if err != nil {
		c.JSON(http.StatusInternalServerError, model.RelatedIncidentsResponse{})
	} else {
		c.JSON(http.StatusOK, relatedIncidents)
	}
}

// @Summary Creates links
// @Tags links
// @Accept  json
// @Produce  json
// @Param request body model.CreateLinksRequest true "Create Links"
// @Success 200 {object} model.CreateLinksResponse
// @Router /links [post]
func (x Server) createLinks(c *gin.Context) {
	var req model.CreateLinksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	var r []db.Relationship
	if req.Infra.DependsOn != nil {
		r = append(r, db.NewInfraDependsOn(req.Infra.Name, req.Infra.DependsOn...))
	}

	if req.Incidents.ActedIn != nil {
		r = append(r, db.NewIncidentHasUsersBy(req.Incidents.Name, req.Incidents.ActedIn...))
	}

	if req.Incidents.CausedBy != nil {
		r = append(r, db.NewIncidentCausedBy(req.Incidents.Name, req.Incidents.CausedBy...))
	}

	if err := x.db.Link(r); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, model.CreateLinksResponse{})
}

// @Summary Creates an incident
// @Tags incidents
// @Accept  json
// @Produce  json
// @Param incident body model.UpsertIncidentRequest true "Create Incident"
// @Success 200 {object} model.CreateIncidentResponse
// @Router /incidents [post]
func (x Server) createIncident(c *gin.Context) {
	var req model.UpsertIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	if err := x.db.InsertIncident(db.Incident{
		Name:     req.Name,
		IsActive: req.IsActive,
	}); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, model.CreateIncidentResponse{})
}

// @Summary Ping
// @Tags diagnostics
// @Accept  json
// @Produce  json
// @Success 200
// @Router /ping [get]
func (x Server) ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func (x Server) Start(address string) error {
	return x.engine.Run(address)
}
