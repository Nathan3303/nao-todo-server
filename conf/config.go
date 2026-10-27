package conf

import (
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

var Conf *Config

type Config struct {
	Server   *Server    `yaml:"server"`
	MySQL    *MySQL     `yaml:"mysql"`
	Redis    *Redis     `yaml:"redis"`
	Log      *LogConfig `yaml:"log"`
	Uploads  *Uploads   `yaml:"uploads"`
	Reminder Reminder   `yaml:"reminder"` // 值类型：配置缺省时取零值，等价于「停用字段改写」
	Session  Session    `yaml:"session"`  // 值类型：配置缺省时取零值，等价于「停用会话清理」
}

// Session 会话相关配置
type Session struct {
	Cleanup SessionCleanup `yaml:"cleanup"`
}

// SessionCleanup 过期会话清理任务配置（nao-todo-server#56 / DEF-14）
// 无 DDL：删除对象仅为 user_sessions 历史行；回滚方式为 enabled=false + 重启。
type SessionCleanup struct {
	// Enabled 是否启用清理任务；零值 false = 停用（缺省配置下不会删除任何行）
	Enabled bool `yaml:"enabled"`
	// Cron 清理任务执行周期，默认每日 03:10（避开 04:00 删注销用户与每分钟提醒扫描）
	Cron string `yaml:"cron"`
	// RetentionDays 过期会话保留宽限期（天）：expired_at < now - RetentionDays 才回收
	RetentionDays int `yaml:"retentionDays"`
	// RevokedRetentionHours 已登出会话保留宽限期（小时）：deleted_at < now - RevokedRetentionHours 才回收
	RevokedRetentionHours int `yaml:"revokedRetentionHours"`
	// BatchSize 单批删除行数上限
	BatchSize int `yaml:"batchSize"`
	// MaxBatches 单次执行最多批数（过期 + 已登出共享该预算）
	MaxBatches int `yaml:"maxBatches"`
	// RunOnStart 进程启动后立即执行一次（仅部署验证窗口临时开启，默认 false）
	RunOnStart bool `yaml:"runOnStart"`
}

// applyDefaults 为缺省字段补默认值（batchSize/maxBatches 为 0 会让清理静默失效，故必须兜底）。
// enabled 不做兜底：零值即停用，缺失配置时宁可不删，避免误删。
func (sc *SessionCleanup) applyDefaults() {
	if sc.Cron == "" {
		sc.Cron = "10 3 * * *"
	}
	if sc.RetentionDays <= 0 {
		sc.RetentionDays = 7
	}
	if sc.RevokedRetentionHours <= 0 {
		sc.RevokedRetentionHours = 24
	}
	if sc.BatchSize <= 0 {
		sc.BatchSize = 500
	}
	if sc.MaxBatches <= 0 {
		sc.MaxBatches = 20
	}
}

// Reminder 提醒相关开关
type Reminder struct {
	// CronFieldWriteBack 提醒 cron 是否把判定结果写回任务字段
	// （清空一次性 remind_at / 推进重复提醒，并推送 /sse/reminders 事件）。
	// 默认 false = 停用：提醒判定下移到客户端，服务端保留到期痕迹以支持客户端补弹；
	// 置 true 即回滚到旧行为。
	CronFieldWriteBack bool `yaml:"cronFieldWriteBack"`
}

type Uploads struct {
	UploadDir   string `yaml:"uploadDir"`   // 文件上传根目录
	AvatarDir   string `yaml:"avatarDir"`   // 头像存储子目录
	MaxFileSize int64  `yaml:"maxFileSize"` // 最大文件大小（字节）
	StaticPath  string `yaml:"staticPath"`  // 静态文件访问路径
	ServerURL   string `yaml:"serverURL"`   // 服务器地址（用于拼接头像完整 URL）
}

// AvatarURL 将相对路径的头像 URL 拼接为完整 URL。
// 若 avatar 为空、已是绝对 URL、或 ServerURL 未配置，则原样返回。
func (u *Uploads) AvatarURL(avatar string) string {
	if avatar == "" || u.ServerURL == "" {
		return avatar
	}
	if strings.HasPrefix(avatar, "http://") ||
		strings.HasPrefix(avatar, "https://") {
		return avatar
	}
	return u.ServerURL + avatar
}

type LogConfig struct {
	Level         string `yaml:"level"`         // 日志级别：debug, info, warn, error, fatal, panic
	FilePath      string `yaml:"filePath"`      // 日志文件路径
	MaxSize       int    `yaml:"maxSize"`       // 单个日志文件最大大小（MB）
	MaxAge        int    `yaml:"maxAge"`        // 日志文件保留天数
	MaxBackups    uint   `yaml:"maxBackups"`    // 保留的日志文件副本数量
	Compress      bool   `yaml:"compress"`      // 是否压缩日志文件
	OutputConsole bool   `yaml:"outputConsole"` // 是否同时输出到控制台
}

type Server struct {
	Ip        string `yaml:"ip"`
	Port      string `yaml:"port"`
	Version   string `yaml:"version"`
	JwtSecret string `yaml:"jwtSecret"`
	GoMaxProc int    `yaml:"goMaxProc"`
	CertFile  string `yaml:"certFile"` // TLS 证书路径（空=HTTP）
	KeyFile   string `yaml:"keyFile"`  // TLS 私钥路径（空=HTTP）
	Debug     bool   `yaml:"debug"`    // 是否开启调试模式
	// MaxRequestBodyBytes 全局请求体上限（字节，SEC-01）；<=0 时取默认 2 MB
	MaxRequestBodyBytes int64 `yaml:"maxRequestBodyBytes"`
	// MaxUploadBodyBytes 上传路由（头像）请求体上限（字节，SEC-01）；<=0 时取默认 5 MB
	MaxUploadBodyBytes int64 `yaml:"maxUploadBodyBytes"`
}

// 请求体上限默认值（SEC-01 / T561）
const (
	DefaultMaxRequestBodyBytes int64 = 2 << 20 // 2 MB
	DefaultMaxUploadBodyBytes  int64 = 5 << 20 // 5 MB（头像上传单独放宽）
)

type MySQL struct {
	Host      string `yaml:"host"`
	Port      string `yaml:"port"`
	Database  string `yaml:"database"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Charset   string `yaml:"charset"`
	ParseTime string `yaml:"parseTime"`
	Loc       string `yaml:"loc"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DB       int    `yaml:"database"`
	Password string `yaml:"password"`
	Network  string `yaml:"network"`
}

func InitConfig() {
	configDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir + "/conf")

	err = viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&Conf)
	if err != nil {
		panic(err)
	}

	overrideWithEnv()
}

func overrideWithEnv() {
	var env string
	if env = os.Getenv("JWT_SECRET"); env != "" {
		Conf.Server.JwtSecret = env
	}
	if env = os.Getenv("MYSQL_HOST"); env != "" {
		Conf.MySQL.Host = env
	}
	if env = os.Getenv("MYSQL_PORT"); env != "" {
		Conf.MySQL.Port = env
	}
	if env = os.Getenv("MYSQL_USER"); env != "" {
		Conf.MySQL.Username = env
	}
	if env = os.Getenv("MYSQL_PASSWORD"); env != "" {
		Conf.MySQL.Password = env
	}
	if env = os.Getenv("MYSQL_DATABASE"); env != "" {
		Conf.MySQL.Database = env
	}
	if env = os.Getenv("REDIS_HOST"); env != "" {
		Conf.Redis.Host = env
	}
	if env = os.Getenv("REDIS_PORT"); env != "" {
		Conf.Redis.Port = env
	}
	if env = os.Getenv("REDIS_PASSWORD"); env != "" {
		Conf.Redis.Password = env
	}
	if env = os.Getenv("REDIS_DB"); env != "" {
		if db, err := strconv.Atoi(env); err == nil {
			Conf.Redis.DB = db
		}
	}
	if env = os.Getenv("APP_PORT"); env != "" {
		Conf.Server.Port = env
	}
	Conf.Server.Debug = os.Getenv("APP_DEBUG") == "true"
	if env = os.Getenv("GOMAXPROCS"); env != "" {
		if maxProc, err := strconv.Atoi(env); err == nil {
			Conf.Server.GoMaxProc = maxProc
		}
	}
	if env = os.Getenv("APP_SERVER_URL"); env != "" {
		Conf.Uploads.ServerURL = env
	}
	if Conf.Server.Debug {
		Conf.Server.CertFile = ""
		Conf.Server.KeyFile = ""
	} else {
		if env = os.Getenv("TLS_CERT_FILE"); env != "" {
			Conf.Server.CertFile = env
		}
		if env = os.Getenv("TLS_KEY_FILE"); env != "" {
			Conf.Server.KeyFile = env
		}
	}
	overrideSessionCleanupWithEnv(&Conf.Session.Cleanup)
	Conf.Session.Cleanup.applyDefaults()
	overrideBodyLimitsWithEnv(Conf.Server)
}

// overrideBodyLimitsWithEnv 用 MAX_REQUEST_BODY_BYTES / MAX_UPLOAD_BODY_BYTES 覆盖请求体上限，
// 非法值（非正整数）忽略；缺省或 <=0 时取默认值（SEC-01 / T561）。
func overrideBodyLimitsWithEnv(server *Server) {
	if server == nil {
		return
	}
	if v := os.Getenv("MAX_REQUEST_BODY_BYTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			server.MaxRequestBodyBytes = n
		}
	}
	if v := os.Getenv("MAX_UPLOAD_BODY_BYTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			server.MaxUploadBodyBytes = n
		}
	}
	if server.MaxRequestBodyBytes <= 0 {
		server.MaxRequestBodyBytes = DefaultMaxRequestBodyBytes
	}
	if server.MaxUploadBodyBytes <= 0 {
		server.MaxUploadBodyBytes = DefaultMaxUploadBodyBytes
	}
}

// overrideSessionCleanupWithEnv 用 SESSION_CLEANUP_* 环境变量覆盖会话清理配置
func overrideSessionCleanupWithEnv(sc *SessionCleanup) {
	var env string
	if env = os.Getenv("SESSION_CLEANUP_ENABLED"); env != "" {
		if enabled, err := strconv.ParseBool(env); err == nil {
			sc.Enabled = enabled
		}
	}
	if env = os.Getenv("SESSION_CLEANUP_CRON"); env != "" {
		sc.Cron = env
	}
	if env = os.Getenv("SESSION_CLEANUP_RETENTION_DAYS"); env != "" {
		if days, err := strconv.Atoi(env); err == nil {
			sc.RetentionDays = days
		}
	}
	if env = os.Getenv("SESSION_CLEANUP_REVOKED_RETENTION_HOURS"); env != "" {
		if hours, err := strconv.Atoi(env); err == nil {
			sc.RevokedRetentionHours = hours
		}
	}
	if env = os.Getenv("SESSION_CLEANUP_BATCH_SIZE"); env != "" {
		if size, err := strconv.Atoi(env); err == nil {
			sc.BatchSize = size
		}
	}
	if env = os.Getenv("SESSION_CLEANUP_MAX_BATCHES"); env != "" {
		if batches, err := strconv.Atoi(env); err == nil {
			sc.MaxBatches = batches
		}
	}
	if env = os.Getenv("SESSION_CLEANUP_RUN_ON_START"); env != "" {
		if runOnStart, err := strconv.ParseBool(env); err == nil {
			sc.RunOnStart = runOnStart
		}
	}
}
