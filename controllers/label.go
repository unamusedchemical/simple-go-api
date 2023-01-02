package controllers

import (
	"awesomeProject/database"
	"awesomeProject/models"
	"github.com/gofiber/fiber/v2"
)

func groupBelongsToCurrentUser(userId uint, groupId uint) (bool, error) {
	var count int64
	if err := database.DB.Raw("SELECT COUNT(*) FROM labels WHERE id = ? AND user_id=?", groupId, userId).Scan(&count).Error; err != nil {
		return false, err
	}

	return count != 0, nil
}

func CreateGroup(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)
	if err != nil {
		return c.Status(401).JSON(err.Error())
	}

	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		c.Status(400).JSON(err.Error())
	}

	group := models.Label{
		Name:   data["name"],
		UserId: userId,
	}

	if err := database.DB.Exec("INSERT INTO labels (name, user_id) VALUES (?, ?) ", group.Name, group.UserId).Error; err != nil {
		return c.Status(500).JSON(err.Error())
	}

	return c.Status(200).JSON("success")
}

func UpdateGroup(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)
	if err != nil {
		return c.Status(401).JSON(err.Error())
	}

	groupId, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(err.Error())
	}
	var group models.Label

	if err := database.DB.Raw("SELECT * FROM labels WHERE id = ? AND user_id = ?", groupId, userId).Scan(&group).Error; err != nil {
		return c.Status(500).JSON(err.Error())
	}

	if group.UserId != userId {
		return c.Status(404).JSON("group not found")
	}

	var data map[string]string
	if err := c.BodyParser(data); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	group.Name = data["name"]

	if err := database.DB.Exec("UPDATE labels SET  name = ? WHERE id = ?", group.Name, group.Id).Error; err != nil {
		return c.Status(500).JSON(err.Error())
	}

	return c.Status(200).JSON("success")
}

func DeleteGroup(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)
	if err != nil {
		return c.Status(401).JSON("unauthorised")
	}

	groupId, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(err.Error())
	}

	belongs, err := groupBelongsToCurrentUser(userId, uint(groupId))
	if err != nil {
		return c.Status(500).JSON(err.Error())
	} else if !belongs {
		return c.Status(404).JSON("group does not belong to user")
	}

	if err := database.DB.Exec("DELETE FROM labels WHERE id = ? AND user_id = ?", groupId, userId).Error; err != nil {
		return c.Status(500).JSON(err.Error())
	}

	return c.Status(200).JSON("success")
}

func getGroupActivities(userId uint, groupId uint) ([]models.Activity, error) {
	sql := "SELECT * FROM activities WHERE user_id = ? AND label_id = ?"

	var activities []models.Activity
	if err := database.DB.Raw(sql, userId, groupId).Scan(&activities).Error; err != nil {
		return []models.Activity{}, err
	}

	return activities, nil
}

func GetGroup(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)

	if err != nil {
		return c.Status(401).JSON(err.Error())
	}

	groupId, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(err.Error())
	}

	var group models.Label
	database.DB.Raw("SELECT * FROM labels WHERE id=? AND user_id=?", groupId, userId).Scan(&group)

	if group.Id == 0 {
		return c.Status(404).JSON("activity not found")
	}

	activities, err := getGroupActivities(userId, uint(groupId))
	if err != nil {
		return c.Status(500).JSON(err.Error())
	}

	return c.Status(200).JSON(fiber.Map{
		"group":      group,
		"activities": activities,
	})
}

func GetGroups(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)

	if err != nil {
		return c.Status(401).JSON(err.Error())
	}

	var groups []models.Label

	if err := database.DB.Raw("SELECT * FROM labels WHERE user_id = ?", userId).Scan(&groups).Error; err != nil {
		return c.Status(500).JSON(err.Error())
	}

	return c.Status(200).JSON(groups)
}
