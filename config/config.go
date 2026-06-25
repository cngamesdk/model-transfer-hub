package config

// System 系统配置
type System struct {
	Addr string `mapstructure:"addr" json:"addr" yaml:"addr"` // 监听地址
	Mode string `mapstructure:"mode" json:"mode" yaml:"mode"` // 运行模式 debug/release
}

// Database 数据库配置
type Database struct {
	Type         string `mapstructure:"type" json:"type" yaml:"type"`
	Host         string `mapstructure:"host" json:"host" yaml:"host"`
	Port         string `mapstructure:"port" json:"port" yaml:"port"`
	DbName       string `mapstructure:"db-name" json:"db-name" yaml:"db-name"`
	Username     string `mapstructure:"username" json:"username" yaml:"username"`
	Password     string `mapstructure:"password" json:"password" yaml:"password"`
	MaxIdleConns int    `mapstructure:"max-idle-conns" json:"max-idle-conns" yaml:"max-idle-conns"`
	MaxOpenConns int    `mapstructure:"max-open-conns" json:"max-open-conns" yaml:"max-open-conns"`
	LogMode      string `mapstructure:"log-mode" json:"log-mode" yaml:"log-mode"`
}

// Zap 日志配置
type Zap struct {
	Level         string `mapstructure:"level" json:"level" yaml:"level"`
	Prefix        string `mapstructure:"prefix" json:"prefix" yaml:"prefix"`
	Format        string `mapstructure:"format" json:"format" yaml:"format"`
	Director      string `mapstructure:"director" json:"director" yaml:"director"`
	EncodeLevel   string `mapstructure:"encode-level" json:"encode-level" yaml:"encode-level"`
	StacktraceKey string `mapstructure:"stacktrace-key" json:"stacktrace-key" yaml:"stacktrace-key"`
	ShowLine      bool   `mapstructure:"show-line" json:"show-line" yaml:"show-line"`
	LogInConsole  bool   `mapstructure:"log-in-console" json:"log-in-console" yaml:"log-in-console"`
	RetentionDay  int    `mapstructure:"retention-day" json:"retention-day" yaml:"retention-day"`
	MaxSize       int    `mapstructure:"max-size" json:"max-size" yaml:"max-size"` // MB
}

// RateLimit 限流配置
type RateLimit struct {
	Enabled    bool `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	DefaultRPM int  `mapstructure:"default_rpm" json:"default_rpm" yaml:"default_rpm"` // 每分钟请求数
	DefaultRPH int  `mapstructure:"default_rph" json:"default_rph" yaml:"default_rph"` // 每小时请求数
}

// Trace 链路追踪配置
type Trace struct {
	HeaderName        string `mapstructure:"header_name" json:"header_name" yaml:"header_name"`
	GenerateIfMissing bool   `mapstructure:"generate_if_missing" json:"generate_if_missing" yaml:"generate_if_missing"`
}

// Server 服务配置
type Server struct {
	System    System     `mapstructure:"system" json:"system" yaml:"system"`
	Database  Database   `mapstructure:"database" json:"database" yaml:"database"`
	Zap       Zap        `mapstructure:"zap" json:"zap" yaml:"zap"`
	Providers []Provider `mapstructure:"providers" json:"providers" yaml:"providers"`
	RateLimit RateLimit  `mapstructure:"rate_limit" json:"rate_limit" yaml:"rate_limit"`
	Trace     Trace      `mapstructure:"trace" json:"trace" yaml:"trace"`
	DevMode   bool       `mapstructure:"dev_mode" json:"dev_mode" yaml:"dev_mode"`
}
