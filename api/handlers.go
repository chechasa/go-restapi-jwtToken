package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var timesEnteredInvalidCredentials int
var timeOut bool = false
var isTenSecondsLoop bool = false
var jwtkey = []byte("secret_key")
var I int
var users = make(map[string]string)

type Result struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Family   string `json:"family"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
	Person_id  int    `json:"person_id"`
	PersonRole string `json:"personRole"`
}

// The function responsable for the user's credentials verification
func Login(w http.ResponseWriter, r *http.Request) {

	// timeOut is going to be true whenever the user enters the wrong credentions 3 times in a row in less then 10 seconds
	if timeOut {

		// If you try to login during timeOut it will return a message and how many second left to end it
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(fmt.Sprintln("You entered invalid credentials 3 times, please wait 30 seconds and try again")))
		w.Write([]byte(fmt.Sprintln(I)))

		// If the user is not in a timeOut it will just continue normally
	} else if !timeOut {
		var results Result

		// Getting the info to connect to the db and make a query
		var dns = "root:1234@tcp(127.0.0.1:3306)/exercise_db?charset=utf8mb4"
		var db, _ = sql.Open("mysql", dns)
		rows, er := db.Query("select * from persons")
		if er != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for rows.Next() {

			// Using the query to create the users
			rows.Scan(&results.Id, &results.Name, &results.Age, &results.Family, &results.Username, &results.Password, &results.Role)
			users[results.Username] = results.Password
		}

		// Decoding the json file with the credentials and making the verification
		var credentials Credentials
		err := json.NewDecoder(r.Body).Decode(&credentials)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		expectedPassword, ok := users[credentials.Username]
		if !ok || expectedPassword != credentials.Password {
			w.WriteHeader(http.StatusForbidden)

			// if verification went wrong it enters the TimeOut function, the details of the function are in the function in the bottom of the script
			go TimeOut(w)

			// the TenSecondsLoop function is responsible for verifying if the credentials were entered wrong 3 times in row in less then 10 seconds, more details in the bottom
			// the function will only run if there's not a timeout or a tensecondloop already running
			if !timeOut || !isTenSecondsLoop {
				go TenSecondsLoop()
			}

			return
		}

		// Setting the 30 minutes duration of the jwt token
		expirationTime := time.Now().Add(time.Minute * 30)

		var result Result

		// Querying the row of the user that is logged in
		resu := db.QueryRow("select * from persons where username = " + "'" + credentials.Username + "'")
		resu.Scan(&result.Id, &result.Name, &result.Age, &result.Family, &result.Username, &result.Password, &result.Role)

		// Setting the claims of the token
		claims := &Claims{
			Username: credentials.Username,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},

			// the person id and role so that it makes it easier to find the user in the db later
			Person_id:  result.Id,
			PersonRole: result.Role,
		}

		// setting the token with its signing method and claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtkey)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// setting the cookie with the token value so that we can query the token from the request in any subsequent requests
		http.SetCookie(w,
			&http.Cookie{
				Name:    "token",
				Value:   tokenString,
				Expires: expirationTime,
			})
	}
}

// function responsible for the creating of new persons
func CreatePerson(w http.ResponseWriter, r *http.Request) {
	var dns = "root:1234@tcp(127.0.0.1:3306)/exercise_db?charset=utf8mb4"
	var db, _ = sql.Open("mysql", dns)

	// querying the token from the request with the name that was set for it
	cookie, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenStr := cookie.Value

	claims := &Claims{}

	// Decoding the token to use data the that was encrypted in it
	tkn, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (interface{}, error) {
			return jwtkey, nil
		})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
	}

	// verifying if the user that is logged in has the role admin
	if claims.PersonRole == "admin" {
		var result Result

		// decoding the json entered with the post request
		err := json.NewDecoder(r.Body).Decode(&result)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// verifying if there's is another user with the same username
		for k, _ := range users {
			if k == result.Username {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("Sorry, there's already an user with this username")))
				return
			}
		}

		// verifying if the username or password field were empty
		if result.Username == "" || result.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Sorry, Username and password field are required")))
			return
		}

		// if everything goes right, then it create a new person
		w.Write([]byte(fmt.Sprintf("gtz you created a person with username: '%s'", result.Username)))

		id_string := strconv.Itoa(result.Id)
		age_string := strconv.Itoa(result.Age)
		if result.Role == "" {
			result.Role = "normal"
		}

		// and insert it into the table
		db.Exec("INSERT INTO persons VALUES(" + id_string + ", '" + result.Name + "'" + ", " + age_string + "" + ", '" + result.Family + "'" + ", '" + result.Username + "'" + ", '" + result.Password + "'" + ", '" + result.Role + "')")
		users[result.Username] = result.Password

		// or if the user is a not an admin, it return this msg and doesn't allow the creating to happen
	} else {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(fmt.Sprintf("Sorry, only admins can create new persons, and your role is , %s\n", claims.PersonRole)))
	}

}

// Everytime the TimeOut is used it will add 1 to the timesEnteredInvalidCredentials and in a case of three time it will give a timeout of 30 seconds
func TimeOut(w http.ResponseWriter) {

	timesEnteredInvalidCredentials++
	if timesEnteredInvalidCredentials == 3 {
		timeOut = true
		for I = 30; I > 0; I-- {
			iString := strconv.Itoa(I)
			time.Sleep(time.Second)
			w.Write([]byte(fmt.Sprintln(iString)))
		}
		timeOut = false
		timesEnteredInvalidCredentials = 0

	}
}

// and the TenSecondsLoop will reset the variable timesEnteredInvalidCredentials every 10 seconds
// only allowing a timeout if the credentials were entered 3 times wrong in a time space of 10 seconds
func TenSecondsLoop() {
	isTenSecondsLoop = true
	time.Sleep(time.Second * 10)
	timesEnteredInvalidCredentials = 0
	isTenSecondsLoop = false
}
