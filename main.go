package main

import (
	handler "goland_exercise/api"
	"goland_exercise/config"
	mysql_pack "goland_exercise/db_mysql"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	// This project is set to work with a database named "exercise_db"
	// the person created in the first migration has the
	// username: first
	// password: 1234
	// http://localhost:9090

	//Loading the configs to start working with the database
	config.LoadConfig()
	//Starts the database and make the first migration
	mysql_pack.GetMysqlConn()

	//HandleFunc where the url path and the function that's going to give the request a functionality is set
	http.HandleFunc("/login", handler.Login)
	http.HandleFunc("/createperson", handler.CreatePerson)
	//The server creation
	log.Fatal(http.ListenAndServe(":9090", nil))
}
