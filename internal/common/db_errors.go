package common

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// pgExclusionViolation is the postgres error code raised by an EXCLUDE constraint.
const pgExclusionViolation = "23P01"

// IsExclusionViolation reports whether err comes from an EXCLUDE constraint,
// e.g. bookings_no_overlap when two bookings of the same employee overlap.
func IsExclusionViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgExclusionViolation
}
