package validation

import (
	"errors"
	"regexp"
)

func ValidatePort(port int) error {
	if port <= 0 || port >= 65535 {
		return errors.New("port number must between 0~25535")
	}

	return nil
}

func ValidateName(name string) error {
	if ok, _ := regexp.MatchString("^[a-zA-Z]{1}[a-zA-Z0-9_-]{4,20}$", name); !ok {
		return errors.New("name can only start with letters, then contains number - and _, length between 4~20")
	}

	return nil
}
