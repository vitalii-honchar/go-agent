package validation

import (
	"errors"
	"fmt"
	"regexp"
)

var ErrValidationFailed = errors.New("validation failed")

const (
	snakeCasePattern     = `^[a-z0-9_]+$`
	maxNameLength        = 64
	maxDescriptionLength = 1024
)

func NameIsValid(name string) error {
	if err := StringIsNotEmpty(name); err != nil {
		return err
	}
	if err := StringIsMaxLength(name, maxNameLength); err != nil {
		return err
	}
	if err := StringMatchesPattern(name, snakeCasePattern); err != nil {
		return err
	}

	return nil
}

func DescriptionIsValid(description string) error {
	if err := StringIsNotEmpty(description); err != nil {
		return err
	}
	if err := StringIsMaxLength(description, maxDescriptionLength); err != nil {
		return err
	}

	return nil
}

func NotNil(value any) error {
	if value == nil {
		return fmt.Errorf("%w: value cannot be nil", ErrValidationFailed)
	}

	return nil
}

func StringIsNotEmpty(s string) error {
	if s == "" {
		return fmt.Errorf("%w: string cannot be empty", ErrValidationFailed)
	}

	return nil
}

func StringIsMaxLength(s string, max int) error {
	if len(s) > max {
		return fmt.Errorf("%w: string cannot be longer than %d characters", ErrValidationFailed, max)
	}

	return nil
}

func StringMatchesPattern(s, pattern string) error {
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		return fmt.Errorf("%w: failed to match string against pattern %s: %w", ErrValidationFailed, pattern, err)
	}
	if !matched {
		return fmt.Errorf("%w: string does not match pattern %s", ErrValidationFailed, pattern)
	}

	return nil
}
