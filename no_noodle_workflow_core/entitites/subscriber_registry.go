package entitites

import "time"

// type ChannelInfo struct {
// 	Channel        string `json:"channel"`
// 	HealthCheckURL string `json:"health_check_url"`
// 	RegisterURL    string `json:"callback_url"`
// }

// session_key VARCHAR(255) PRIMARY KEY,
// process_id VARCHAR(255) NOT NULL,
// task VARCHAR(255) NOT NULL,
// health_check_url TEXT NOT NULL,
// callback_url TEXT NOT NULL,
// create_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
// FOREIGN KEY (process_id) REFERENCES process_config (process_id)

type SubscriberRegistry struct {
	SessionKey     string    `json:"session_key"`
	ProcessID      string    `json:"process_id"`
	Task           string    `json:"task"`
	HealthCheckURL string    `json:"health_check_url"`
	CallbackURL    string    `json:"callback_url"`
	CreateDate     time.Time `json:"create_date"`
}
