package repository

import (
	"llm-service/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func uuidToPgtype(id domain.ID) pgtype.UUID {
	return pgtype.UUID{Bytes: uuid.UUID(id), Valid: id != domain.ID{}}
}

func uuidsToPgtype(ids []domain.ID) []pgtype.UUID {
	result := make([]pgtype.UUID, 0, len(ids))
	for _, id := range ids {
		result = append(result, uuidToPgtype(id))
	}

	return result
}

func timeToPgtype(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: !t.IsZero()}
}

func durationFromPgtype(d pgtype.Interval) time.Duration {
	return time.Duration(d.Microseconds) * time.Microsecond
}

func intervalToPgtype(d time.Duration) pgtype.Interval {
	return pgtype.Interval{Microseconds: int64(d / time.Microsecond), Valid: d != 0}
}

func timeFromPgtype(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}

	return t.Time
}

func arrayToSlice[T any, U any](arr pgtype.FlatArray[T], f func(T) U) []U {
	result := make([]U, 0, len(arr.Dimensions()))
	for _, elem := range arr {
		result = append(result, f(elem))
	}

	return result
}

// nullableIDToString конвертирует nullable ID в *string
func nullableIDToString(id *domain.ID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

// parseTimestamp парсит timestamp строку в time.Time
func parseTimestamp(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func sliceToArray[T any, U any](slice []T, f func(T) U) pgtype.FlatArray[U] {
	result := make(pgtype.FlatArray[U], 0, len(slice))
	for _, elem := range slice {
		result = append(result, f(elem))
	}

	return result
}
