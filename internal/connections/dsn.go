package connections

import (
	"fmt"
	"sql-sharding-v2/internal/repository"
)

func buildDSN(s repository.ShardConnection) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		s.Username,
		s.Password,
		s.Host,
		s.Port,
		s.DatabaseName,
	)
}
