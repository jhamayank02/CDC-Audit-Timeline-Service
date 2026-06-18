package config

import "log/slog"

type Config struct {
	HTTP  HTTPConfig
	DB    DBConfig
	Kafka KafkaConfig
}

type HTTPConfig struct {
	Addr string
}

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
	SSLMode  string
}

type KafkaConfig struct {
	Brokers []string
	Topics  []string
	Topic   string
	GroupID string
}

func Load(logger *slog.Logger) (*Config, error) {
	env, err := NewEnv(logger)
	if err != nil {
		return nil, err
	}

	return &Config{
		HTTP: HTTPConfig{
			Addr: env.GetString("PORT", ":8080"),
		},
		DB: DBConfig{
			User:     env.GetString("DB_USER", "postgres"),
			Password: env.GetString("DB_PASSWORD", "postgres"),
			Host:     env.GetString("DB_HOST", "localhost"),
			Port:     env.GetString("DB_PORT", "5432"),
			Name:     env.GetString("DB_NAME", "postgres"),
			SSLMode:  env.GetString("DB_SSLMODE", "disable"),
		},
		Kafka: KafkaConfig{
			Brokers: env.GetStringSlice("KAFKA_BROKERS", []string{"localhost:29092"}),
			Topics:  env.GetStringSlice("KAFKA_TOPICS", []string{env.GetString("KAFKA_TOPIC", "cdc.public.users")}),
			Topic:   env.GetString("KAFKA_TOPIC", "cdc.public.users"),
			GroupID: env.GetString("KAFKA_GROUP_ID", "cdc-audit-consumer"),
		},
	}, nil
}
