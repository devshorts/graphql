package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSnapshotDB(t *testing.T) {
	/*
		Between two incidents what is the shared path
		using only the relationship types defined
		with a path limit from 1 to 10:

		MATCH
		(incident1:Incident { id: 'ir-root' }),
		(incident2:Incident { id: 'ir-silver-neon' }),
		p = allShortestPaths((incident1)-[:CAUSED_BY | :DEPENDS_ON *1..10]-(incident2))
		RETURN p

		Incidents that are failing and connected to each other somehow:

		match (root:Incident)-[*]-(failing:Incident)
		where root.id="ir-root"
		return failing, root
	*/

	db, err := NewDefaultDB("bolt://localhost:7687/", "", "")
	assert.NoError(t, err)
	assert.NoError(t, db.Drop())

	graphConfig, err := LoadConfigFromFile("./graph.json")
	assert.NoError(t, err)

	createIncidents(t, graphConfig, db)
	createInfraDependencies(t, graphConfig, db)
	createUserToTeam(t, graphConfig, db)
	createServiceToTeam(t, graphConfig, db)
}

func TestDB_RelatedIncidents(t *testing.T) {
	_, err := NewDefaultDB("bolt://localhost:7687/", "", "")
	assert.NoError(t, err)

	graphConfig, err := LoadConfigFromFile("./graph.json")
	assert.NoError(t, err)
	fmt.Println(graphConfig)

	//res, err := db.RelatedIncidents(graphConfig.Incidents[0].Name)
	//assert.NoError(t, err)
	//assert.NotEmpty(t, res)
}

func TestDB_Hotspots(t *testing.T) {
	db, err := NewDefaultDB("bolt://localhost:7687/", "", "")
	assert.NoError(t, err)

	res, err := db.Hotspots()
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestDB_Communities(t *testing.T) {
	db, err := NewDefaultDB("bolt://localhost:7687/", "", "")
	assert.NoError(t, err)

	res, err := db.Communities()
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func createInfraDependencies(t *testing.T, config *GraphConfig, db *DB) {
	t.Helper()

	for _, deps := range config.Dependencies {
		assert.NoError(t, db.Link([]Relationship{
			NewInfraDependsOn(deps.Node, deps.DependsOn...),
		}))
	}
}

func createUserToTeam(t *testing.T, config *GraphConfig, db *DB) {
	t.Helper()

	for _, userToTeam := range config.UserTeams {
		assert.NoError(t, db.Link([]Relationship{
			NewUserToTeam(userToTeam.Node, userToTeam.Team),
		}))
	}
}

func createServiceToTeam(t *testing.T, config *GraphConfig, db *DB) {
	t.Helper()

	for _, service := range config.ServiceTeams {
		assert.NoError(t, db.Link([]Relationship{
			NewServiceToTeam(service.Team, service.Service),
		}))
	}
}

func createIncidents(t *testing.T, config *GraphConfig, db *DB) {
	t.Helper()

	for _, incident := range config.Incidents {

		assert.NoError(t, db.InsertIncident(Incident{
			Name:       incident.Name,
			IsActive:   incident.IsActive,
			Resolution: incident.Resolution,
		}))

		assert.NoError(t, db.Link([]Relationship{
			NewSeverity(incident.Name, incident.Severity),
			NewIncidentCausedBy(incident.Name, incident.CausedBy...),
			NewIncidentHasUsersBy(incident.Name, incident.Users...),
		}))

	}
}

func LoadConfigFromFile(configFilePath string) (*GraphConfig, error) {
	absPath, err := filepath.Abs(configFilePath)
	if err != nil {
		return nil, err
	}

	jsonConfig, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	var cfg GraphConfig
	err = json.Unmarshal(jsonConfig, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
