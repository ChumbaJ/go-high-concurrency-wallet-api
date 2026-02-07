package env

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Env struct {
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	DBTimezone      string
	DBMaxOpenConns  int
	DBMaxIdleConns  int
	DBConnMaxLifetime time.Duration
	RequestTimeout  time.Duration
	RateLimit        int
	RateLimitPeriod  time.Duration
	QueueBuffSize    int
	QueueFlushPeriod time.Duration
}

func LoadFromFile(path string) (*Env, error) {
	m, err := parseEnvFile(path)
	if err != nil {
		return nil, err
	}
	
	getEnv := func(key string) string {
		if val := os.Getenv(key); val != "" {
			return val
		}
		return m[key]
	}
	
	e := &Env{
		DBHost:     getEnv("DB_HOST"),
		DBPort:     getEnv("DB_PORT"),
		DBUser:     getEnv("DB_USER"),
		DBPassword: getEnv("DB_PASSWORD"),
		DBName:     getEnv("DB_NAME"),
		DBSSLMode:  getEnv("DB_SSLMODE"),
		DBTimezone: defaultString(getEnv("DB_TIMEZONE"), "UTC"),
	}

	maxOpenConns := defaultString(m["DB_MAX_OPEN_CONNS"], "50")
	if err := parseInt(maxOpenConns, &e.DBMaxOpenConns); err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_OPEN_CONNS: %w", err)
	}

	maxIdleConns := defaultString(m["DB_MAX_IDLE_CONNS"], "25")
	if err := parseInt(maxIdleConns, &e.DBMaxIdleConns); err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_IDLE_CONNS: %w", err)
	}

	connMaxLifetimeStr := defaultString(m["DB_CONN_MAX_LIFETIME"], "15m")
	connMaxLifetime, err := time.ParseDuration(connMaxLifetimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME: %w", err)
	}
	e.DBConnMaxLifetime = connMaxLifetime

	timeoutStr := defaultString(m["REQUEST_TIMEOUT"], "30s")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REQUEST_TIMEOUT: %w", err)
	}
	e.RequestTimeout = timeout
	
	rateLimitStr := defaultString(m["RATE_LIMIT"], "100")
	if err := parseInt(rateLimitStr, &e.RateLimit); err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT: %w", err)
	}

	rateLimitPeriodStr := defaultString(m["RATE_LIMIT_PERIOD"], "1m")
	rateLimitPeriod, err := time.ParseDuration(rateLimitPeriodStr)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_PERIOD: %w", err)
	}
	e.RateLimitPeriod = rateLimitPeriod

	buffSizeStr := defaultString(getEnv("QUEUE_BUFF_SIZE"), "50")
	if err := parseInt(buffSizeStr, &e.QueueBuffSize); err != nil {
		return nil, fmt.Errorf("invalid QUEUE_BUFF_SIZE: %w", err)
	}

	flushPeriodStr := defaultString(getEnv("QUEUE_FLUSH_PERIOD"), "100ms")
	flushPeriod, err := time.ParseDuration(flushPeriodStr)
	if err != nil {
		return nil, fmt.Errorf("invalid QUEUE_FLUSH_PERIOD: %w", err)
	}
	e.QueueFlushPeriod = flushPeriod

	if err := e.Validate(); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *Env) Validate() error {
	var missing []string

	if e.DBHost == "" {
		missing = append(missing, "DB_HOST")
	}
	if e.DBPort == "" {
		missing = append(missing, "DB_PORT")
	}
	if e.DBUser == "" {
		missing = append(missing, "DB_USER")
	}
	if e.DBPassword == "" {
		missing = append(missing, "DB_PASSWORD")
	}
	if e.DBName == "" {
		missing = append(missing, "DB_NAME")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}

	if e.DBMaxOpenConns <= 0 {
		return fmt.Errorf("DB_MAX_OPEN_CONNS must be > 0")
	}
	if e.DBMaxIdleConns < 0 {
		return fmt.Errorf("DB_MAX_IDLE_CONNS must be >= 0")
	}
	if e.QueueBuffSize <= 0 {
		return fmt.Errorf("QUEUE_BUFF_SIZE must be > 0")
	}

	return nil
}

func parseEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		i := strings.Index(line, "=")
		if i < 0 {
			continue
		}
		key := strings.TrimSpace(line[:i])
		val := strings.TrimSpace(line[i+1:])
		val = strings.Trim(val, `"`)
		if key != "" {
			m[key] = val
		}
	}
	return m, scanner.Err()
}

func parseInt(s string, dst *int) error {
	val, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*dst = val
	return nil
}

func defaultString(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
