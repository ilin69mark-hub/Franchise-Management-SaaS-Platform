package handlers

import (
	"html"
	"regexp"
	"strings"
)

var (
	emailRegex   = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegex   = regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	alphanumeric = regexp.MustCompile(`[^a-zA-Z0-9а-яА-ЯёЁ\s\-_]`)
)

func SanitizeString(input string) string {
	return strings.TrimSpace(input)
}

func SanitizeHTML(input string) string {
	return html.EscapeString(input)
}

func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func ValidatePhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

func StripDangerousChars(input string) string {
	return alphanumeric.ReplaceAllString(input, "")
}