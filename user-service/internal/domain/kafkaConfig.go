package domain

type KafkaConfig struct {
	GroupID string `json:"omitempty"`
	Topic   string
	Brokers []string
}
