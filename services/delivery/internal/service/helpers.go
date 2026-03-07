package service

import "strings"

func redactValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 6 {
		return "******"
	}
	return value[:3] + "..." + value[len(value)-3:]
}
