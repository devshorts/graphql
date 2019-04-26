package model

import "github.com/devshorts/incidentql/pkg/graph/db"

type UpsertIncidentRequest struct {
	Name     db.IncidentName `json:"name" binding:"required" example:"ir-foo"`
	IsActive bool            `json:"active" example:"true"`
}

type CreateIncidentResponse struct{}

type CreateInfraRequest struct {
	Name db.Infra `json:"name" binding:"required"`
}

type CreateInfraResponse struct{}

type CreateLinksRequest struct {
	Infra     InfraLink    `json:"infra"`
	Incidents IncidentLink `json:"incidents"`
}

type CreateLinksResponse struct{}

type InfraLink struct {
	Name      db.Infra   `json:"name" example:"term-health-srv"`
	DependsOn []db.Infra `json:"depends_on" swaggertype:"array,string" example:"envoy"`
}

type IncidentLink struct {
	Name     db.IncidentName `json:"name"  example:"ir-foo"`
	CausedBy []db.Infra      `json:"caused_by"  swaggertype:"array,string" example:"envoy"`
	ActedIn  []db.User       `json:"acted_in" swaggertype:"array,string" example:"suhas"`
}

type RelatedIncidentsResponse struct {
	Incidents []db.Incident `json:"incidents"`
}

type InfraHotSpotsResponse struct {
	Hotspots []db.Hotspot `json:"hotspots"`
}

type InfraCommunitiesResponse struct {
	Communities [][]db.Infra `json:"communities"`
}

type SharedInfraResponse struct {
	Paths [][]string `json:"paths"`
}
