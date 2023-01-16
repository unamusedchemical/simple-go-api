package controllers

import (
	"awesomeProject/database"
	"awesomeProject/models"
	sql2 "database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"math"
	"strconv"
	"time"
)

type ActivityJSON struct {
	Id       int64      `json:"id"`
	Title    string     `json:"title"`
	Body     string     `json:"body"`
	ClosedOn *time.Time `json:"closed_on"`
	OpenedOn time.Time  `json:"opened_on"`
	Due      *time.Time `json:"due"`
	GroupId  *int64     `json:"group_id"`
}

type ActivityJoinJSON struct {
	ActivityJSON
	GroupName *string `json:"group_name"`
}

// turn data from the database into json
func (a *ActivityJoinJSON) initJSON(activity models.Activity, groupName sql2.NullString) {
	a.Id = activity.Id
	a.Body = activity.Body
	a.Title = activity.Title
	a.OpenedOn = activity.OpenedOn

	if activity.ClosedOn.Valid {
		a.ClosedOn = new(time.Time)
		println(activity.ClosedOn.Time.String())
		*a.ClosedOn = activity.ClosedOn.Time
	}

	if activity.Due.Valid {
		a.Due = new(time.Time)
		*a.Due = activity.Due.Time
	}

	if activity.GroupId.Valid {
		a.GroupId = new(int64)
		*a.GroupId = activity.GroupId.Int64
	}

	if groupName.Valid {
		a.GroupName = new(string)
		*a.GroupName = groupName.String
	}
}

// check if activity belongs to current user
func ActivityBelongsToCurrentUser(userId int64, activityId int64) (bool, error) {
	var count int64

	println(activityId)
	// we check if it does by seeing the number of elements where their id = activityId and userid = userId
	rows, err := database.DB.Query("SELECT COUNT(*) FROM Activity WHERE Id = ? AND UserId = ?", activityId, userId)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	rows.Next()
	err = rows.Scan(&count)
	if err != nil {
		return false, err
	}

	return count != 0, nil
}

// create a new activity
func CreateActivity(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)
	if err != nil {
		return c.Status(401).JSON("User is not logged in!")
	}
	// parse input json
	var json ActivityJSON
	if err := c.BodyParser(&json); err != nil {
		return c.Status(400).JSON(err.Error())
	}
	// prepare statement for insertion into the database
	stmt, err := database.DB.Prepare("INSERT INTO Activity (Title, Body, OpenedOn, Due, UserId, GroupId) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	// check if the group belongs to the user
	if json.GroupId != nil && *json.GroupId != 0 {
		belongs, err := GroupBelongsToCurrentUser(userId, int64(*json.GroupId))
		if err != nil {
			println(err.Error())
			return c.SendStatus(500)
		} else if !belongs {
			return c.Status(404).JSON("Cannot add an activity to a group that does not exits!")
		}
	}

	json.OpenedOn = time.Now()
	// ececute prepared statement with parameters, by using the `?` we make sure that no sql injection can happen
	result, err := stmt.Exec(json.Title, json.Body, json.OpenedOn, json.Due, userId, json.GroupId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	// get the id of the newly inserted object
	json.Id, err = result.LastInsertId()
	if err != nil {
		println(err.Error())
		c.SendStatus(500)
	}
	json.ClosedOn = nil

	// return the newly inserted object
	return c.Status(200).JSON(json)
}

func DeleteActivity(c *fiber.Ctx) error {
	// check if user is logged in
	userId, err := GetCurrentUserId(c)
	if err != nil {
		c.Status(401).JSON("User is not logged in!")
	}

	// parse input data
	aId, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(err.Error())
	}
	// prepare statement for deletion
	stmt, err := database.DB.Prepare("DELETE FROM Activity WHERE Id = ? AND UserId = ?")
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	// execute statement
	result, err := stmt.Exec(aId, userId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	// if no rows are affected then there is no activity with userId of the current user and the provided activity id
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return c.Status(404).JSON("Cannot delete an activity that does not exist!")
	}

	// if everything is fine, return status 200
	return c.SendStatus(200)
}

func UpdateActivity(c *fiber.Ctx) error {
	// check if user is logged in
	userId, err := GetCurrentUserId(c)
	if err != nil {
		c.Status(401).JSON("User not logged in!")
	}

	// parse request data
	var json ActivityJSON
	if err := c.BodyParser(&json); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	// check if activity belongs to user
	belongs, err := ActivityBelongsToCurrentUser(userId, json.Id)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	} else if !belongs {
		return c.Status(404).JSON("Activity does not exist!")
	}

	// prepare query to execute
	stmt, err := database.DB.Prepare("UPDATE Activity SET Title = ?, Body = ?, Due = ? WHERE Id = ?")
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	// execute query with parameters
	_, err = stmt.Exec(json.Title, json.Body, json.Due, json.Id)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	json.ClosedOn = nil

	// return 200 at success
	return c.Status(200).JSON(json)
}

// add or change the group of an activity
func EditGroupActivity(c *fiber.Ctx) error {
	// check if user is logged in and get their id
	userId, err := GetCurrentUserId(c)

	if err != nil {
		c.SendStatus(401)
	}

	// get the input data - the id of the post and the id of the group
	var data map[string]int64
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON("An error occurred while parsing JSON")
	}

	// validate the data
	activityId, ok := data["activity_id"]
	if !ok {
		return c.Status(400).JSON("`activity_id` needs to be provided")
	}
	groupId, ok := data["group_id"]
	if !ok {
		return c.Status(400).JSON("`group_id` needs to be provided")
	}

	// check if they belong to the current user
	belongs, err := ActivityBelongsToCurrentUser(userId, activityId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	} else if !belongs {
		return c.Status(404).JSON("Activity does not exist!")
	}

	belongs, err = GroupBelongsToCurrentUser(userId, groupId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	} else if !belongs {
		return c.Status(404).JSON("Group does not exist!")
	}

	// prepare a query for execution
	stmt, err := database.DB.Prepare("UPDATE Activity SET GroupId = ? WHERE Id = ?")
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	// execute query
	_, err = stmt.Exec(groupId, activityId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	return c.SendStatus(200)
}

func Activities(c *fiber.Ctx) error {
	// get current user, if no one is logged in an error will be returned
	userId, err := GetCurrentUserId(c)
	if err != nil {
		return c.Status(401).JSON("User is not logged in!")
	}

	// get total number of activities, this is needed to paginate the data
	limit := 10
	var count int64
	rows, err := database.DB.Query("SELECT COUNT(*) FROM Activity WHERE UserId = ?", userId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	// get the number
	rows.Next()
	rows.Scan(&count)
	rows.Close()

	// calculate the number of pages
	pages := int(math.Ceil(float64(count) / float64(limit)))
	if pages < 1 {
		pages = 1
	}

	// get the url parameters
	startQ := c.Query("start", "1")
	descQ := c.Query("desc", "true")
	all := c.Query("all", "true")
	search := c.Query("search", "")

	//validate url params
	start, err := strconv.Atoi(startQ)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	if start < 1 {
		start = 1
	} else if start > pages {
		start = pages
	}

	//validate url params
	desc, err := strconv.ParseBool(descQ)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	// create a separate variable to add to the sql query, based on the user input
	// it is not a good practice to pass user input directly
	var order string
	if !desc {
		order = "ASC"
	} else {
		order = "DESC"
	}

	// get initial sql string
	sql := `SELECT a.Id, a.Title, a.Body, a.ClosedOn, a.OpenedOn, a.Due, g.Id, g.Name
		FROM Activity AS a
		LEFT JOIN ActivityGroup AS g ON a.GroupId = g.Id
		WHERE a.UserId = ? `

	if all == "false" {
		// if the data is going to be filtered; apply the order in which it is going to be sorted
		sql += fmt.Sprintf("AND MATCH(a.Title, a.Body) AGAINST(? IN NATURAL LANGUAGE MODE) ORDER BY a.OpenedOn %s LIMIT ? OFFSET ?", order)
		rows, err = database.DB.Query(sql, userId, search, limit, (start-1)*limit)
	} else {
		// if the data is not going to be filtered; apply the order in which it is going to be sorted
		sql += fmt.Sprintf("ORDER BY a.OpenedOn %s LIMIT ? OFFSET ?", order)
		rows, err = database.DB.Query(sql, userId, limit, (start-1)*limit)
	}

	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	// set the query to be inaccessible, once the method finishes
	defer rows.Close()

	var activities []ActivityJoinJSON
	for rows.Next() {
		var activity models.Activity
		var groupName sql2.NullString
		// get the data from the database
		err := rows.Scan(&activity.Id, &activity.Title, &activity.Body, &activity.ClosedOn, &activity.OpenedOn, &activity.Due, &activity.GroupId, &groupName)
		if err != nil {
			println(err.Error())
			return c.SendStatus(500)
		}
		// turn the data into json
		var activityJoinJSON ActivityJoinJSON
		(&activityJoinJSON).initJSON(activity, groupName)

		activities = append(activities, activityJoinJSON)
	}

	// return json
	return c.Status(200).JSON(fiber.Map{
		"limit":            limit,
		"curr_page":        start,
		"last_page":        pages,
		"total_activities": count,
		"data":             activities,
	})
}

func CloseActivity(c *fiber.Ctx) error {
	// get current user id
	userId, err := GetCurrentUserId(c)
	if err != nil {
		c.Status(401).JSON(err.Error())
	}

	// get id
	aId, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON("Error while parsing params!")
	}

	// check if activity belongs to user
	belongs, err := ActivityBelongsToCurrentUser(userId, int64(aId))

	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	} else if !belongs {
		return c.Status(403).JSON("Activity does not exist!")
	}

	// prepare statement to execute
	stmt, err := database.DB.Prepare("UPDATE Activity SET ClosedOn=NOW() WHERE Id = ? AND UserId = ?")
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	// execute statement
	_, err = stmt.Exec(aId, userId)
	if err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}

	return c.SendStatus(200)
}
