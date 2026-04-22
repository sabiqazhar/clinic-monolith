package helper

import (
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func ToPgNumeric(f float64) pgtype.Numeric {
	if f == 0 {
		return pgtype.Numeric{Int: big.NewInt(0), Valid: true}
	}
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
