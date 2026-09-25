# Fix Character Limit Logic: Byte Count → Character (Rune) Count

## Context

All string length validation uses Go's `len(string)` which counts **bytes**, not characters. For CJK text (3 bytes/UTF-8 char), a "512 byte" limit only allows ~170 Chinese characters. But the user wants "512" to mean 512 **Chinese characters** (runes).

MySQL's `VARCHAR(N)` with `utf8mb4` charset already counts by characters, so no DB migration needed. Only the Go application layer needs fixing.

Additionally, several domain VO limits are inconsistent with their DB column sizes — fix those at the same time.

## Approach

1. Create `infrastructure/utils/strings.go` with a `RuneLength(s string) int` function wrapping `utf8.RuneCountInString()`
2. Replace all `len(s)` text-length checks with `utils.RuneLength(s)` across all domain VOs, entities, and controllers
3. Fix limit mismatches between domain VOs and DB column sizes
4. Error messages updated to match the corrected limits

## Files to Create

- `infrastructure/utils/strings.go` — `RuneLength()` utility

## Files to Modify

### Domain entities
- `domain/user/entities/user.go` — `IsNicknameValid()`: `len` → `RuneLength`

### Domain value objects (14 files)
- `domain/auth/valueobjects/createUser.go` — Nickname check: `len` → `RuneLength`
- `domain/comment/valueobjects/createComment.go` — Content: `len` → `RuneLength`
- `domain/comment/valueobjects/updateComment.go` — Content: `len` → `RuneLength`
- `domain/event/valueobjects/createEvent.go` — Name + Description: `len` → `RuneLength`; Description limit 256→512
- `domain/event/valueobjects/updateEvent.go` — Name + Description: `len` → `RuneLength`; Description limit 256→512
- `domain/event/valueobjects/batchUpdateEvent.go` — Name + Description: `len` → `RuneLength`; Description limit 256→512
- `domain/project/valueobjects/createProject.go` — Name + Description: `len` → `RuneLength`; Description limit 256→512
- `domain/project/valueobjects/updateProject.go` — Name + Description: `len` → `RuneLength`; Description limit 256→512
- `domain/project/valueobjects/batchUpdateProject.go` — Name + Description: `len` → `RuneLength`; Description limit 256→512
- `domain/tag/valueobjects/createTag.go` — Name + Description + Color: `len` → `RuneLength`; Name limit 128→64; Description limit 256→512
- `domain/tag/valueobjects/updateTag.go` — Name + Description + Color: `len` → `RuneLength`; Name limit 128→64; Description limit 256→512
- `domain/tag/valueobjects/batchUpdateTag.go` — Name + Description + Color: `len` → `RuneLength`; Name limit 128→64; Description limit 256→512
- `domain/task/valueobjects/createTask.go` — Description: `len` → `RuneLength`; Desc limit 256→512; Add Name max 256 check
- `domain/task/valueobjects/updateTask.go` — Name + Description: `len` → `RuneLength` (limits already match DB)

### Controllers
- `interfaces/controllers/user.go` — Nickname + Password: `len` → `RuneLength`; Nickname limit 3-32 → 2-20

## Limit Corrections Summary

| Field | Old Limit | New Limit | DB | Reason |
|-------|----------|-----------|-----|--------|
| Event.Description | 256 (code, but err msg said 512) | 512 | 512 | Fix code to match error message + DB |
| Project.Description | 256 | 512 | 512 | Match DB |
| Tag.Description | 256 | 512 | 512 | Match DB |
| Tag.Name | 128 | 64 | 64 | Match DB |
| Task.Description (create) | 256 | 512 | 512 | Match DB + Update path |
| Task.Name (create) | none | 256 | 256 | Match Update path + DB |
| User.Nickname | 32 (VO), 3-32 (controller), 2-20 (entity) | 2-20 | 64 | Standardize on entity's existing limit |

## NOT Changed

- `len()` used on slices/arrays (attachments count, batch items count, etc.) — these are already correct
- DB columns (`gorm:"size:N"`) — `VARCHAR(N)` with `utf8mb4` already stores N characters, not bytes
- `CommentUser.Avatar` and `CommentUser.Nickname` GORM models — these have no `size` tag (become `longtext`), no issue

## Verification

1. `go build ./...` — must compile
2. `go vet ./...` — must pass
3. `go test ./...` — must pass (no existing tests, but verify no compilation errors)
4. Manual sanity check: a 512-Chinese-character string (1536 bytes) should pass validation, while 513 Chinese characters should be rejected
