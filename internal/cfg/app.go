package cfg

import (
	"log"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type GeneralConfig struct {
	AppConfig    AppConfig    `env-prefix:"APP_"`
	BucketConfig BucketConfig `env-prefix:"BUCKET_"`
	TasksConfig  TasksConfig  `env-prefix:"TASKS_"`
}

type AppConfig struct {
	Port              string        `env:"PORT" env-default:"8080"`
	IdempotencyTTL    time.Duration `env:"IDEMPOTENCY_TTL" env-default:"1m"`
	ReadTimeout       time.Duration `env:"READ_TIMEOUT" env-default:"15s"`
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT" env-default:"1s"`
	WriteTimeout      time.Duration `env:"WRITE_TIMEOUT" env-default:"30s"`
	IdleTimeout       time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
	ReplacedAudit     string        `env:"REPLACED_AUDIT" env-default:"[REDACTED]"`
}

type BucketConfig struct {
	Name            string `env:"NAME" env-default:"auditory-bucket"`
	Endpoint        string `env:"ENDPOINT" env-default:"http://localhost:9000"`
	AccessKeyID     string `env:"ACCESS_KEY_ID" env-default:"minioadmin"`
	SecretAccessKey string `env:"SECRET_ACCESS_KEY" env-default:"minioadmin"`
	Region          string `env:"REGION" env-default:"us-east-1"`
	UsePathStyle    bool   `env:"USE_PATH_STYLE" env-default:"true"`

	ExpiresBackupDays int `env:"EXPIRES_BACKUP_DAYS" env-default:"2"`
	ExpiresStoreDays  int `env:"EXPIRES_STORE_DAYS" env-default:"365"`
}

type TasksConfig struct {
	IdempotencyClearPeriod time.Duration `env:"IDEMPOTENCY_CLEAR_PERIOD" env-default:"1m"`
	BackupPeriod           time.Duration `env:"BACKUP_PERIOD" env-default:"30m"`
	StorePeriod            time.Duration `env:"STORE_PERIOD" env-default:"1h"`
}

var (
	cfg  *GeneralConfig
	once sync.Once
)

func InitConfig() {
	once.Do(func() {
		cfg = &GeneralConfig{}
		if err := cleanenv.ReadEnv(cfg); err != nil {
			log.Fatalf("failed to read config: %v", err)
		}
	})
}

func GetConfig() *GeneralConfig {
	return cfg
}

func SetConfig(c *GeneralConfig) {
	cfg = c
}
