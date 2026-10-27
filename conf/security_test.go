package conf

import (
	"strings"
	"testing"
)

// compliantConfig 返回一份凭据合规的配置（仅用于启动校验用例）。
func compliantConfig() *Config {
	return &Config{
		Server: &Server{JwtSecret: strings.Repeat("a", minJWTSecretBytes)},
		MySQL:  &MySQL{Password: strings.Repeat("m", 24)},
		Redis:  &Redis{Password: strings.Repeat("r", 24)},
	}
}

// TestValidateSecurityAllowsCompliant 合规凭据 ⇒ 允许启动
func TestValidateSecurityAllowsCompliant(t *testing.T) {
	if err := compliantConfig().ValidateSecurity(); err != nil {
		t.Fatalf("合规配置不应拒绝启动，实际: %v", err)
	}
}

// TestValidateSecurityRejectsRepoDefaults 仓库默认凭据 ⇒ 三个变量全部报错，且不回显值
func TestValidateSecurityRejectsRepoDefaults(t *testing.T) {
	cfg := compliantConfig()
	cfg.Server.JwtSecret = repoDefaultJWTSecret
	cfg.MySQL.Password = repoDefaultMySQLPassword
	cfg.Redis.Password = repoDefaultRedisPassword

	err := cfg.ValidateSecurity()
	if err == nil {
		t.Fatal("仓库默认凭据必须拒绝启动")
	}
	msg := err.Error()
	for _, want := range []string{"JWT_SECRET", "MYSQL_PASSWORD", "REDIS_PASSWORD", "仍是仓库内默认值"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("错误信息缺少 %q，实际: %s", want, msg)
		}
	}
	assertNoSecretLeak(t, msg, repoDefaultJWTSecret, repoDefaultMySQLPassword, repoDefaultRedisPassword)
}

// TestValidateSecurityRejectsShortJWTSecret 长度 <32 字节 ⇒ 拒绝启动（提示生成方式）
func TestValidateSecurityRejectsShortJWTSecret(t *testing.T) {
	cfg := compliantConfig()
	cfg.Server.JwtSecret = strings.Repeat("s", minJWTSecretBytes-1)

	err := cfg.ValidateSecurity()
	if err == nil {
		t.Fatal("短 JWT 密钥必须拒绝启动")
	}
	msg := err.Error()
	for _, want := range []string{"JWT_SECRET", "长度不足", "openssl rand -hex 32"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("错误信息缺少 %q，实际: %s", want, msg)
		}
	}
	assertNoSecretLeak(t, msg, cfg.Server.JwtSecret)
}

// TestValidateSecurityRejectsEmptyCredentials 空 JWT / 空库口令 ⇒ 全部拒绝，且不回显值
func TestValidateSecurityRejectsEmptyCredentials(t *testing.T) {
	cfg := &Config{Server: &Server{}, MySQL: &MySQL{}, Redis: &Redis{}}

	err := cfg.ValidateSecurity()
	if err == nil {
		t.Fatal("空凭据必须拒绝启动")
	}
	msg := err.Error()
	for _, want := range []string{"JWT_SECRET 未设置", "MySQL 未设置", "Redis 未设置"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("错误信息缺少 %q，实际: %s", want, msg)
		}
	}
}

// TestValidateSecurityDoesNotLeakCustomSecrets 自定义强口令出错时也不得出现在错误信息里
func TestValidateSecurityDoesNotLeakCustomSecrets(t *testing.T) {
	jwtSecret := strings.Repeat("j", minJWTSecretBytes)
	mysqlPass := strings.Repeat("m", 40)
	cfg := &Config{
		Server: &Server{JwtSecret: jwtSecret},
		MySQL:  &MySQL{Password: mysqlPass},
		Redis:  &Redis{Password: ""},
	}

	err := cfg.ValidateSecurity()
	if err == nil {
		t.Fatal("Redis 空口令必须拒绝启动")
	}
	assertNoSecretLeak(t, err.Error(), jwtSecret, mysqlPass)
}

// assertNoSecretLeak 断言错误信息不含任何凭据值（避免日志二次泄露）
func assertNoSecretLeak(t *testing.T, msg string, secrets ...string) {
	t.Helper()
	for _, secret := range secrets {
		if secret != "" && strings.Contains(msg, secret) {
			t.Fatalf("错误信息泄露了凭据值（长度 %d）", len(secret))
		}
	}
}

// TestOverrideBodyLimitsWithEnv 请求体上限：env 覆盖 + 缺省兜底（SEC-01）
func TestOverrideBodyLimitsWithEnv(t *testing.T) {
	t.Setenv("MAX_REQUEST_BODY_BYTES", "1048576")
	t.Setenv("MAX_UPLOAD_BODY_BYTES", "8388608")

	server := &Server{}
	overrideBodyLimitsWithEnv(server)
	if server.MaxRequestBodyBytes != 1048576 {
		t.Fatalf("MaxRequestBodyBytes = %d, want 1048576", server.MaxRequestBodyBytes)
	}
	if server.MaxUploadBodyBytes != 8388608 {
		t.Fatalf("MaxUploadBodyBytes = %d, want 8388608", server.MaxUploadBodyBytes)
	}
}

// TestOverrideBodyLimitsDefaults env 缺省/非法 ⇒ 回落到 2 MB / 5 MB
func TestOverrideBodyLimitsDefaults(t *testing.T) {
	t.Setenv("MAX_REQUEST_BODY_BYTES", "")
	t.Setenv("MAX_UPLOAD_BODY_BYTES", "not-a-number")

	server := &Server{}
	overrideBodyLimitsWithEnv(server)
	if server.MaxRequestBodyBytes != DefaultMaxRequestBodyBytes {
		t.Fatalf("MaxRequestBodyBytes = %d, want %d", server.MaxRequestBodyBytes, DefaultMaxRequestBodyBytes)
	}
	if server.MaxUploadBodyBytes != DefaultMaxUploadBodyBytes {
		t.Fatalf("MaxUploadBodyBytes = %d, want %d", server.MaxUploadBodyBytes, DefaultMaxUploadBodyBytes)
	}
}
