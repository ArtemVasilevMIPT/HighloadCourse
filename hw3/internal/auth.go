package internal

import (
	"database/sql"
	"fmt"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"net/http"
	"time"
)

var TokenAuth *jwtauth.JWTAuth
var ExpirationDuration = 2 * time.Hour

func InitAuth() {
	TokenAuth = jwtauth.New("HS256", []byte("secret"), nil)
	fmt.Println("Initialized auth")
}

func Register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Printf("Failed to parse form data: %v\n", err)
	}
	email := r.PostFormValue("email")
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	if email == "" || username == "" || password == "" {
		http.Error(w, "All form fields should be filled", http.StatusBadRequest)
		return
	}
	//fmt.Printf("New User:\nEmail: %s\nUsername: %s\nPassword: %s\n", email, username, password)
	// check if email/username exists
	err = Db.QueryRow("SELECT Username FROM Users WHERE Username = ? OR Email = ?", username, email).Scan()
	if err == nil || err != sql.ErrNoRows {
		http.Error(w, "Username of Email are already in use", http.StatusBadRequest)
		return
	}
	err = Db.QueryRow("SELECT Username FROM Verification WHERE Username = ? OR Email = ?", username, email).Scan()
	if err == nil || err != sql.ErrNoRows {
		http.Error(w, "Username of Email are already in use", http.StatusBadRequest)
		return
	}
	//
	_, tokenString, _ := TokenAuth.Encode(map[string]interface{}{"username": username, jwt.ExpirationKey: time.Now().Add(ExpirationDuration)})
	// write data to db
	_, err = Db.Exec("INSERT OR REPLACE INTO Verification (Email, Username, Password, Token) VALUES (?, ?, ?, ?)", email, username, password, tokenString)
	if err != nil {
		fmt.Printf("Couldn't add user %s to verification database\n", username)
		http.Error(w, "Couldn't add user to database", http.StatusInternalServerError)
		return
	}
	//
	SendRegistrationEmail(email, tokenString)
}

func ConfirmRegistration(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("jwt")
	token, err := TokenAuth.Decode(tokenString)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	userInter, ok := token.Get("username")
	if !ok {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	var (
		email    string
		username string
		password string
		tokenDb  string
	)
	err = Db.QueryRow("SELECT * FROM Verification WHERE Username = ?", userInter.(string)).Scan(&email, &username, &password, &tokenDb)
	if err != nil {
		fmt.Printf("Couldn't get user %s from verification database\n", email)
		http.Error(w, "Couldn't get data from database", http.StatusInternalServerError)
		return
	}
	/*
		if tokenDb != tokenString {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
	*/
	_, err = Db.Exec("INSERT OR FAIL INTO Users (Username, Email, Password) VALUES (?, ?, ?)", username, email, password)
	if err != nil {
		fmt.Printf("Couldn't add user (%s, %s) to users database\n", username, email)
		http.Error(w, "Couldn't add user to database", http.StatusInternalServerError)
		return
	}
	_, err = Db.Exec("DELETE FROM Verification WHERE Email = ?", email)
	if err != nil {
		fmt.Printf("Couldn't remove user (%s, %s) from verification database", username, email)
		http.Error(w, "Couldn't update database", http.StatusInternalServerError)
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Printf("Failed to parse form data: %v\n", err)
	}
	usernameForm := r.PostFormValue("username")
	passwordForm := r.PostFormValue("password")

	var password string

	err = Db.QueryRow("SELECT Password FROM Users WHERE Username = ?", usernameForm).Scan(&password)
	if err != nil {
		fmt.Printf("Couldn't get user %s from verification database: %v\n", usernameForm, err)
		if err == sql.ErrNoRows {
			http.Error(w, "User"+usernameForm+"doesn't exist", http.StatusBadRequest)
		} else {
			http.Error(w, "Couldn't get data from database", http.StatusInternalServerError)
		}
		return
	}
	if password != passwordForm {
		http.Error(w, "Invalid password", http.StatusBadRequest)
		return
	}

	token, tokenString, _ := TokenAuth.Encode(map[string]interface{}{"username": usernameForm, jwt.ExpirationKey: time.Now().Add(ExpirationDuration)})
	fmt.Printf("Produced token:\nValue: %+v\nString: %+v\n", token, tokenString)
	_, err = Db.Exec("UPDATE Users SET Token = ? WHERE Username = ?", tokenString, usernameForm)
	if err != nil {
		fmt.Printf("Couldn't update token for user %s\n", usernameForm)
		http.Error(w, "Couldn't update token", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Authorization", "Bearer "+tokenString)
}

func ResetPassword(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Printf("Failed to parse form data: %v\n", err)
	}
	usernameForm := r.PostFormValue("username")

	var (
		email string
	)

	err = Db.QueryRow("SELECT Email FROM Users WHERE Username = ?", usernameForm).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User"+usernameForm+"doesn't exist", http.StatusBadRequest)
		} else {
			fmt.Printf("Couldn't get user %s from verification database: %v\n", usernameForm, err)
			http.Error(w, "Couldn't get data from database", http.StatusInternalServerError)
		}
		return
	}

	_, tokenString, _ := TokenAuth.Encode(map[string]interface{}{"username": usernameForm, jwt.ExpirationKey: time.Now().Add(ExpirationDuration)})
	_, err = Db.Exec("INSERT OR REPLACE INTO Verification (Email, Username,  Token) VALUES (?, ?,  ?)", email, usernameForm, tokenString)
	if err != nil {
		fmt.Printf("Couldn't add user %s to verification database\n", usernameForm)
		http.Error(w, "Couldn't add user to database", http.StatusInternalServerError)
		return
	}

	SendResetEmail(email, tokenString)
}

func ConfirmResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		fmt.Println("Failed to get token from context")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	username := claims["username"].(string)
	if username == "" {
		fmt.Println("Failed to get username from token")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = r.ParseForm()
	if err != nil {
		fmt.Printf("Failed to parse form data: %v\n", err)
	}
	passwordForm := r.PostFormValue("password")
	if passwordForm == "" {
		http.Error(w, "Invalid password", http.StatusBadRequest)
		return
	}
	_, err = Db.Exec("DELETE FROM Verification WHERE Username = ?", username)
	if err != nil {
		fmt.Printf("Couldn't remove user %s from verification database\n", username)
		http.Error(w, "Couldn't update database", http.StatusInternalServerError)
		return
	}
	_, err = Db.Exec("UPDATE Users SET Password = ? WHERE Username = ?", passwordForm, username)
	if err != nil {
		fmt.Printf("Couldn't update password for user %s\n", username)
		http.Error(w, "Couldn't update password", http.StatusInternalServerError)
		return
	}
}

func Verifier(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return jwtauth.Verify(ja, jwtauth.TokenFromQuery, jwtauth.TokenFromHeader, jwtauth.TokenFromCookie)(next)
	}
}

func UserAuthenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())

		if err != nil {
			fmt.Printf("Failed to get token from context\n")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if token == nil || jwt.Validate(token) != nil {
			fmt.Printf("Failed to validate token\n")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		username, ok := token.Get("username")
		if !ok {
			fmt.Printf("Invalid token format\n")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		var tokenDbString string
		err = Db.QueryRow("SELECT Token FROM Users WHERE Username = ?", username.(string)).Scan(&tokenDbString)
		if err != nil {
			fmt.Printf("Failed to retrieve user %s\n", username.(string))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		/*
			tokenDb, err := TokenAuth.Decode(tokenDbString)
			if err != nil {
				fmt.Printf("Failed to decode token\n")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		*/
		//fmt.Printf("Comparing tokens:\nCookie: %+v\nDb string: %+v\nDb: %+v\n", token, tokenDbString, tokenDb)
		/*
			if token != tokenDb {
				fmt.Printf("Tokens don't match\n")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		*/
		next.ServeHTTP(w, r)
	})
}

func VerificationAuthenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())

		if err != nil {
			fmt.Printf("Couldn't get token from context\n")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if token == nil || jwt.Validate(token) != nil {
			fmt.Printf("Couldn't validate token\n")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		username, ok := token.Get("username")
		if !ok {
			fmt.Printf("Couldn't get username from token\n")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		var tokenDbString string
		err = Db.QueryRow("SELECT Token FROM Verification WHERE Username = ?", username.(string)).Scan(&tokenDbString)
		if err != nil {
			fmt.Printf("Failed to retrieve user %s\n", username.(string))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		//fmt.Println(tokenDbString)
		tokenString := r.URL.Query().Get("jwt")
		if tokenDbString != tokenString {
			fmt.Printf("Tokens don't match\n")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
