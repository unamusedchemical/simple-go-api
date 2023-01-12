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
	result, err := database.DB.Query("SELECT * FROM User WHERE Email = ?", email)

	if err != nil {
		return models.User{}, err
	}

	defer result.Close()

	var user models.User

	for result.Next() {
		err := result.Scan(&user.Id, &user.Username, &user.Email, &user.Password)
		if err != nil {
			return models.User{}, err
		}
	}

	return user, nil
}

func Register(c *fiber.Ctx) error {
	var json UserJSON

	if err := c.BodyParser(&json); err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	user, err := getUser(json.Email)
	if err != nil {
		println(err.Error())
		c.SendStatus(500)
	} else if user.Id != 0 {
		return c.Status(409).JSON(fiber.Map{
			"message": "user with such email already exists",
		})
	}

	password, err := bcrypt.GenerateFromPassword([]byte(json.Password), 14)

	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	stmt, err := database.DB.Prepare("INSERT INTO User(Username, Email, Password) VALUES (?, ?, ?)")
	if err != nil {
		println(err.Error())
		c.SendStatus(500)
	}

	temp, err := stmt.Exec(json.Username, json.Email, password)
	if err != nil {
		println(err.Error())
		c.SendStatus(500)
	}

	json.Id, err = temp.LastInsertId()
	if err != nil {
		println(err.Error())
		c.SendStatus(500)
	}

	return c.Status(200).JSON(json)
}

func Login(c *fiber.Ctx) error {
	var json UserJSON

	if err := c.BodyParser(&json); err != nil {
		println(err.Error())
		return c.SendStatus(400)
	}

	user, err := getUser(json.Email)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	} else if user.Id == 0 {
		return c.Status(404).JSON("User not found!")
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(json.Password)); err != nil {
		return c.Status(403).JSON("Incorrect password!")
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    strconv.Itoa(int(user.Id)),
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), //1 day
	})

	token, err := claims.SignedString([]byte(SecretKey))

	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.SendStatus(200)
}

func GetCurrentUserId(c *fiber.Ctx) (int64, error) {
	cookie := c.Cookies("jwt")

	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err != nil {
		return 0, err
	}

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
		return c.Status(401).JSON(fiber.Map{"message": "unauthorised"})
	}

	result, err := database.DB.Query("SELECT Id, Username, Email FROM User WHERE Id = ?", userId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	defer result.Close()

	var json UserJSON
	for result.Next() {
		err := result.Scan(&json.Id, &json.Username, &json.Email)
		if err != nil {
			println(err.Error())
			return c.SendStatus(500)
		}
	}

	return c.Status(200).JSON(json)
}

func UpdateUser(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)

	if err != nil {
		return c.SendStatus(401)
	}

	var json UserJSON
	if err := c.BodyParser(&json); err != nil {
		return err
	}

	if userId != json.Id {
		return c.Status(403).JSON("cannot update another user")
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(json.Password), 14)
	user := models.User{
		Id:       json.Id,
		Username: json.Username,
		Email:    json.Email,
		Password: password,
	}

	stmt, err := database.DB.Prepare("UPDATE User SET Username=?, Email=?, Password=? WHERE Id=?")
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	stmt.Exec(user.Username, user.Email, user.Password, user.Id)

	return c.Status(200).JSON(user)
}

func DeleteUser(c *fiber.Ctx) error {
	Logout(c)

	userId, err := GetCurrentUserId(c)

	if err != nil {
		return c.SendStatus(401)
	}

	stmt, err := database.DB.Prepare("DELETE FROM User WHERE Id = ?")
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	_, err = stmt.Exec(userId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	return c.SendStatus(200)
}

func Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.SendStatus(200)
}
