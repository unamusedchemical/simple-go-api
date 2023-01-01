package controllers

import (
	"awesomeProject/database"
	"awesomeProject/models"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"time"
)

const SecretKey = "secret"

func userExists(email string) (models.User, error) {
	var user models.User
	database.DB.Raw("SELECT * FROM users WHERE email = ?", email).Scan(&user)
	if user.Id == 0 {
		return models.User{}, errors.New("user not found")
	}

	return user, nil
}

func Register(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return c.SendStatus(500)
	}

	_, err := userExists(data["email"])
	if err == nil {
		return c.SendStatus(409)
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)

	user := models.User{
		Username: data["username"],
		Email:    data["email"],
		Password: password,
	}
	database.DB.Exec("INSERT INTO users(username, email, password) VALUES (?, ?, ?)", user.Username, user.Email, user.Password)

	return c.SendStatus(200)
}

func Login(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	user, err := userExists(data["email"])
	if err != nil {
		println(err.Error())
		return c.SendStatus(404)
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(data["password"])); err != nil {
		return c.SendStatus(403)
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

func GetCurrentUserId(c *fiber.Ctx) (uint, error) {
	cookie := c.Cookies("jwt")

	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err != nil {
		println("hello")
		return 0, err
	}

	claims := token.Claims.(*jwt.StandardClaims)
	userId, err := strconv.Atoi(claims.Issuer)
	if err != nil {
		return 0, err
	}

	return uint(userId), nil
}

func User(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)

	if err != nil {
		return c.Status(401).JSON("unauthorised")
	}

	var user models.User
	database.DB.Raw("SELECT * FROM users WHERE id = ?", userId).Scan(&user)

	return c.Status(200).JSON(user)
}

func UpdateUser(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)

	if err != nil {
		return c.SendStatus(401)
	}

	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		return err
	}

	id, _ := strconv.Atoi(data["id"])
	password, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)
	user := models.User{
		Id:       uint(id),
		Username: data["username"],
		Email:    data["email"],
		Password: password,
	}

	if data["pfp"] != "" {
		*user.ProfilePicture = data["pfp"]
	} else {
		user.ProfilePicture = nil
	}

	if userId != user.Id {
		return c.Status(403).JSON("cannot update another user")
	}

	database.DB.Exec("UPDATE users SET username=?, email=?, password=?, profile_picture=? WHERE id=?", user.Username, user.Email, user.Password, user.ProfilePicture, userId)
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
