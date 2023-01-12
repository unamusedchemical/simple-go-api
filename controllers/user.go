package controllers

import (
	"awesomeProject/database"
	"awesomeProject/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"time"
)

const SecretKey = "secret"

type UserJSON struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

func getUser(email string) (models.User, error) {
	// create a database query and get the result
	result, err := database.DB.Query("SELECT * FROM User WHERE Email = ?", email)

	if err != nil {
		return models.User{}, err
	}
	// set the result to be inaccessible once the function call ends
	defer result.Close()

	var user models.User

	// iterate through the result and get the data
	for result.Next() {
		err := result.Scan(&user.Id, &user.Username, &user.Email, &user.Password)
		if err != nil {
			return models.User{}, err
		}
	}

	return user, nil
}

// register new user
func Register(c *fiber.Ctx) error {
	var json UserJSON

	// parse input json data into a UserJSON object
	if err := c.BodyParser(&json); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	// check if the user already exists
	user, err := getUser(json.Email)
	if err != nil {
		println(err.Error())
		c.SendStatus(500)
	} else if user.Id != 0 {
		return c.Status(409).JSON(
			"User with such email already exists!",
		)
	}

	// generate a password hash
	password, err := bcrypt.GenerateFromPassword([]byte(json.Password), 14)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	// prepare an insert statement for the new user
	stmt, err := database.DB.Prepare("INSERT INTO User(Username, Email, Password) VALUES (?, ?, ?)")
	if err != nil {
		println(err.Error())
		c.SendStatus(500)
	}

	// execute the statement with the parsed input data
	temp, err := stmt.Exec(json.Username, json.Email, password)
	if err != nil {
		println(err.Error())
		c.SendStatus(500)
	}

	// get the id of the newly inserted data
	json.Id, err = temp.LastInsertId()
	if err != nil {
		println(err.Error())
		c.SendStatus(500)
	}

	// return the inserted user
	return c.Status(200).JSON(json)
}

func Login(c *fiber.Ctx) error {
	var json UserJSON
	// parse input data
	if err := c.BodyParser(&json); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	// check if user exists
	user, err := getUser(json.Email)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	} else if user.Id == 0 {
		return c.Status(404).JSON("User not found!")
	}

	// compare hashed passwords
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(json.Password)); err != nil {
		return c.Status(403).JSON("Incorrect password!")
	}

	// create a new jwt token, that expires in 24 hours and that contains the userId
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    strconv.Itoa(int(user.Id)),
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), //1 day
	})

	// sign the token with the secret key
	token, err := claims.SignedString([]byte(SecretKey))
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	// create an httponly cookie
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}
	// set created cookie
	c.Cookie(&cookie)

	return c.SendStatus(200)
}

func GetCurrentUserId(c *fiber.Ctx) (int64, error) {
	// get cookie holding jwt token
	cookie := c.Cookies("jwt")

	// parse token
	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err != nil {
		return 0, err
	}

	// get userId from token
	claims := token.Claims.(*jwt.StandardClaims)
	userId, err := strconv.Atoi(claims.Issuer)
	if err != nil {
		return 0, err
	}

	return int64(userId), nil
}

func GetUser(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)
	if err != nil {
		return c.Status(401).JSON("User is not logged in!")
	}

	// create a query to get the current user
	result, err := database.DB.Query("SELECT Id, Username, Email FROM User WHERE Id = ?", userId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	// set the result of the query to be inaccessible, once the function returns
	defer result.Close()

	var json UserJSON
	// iterate and get the result
	for result.Next() {
		err := result.Scan(&json.Id, &json.Username, &json.Email)
		if err != nil {
			println(err.Error())
			return c.SendStatus(500)
		}
	}

	// return the result
	return c.Status(200).JSON(json)
}

func UpdateUser(c *fiber.Ctx) error {
	// check if the user is logged in
	userId, err := GetCurrentUserId(c)
	if err != nil {
		return c.Status(401).JSON("User is not logged in!")
	}

	// parse json
	var json UserJSON
	if err := c.BodyParser(&json); err != nil {
		return err
	}

	// check if the user to update is the current user
	if userId != json.Id {
		return c.Status(403).JSON("Cannot update another user!")
	}

	// create the updated user object
	password, _ := bcrypt.GenerateFromPassword([]byte(json.Password), 14)
	user := models.User{
		Id:       json.Id,
		Username: json.Username,
		Email:    json.Email,
		Password: password,
	}

	// prepare statement to execute the query
	stmt, err := database.DB.Prepare("UPDATE User SET Username=?, Email=?, Password=? WHERE Id=?")
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	// execute the query
	stmt.Exec(user.Username, user.Email, user.Password, user.Id)

	return c.Status(200).JSON(user)
}

func DeleteUser(c *fiber.Ctx) error {
	// get the current user id
	userId, err := GetCurrentUserId(c)
	if err != nil {
		return c.Status(401).JSON("User is not logged in!")
	}

	// logout
	Logout(c)

	// prepare statement to delete user
	stmt, err := database.DB.Prepare("DELETE FROM User WHERE Id = ?")
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	// execute prepared statement
	_, err = stmt.Exec(userId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	return c.SendStatus(200)
}

func Logout(c *fiber.Ctx) error {
	// make cookie invalid
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	}
	c.Cookie(&cookie)

	return c.SendStatus(200)
}
