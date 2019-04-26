package db

type IncidentName string
type User string
type Team string
type RelationshipName string
type Infra string
type IncidentLevel string

type link struct {
	rootIDValue   string
	rootType      NodeType
	relationships Relationships
	targetIDValue string
	targetType    NodeType
}

type Hotspot struct {
	Infra      Infra   `json:"infra"`
	Centrality float64 `json:"centrality"`
}

type GraphConfig struct {
	Incidents    []BatchIncident  `json:"incidents"`
	UserTeams    []UserToTeam     `json:"user_teams"`
	Dependencies []InfraDependsOn `json:"infrastructure"`
	ServiceTeams []ServiceToTeam  `json:"service_teams"`
}

type ServiceToTeam struct {
	Team    string `json:"owns"`
	Service string `json:"service"`
}

func NewServiceToTeam(Team string, Service string) *ServiceToTeam {
	return &ServiceToTeam{
		Team:    Team,
		Service: Service,
	}
}

func (ServiceToTeam) Relationship() Relationships {
	return Relationships{
		forward: "OWNS",
	}
}

func (s ServiceToTeam) links() []link {
	var l []link

	l = append(l,
		link{
			rootIDValue:   string(s.Team),
			rootType:      TeamType,
			relationships: s.Relationship(),
			targetIDValue: string(s.Service),
			targetType:    InfraType,
		})

	return l
}

type Incident struct {
	Name       IncidentName `json:"name" mapstructure:"id"`
	IsActive   bool         `json:"is_active" mapstructure:"active"`
	Resolution string       `json:"resolution"`
}

type BatchIncident struct {
	Incident
	Severity IncidentLevel `json:"severity"`
	CausedBy []Infra       `json:"caused_by"`
	Users    []User        `json:"users"`
}

type Relationships struct {
	forward   RelationshipName
	backwards RelationshipName
}

type Relationship interface {
	Relationship() Relationships
	links() []link
}

type UserToTeam struct {
	Node User `json:"user"`
	Team Team `json:"team"`
}

type InfraDependsOn struct {
	Node      Infra   `json:"name"`
	DependsOn []Infra `json:"depends_on"`
}

func NewUserToTeam(Node User, Team Team) *UserToTeam {
	return &UserToTeam{
		Node: Node,
		Team: Team,
	}
}

func (UserToTeam) Relationship() Relationships {
	return Relationships{
		forward: "BELONG_TO",
	}
}

func (u UserToTeam) links() []link {
	var l []link

	l = append(l,
		link{
			rootIDValue:   string(u.Node),
			rootType:      UserType,
			relationships: u.Relationship(),
			targetIDValue: string(u.Team),
			targetType:    TeamType,
		})

	return l
}

func NewInfraDependsOn(node Infra, dependsOn ...Infra) *InfraDependsOn {
	return &InfraDependsOn{Node: node, DependsOn: dependsOn}
}

func (v InfraDependsOn) links() []link {
	var l []link

	for _, depends := range v.DependsOn {
		l = append(l,
			link{
				rootIDValue:   string(v.Node),
				rootType:      InfraType,
				relationships: v.Relationship(),
				targetIDValue: string(depends),
				targetType:    InfraType,
			})
	}

	return l
}

func (InfraDependsOn) Relationship() Relationships {
	return Relationships{
		forward: "DEPENDS_ON",
	}
}

var _ Relationship = InfraDependsOn{}

type Severity struct {
	Node     IncidentName
	Severity IncidentLevel
}

func NewSeverity(node IncidentName, severity IncidentLevel) *Severity {
	return &Severity{
		Node:     node,
		Severity: severity,
	}
}

func (Severity) Relationship() Relationships {
	return Relationships{
		forward: "SEVERITY",
	}
}

func (s Severity) links() []link {
	var l []link

	l = append(l,
		link{
			rootIDValue:   string(s.Node),
			rootType:      IncidentType,
			relationships: s.Relationship(),
			targetIDValue: string(s.Severity),
			targetType:    SeverityType,
		})

	return l
}

var _ Relationship = &Severity{}

type IncidentCausedBy struct {
	Node     IncidentName
	CausedBy []Infra
}

func NewIncidentCausedBy(node IncidentName, causedBy ...Infra) *IncidentCausedBy {
	return &IncidentCausedBy{Node: node, CausedBy: causedBy}
}

func (v IncidentCausedBy) links() []link {
	var l []link

	for _, causedBy := range v.CausedBy {
		l = append(l,
			link{
				rootIDValue:   string(v.Node),
				rootType:      IncidentType,
				relationships: v.Relationship(),
				targetIDValue: string(causedBy),
				targetType:    InfraType,
			})
	}

	return l
}

func (IncidentCausedBy) Relationship() Relationships {
	return Relationships{
		forward: "CAUSED_BY",
	}
}

var _ Relationship = IncidentCausedBy{}

type IncidentHasUsersBy struct {
	Node  IncidentName
	Users []User
}

func NewIncidentHasUsersBy(node IncidentName, user ...User) *IncidentHasUsersBy {
	return &IncidentHasUsersBy{Node: node, Users: user}
}

func (v IncidentHasUsersBy) links() []link {
	var l []link

	for _, user := range v.Users {
		l = append(l,
			link{
				rootIDValue:   string(v.Node),
				rootType:      IncidentType,
				relationships: v.Relationship(),
				targetIDValue: string(user),
				targetType:    UserType,
			})
	}

	return l
}

func (IncidentHasUsersBy) Relationship() Relationships {
	return Relationships{
		forward: "ACTED_IN",
	}
}

var _ Relationship = IncidentHasUsersBy{}
