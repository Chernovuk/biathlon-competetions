package biathlon

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config structure represents configuration
// that can be read from json file.
type Config struct {
	Laps        int      `json:"laps"`
	LapLen      float64  `json:"lapLen"`
	PenaltyLen  float64  `json:"penaltyLen"`
	FiringLines int      `json:"firingLines"`
	Start       justTime `json:"start"`
	StartDelta  duration `json:"startDelta"`
}

// justTime represents time.Time without date parameters
// to satisfy start time from config.json.
type justTime time.Time

// durationTime redefines time.Duration to be able to parse
// startDelta time as a correct duration parameter.
type duration time.Duration

// ParseConfig converts json data into Config struct.
func ParseConfig(filePath string) (Config, error) {
	configFile, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config file: %w", err)
	}

	settings := Config{}
	err = json.Unmarshal(configFile, &settings)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse config data: %w", err)
	}

	return settings, nil
}

func (c *justTime) UnmarshalJSON(b []byte) error {
	value := strings.Trim(string(b), `"`)
	if value == "" || value == "null" {
		return nil
	}

	t, err := time.Parse(time.TimeOnly, value)
	if err != nil {
		return err
	}
	*c = justTime(t)
	return nil
}

// func (t justTime) MarshalJSON() ([]byte, error) {
// 	return []byte(`"` + time.Time(t).Format("21:54:42.123") + `"`), nil
// }

func (d *duration) UnmarshalJSON(b []byte) error {
	value := strings.Trim(string(b), `"`)
	if value == "" || value == "null" {
		return nil
	}

	t, err := time.Parse(time.TimeOnly, value)
	if err != nil {
		return err
	}

	zeroTime, err := time.Parse(time.TimeOnly, "00:00:00")
	if err != nil {
		return err
	}
	tmp := t.Sub(zeroTime)
	*d = duration(tmp.Seconds())
	return nil
}

// func (d duration) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(time.Duration(d).String())
// }
