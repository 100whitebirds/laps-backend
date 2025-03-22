package validator

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegex    = regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	passwordRegex = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?]{6,}$`)
)

func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func ValidatePhone(phone string) bool {
	cleanPhone := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) || r == '+' {
			return r
		}
		return -1
	}, phone)

	return phoneRegex.MatchString(cleanPhone)
}

func ValidatePassword(password string) bool {
	if len(password) < 6 {
		return false
	}

	return passwordRegex.MatchString(password)
}

func ValidateNamePart(name string) bool {
	if len(name) < 2 {
		return false
	}

	for _, r := range name {
		if !unicode.IsLetter(r) && r != '-' && r != ' ' && r != '\'' {
			return false
		}
	}

	return true
}

func FormatPhone(phone string) string {
	cleanPhone := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) || r == '+' {
			return r
		}
		return -1
	}, phone)

	if !strings.HasPrefix(cleanPhone, "+") {
		if strings.HasPrefix(cleanPhone, "8") {
			cleanPhone = "+7" + cleanPhone[1:]
		} else if !strings.HasPrefix(cleanPhone, "7") {
			cleanPhone = "+7" + cleanPhone
		} else {
			cleanPhone = "+" + cleanPhone
		}
	}

	return cleanPhone
}

func FormatName(name string) string {
	if len(name) == 0 {
		return ""
	}

	parts := strings.Fields(name)
	for i, part := range parts {
		if strings.Contains(part, "-") {
			subparts := strings.Split(part, "-")
			for j, subpart := range subparts {
				if len(subpart) > 0 {
					subparts[j] = strings.ToUpper(subpart[:1]) + strings.ToLower(subpart[1:])
				}
			}
			parts[i] = strings.Join(subparts, "-")
		} else if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}

	return strings.Join(parts, " ")
}

func SanitizeString(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '<' || r == '>' || r == '&' || r == '"' || r == '\'' || r == '`' || r == ';' {
			return -1
		}
		return r
	}, s)
}
