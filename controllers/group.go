package controllers

import (
	"awesomeProject/database"
	"github.com/gofiber/fiber/v2"
)

type GroupJSON struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type GroupCountJSON struct {
	GroupJSON
	Count int64 `json:"count"`
}

func GroupBelongsToCurrentUser(userId int64, groupId int64) (bool, error) {
	var count int64

	rows, err := database.DB.Query("SELECT COUNT(*) FROM ActivityGroup WHERE Id = ? AND UserId = ?", groupId, userId)
	if err != nil {
		return false, err
	}

	defer rows.Close()
	println(userId, groupId)
	rows.Next()
	err = rows.Scan(&count)
	if err != nil {
		return false, err
	}
	return count != 0, nil
}

func CreateGroup(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)
	if err != nil {
		return c.Status(401).JSON("User is not logged in!")
	}

	var json GroupJSON

	if err := c.BodyParser(&json); err != nil {
		c.Status(400).JSON(err.Error())
	}

	stmt, err := database.DB.Prepare("INSERT INTO ActivityGroup(Name, UserId) VALUES (?, ?)")
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	result, err := stmt.Exec(json.Name, userId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	json.Id, err = result.LastInsertId()
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	return c.Status(200).JSON(json)
}

func UpdateGroup(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)
	if err != nil {
		return c.Status(401).JSON("No user logged in!")
	}

	var group GroupJSON
	if err := c.BodyParser(&group); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	belongs, err := GroupBelongsToCurrentUser(userId, group.Id)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	} else if !belongs {
		return c.Status(404).JSON("Group does not exist!")
	}

	stmt, err := database.DB.Prepare("UPDATE ActivityGroup SET Name = ? WHERE UserId = ?")
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	_, err = stmt.Exec(group.Name, userId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	return c.Status(200).JSON(group)
}

func DeleteGroup(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)
	if err != nil {
		return c.Status(401).JSON("No user logged in!")
	}

	groupId, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(err.Error())
	}

	stmt, err := database.DB.Prepare("DELETE FROM ActivityGroup WHERE Id = ? AND UserId = ?")
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	result, err := stmt.Exec(groupId, userId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.Status(404).JSON("Group does not exist")
	}

	return c.SendStatus(200)
}

func GetGroup(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)

	if err != nil {
		return c.Status(401).JSON("User not found!")
	}

	groupId, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(err.Error())
	}

	sql := `SELECT g.id, g.name, COUNT(a.Id)
		FROM ActivityGroup AS g
		LEFT JOIN Activity AS a
		ON a.GroupId = g.Id
		WHERE g.UserId = ? AND g.Id = ?
		GROUP BY g.id, g.name`

	println(groupId, userId)
	rows, err := database.DB.Query(sql, groupId, userId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	defer rows.Close()

	var group GroupCountJSON
	for rows.Next() {
		err := rows.Scan(&group.Id, &group.Name, &group.Count)
		if err != nil {
			err.Error()
			return c.SendStatus(500)
		}
	}

	if group.Id == 0 {
		return c.Status(404).JSON("Group not found!")
	}

	return c.Status(200).JSON(group)
}

func GetGroups(c *fiber.Ctx) error {

	userId, err := GetCurrentUserId(c)

	if err != nil {
		return c.Status(401).JSON("User is not logged in!")
	}

	sql := `SELECT g.id, g.name, COUNT(a.Id)
		FROM ActivityGroup AS g
		LEFT JOIN Activity AS a
		ON a.GroupId = g.Id
		WHERE g.UserId = ?
		GROUP BY g.id, g.name`

	rows, err := database.DB.Query(sql, userId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	defer rows.Close()

	var groups []GroupCountJSON

	for rows.Next() {
		var group GroupCountJSON
		err := rows.Scan(&group.Id, &group.Name, &group.Count)
		if err != nil {
			println(err.Error())
			return c.SendStatus(500)
		}
		groups = append(groups, group)
	}

	return c.Status(200).JSON(groups)
}
