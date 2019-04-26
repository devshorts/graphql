## Setup

Build seabolt C bindings for neo4j go driver

```
ROOT=$HOME/src/lib
mkdir -p $ROOT
cd $ROOT
brew install pkg-config
brew install cmake
git clone https://github.com/neo4j-drivers/seabolt.git
export OPENSSL_ROOT_DIR=/usr/local/opt/openssl
cd seabolt
./make_release.sh
```

In your shell dotconfig now add

```
ROOT=$HOME/src/lib
export PKG_CONFIG_PATH=$ROOT/seabolt/build/dist/share/pkgconfig
export DYLD_LIBRARY_PATH=$ROOT/seabolt/build/dist/lib
export LD_LIBRARY_PATH=$ROOT/seabolt/build/dist/lib64
```

## 

Start the DB

```
docker-compose up

```
Run the service

```
go generate ./...
go run cmd/api/main.go
```

Hit swagger

```
http://localhost:8080/swagger/index.html
```

## Get jupyter notebook working

Once docker-compose up is running, go to the jupyter page (localhost:8888...) and open a new terminal
Run `cd work && pip install -r requirements.txt`

Now you can access neo4j and the vis.js drawing API

## Notes

```
// incidents and keywords from the root incident to incidents related to bapi
MATCH (incident1:Incident)-[:CAUSED_BY]->(keyword1)<-[:CAUSED_BY]-(coIncident:Incident),
   (coIncident)-[:CAUSED_BY]->(keyword2)<-[:CAUSED_BY]-(incident2:Incident)   
   where incident1.id = "ir-green-question" and keyword2.id = "bapi"
RETURN coIncident, keyword1, incident1, incident2, keyword2
```

```
Degrees of separation between incidents

MATCH (user1:User)-[:INVOLVED_IN]->(incident1)<-[:INVOLVED_IN]-(coUser:User),
         (coUser)-[:INVOLVED_IN]->(incident2)<-[:INVOLVED_IN]-(user2:User)
         where user1.id = "akropp" and user2.id = "derekl"
AND   NOT    (user1)-[:INVOLVED_IN]->(incident2)
RETURN coUser, user1, incident1, incident2, user2
```

```
failing infrastructures:

match (root:Incident)-[*]-(failing:Incident)
where root.id="ir-root"
return failing, root
```

```
dependent infra between two incidents:
MATCH
(incident1:Incident { id: 'ir-root' }),
(incident2:Incident { id: 'ir-orange-tool' }), 
p = allShortestPaths((incident1)-[*]-(incident2))
RETURN p
```

```
most dependent infra (name - # dependencies)
MATCH (b:Infra)
 WITH b, SIZE(()-[:DEPENDS_ON]->(b)) as dep
 ORDER BY dep DESC LIMIT 10
 RETURN b.id, dep
```

```
hot spots (infra between the most amount of stuff)
CALL algo.betweenness.stream('Infra',null, {direction:'both'})
YIELD nodeId, centrality

MATCH (i:Infra) WHERE id(i) = nodeId

RETURN i.id AS i,centrality
ORDER BY centrality DESC limit 5
```

```
communities:

CALL algo.louvain.stream('Infra', 'DEPENDS_ON', {})
YIELD nodeId, community
MATCH (n:Infra) WHERE id(n)=nodeId
RETURN community,
       count(*) as communitySize, 
       collect(n.id) as members 
order by communitySize desc limit 100
```
