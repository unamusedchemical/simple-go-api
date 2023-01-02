package controllers

import (
	"awesomeProject/database"
	"awesomeProject/models"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"math"
	"strconv"
	"time"
)

func activityBelongsToCurrentUser(userId uint, postId uint) (bool, error) {
	var count int64
	if err := database.DB.Raw("SELECT COUNT(*) FROM activities WHERE id = ? AND user_id = ?", postId, userId).Scan(&count).Error; err != nil {
		return false, nil
	}

	return count != 0, nil
}

func CreateActivity(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)
	if err != nil {
		return c.SendStatus(401)
	}

	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	activity := models.Activity{
		ActivityName:    data["name"],
		ActivityContent: data["content"],
		UserId:          userId,
	}

	if data["due"] != "" {
		*activity.Due, _ = time.Parse(models.DataTimeFormat, data["due"])
	} else {
		activity.Due = nil
	}

	if err := database.DB.Exec("INSERT INTO activities(activity_name, activity_content, opened_on, due, user_id) VALUES (?, ?, NOW(), ?, ?)",
		activity.ActivityName, activity.ActivityContent, activity.Due, userId).Error; err != nil {
		return c.Status(500).JSON(err.Error())
	}

	return c.Status(200).JSON("success")
}

func GetActivity(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)

	if err != nil {
		return c.Status(401).JSON(err.Error())
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(err.Error())
	}

	var activity models.Activity
	database.DB.Raw("SELECT * FROM activities WHERE id=? AND user_id=?", id, userId).Scan(&activity)

	if activity.Id == 0 {
		return c.SendStatus(404)
	}

	return c.SendStatus(200)
}

func DeleteActivity(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)

	if err != nil {
		c.SendStatus(401)
	}

	aId, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(400)
	}

	status := database.DB.Exec("DELETE FROM activities WHERE id = ? AND user_id = ?", aId, userId)

	if status.RowsAffected == 0 {
		return c.SendStatus(404)
	}

	return c.SendStatus(200)
}

func UpdateActivity(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)

	if err != nil {
		c.SendStatus(401)
	}

	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return c.SendStatus(400)
	}

	aId, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(400)
	}

	activityBelongs, err := activityBelongsToCurrentUser(userId, uint(aId))
	if err != nil {
		return c.SendStatus(500)
	} else if !activityBelongs {
		return c.SendStatus(404)
	}

	activity := models.Activity{
		ActivityName:    data["name"],
		ActivityContent: data["content"],
	}
	if data["label"] != "" {
		labelId, _ := strconv.Atoi(data["label"])
		labelBelongs, err := groupBelongsToCurrentUser(userId, uint(labelId))
		if err != nil {
			return c.Status(500).JSON(err.Error())
		} else if !labelBelongs {
			return c.SendStatus(404)
		}
		*activity.LabelId = uint(labelId)
	} else {
		activity.LabelId = nil
	}

	if err := database.DB.Exec("UPDATE activities SET activity_name=?, activity_content=?, label_id=? WHERE id = ? AND user_id = ?", activity.ActivityName, activity.ActivityContent, activity.LabelId, aId, userId).Error; err != nil {
		return c.SendStatus(500)
	}

	if err != nil {
		return c.SendStatus(500)
	}

	return c.SendStatus(200)
}

func Activities(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)

	if err != nil {
		c.SendStatus(401)
	}

	limit := 10
	var count int64
	if err := database.DB.Raw("SELECT COUNT(*) FROM activities WHERE user_id = ?", userId).Scan(&count).Error; err != nil {
		println(err.Error())
		return c.SendStatus(500)
	}
	pages := int(math.Ceil(float64(count) / float64(limit)))

	startQ := c.Query("start", "1")
	descQ := c.Query("desc", "true")
	search := c.Query("search", "")

	// validate query params
	start, err := strconv.Atoi(startQ)
	if err != nil {
		return c.SendStatus(400)
	}

	if start > pages {
		start = pages
	} else if start < 1 {
		start = 1
	}

	// validate query params
	desc, err := strconv.ParseBool(descQ)
	if err != nil {
		return c.SendStatus(400)
	}

	var order string
	if !desc {
		order = "ASC"
	} else {
		order = "DESC"
	}
	var sql string
	var activities []models.Activity
	if search != "" {
		sql = fmt.Sprintf("SELECT * FROM activities WHERE user_id = ? AND MATCH(activity_name) AGAINST(? IN NATURAL LANGUAGE MODE) ORDER BY closed_on %s LIMIT ? OFFSET ?", order)
		if err := database.DB.Raw(sql, userId, search, limit, (start-1)*limit).Scan(&activities).Error; err != nil {
			println(err.Error())
			return c.SendStatus(500)
		}
	} else {
		sql = fmt.Sprintf("SELECT * FROM activities WHERE user_id = ? ORDER BY closed_on %s LIMIT ? OFFSET ?", order)
		if err := database.DB.Raw(sql, userId, limit, (start-1)*limit).Scan(&activities).Error; err != nil {
			println(err.Error())
			return c.SendStatus(500)
		}
	}

	return c.Status(200).JSON(fiber.Map{
		"limit":            limit,
		"curr_page":        start,
		"last_page":        pages,
		"total_activities": count,
		"data":             activities,
	})
}

func CloseActivity(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)

	if err != nil {
		c.SendStatus(401)
	}

	aId, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(400)
	}

	belongs, err := activityBelongsToCurrentUser(userId, uint(aId))

	if err != nil {
		return c.SendStatus(500)
	}

	if !belongs {
		return c.SendStatus(403)
	}

	database.DB.Exec("UPDATE activities SET closed_on=NOW() WHERE id=? AND user_id=?", aId, userId)

	return c.SendStatus(200)
}
