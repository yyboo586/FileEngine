package dbaccess

import "database/sql"

var (
	dbPool *sql.DB
)

func SetDBPool(i *sql.DB) {
	dbPool = i
}
