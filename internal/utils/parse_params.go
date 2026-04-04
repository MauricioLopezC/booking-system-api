package utils

import (
	"fmt"
	"strings"
)

// ParseSortParam Parse query param for sorting, allows only one orderBy param, for example:
// ?orderBy=price
// pass only allowedFields as a key in map with true value
// NOTE: This function should only verify that the parameter has the correct format.
// Then, parsing to GORM format should be done in the GORM-specific repository.
func ParseSortParam(sortParam string, allowedFields map[string]bool) (string, error) {
	var gormOrderBy string
	fieldName := sortParam

	if strings.HasPrefix(sortParam, "-") || strings.HasPrefix(sortParam, "+") {
		fieldName = sortParam[1:]
	}

	if !allowedFields[fieldName] {
		return "", fmt.Errorf("field %s is not allowed", fieldName)
	}

	if strings.HasPrefix(sortParam, "-") {
		gormOrderBy = fieldName + " desc"
	} else {
		gormOrderBy = fieldName + " asc"
	}

	return gormOrderBy, nil
}
