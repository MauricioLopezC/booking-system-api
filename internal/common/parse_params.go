package common

import (
	"errors"
	"fmt"
	"strings"
)

// Parse query param for sorting, allows only one orderBy param, for example:
// ?orderBy=price
// pass only allowedFields as a key in map with true value
func ParseSortParam(sortParam string, allowedFields map[string]bool) (string, error) {
	var gormOrderBy string
	fieldName := sortParam

	if strings.HasPrefix(sortParam, "-") || strings.HasPrefix(sortParam, "+") {
		fieldName = sortParam[1:]
	}

	if !allowedFields[fieldName] {
		return "", errors.New(fmt.Sprintf("field %s is not allowed", fieldName))
	}

	if strings.HasPrefix(sortParam, "-") {
		gormOrderBy = fieldName + " desc"
	} else {
		gormOrderBy = fieldName + " asc"
	}

	return gormOrderBy, nil
}
