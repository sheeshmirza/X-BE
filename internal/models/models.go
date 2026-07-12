package models

import (
	"strings"
	"time"
)

type CareerCreateRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	Department string `json:"department"`
	Message    string `json:"message"`
}

type CareerApplication struct {
	Name       string    `bson:"name" json:"name"`
	Email      string    `bson:"email" json:"email"`
	Phone      string    `bson:"phone" json:"phone"`
	City       string    `bson:"city" json:"city"`
	State      string    `bson:"state" json:"state"`
	Country    string    `bson:"country" json:"country"`
	Department string    `bson:"department" json:"department"`
	Message    string    `bson:"message" json:"message"`
	Date       time.Time `bson:"date" json:"date"`
	Status     string    `bson:"status" json:"status"`
	Source     string    `bson:"source" json:"source"`
}

type NewsletterSubscriptionRequest struct {
	Email string `json:"email"`
}

type SaleCreateRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Query string `json:"query"`
}

type SaleLead struct {
	Name   string    `bson:"name" json:"name"`
	Email  string    `bson:"email" json:"email"`
	Query  string    `bson:"query" json:"query"`
	Date   time.Time `bson:"date" json:"date"`
	Status string    `bson:"status" json:"status"`
	Source string    `bson:"source" json:"source"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
