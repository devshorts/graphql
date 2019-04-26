package db

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/sirupsen/logrus"
)

type NodeType string

const (
	IncidentType NodeType = "Incident"
	UserType              = "User"
	InfraType             = "Infra"
	SeverityType          = "Severity"
	TeamType              = "Team"
)

type DB struct {
	driver  neo4j.Driver
	session neo4j.Session
}

func NewDB(driver neo4j.Driver, session neo4j.Session) *DB {
	return &DB{driver: driver, session: session}
}

func NewDefaultDB(uri, username, password string) (*DB, error) {
	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, err
	}

	session, err := driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return nil, err
	}

	return NewDB(driver, session), nil
}

func (x DB) Drop() error {
	return x.transact(func(transaction neo4j.Transaction) (result neo4j.Result, e error) {
		return transaction.Run("match (n) detach delete n", nil)
	})
}

func (x DB) Link(relationships []Relationship) error {
	transaction, err := x.session.BeginTransaction()
	defer func() {
		if err := transaction.Close(); err != nil {
			logrus.Error(err)
		}
	}()
	if err != nil {
		return err
	}

	for _, r := range relationships {
		for _, link := range r.links() {
			if err := x.createLink(transaction, link); err != nil {
				return err
			}
		}
	}

	return transaction.Commit()
}

func (x DB) createLink(transaction neo4j.Transaction, l link) error {

	var baseCommand = `
MERGE (root:%s { id: $rootIDName })
MERGE (target:%s { id: $targetIDName })
CREATE UNIQUE (root)-[:%s]->(target)`

	if l.relationships.backwards != "" {
		baseCommand += fmt.Sprintf(`CREATE (root)<-[:%s]-(target)`, l.relationships.backwards)
	}

	result, err := transaction.Run(fmt.Sprintf(baseCommand,
		l.rootType,
		l.targetType,
		l.relationships.forward,
	), map[string]interface{}{
		"rootIDName":   l.rootIDValue,
		"targetIDName": l.targetIDValue,
	})
	if err != nil {
		return err
	}

	return x.consume(result)
}

func (x DB) InsertIncident(incident Incident) error {
	return x.insert(IncidentType, string(incident.Name), map[string]interface{}{
		"active":     incident.IsActive,
		"resolution": incident.Resolution,
	})
}

func (x DB) consume(result neo4j.Result) error {
	for result.Next() {
	}
	if err := result.Err(); err != nil {
		return err // handle error
	}

	return nil
}

func (x DB) transact(block func(transaction neo4j.Transaction) (neo4j.Result, error)) error {
	transaction, err := x.session.BeginTransaction()
	defer func() {
		if err := transaction.Close(); err != nil {
			logrus.Error(err)
		}
	}()
	if err != nil {
		return err
	}

	result, err := block(transaction)
	if err != nil {
		return err // handle error
	}

	if err := x.consume(result); err != nil {
		return err
	}

	return transaction.Commit()
}

// insert generically inserts a node type with the ID and sets the properties
// from the map on the object when set.  Requires an ID field
func (x DB) insert(nodeType NodeType, id string, properties map[string]interface{}) error {
	return x.transact(func(transaction neo4j.Transaction) (neo4j.Result, error) {
		const matchRef = "i"

		var props []string
		for propName := range properties {
			props = append(props, fmt.Sprintf("%s.%s = $%s", matchRef, propName, propName))
		}

		properties["id"] = id

		return transaction.Run(fmt.Sprintf(`
MERGE (%s:%v { id: $id })
ON CREATE SET %s
ON MATCH SET %s
`, matchRef, nodeType,
			strings.Join(props, " , "),
			strings.Join(props, " , ")), properties)
	})
}

func (x DB) retrieve(block func(transaction neo4j.Transaction) (neo4j.Result, error)) (neo4j.Result, error) {
	transaction, err := x.session.BeginTransaction()
	defer func() {
		if err := transaction.Close(); err != nil {
			logrus.Error(err)
		}
	}()
	if err != nil {
		return nil, err
	}

	result, err := block(transaction)
	if err != nil {
		return nil, err // handle error
	}

	return result, nil
}

func (x DB) Communities() ([][]Infra, error) {
	cql := `CALL algo.louvain.stream('Infra', 'DEPENDS_ON', {})
YIELD nodeId, community
MATCH (n:Infra) WHERE id(n)=nodeId
RETURN community,
       collect(n.id) as members 
 limit 100`
	result, err := x.retrieve(func(transaction neo4j.Transaction) (neo4j.Result, error) {

		return transaction.Run(cql, nil)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve communities err: %v", err)
	}

	var communities [][]Infra

	for result.Next() {
		r := result.Record().GetByIndex(1).([]interface{})
		communities = append(communities, toInfra(r))
	}

	return communities, nil

}

func toInfra(communities []interface{}) []Infra {
	var result []Infra
	for _, community := range communities {
		result = append(result, Infra(fmt.Sprintf("%v", community)))
	}
	return result
}

func (x DB) Hotspots() ([]Hotspot, error) {
	cql := `CALL algo.betweenness.stream('Infra',null, {direction:'both'})
YIELD nodeId, centrality

MATCH (i:Infra) WHERE id(i) = nodeId

RETURN i.id AS i,centrality
ORDER BY centrality DESC limit 5`

	result, err := x.retrieve(func(transaction neo4j.Transaction) (neo4j.Result, error) {

		return transaction.Run(cql, nil)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve hotspots err: %v", err)
	}

	var hotspots []Hotspot
	for result.Next() {
		spot := Hotspot{
			Infra:      Infra(result.Record().GetByIndex(0).(string)),
			Centrality: result.Record().GetByIndex(1).(float64),
		}

		hotspots = append(hotspots, spot)

	}
	return hotspots, nil

}

func (x DB) RelatedIncidents(name IncidentName) ([]Incident, error) {
	incidentCypher := `
match (root:Incident {id: $id})-[:CAUSED_BY]->()<-[:DEPENDS_ON *1..4]-()<-[:CAUSED_BY]-(failing:Incident {active: true})
where root.id <> failing.id 
return distinct failing`
	result, err := x.retrieve(func(transaction neo4j.Transaction) (neo4j.Result, error) {

		return transaction.Run(incidentCypher, map[string]interface{}{
			"id": name,
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve related incidents for incident: %s, err: %v", name, err)
	}

	var relatedIncidents []Incident
	for result.Next() {
		r := result.Record().GetByIndex(0).(neo4j.Node)
		var incident Incident
		err = mapstructure.Decode(r.Props(), &incident)
		if err != nil {
			return nil, fmt.Errorf("failed to decode the Incident response from DB, err: %v", err)
		}
		relatedIncidents = append(relatedIncidents, incident)

	}
	return relatedIncidents, nil
}

func (x DB) SharedInfra(source IncidentName, target IncidentName) ([][]string, error) {
	incidentCypher := `
match path=(root:Incident {id: $source})-[c:CAUSED_BY]->(m)<-[:DEPENDS_ON *1..4]-(k)<-[c2:CAUSED_BY]-(failing:Incident {id: $target})
return extract(n in nodes(path) | n.id)`

	result, err := x.retrieve(func(transaction neo4j.Transaction) (neo4j.Result, error) {

		return transaction.Run(incidentCypher, map[string]interface{}{
			"source": source,
			"target": target,
		})
	})

	if err != nil {
		return nil, err
	}

	var paths [][]string
	for result.Next() {
		var curr []string
		for _, arr := range result.Record().Values() {
			for _, value := range arr.([]interface{}) {
				curr = append(curr, value.(string))
			}
		}

		paths = append(paths, curr)

	}
	return paths, nil
}
