package databases

import (
	"fmt"

	postgresql_physical "databasus-backend/internal/features/databases/databases/postgresql/physical"
	"databasus-backend/internal/util/availability"
)

// RequireTypeAvailable is checked before anything connects to the database, so a database this
// instance cannot back up is refused up front instead of failing at its first backup.
func (d *Database) RequireTypeAvailable() error {
	switch d.Type {
	case DatabaseTypePostgresLogical:
		return availability.RequireEngine(availability.EnginePostgresql)

	case DatabaseTypePostgresPhysical:
		if err := availability.RequireFeature(availability.FeaturePhysicalBackup); err != nil {
			return err
		}

		if d.PostgresqlPhysical != nil &&
			d.PostgresqlPhysical.BackupType == postgresql_physical.BackupTypeFullIncrementalAndWalStream {
			return availability.RequireFeature(availability.FeaturePitr)
		}

		return nil

	case DatabaseTypeMysql:
		return availability.RequireEngine(availability.EngineMysql)

	case DatabaseTypeMariadb:
		return availability.RequireEngine(availability.EngineMariadb)

	case DatabaseTypeMongodb:
		return availability.RequireEngine(availability.EngineMongodb)

	default:
		return fmt.Errorf("invalid database type: %s", d.Type)
	}
}
