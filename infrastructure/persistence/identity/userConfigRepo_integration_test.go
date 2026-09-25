//go:build integration

// TASK-26 / T130 服务端契约集成：user_configs.preferences（JSON blob，哑存储）
// + 服务端权威 updated_at（LWW 判据）。
//
// 覆盖 qa 用例：
//   - CT-01 设置面回传：PUT /user/config 接受全量快照 preferences（推送时装配）
//   - CT-03 GET /user/config 出参含 updatedAt（服务端权威时间）
//   - CT-05 preferences 列 nullable + versioned + 服务端不解析（哑存储）
//   - CT-06 appearance 独立不动（三态：仅 appearance / 仅 preferences / 两者）
//   - CT-08 内建偏好不复用业务实体（不写 project_preferences）
package identity

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"naotodoserver/domain/identity/valueobjects"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/models"
)

// blobSample 全量偏好快照（客户端推送时装配的形态）
const blobSample = `{"version":1,"builtInProjectPreferences":{"all":{"viewType":"kanban"}},` +
	`"asideWidth":280,"calendar":{"weekStart":1,"pomodoroBadge":true,"dayZoom":1.5}}`

func strPtr(s string) *string { return &s }

func cleanUserConfigs(t *testing.T) {
	t.Helper()
	for _, stmt := range []string{
		// 仅清理本测试使用的用户区间，避免与其它包的集成测试互删数据
		"DELETE FROM project_preferences WHERE user_id BETWEEN 2900 AND 2999",
		"DELETE FROM user_configs WHERE user_id BETWEEN 2900 AND 2999",
	} {
		if err := testDB.Exec(stmt).Error; err != nil {
			t.Fatalf("清理失败 %s: %v", stmt, err)
		}
	}
}

// assertJSONEqual 语义比较两段 JSON（MySQL JSON 列会规范化键序/空白）
func assertJSONEqual(t *testing.T, got, want string) {
	t.Helper()
	var g, w any
	if err := json.Unmarshal([]byte(got), &g); err != nil {
		t.Fatalf("解析 got JSON 失败: %v (%s)", err, got)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("解析 want JSON 失败: %v (%s)", err, want)
	}
	if !reflect.DeepEqual(g, w) {
		t.Fatalf("JSON 语义不一致:\n got=%s\nwant=%s", got, want)
	}
}

// TestUserConfig_PreferencesDumbStorageRoundTrip CT-01 / CT-05：
// 全量快照原样落库并回读；服务端不解析内容；未携带 appearance 时保留默认值。
func TestUserConfig_PreferencesDumbStorageRoundTrip(t *testing.T) {
	cleanUserConfigs(t)
	repo := NewUserRepo(testDB, testCache)
	ctx := context.Background()
	const userId = 2901

	if err := repo.UpdateConfig(ctx, types.UserID(userId), valueobjects.UpdateUserConfig{
		Preferences: strPtr(blobSample),
	}); err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}

	config, err := repo.GetConfig(ctx, types.UserID(userId))
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if config.Appearance != "auto" {
		t.Fatalf("未携带 appearance 时不应覆盖默认值，got %q", config.Appearance)
	}
	assertJSONEqual(t, config.Preferences, blobSample)
	if config.UpdatedAt.IsZero() {
		t.Fatal("updated_at 应由服务端写入")
	}

	// 直接回读模型：确认 JSON 列真实落库（而非仅在缓存中）
	var model models.UserConfig
	if err := testDB.Where("user_id = ?", userId).First(&model).Error; err != nil {
		t.Fatalf("回读模型: %v", err)
	}
	if model.Preferences == nil {
		t.Fatal("preferences 列应已写入")
	}
	assertJSONEqual(t, *model.Preferences, blobSample)
}

// TestUserConfig_PreferencesNullableByDefault CT-05：未设置偏好时列可空（NULL），实体为空串。
func TestUserConfig_PreferencesNullableByDefault(t *testing.T) {
	cleanUserConfigs(t)
	repo := NewUserRepo(testDB, testCache)
	ctx := context.Background()
	const userId = 2902

	if err := repo.UpdateConfig(ctx, types.UserID(userId), valueobjects.UpdateUserConfig{
		Appearance: strPtr("dark"),
	}); err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}

	config, err := repo.GetConfig(ctx, types.UserID(userId))
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if config.Preferences != "" {
		t.Fatalf("未设置偏好时实体应为空串，got %q", config.Preferences)
	}

	var isNull bool
	if err := testDB.
		Raw("SELECT preferences IS NULL FROM user_configs WHERE user_id = ?", userId).
		Scan(&isNull).Error; err != nil {
		t.Fatalf("查询列可空性: %v", err)
	}
	if !isNull {
		t.Fatal("未设置偏好时 preferences 列应为 NULL")
	}
}

// TestUserConfig_UpdatedAtIsServerOwnedAndMonotonic CT-03 / PS-8：
// updated_at 为服务端时间且更新后严格递增（LWW 判据）。
func TestUserConfig_UpdatedAtIsServerOwnedAndMonotonic(t *testing.T) {
	cleanUserConfigs(t)
	repo := NewUserRepo(testDB, testCache)
	ctx := context.Background()
	const userId = 2903

	lowerBound := time.Now().Add(-1 * time.Second)
	if err := repo.UpdateConfig(ctx, types.UserID(userId), valueobjects.UpdateUserConfig{
		Appearance: strPtr("dark"),
	}); err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}
	first, err := repo.GetConfig(ctx, types.UserID(userId))
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if first.UpdatedAt.Before(lowerBound) || first.UpdatedAt.After(time.Now().Add(time.Second)) {
		t.Fatalf("updated_at 应为服务端当前时间，got %v", first.UpdatedAt)
	}

	time.Sleep(5 * time.Millisecond)
	if err := repo.UpdateConfig(ctx, types.UserID(userId), valueobjects.UpdateUserConfig{
		Preferences: strPtr(blobSample),
	}); err != nil {
		t.Fatalf("UpdateConfig(二次): %v", err)
	}
	second, err := repo.GetConfig(ctx, types.UserID(userId))
	if err != nil {
		t.Fatalf("GetConfig(二次): %v", err)
	}
	if !second.UpdatedAt.After(first.UpdatedAt) {
		t.Fatalf("二次更新后 updated_at 应严格递增: %v -> %v", first.UpdatedAt, second.UpdatedAt)
	}
}

// TestUserConfig_AppearanceOnlyUpdatePreservesPreferences CT-06：仅改 appearance 不动 preferences。
func TestUserConfig_AppearanceOnlyUpdatePreservesPreferences(t *testing.T) {
	cleanUserConfigs(t)
	repo := NewUserRepo(testDB, testCache)
	ctx := context.Background()
	const userId = 2904

	if err := repo.UpdateConfig(ctx, types.UserID(userId), valueobjects.UpdateUserConfig{
		Preferences: strPtr(blobSample),
	}); err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}
	if err := repo.UpdateConfig(ctx, types.UserID(userId), valueobjects.UpdateUserConfig{
		Appearance: strPtr("dark"),
	}); err != nil {
		t.Fatalf("UpdateConfig(appearance): %v", err)
	}

	config, err := repo.GetConfig(ctx, types.UserID(userId))
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if config.Appearance != "dark" {
		t.Fatalf("appearance 未更新，got %q", config.Appearance)
	}
	assertJSONEqual(t, config.Preferences, blobSample)
}

// TestUserConfig_PreferencesOnlyUpdatePreservesAppearance CT-06：仅改 preferences 不动 appearance。
func TestUserConfig_PreferencesOnlyUpdatePreservesAppearance(t *testing.T) {
	cleanUserConfigs(t)
	repo := NewUserRepo(testDB, testCache)
	ctx := context.Background()
	const userId = 2905

	if err := repo.UpdateConfig(ctx, types.UserID(userId), valueobjects.UpdateUserConfig{
		Appearance: strPtr("light"),
	}); err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}
	if err := repo.UpdateConfig(ctx, types.UserID(userId), valueobjects.UpdateUserConfig{
		Preferences: strPtr(blobSample),
	}); err != nil {
		t.Fatalf("UpdateConfig(preferences): %v", err)
	}

	config, err := repo.GetConfig(ctx, types.UserID(userId))
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if config.Appearance != "light" {
		t.Fatalf("appearance 不应被 preferences 更新覆盖，got %q", config.Appearance)
	}
	assertJSONEqual(t, config.Preferences, blobSample)
}

// TestUserConfig_ExplicitEmptyPreferencesClears CT-05：空串视为清除（写回 NULL）。
func TestUserConfig_ExplicitEmptyPreferencesClears(t *testing.T) {
	cleanUserConfigs(t)
	repo := NewUserRepo(testDB, testCache)
	ctx := context.Background()
	const userId = 2906

	if err := repo.UpdateConfig(ctx, types.UserID(userId), valueobjects.UpdateUserConfig{
		Preferences: strPtr(blobSample),
	}); err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}
	if err := repo.UpdateConfig(ctx, types.UserID(userId), valueobjects.UpdateUserConfig{
		Preferences: strPtr(""),
	}); err != nil {
		t.Fatalf("UpdateConfig(清空): %v", err)
	}

	config, err := repo.GetConfig(ctx, types.UserID(userId))
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if config.Preferences != "" {
		t.Fatalf("清空后实体应为空串，got %q", config.Preferences)
	}

	var isNull bool
	if err := testDB.
		Raw("SELECT preferences IS NULL FROM user_configs WHERE user_id = ?", userId).
		Scan(&isNull).Error; err != nil {
		t.Fatalf("查询列可空性: %v", err)
	}
	if !isNull {
		t.Fatal("清空后 preferences 列应为 NULL")
	}
}

// TestUserConfig_BuiltInPreferencesDoNotTouchProjectPreferences CT-08 / PS-4（负向）：
// 内建清单偏好走 UserConfig.preferences，不得伪造 project_preferences 行。
func TestUserConfig_BuiltInPreferencesDoNotTouchProjectPreferences(t *testing.T) {
	cleanUserConfigs(t)
	repo := NewUserRepo(testDB, testCache)
	ctx := context.Background()
	const userId = 2907

	builtInBlob := `{"version":1,"builtInProjectPreferences":` +
		`{"all":{"viewType":"kanban"},"today":{"viewType":"list"}}}`
	if err := repo.UpdateConfig(ctx, types.UserID(userId), valueobjects.UpdateUserConfig{
		Preferences: strPtr(builtInBlob),
	}); err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}

	var count int64
	if err := testDB.
		Model(&models.ProjectPreference{}).
		Where("user_id = ?", userId).
		Count(&count).Error; err != nil {
		t.Fatalf("统计 project_preferences: %v", err)
	}
	if count != 0 {
		t.Fatalf("内建偏好不得写入 project_preferences，got %d 行", count)
	}
}
