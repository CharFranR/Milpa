package validate

import (
	"fmt"

	"github.com/google/uuid"
)

type Rule struct {
	Field string
	Value any
}

func Request(rules []Rule) error {
	for _, r := range rules {
		switch v := r.Value.(type) {
		case string:
			if v == "" {
				return fmt.Errorf("%s: cannot be blank", r.Field)
			}
		case uuid.UUID:
			if v == uuid.Nil {
				return fmt.Errorf("%s: cannot be blank", r.Field)
			}
		}
	}
	return nil
}
