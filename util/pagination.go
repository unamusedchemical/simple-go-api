package util

import (
	"awesomeProject/models"
	"encoding/base64"
	"encoding/json"
	"time"
)

type PaginationInfo struct {
	NextCursor string `json:"next_cursor"`
	PrevCursor string `json:"prev_cursor"`
}

type Cursor map[string]interface{}

func CreateCursor(id uint, createdAt time.Time, pointsNext bool) Cursor {
	return Cursor{
		"id":          id,
		"opened_on":   createdAt,
		"points_next": pointsNext,
	}
}

func GeneratePager(next Cursor, prev Cursor) PaginationInfo {
	return PaginationInfo{
		NextCursor: encodeCursor(next),
		PrevCursor: encodeCursor(prev),
	}
}

func encodeCursor(cursor Cursor) string {
	if len(cursor) == 0 {
		return ""
	}
	serializedCursor, err := json.Marshal(cursor)
	if err != nil {
		return ""
	}
	encodedCursor := base64.StdEncoding.EncodeToString(serializedCursor)
	return encodedCursor
}

func DecodeCursor(cursor string) (Cursor, error) {
	decodedCursor, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}

	var cur Cursor
	if err := json.Unmarshal(decodedCursor, &cur); err != nil {
		return nil, err
	}
	return cur, nil
}

func Reverse[T any](s []T) []T {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func GetPaginationOperator(pointsNext bool, sortOrder string) (string, string) {
	if pointsNext && sortOrder == "asc" {
		return ">", ""
	}
	if pointsNext && sortOrder == "desc" {
		return "<", ""
	}
	if !pointsNext && sortOrder == "asc" {
		return "<", "desc"
	}
	if !pointsNext && sortOrder == "desc" {
		return ">", "asc"
	}

	return "", ""
}

func CalculatePagination(isFirstPage bool, hasPagination bool, limit int, activities []models.Activity, pointsNext bool) PaginationInfo {
	pagination := PaginationInfo{}
	nextCur := Cursor{}
	prevCur := Cursor{}
	if isFirstPage {
		if hasPagination {
			nextCur := CreateCursor(activities[limit-1].Id, activities[limit-1].OpenedOn, true)
			pagination = GeneratePager(nextCur, nil)
		}
	} else {
		if pointsNext {
			// if pointing next, it always has prev but it might not have next
			if hasPagination {
				nextCur = CreateCursor(activities[limit-1].Id, activities[limit-1].OpenedOn, true)
			}
			prevCur = CreateCursor(activities[0].Id, activities[0].OpenedOn, false)
			pagination = GeneratePager(nextCur, prevCur)
		} else {
			// this is case of prev, there will always be nest, but prev needs to be calculated
			nextCur = CreateCursor(activities[limit-1].Id, activities[limit-1].OpenedOn, true)
			if hasPagination {
				prevCur = CreateCursor(activities[0].Id, activities[0].OpenedOn, false)
			}
			pagination = GeneratePager(nextCur, prevCur)
		}
	}
	return pagination
}
