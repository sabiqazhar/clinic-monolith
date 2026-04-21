package helper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func ToPgNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(f)
	return n
}

func ToPgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func ToPgTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
