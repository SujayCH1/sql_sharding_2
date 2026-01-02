package config

import "database/sql"

// store the connection credentials of the app db
type ApplicationDatabaseCreds struct {
	DB_HOST string
	DB_PORT string
	DB_USER string
	DB_NAME string
	DB_PASS string
}

var ApplicationDatabaseCredentials ApplicationDatabaseCreds

// instance of app db connection
type AppicationDatabaseConn struct {
	ConnInst *sql.DB
}

var ApplicationDatabaseConnection AppicationDatabaseConn
