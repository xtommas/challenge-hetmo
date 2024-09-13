package models

import "time"

type Event struct {
	Id               int64     `json:"id"`
	Title            string    `json:"title" validate:"required,min=3,max=100"`
	LongDescription  string    `json:"long_description" validate:"required,min=10"`
	ShortDescription string    `json:"short_description" validate:"required,max=200"`
	DateAndTime      time.Time `json:"date_and_time" validate:"required,gt=now"`
	Organizer        string    `json:"organizer" validate:"required"`
	Location         string    `json:"location" validate:"required"`
	Status           string    `json:"status" validate:"required,oneof=draft published"`
}
