package conf

type MigratorAppConf struct {
	WorkersNum                 int `mapstructure:"workersNum"`
	MigrateStepMinutes         int `mapstructure:"migrateStepMinutes"`
	MigrateSucessExpireMinutes int `mapstructure:"migrateSuccessExpireMinutes"`
	MigrateTryLockMinutes      int `mapstructure:"migrateTryLockMinutes"`
	TimerDetailCacheMinutes    int `mapstructure:"timerDetailCacheMinutes"`
}

var defaultMigratorAppConfProvider *MigratorAppConfProvider

type MigratorAppConfProvider struct {
	conf *MigratorAppConf
}

func NewMigratorAppConfProvider(conf *MigratorAppConf) *MigratorAppConfProvider {
	return &MigratorAppConfProvider{
		conf: conf,
	}
}

func (m *MigratorAppConfProvider) Get() *MigratorAppConf {
	return m.conf
}

func DefaultMigratorAppConfProvider() *MigratorAppConfProvider {
	return defaultMigratorAppConfProvider
}
