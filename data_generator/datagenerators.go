package data_generator

import (
	uuid "github.com/satori/go.uuid"
	"strconv"
	"time"
)

type DataGenerators struct{}

func (g *DataGenerators) GenerateTime() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func (g *DataGenerators) GenerateUUID() string {
	return uuid.NewV4().String()
}
