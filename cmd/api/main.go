package main

import (
	"github.com/devshorts/incidentql/pkg/graph/db"
	"github.com/devshorts/incidentql/pkg/server"
	"github.com/sirupsen/logrus"
)

func main() {
	defaultDB, err := db.NewDefaultDB("bolt://localhost:7687/", "", "")
	if err != nil {
		logrus.Fatal(err)
	}

	if err := server.NewServer(defaultDB).Start("localhost:8080"); err != nil {
		logrus.Fatal(err)
	}
}
