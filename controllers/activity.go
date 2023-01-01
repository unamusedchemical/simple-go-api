package controllers

import (
	"awesomeProject/database"
	"awesomeProject/models"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"time"
)

type ResponseActivity struct {
	count   uint
	id      uint
	name    string
	content string
	closed  *time.Time
	opened  time.Time
	due     *time.Time
	groupId *int
}

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
		return c.Status(404).JSON("activity not found")
	}

	return c.Status(200).JSON(activity)
}

func DeleteActivity(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)

	if err != nil {
		c.Status(401).JSON("unauthenticated")
	}

	aId, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(err.Error())
	}

	status := database.DB.Exec("DELETE FROM activities WHERE id = ? AND user_id = ?", aId, userId)

	if status.RowsAffected == 0 {
		return c.SendStatus(404)
	}

	return c.Status(200).JSON("Success")
}

func UpdateActivity(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)

	if err != nil {
		c.Status(401).JSON(err.Error())
	}

	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	cId, _ := strconv.Atoi(data["group"])
	groupBelongs, err := groupBelongsToCurrentUser(userId, uint(cId))
	if err != nil {
		return c.Status(500).JSON(err.Error())
	} else if !groupBelongs {
		return c.Status(404).JSON("group does not belong to user")
	}

	//aId, err := c.ParamsInt("id")
	//if err != nil {
	//	return c.Status(400).JSON(err.Error())
	//}
	//
	//activityBelongs, err := activityBelongsToCurrentUser(userId, uint(aId))
	//if err != nil {
	//	return c.Status(500).JSON(err.Error())
	//} else if !activityBelongs {
	//	return c.Status(404).JSON("activity does not belong to user")
	//}
	//
	//activity := models.Activity{
	//	ActivityName:    data["name"],
	//	ActivityContent: data["content"],
	//}
	//*activity.CollectionId = uint(cId)

	var updatedActivity models.Activity

	//if err := database.DB.Exec("UPDATE activities SET activity_name=?, activity_content=?, group_id=? WHERE id = ? AND user_id = ?", activity.ActivityName, activity.ActivityContent, groupId, aId, userId).Error; err != nil {
	//	// returning err will cause rollback
	//	return c.Status(500).JSON(err.Error())
	//}

	if err != nil {
		return c.Status(500).JSON(err.Error())
	}

	return c.Status(200).JSON(updatedActivity)
}

func Activities(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)

	if err != nil {
		c.Status(401).JSON("unauthenticated")
	}

	var activities []models.Activity
	database.DB.Raw("SELECT * FROM activities WHERE user_id = ? ORDER BY closed_on ASC", userId).Scan(&activities)

	return c.Status(200).JSON(activities)
}

//func GetPaginatedActivities(c *fiber.Ctx) error {
//	userId, err := GetCurrentUserId(c)
//	if err != nil {
//		return c.SendStatus(401)
//	}
//
//	activities := []models.Activity{}
//	perPage := c.Query("per_page", "10")
//	sortOrder := c.Query("sort_order", "DESC")
//	sortOrder = strings.ToUpper(sortOrder)
//	cursor := c.Query("cursor", "")
//	limit, err := strconv.ParseInt(perPage, 10, 64)
//	if limit < 1 || limit > 25 {
//		limit = 10
//	}
//	if err != nil {
//		return c.Status(500).JSON("Invalid per_page option")
//	}
//
//	isFirstPage := cursor == ""
//	pointsNext := false
//
//	query := database.DB
//	if cursor != "" {
//		decodedCursor, err := util.DecodeCursor(cursor)
//		if err != nil {
//			fmt.Println(err)
//			return c.SendStatus(500)
//		}
//		pointsNext = decodedCursor["points_next"] == true
//
//		operator, order := util.GetPaginationOperator(pointsNext, sortOrder)
//		order = strings.ToUpper(order)
//		sql := fmt.Sprintf("SELECT * FROM activities WHERE user_id = ? AND (opened_on %s ? OR (opened_on = ? AND id %s ?))", operator, operator)
//		if order == "DESC" || order == "ASC" || order == "" {
//			query.Raw(sql+"ORDER BY opened_on "+order+" LIMIT ?", userId, decodedCursor["created_at"], decodedCursor["created_at"], decodedCursor["id"], sortOrder, int(limit)+1).Scan(&activities)
//		} else {
//			return c.SendStatus(400)
//		}
//
//	} else if sortOrder == "DESC" {
//		query.Raw("SELECT * FROM activities ORDER BY opened_on "+sortOrder+" LIMIT ?", int(limit)).Scan(&activities)
//	} else if sortOrder == "ASC" {
//		query.Raw("SELECT * FROM activities ORDER BY opened_on "+sortOrder+" LIMIT ?", int(limit)).Scan(&activities)
//	} else {
//		return c.SendStatus(400)
//	}
//
//	hasPagination := len(activities) > int(limit)
//
//	if hasPagination {
//		activities = activities[:limit]
//	}
//
//	if !isFirstPage && !pointsNext {
//		activities = util.Reverse(activities)
//	}
//
//	pageInfo := util.CalculatePagination(isFirstPage, hasPagination, int(limit), activities, pointsNext)
//
//	return c.Status(fiber.StatusOK).JSON(fiber.Map{
//		"success":    true,
//		"data":       activities,
//		"pagination": pageInfo,
//	})
//}

func CloseActivity(c *fiber.Ctx) error {
	userId, err := GetCurrentUserId(c)

	if err != nil {
		c.Status(401).JSON("unauthenticated")
	}

	aId, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(err.Error())
	}

	belongs, err := activityBelongsToCurrentUser(userId, uint(aId))

	if err != nil {
		return c.Status(500).JSON(err.Error())
	}

	if !belongs {
		return c.Status(403).JSON("cannot close a post that does not belong to the current user")
	}

	database.DB.Exec("UPDATE activities SET closed_on=NOW() WHERE id=? AND user_id=?", aId, userId)

	return c.Status(200).JSON("success")
}
