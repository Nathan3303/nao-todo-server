package conf

import (
	"errors"
	"fmt"
	"strings"
)

// 启动期安全校验（SEC-02 / T561）。
//
// 背景：conf/config.yaml 内提交了非占位符的默认凭据，overrideWithEnv 只在 env 非空时覆盖，
// 因此部署 .env 缺项时会静默回落到仓库默认值。这里在监听之前 fail-fast：
// 凭据为空、过短或仍等于仓库默认值 ⇒ 拒绝启动。
//
// ⛔ 校验输出只说明「哪个变量不合格 + 期望」，绝不包含变量值本身（避免日志二次泄露）。

// 仓库内 conf/config.yaml 的默认凭据（仅用于比对，禁止写入日志/错误/响应）。
const (
	// 仓库内 conf/config.yaml 的默认凭据字面量：仅用于启动校验比对。
	// 它们不是线上凭据，也不会进入日志/错误/响应（归属见下方 nolint 说明）。
	repoDefaultJWTSecret     = "jwt@naotodo-test.secret" //nolint:gosec,lll // 仅比对仓库默认值
	repoDefaultMySQLPassword = "naotodo-test@mysql.pass" //nolint:gosec,lll // 仅比对仓库默认值
	repoDefaultRedisPassword = "naotodo-test@redis.pass" //nolint:gosec,lll // 仅比对仓库默认值
)

// minJWTSecretBytes JWT 密钥最小长度（字节）
const minJWTSecretBytes = 32

// jwtEnvHint JWT 密钥的可操作提示（统一文案）
const jwtEnvHint = "请通过环境变量 JWT_SECRET 提供随机串，生成方式：openssl rand -hex 32"

// ValidateSecurity 校验启动所需的安全默认值；返回 nil 表示可以启动。
// 一次性汇总所有问题，便于部署时一轮修完；错误信息不含任何密钥/口令值。
func (c *Config) ValidateSecurity() error {
	if c == nil {
		return errors.New("安全校验失败（SEC-02），拒绝启动：配置未加载")
	}

	problems := make([]string, 0, 3)
	problems = append(problems, validateJWTSecret(c.Server)...)
	problems = append(problems, validateCredential(
		"MySQL", "MYSQL_PASSWORD", passwordOfMySQL(c.MySQL), repoDefaultMySQLPassword,
	)...)
	problems = append(problems, validateCredential(
		"Redis", "REDIS_PASSWORD", passwordOfRedis(c.Redis), repoDefaultRedisPassword,
	)...)

	if len(problems) > 0 {
		return fmt.Errorf(
			"安全默认值校验未通过（SEC-02），拒绝启动：\n- %s",
			strings.Join(problems, "\n- "),
		)
	}
	return nil
}

// validateJWTSecret 校验 JWT 密钥：空 / 过短 / 等于仓库默认值均拒绝。
// 先判「等于默认值」，避免默认值（23 字节）被误报成单纯长度不足。
func validateJWTSecret(server *Server) []string {
	if server == nil {
		return []string{"配置缺少 server 段，无法校验 JWT 密钥"}
	}
	switch {
	case server.JwtSecret == "":
		return []string{"环境变量 JWT_SECRET 未设置（当前为空）；" + jwtEnvHint}
	case server.JwtSecret == repoDefaultJWTSecret:
		return []string{"环境变量 JWT_SECRET 仍是仓库内默认值（必须轮换）；" + jwtEnvHint}
	case len(server.JwtSecret) < minJWTSecretBytes:
		return []string{fmt.Sprintf(
			"环境变量 JWT_SECRET 长度不足（%d 字节，要求 ≥%d 字节）；%s",
			len(server.JwtSecret), minJWTSecretBytes, jwtEnvHint,
		)}
	default:
		return nil
	}
}

// validateCredential 校验 DB/Redis 口令：空或等于仓库默认值均拒绝。
// @param name 组件名（仅用于提示文案）
// @param envName 环境变量名（提示用户改哪里）
// @param value 当前口令
// @param repoDefault 仓库内默认口令（比对用，不输出）
func validateCredential(name, envName, value, repoDefault string) []string {
	if value == "" {
		return []string{fmt.Sprintf(
			"环境变量 %s 未设置（%s 口令为空）；请通过环境变量 %s 配置强口令",
			name, envName, envName,
		)}
	}
	if value == repoDefault {
		return []string{fmt.Sprintf(
			"环境变量 %s 口令仍是仓库内默认值（必须轮换）；请重新配置 %s", name, envName,
		)}
	}
	return nil
}

func passwordOfMySQL(mysql *MySQL) string {
	if mysql == nil {
		return ""
	}
	return mysql.Password
}

func passwordOfRedis(redis *Redis) string {
	if redis == nil {
		return ""
	}
	return redis.Password
}
