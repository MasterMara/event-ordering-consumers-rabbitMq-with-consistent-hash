package util

import (
	"github.com/google/uuid"
)

func GetGuid() string {
	return uuid.New().String()
}
