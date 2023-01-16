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

// check if a group belongs to the current user
func GroupBelongsToCurrentUser(userId int64, groupId int64) (bool, error) {
	var count int64

	// execute query to count the number of groups that are associated with groupId and userId
	rows, err := database.DB.Query("SELECT COUNT(*) FROM ActivityGroup WHERE Id = ? AND UserId = ?", groupId, userId)
	if err != nil {
		return false, err
	}

	// set the data returned by the query to be inaccessible once the function ends
	defer rows.Close()
	// get the data
	rows.Next()
	err = rows.Scan(&count)
	// check for error with the database
	if err != nil {
		return false, err
	}
	// return whether there is an element that is associated with groupId and userId
	return count != 0, nil
}

// create a group
func CreateGroup(c *fiber.Ctx) error {
	// get logged in user id
	userId, err := GetCurrentUserId(c)
	// check for errors
	if err != nil {
		return c.Status(401).JSON("User is not logged in!")
	}

	// parse input data
	var json GroupJSON
	// check for errors
	if err := c.BodyParser(&json); err != nil {
		c.Status(400).JSON(err.Error())
	}

	// prepare statement to insert
	stmt, err := database.DB.Prepare("INSERT INTO ActivityGroup(Name, UserId) VALUES (?, ?)")
	// check for errors
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	// execute statement
	result, err := stmt.Exec(json.Name, userId)
	// check for errors
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	// get last inserted id
	json.Id, err = result.LastInsertId()
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	// return the inserted element
	return c.Status(200).JSON(json)
}

// update group
func UpdateGroup(c *fiber.Ctx) error {
	// get logged in user id
	userId, err := GetCurrentUserId(c)
	if err != nil {
		return c.Status(401).JSON("No user logged in!")
	}

	// parse user input
	var group GroupJSON
	if err := c.BodyParser(&group); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	// check if it belongs to the current user
	belongs, err := GroupBelongsToCurrentUser(userId, group.Id)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	} else if !belongs {
		return c.Status(404).JSON("Group does not exist!")
	}

	// prepare update statement
	stmt, err := database.DB.Prepare("UPDATE ActivityGroup SET Name = ? WHERE UserId = ? AND Id = ?")
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	// execute statement
	_, err = stmt.Exec(group.Name, userId, group.Id)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	// return the modified group
	return c.Status(200).JSON(group)
}

func DeleteGroup(c *fiber.Ctx) error {
	// get logged in user id
	userId, err := GetCurrentUserId(c)
	if err != nil {
		return c.Status(401).JSON("No user logged in!")
	}

	// get group id from url
	groupId, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(err.Error())
	}

	// prepare delete statement
	stmt, err := database.DB.Prepare("DELETE FROM ActivityGroup WHERE Id = ? AND UserId = ?")
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	// execute delete statement - delete row where userId and groupId
	result, err := stmt.Exec(groupId, userId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	// if no rows were deleted then the group does not exist
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.Status(404).JSON("Group does not exist")
	}

	return c.SendStatus(200)
}

func GetGroup(c *fiber.Ctx) error {
	// get logged in user id
	userId, err := GetCurrentUserId(c)
	if err != nil {
		return c.Status(401).JSON("User not found!")
	}

	// get id from url
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

	// execute query
	rows, err := database.DB.Query(sql, groupId, userId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	// set the data to be inaccessible once the function terminates
	defer rows.Close()

	// get the query data
	var group GroupCountJSON
	for rows.Next() {
		err := rows.Scan(&group.Id, &group.Name, &group.Count)
		if err != nil {
			err.Error()
			return c.SendStatus(500)
		}
	}

	// if the id is 0 (it starts from 1 and goes upwards), then no such group exists
	if group.Id == 0 {
		return c.Status(404).JSON("Group not found!")
	}

	// return query data
	return c.Status(200).JSON(group)
}

func GetGroups(c *fiber.Ctx) error {
	// get current user id
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

	// execute query
	rows, err := database.DB.Query(sql, userId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	// set the data to be inaccessible once the function terminates
	defer rows.Close()

	// get query data
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

	// return query data
	return c.Status(200).JSON(groups)
}
