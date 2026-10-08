#!/usr/bin/env bash
# =============================================================================
# T507 · 用例先行 + 独立验证脚本
#   nao-todo-server 迁移到 nao-skill 0.12.0（pi 原生包形态 / shim 转发）
#
# 断言来源（正文权威）：docs/prds/2026-10-08-nao-fleet-0.12.0-migration.md
#   §7 AC1–AC8 · §8 V1–V5 · §5 BR1–BR5 · §6 NFR1–NFR4 · §2 目标指标 · §3 范围
#
# AC3「引用面」口径（PRD §7 AC3 + §8 V5，PM 2026-10-08 裁定）：
#   统计域 = **入库文件**（`git grep`）；排除本批迁移文档自身；排除 `.pi/npm/**`
#   （本机物化产物，gitignored、fresh clone 不存在，天然不在 git 索引中）；
#   `.gitignore` 的忽略规则行不计入（其存在由 AC4 断言）。判据 = 0 → 0。
#
# 用法
#   bash docs/reports/2026-10-08-T507-nao-fleet-migration-verify.sh            # 迁移后验收（默认）
#   bash docs/reports/2026-10-08-T507-nao-fleet-migration-verify.sh --baseline # 迁移前基线（只读）
#   bash docs/reports/2026-10-08-T507-nao-fleet-migration-verify.sh [项目根] [--baseline]
#
# 环境开关（仅迁移后模式生效）
#   T507_SKIP_GATES=1        跳过 go 全门禁（批末统一跑时用）
#   T507_REAL_DEGRADED=1     额外在真树做「临时移走 .pi/npm」负向（须独占窗口；自动恢复）
#   T507_ROLLBACK=1          额外做「git worktree 里 revert 本批提交后 check exit 0」
#
# 纪律：默认只读被测树。负向用 temp fixture；回滚用 git worktree（不动主工作区）。
#       迁移前跑默认模式必然大面积 FAIL（RED = 用例先行的预期）。
# =============================================================================
set -uo pipefail

BASELINE=0
PROJ=""
for a in "$@"; do
  case "$a" in
    --baseline) BASELINE=1 ;;
    *) PROJ="$a" ;;
  esac
done
[ -n "$PROJ" ] || PROJ="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
PROJ="$(cd "$PROJ" && pwd)"
cd "$PROJ"

# --- 迁移前基线常量（本次实测固化；AC3/AC4 用 md5 判「逐字节不变」） ---
MD5_AGENTS_MD="587f47ce7cfc7be8e64e350a3b5d1513"        # AGENTS.md
MD5_CLAUDE_MD="1bab8f84e6274270b378ee457d4d24a8"        # CLAUDE.md
MD5_APPEND_SYSTEM="8a7082f6cccf32abcfb5e1e851e1a0e7"    # .pi/APPEND_SYSTEM.md
MD5_SKILLS_LOCK="73a5ce33756be75fa19984d0832c1138"      # skills-lock.json（AC2 原样）
BASE_NAO_VERSION="0.11.0"
TARGET_NAO_VERSION="0.12.0"
PIN_SPEC="npm:@nathan33/nao-skill@0.12.0"
# 迁移批次自身产出的文档（含机制字样，属 AC3「引用面」统计的排除项）
REF_EXCLUDE_RE='docs/prds/2026-10-08-nao-fleet-0.12.0-migration\.md|docs/reports/2026-10-08-T507'
# git pathspec 版排除项（AC3 口径）：.agents/ 本体 / .gitignore 规则行 / 本批迁移文档
REF_EXCLUDES=( ':(exclude).agents/' ':(exclude).gitignore'
  ':(exclude)docs/prds/2026-10-08-nao-fleet-0.12.0-migration.md' ':(exclude)docs/reports/' )
# 迁移前基线提交（用于 AC3 的 0 → 0 对照）
BASE_REV="$(git merge-base origin/main HEAD 2>/dev/null || git rev-parse origin/main 2>/dev/null || git rev-parse HEAD~2 2>/dev/null || echo '')"
# AC8 非范围路径（本批不得触碰）
NON_SCOPE_RE='^(application|domain|infrastructure|interfaces|cmd|scripts)/|^go\.(mod|sum)$|^\.github/workflows/ci\.yml$'

TMP="$(mktemp -d "${TMPDIR:-/tmp}/t507-verify.XXXXXX")"
EMPTY_HOME="$TMP/emptyhome"; mkdir -p "$EMPTY_HOME"
restore_npm() { [ -d "$PROJ/.pi/npm.t507bak" ] && mv "$PROJ/.pi/npm.t507bak" "$PROJ/.pi/npm" 2>/dev/null || true; }
cleanup(){ restore_npm; rm -rf "$TMP"; }
trap cleanup EXIT

PASS=0; FAIL=0; WARN=0
G=$'\033[32m'; R=$'\033[31m'; Y=$'\033[33m'; D=$'\033[2m'; N=$'\033[0m'
pass(){ PASS=$((PASS+1)); printf '%s[PASS]%s %s\n' "$G" "$N" "$*"; }
fail(){ FAIL=$((FAIL+1)); printf '%s[FAIL]%s %s\n' "$R" "$N" "$*"; }
warn(){ WARN=$((WARN+1)); printf '%s[WARN]%s %s\n' "$Y" "$N" "$*"; }
info(){ printf '%s[INFO]%s %s\n' "$D" "$N" "$*"; }
hdr(){ printf '\n%s===== %s =====%s\n' "$D" "$*" "$N"; }

assert_eq(){ [ "$1" = "$2" ] && pass "$3 (= $2)" || fail "$3 (expected [$1] got [$2])"; }
assert_ne(){ [ "$1" != "$2" ] && pass "$3 (= $2)" || fail "$3 (unexpectedly = $1)"; }
assert_file(){ [ -f "$1" ] && pass "$2" || fail "$2 (missing file $1)"; }
assert_dir(){ [ -d "$1" ] && pass "$2" || fail "$2 (missing dir $1)"; }
assert_absent(){ [ ! -e "$1" ] && pass "$2" || fail "$2 (still exists: $1)"; }
assert_contains(){ case "$1" in *"$2"*) pass "$3";; *) fail "$3 (output lacks [$2])";; esac; }
assert_not_contains(){ case "$1" in *"$2"*) fail "$3 (output unexpectedly has [$2])";; *) pass "$3";; esac; }
md5(){ md5sum "$1" 2>/dev/null | awk '{print $1}'; }

# 经项目内 shim 调 fleet（清掉可能干扰的显式覆盖；返回 RC/OUT/ERR）
shim(){ local d="$1"; shift
  ( cd "$d" && env -u NAO_SKILLS -u NAO_SHIM_ENTERED -u PI_CODING_AGENT_DIR \
      bash .agents/scripts/nao-fleet.sh "$@" ) >"$TMP/o" 2>"$TMP/e"
  RC=$?; OUT="$(cat "$TMP/o")"; ERR="$(cat "$TMP/e")"
}

# .agents/ 之外对机制的引用数（AC3 口径：入库文件 + 排除项；$1 = 可选修订）
ref_hits(){ git grep -I -e '\.agents' -e 'nao-fleet' -e 'NAO_SKILLS' ${1:+"$1"} -- . \
    "${REF_EXCLUDES[@]}" 2>/dev/null | wc -l | tr -d ' '; }

# 解析 pin 的机制包根（仅用于取包内脚本路径）
PKG_ROOT="$(node -e '
  const fs=require("fs"),path=require("path");
  const proj=process.argv[1];
  let s; try{ s=JSON.parse(fs.readFileSync(path.join(proj,".pi/settings.json"),"utf8")); }catch{ process.exit(0); }
  for(const raw of (s.packages||[])){
    const m=/^npm:(.+?)(?:@([^@\/]+))?$/.exec(String(raw)); if(!m) continue;
    let name=m[1]; if(name.startsWith("@")&&name.includes("/@")) name=name.slice(0,name.indexOf("/@"));
    try{
      const pj=require.resolve(name+"/package.json",{paths:[path.join(proj,".pi/npm/node_modules")]});
      const root=path.dirname(pj);
      if(fs.existsSync(path.join(root,".agents","scripts","nao-fleet.sh"))){ process.stdout.write(root); break; }
    }catch{}
  }
' "$PROJ" 2>/dev/null)"

printf 'T507 独立验证 · 项目根 = %s · 模式 = %s\n' "$PROJ" "$( [ "$BASELINE" = 1 ] && echo baseline || echo post-migration )"

# =============================================================================
if [ "$BASELINE" = 1 ]; then
hdr "V1/V2 + 基线：迁移前只读体检（现树应停在 0.11.0 态）"

assert_eq "$BASE_NAO_VERSION" "$(tr -d '[:space:]' < "$PROJ/.agents/.nao-version" 2>/dev/null)" "基线: .agents/.nao-version = $BASE_NAO_VERSION"
assert_absent "$PROJ/.agents/.nao-migrated" "基线: .agents/.nao-migrated 尚不存在（未迁移）"

top="$(ls -A "$PROJ/.agents" 2>/dev/null | sort | tr '\n' ' ')"
info "基线: .agents/ 顶层实测 [$top]（PRD §1/§2 写「7 项」，实测 $(ls -A .agents | wc -l | tr -d ' ') 项）"
for e in prompts common checklists templates roles.yaml scripts skills; do
  [ -e "$PROJ/.agents/$e" ] && pass "基线: 0.11.0 机制项在树 .agents/$e" || fail "基线: 机制项缺失 .agents/$e"
done
n_files="$(find "$PROJ/.agents" -type f | wc -l | tr -d ' ')"
n_git="$(git ls-files .agents | wc -l | tr -d ' ')"
assert_eq "$n_files" "$n_git" "基线: .agents/ 文件数 = git 入库数（$n_files）"
info "基线: .agents/ 文件计数 = $n_files"

assert_eq "0" "$(ref_hits "$BASE_REV")" "基线/AC3: 基线提交 0.11.0 态引用面 0 命中（AC3 口径：入库文件 + 排除项）"
assert_eq "$MD5_AGENTS_MD" "$(md5 "$PROJ/AGENTS.md")" "基线/AC3: AGENTS.md md5 基线"
assert_eq "$MD5_CLAUDE_MD" "$(md5 "$PROJ/CLAUDE.md")" "基线/AC3: CLAUDE.md md5 基线"
assert_eq "$MD5_APPEND_SYSTEM" "$(md5 "$PROJ/.pi/APPEND_SYSTEM.md")" "基线/V3: .pi/APPEND_SYSTEM.md md5 基线"

info "基线/AC4: .gitignore 中 .pi 行 = $(grep -n '^\.pi$' "$PROJ/.gitignore" | head -1)"
assert_eq ".pi" "$(sed -n '68p' "$PROJ/.gitignore" | tr -d '\r')" "基线/AC4: .gitignore 第 68 行原文 = 裸 .pi"
assert_eq "0" "$(git ls-files .pi | wc -l | tr -d ' ')" "基线/AC4: git ls-files .pi = 0（.pi 全忽略）"
assert_absent "$PROJ/.pi/settings.json" "基线/AC4: .pi/settings.json 尚不存在"
if git check-ignore -q ".pi/APPEND_SYSTEM.md"; then pass "基线/V3: .pi/APPEND_SYSTEM.md 现被忽略"; else fail "基线/V3: .pi/APPEND_SYSTEM.md 未被忽略"; fi

lock_keys="$(node -e 'const s=require(process.argv[1]);process.stdout.write(Object.keys(s.skills||{}).sort().join(","))' "$PROJ/skills-lock.json" 2>/dev/null)"
assert_eq "agent-browser,find-skills,skill-creator" "$lock_keys" "基线/AC2: skills-lock 键 = agent-browser,find-skills,skill-creator"
assert_not_contains "$lock_keys" "frontend-design" "基线/AC2: skills-lock 不含 frontend-design（无需去重）"
assert_dir "$PROJ/.agents/skills/frontend-design" "基线/AC2: nao 资产 .agents/skills/frontend-design 现状存在（待迁移备份）"
own_skill="$(ls -A "$PROJ/.agents/skills" 2>/dev/null | grep -vE '\.md$|^frontend-design$' | tr '\n' ' ' | sed 's/ *$//')"
info "基线/AC2: .agents/skills/ 内非 nao 资产条目 = [${own_skill:-<无>}]（预期无自有 skill ⇒ 迁移后该目录可为空）"

hdr "V1 本机 golangci-lint 可用性"
if command -v golangci-lint >/dev/null 2>&1; then
  glv="$(golangci-lint --version 2>&1 | head -1)"
  pass "V1: golangci-lint 可用 → $glv"
  case "$glv" in *2.14.0*) pass "V1: 版本命中 v2.14.0（AC5 可按本机 rc=0 记）";; *) warn "V1: 版本非 v2.14.0（AC5 以 CI 为准复核）";; esac
else
  warn "V1: 本机无 golangci-lint（AC5 该项如实申报 + 以 CI 为准，不得静默略过）"
fi

hdr "V2 go 门禁可跑性（build / vet / gofmt / test）"
timeout 600 go build ./... >"$TMP/b.log" 2>&1; assert_eq "0" "$?" "V2: go build ./... exit=0"
timeout 600 go vet ./...   >"$TMP/v.log" 2>&1; assert_eq "0" "$?" "V2: go vet ./... exit=0"
assert_eq "" "$(gofmt -l . 2>/dev/null)" "V2: gofmt -l . 输出为空"
timeout 900 go test ./... -count=1 >"$TMP/t.log" 2>&1; rc=$?
assert_eq "0" "$rc" "V2: go test ./... -count=1 exit=0（无 MySQL/Redis 依赖）"
info "V2: go test 结果分布 = $(awk '{print $1}' "$TMP/t.log" | sort | uniq -c | tr '\n' ' ')"
if [ "$BASELINE" = 1 ]; then hdr "汇总（baseline）"; printf 'PASS=%d  FAIL=%d  WARN=%d\n' "$PASS" "$FAIL" "$WARN"; [ "$FAIL" -eq 0 ] && { printf '%sBASELINE GREEN%s\n' "$G" "$N"; exit 0; } || { printf '%sFAIL=%d%s\n' "$R" "$FAIL" "$N"; exit 1; }; fi
fi

# =============================================================================
hdr "前置：迁移态与物化"
assert_file "$PROJ/.agents/.nao-migrated" "前置: .agents/.nao-migrated 存在（迁移已执行）"
assert_file "$PROJ/.pi/settings.json" "前置: .pi/settings.json 存在（pin 入库）"
if [ -n "$PKG_ROOT" ]; then
  pass "前置: 机制包已物化（$PKG_ROOT）"
  assert_eq "$TARGET_NAO_VERSION" "$(node -p 'require(process.argv[1]).version' "$PKG_ROOT/package.json" 2>/dev/null)" "前置: 机制包版本 $TARGET_NAO_VERSION"
else
  fail "前置: 机制包未物化（.pi/npm 缺 @nathan33/nao-skill）"
fi

# ---------------------------------------------------------------------------
hdr "AC1 主路径：足迹收敛 + shim check"
for d in prompts common checklists templates; do
  assert_absent "$PROJ/.agents/$d" "AC1: .agents/$d 已移除（机制副本不落项目）"
done
assert_absent "$PROJ/.agents/roles.yaml" "AC1: .agents/roles.yaml 已移除"
for s in intercom-probe.mts qq-notify ui-tokens-check.sh; do
  assert_absent "$PROJ/.agents/scripts/$s" "AC1: 机制脚本 .agents/scripts/$s 已移除"
done
assert_eq "nao-fleet.sh" "$(ls -A "$PROJ/.agents/scripts" 2>/dev/null | tr '\n' ' ' | sed 's/ *$//')" "AC1: .agents/scripts/ 仅 shim nao-fleet.sh"
assert_contains "$(head -40 "$PROJ/.agents/scripts/nao-fleet.sh" 2>/dev/null)" "NAO_SHIM_ENTERED" "AC1: 项目内 nao-fleet.sh 是 shim（含 NAO_SHIM_ENTERED）"
assert_not_contains "$(head -80 "$PROJ/.agents/scripts/nao-fleet.sh" 2>/dev/null)" "cmd_ensure" "AC1: shim 非旧版全量脚本（前 80 行无 cmd_ensure）"
assert_eq "$TARGET_NAO_VERSION" "$(tr -d '[:space:]' < "$PROJ/.agents/.nao-version" 2>/dev/null)" "AC1: .agents/.nao-version = $TARGET_NAO_VERSION"
assert_eq "$TARGET_NAO_VERSION" "$(tr -d '[:space:]' < "$PROJ/.agents/.nao-migrated" 2>/dev/null)" "AC1: .agents/.nao-migrated = $TARGET_NAO_VERSION"
assert_dir "$PROJ/.agents/.nao-obsolete" "AC1: .agents/.nao-obsolete 存在（备份）"

# .nao-obsolete/<stamp>/ 备份 stamp
stamp_dir="$(ls -d "$PROJ/.agents/.nao-obsolete"/*/ 2>/dev/null | head -1)"
[ -n "$stamp_dir" ] && pass "AC1/AC2: 备份 stamp 存在（${stamp_dir#$PROJ/}）" || fail "AC1/AC2: 未找到 .nao-obsolete/<stamp>/"

# 顶层残留项 = 4 机制项 + skills/（本仓无自有 skill ⇒ skills/ 为空，不硬记项数）
top="$(ls -A "$PROJ/.agents" 2>/dev/null | sort | tr '\n' ' ' | sed 's/ *$//')"
info "AC1: .agents/ 顶层实测 = [$top]（计数 $(ls -A "$PROJ/.agents" | wc -l | tr -d ' ')）"
assert_eq ".nao-migrated .nao-obsolete .nao-version scripts skills" "$top" "AC1: 顶层 = 4 机制项 + skills/（空）"
for e in .nao-migrated .nao-obsolete .nao-version scripts; do
  [ -e "$PROJ/.agents/$e" ] && pass "AC1: 顶层机制项存在 $e" || fail "AC1: 顶层机制项缺失 $e"
done
if [ -d "$PROJ/.agents/skills" ]; then
  sk="$(ls -A "$PROJ/.agents/skills" 2>/dev/null | tr '\n' ' ' | sed 's/ *$//')"
  assert_eq "" "$sk" "AC2: .agents/skills/ 为空（本仓无自有 skill，不视为缺陷）"
else
  pass "AC2: .agents/skills/ 已移除（本仓无自有 skill，不视为缺陷）"
fi

# .agents/ 入库文件集（本仓实测 3；fresh clone 下空 skills/ 不入库）
tracked="$(git ls-files .agents | sort | tr '\n' ' ' | sed 's/ *$//')"
info "AC1: .agents/ 入库文件 = [$tracked]（计数 $(git ls-files .agents | wc -l | tr -d ' ')）"
assert_eq ".agents/.nao-migrated .agents/.nao-version .agents/scripts/nao-fleet.sh" "$tracked" "AC1/BR1: .agents/ 入库仅 3 文件（迁移标记 + 版本 + shim）"

shim "$PROJ" check
assert_eq "0" "$RC" "AC1: 经 shim 的 fleet check exit=0"
assert_contains "$OUT$ERR" "check: OK" "AC1: check 输出含 'check: OK'"
assert_contains "$OUT$ERR" "roles=6" "AC1: check 报 roles=6（BR2）"
assert_not_contains "$OUT$ERR" "DEGRADED" "AC1: 正常路径无 DEGRADED"

# ---------------------------------------------------------------------------
hdr "AC2 边界：nao 资产移除且已备份 / skills-lock 原样"
assert_absent "$PROJ/.agents/skills/frontend-design" "AC2: nao 资产 .agents/skills/frontend-design 已从 live 移除"
fd_bak="$(find "$PROJ/.agents/.nao-obsolete" -path '*frontend-design*' 2>/dev/null | head -1)"
[ -n "$fd_bak" ] && pass "AC2: 备份含 frontend-design（${fd_bak#$PROJ/}）" || fail "AC2: 备份未含 frontend-design"
mech_bak="$(find "$PROJ/.agents/.nao-obsolete" \( -name 'roles.yaml' -o -name 'product-manager.md' -o -name 'qa.md' -o -name 'nao-fleet.sh' \) 2>/dev/null | head -1)"
[ -n "$mech_bak" ] && pass "AC2: 备份含旧机制资产（${mech_bak#$PROJ/}）" || fail "AC2: 备份未含旧机制资产"
ob_n="$(find "$PROJ/.agents/.nao-obsolete" -type f 2>/dev/null | wc -l | tr -d ' ')"
info "AC2: .nao-obsolete 备份文件数 = $ob_n（旧树 43 → 期望 >= 5）"
[ "$ob_n" -ge 5 ] && pass "AC2: 备份文件数 >= 5" || fail "AC2: 备份文件数不足（$ob_n）"
lock_keys="$(node -e 'const s=require(process.argv[1]);process.stdout.write(Object.keys(s.skills||{}).sort().join(","))' "$PROJ/skills-lock.json" 2>/dev/null)"
assert_eq "agent-browser,find-skills,skill-creator" "$lock_keys" "AC2: skills-lock 键保持原样（无需去重）"
assert_eq "$MD5_SKILLS_LOCK" "$(md5 "$PROJ/skills-lock.json")" "AC2: skills-lock.json 逐字节不变（md5 基线）"
assert_absent "$PROJ/.agents/skills/frontend-design" "AC2: frontend-design 未残留 live 树"
if git check-ignore -q ".agents/.nao-obsolete/"; then pass "AC2/BR4: .agents/.nao-obsolete/ 被 git 忽略（不入库）"; else fail "AC2/BR4: .agents/.nao-obsolete/ 未被 git 忽略"; fi

# ---------------------------------------------------------------------------
hdr "AC3 设计一致性：引用面 0 不变 + 守则逐字节不变"
assert_eq "0" "$(ref_hits)" "AC3: 入库文件机制引用 0 命中（AC3 口径：git grep + 排除 .agents//.gitignore/迁移文档）"
if [ -n "$BASE_REV" ]; then
  base_hits="$(ref_hits "$BASE_REV")"
  assert_eq "0" "$base_hits" "AC3: 基线提交（0.11.0 态）引用面 0（对照）"
  assert_eq "$base_hits" "$(ref_hits)" "AC3: 引用面 0 → 0（迁移未产生机制路径改写）"
fi
assert_eq "$MD5_AGENTS_MD" "$(md5 "$PROJ/AGENTS.md")" "AC3: AGENTS.md 逐字节不变（md5 基线）"
assert_eq "$MD5_CLAUDE_MD" "$(md5 "$PROJ/CLAUDE.md")" "AC3: CLAUDE.md 逐字节不变（md5 基线）"

# ---------------------------------------------------------------------------
hdr "AC4 配置：.gitignore 放行 pin / settings 内容 / APPEND_SYSTEM 未动"
if git check-ignore -q ".pi/settings.json"; then fail "AC4: .pi/settings.json 仍被忽略（应为 !.pi/settings.json 放行）"; else pass "AC4: .pi/settings.json 未被忽略（可入库）"; fi
if git check-ignore -q ".pi/npm/node_modules/@nathan33/nao-skill/package.json"; then pass "AC4: .pi/npm/** 仍被忽略"; else fail "AC4: .pi/npm/** 未被忽略"; fi
if git check-ignore -q ".pi/APPEND_SYSTEM.md"; then pass "AC4/V3: .pi/APPEND_SYSTEM.md 仍被忽略"; else fail "AC4/V3: .pi/APPEND_SYSTEM.md 未被忽略"; fi
assert_eq ".pi/*" "$(grep -n '^\.pi/\*$' "$PROJ/.gitignore" | head -1 | cut -d: -f2)" "AC4: .gitignore 出现 .pi/* 规则"
assert_eq "!.pi/settings.json" "$(grep -n '^!\.pi/settings\.json$' "$PROJ/.gitignore" | head -1 | cut -d: -f2)" "AC4: .gitignore 出现 !.pi/settings.json 放行"
grep -q '^\.agents/\.nao-obsolete/$' "$PROJ/.gitignore" && pass "AC4: .gitignore 新增 .agents/.nao-obsolete/ 忽略" || fail "AC4: .gitignore 缺 .agents/.nao-obsolete/ 忽略"
assert_eq "" "$(grep -n '^\.pi$' "$PROJ/.gitignore" | head -1)" "AC4: 裸 .pi 规则行已被替换（无残留）"
assert_eq ".pi/settings.json" "$(git ls-files .pi | tr '\n' ' ' | sed 's/ *$//')" "AC4: git ls-files .pi = 恰好 + .pi/settings.json（基线 0）"
if git ls-files --error-unmatch .pi/settings.json >/dev/null 2>&1; then pass "AC4: .pi/settings.json 已入库（团队 clone 即知 pin）"; else warn "AC4: .pi/settings.json 尚未入库（若验证时未提交，请复核）"; fi
pin="$(node -e 'const s=require(process.argv[1]);const a=(s.packages||[]).filter(p=>/nao-skill/.test(p));process.stdout.write(a.join(","))' "$PROJ/.pi/settings.json" 2>/dev/null)"
assert_eq "$PIN_SPEC" "$pin" "AC4: settings pin = $PIN_SPEC"
compact="$(tr -d '[:space:]' < "$PROJ/.pi/settings.json" 2>/dev/null)"
if [ "$compact" = '{"packages":["npm:@nathan33/nao-skill@0.12.0"]}' ]; then
  pass "AC4: settings 原文与 AC4 字面一致"
else
  info "AC4: settings 原文为 pi install 生成的多行格式（与 AC4 字面存空白差异，JSON 等价；pin 值相符）"
fi
assert_eq "$MD5_APPEND_SYSTEM" "$(md5 "$PROJ/.pi/APPEND_SYSTEM.md")" "AC4/V3: .pi/APPEND_SYSTEM.md 内容未变（md5 基线）"

# ---------------------------------------------------------------------------
hdr "AC5 门禁（fleet check + go build/vet/test/gofmt/golangci-lint）"
shim "$PROJ" check
assert_eq "0" "$RC" "AC5: fleet check exit=0"
if [ "${T507_SKIP_GATES:-}" = "1" ]; then
  info "AC5: 跳过 go 门禁（T507_SKIP_GATES=1；由批末全仓门禁统一跑）"
else
  timeout 600 go build ./... >"$TMP/b.log" 2>&1; assert_eq "0" "$?" "AC5: go build ./... exit=0"
  timeout 600 go vet ./...   >"$TMP/v.log" 2>&1; assert_eq "0" "$?" "AC5: go vet ./... exit=0"
  timeout 900 go test ./... -count=1 >"$TMP/t.log" 2>&1; assert_eq "0" "$?" "AC5: go test ./... -count=1 exit=0"
  assert_eq "" "$(gofmt -l . 2>/dev/null)" "AC5: gofmt -l . 输出为空"
  if command -v golangci-lint >/dev/null 2>&1; then
    v="$(golangci-lint --version 2>&1 | head -1)"
    if printf '%s' "$v" | grep -q '2\.14\.0'; then
      timeout 900 golangci-lint run >"$TMP/l.log" 2>&1; rc=$?
      assert_eq "0" "$rc" "AC5: golangci-lint run exit=0（本机 v2.14.0）"
      tail -2 "$TMP/l.log" | sed 's/^/[lint] /'
    else
      warn "AC5: 本机 golangci-lint 非 v2.14.0（$v）→ 如实申报，以 CI 为准"
    fi
  else
    warn "AC5: 本机无 golangci-lint → 如实申报（AC5 该项不静默略过），以 CI 为准"
  fi
fi

# ---------------------------------------------------------------------------
hdr "AC6 负向闭环：旧命令经 shim / qq-notify / 缺包降级"
shim "$PROJ" status
assert_eq "0" "$RC" "AC6: shim status exit=0（旧命令仍可用）"
shim "$PROJ" ensure __t507_probe__
assert_ne "0" "$RC" "AC6: shim ensure 对未知角色非 0（证明转发到包内 fleet，未静默）"
assert_contains "$OUT$ERR" "未知角色" "AC6: ensure 报错来自包内 fleet（未知角色）"

if [ -n "$PKG_ROOT" ] && [ -x "$PKG_ROOT/.agents/scripts/qq-notify" ]; then
  pass "AC6: \$NAO_SKILLS/.agents/scripts/qq-notify 存在且可执行"
  ( "$PKG_ROOT/.agents/scripts/qq-notify" --dry-run "T507 qa dry-run" ) >"$TMP/qq.log" 2>&1; rc=$?
  assert_eq "0" "$rc" "AC6: qq-notify --dry-run \"文本\" exit=0（连通性自检）"
  assert_contains "$(cat "$TMP/qq.log")" "dry-run" "AC6: qq-notify dry-run 输出可辨认"
  ( "$PKG_ROOT/.agents/scripts/qq-notify" --dry-run ) >"$TMP/qq2.log" 2>&1; rc2=$?
  info "AC6 口径：裸 \`qq-notify --dry-run\`（无文本）实测 rc=$rc2 —— 契约形式须带文本（PRD 已注明，不作失败）"
else
  fail "AC6: 包内 qq-notify 不存在或不可执行"
fi

# 缺包降级：temp fixture（不动真树）
fx="$TMP/fixture-proj"; mkdir -p "$fx/.agents/scripts" "$fx/.pi" "$fx/home"
cp "$PROJ/.agents/scripts/nao-fleet.sh" "$fx/.agents/scripts/nao-fleet.sh"
printf '{"packages":["%s"]}\n' "$PIN_SPEC" > "$fx/.pi/settings.json"
( cd "$fx" && env -u NAO_SKILLS -u NAO_SHIM_ENTERED -u PI_CODING_AGENT_DIR HOME="$fx/home" \
    bash .agents/scripts/nao-fleet.sh check ) >"$TMP/d.log" 2>&1; rc=$?
joined="$(cat "$TMP/d.log")"
assert_eq "2" "$rc" "AC6/NFR3: 缺包场景 shim exit=2"
assert_eq "1" "$(printf '%s\n' "$joined" | grep -cE '^DEGRADED:' || true)" "AC6/NFR3: 恰一行 DEGRADED:"
assert_contains "$joined" "pi install" "AC6/NFR3: 含可复制恢复命令（pi install）"
assert_not_contains "$joined" "check: OK" "AC6/NFR3: 不静默成功"

if [ "${T507_REAL_DEGRADED:-}" = "1" ]; then
  if [ -d "$PROJ/.pi/npm" ]; then
    mv "$PROJ/.pi/npm" "$PROJ/.pi/npm.t507bak"
    ( cd "$PROJ" && env -u NAO_SKILLS -u NAO_SHIM_ENTERED -u PI_CODING_AGENT_DIR HOME="$EMPTY_HOME" \
        bash .agents/scripts/nao-fleet.sh check ) >"$TMP/rd.log" 2>&1; rc=$?
    restore_npm
    joined="$(cat "$TMP/rd.log")"
    assert_eq "2" "$rc" "AC6/NFR3(真树): 移走 .pi/npm 后 shim exit=2"
    assert_eq "1" "$(printf '%s\n' "$joined" | grep -cE '^DEGRADED:' || true)" "AC6/NFR3(真树): 恰一行 DEGRADED:"
    assert_contains "$joined" "pi install" "AC6/NFR3(真树): 含恢复命令"
    info "AC6/NFR3(真树): .pi/npm 已恢复（$( [ -d "$PROJ/.pi/npm" ] && echo ok || echo MISSING )）"
  else
    warn "AC6/NFR3(真树): .pi/npm 不存在，跳过"
  fi
else
  info "AC6/NFR3(真树): 未启用（T507_REAL_DEGRADED=1 时在独占窗口执行）"
fi

# ---------------------------------------------------------------------------
hdr "AC6 / NFR1 回滚：git worktree 内 revert 本批提交后 fleet check exit=0"
if [ "${T507_ROLLBACK:-}" = "1" ]; then
  base="$(git merge-base origin/main HEAD 2>/dev/null || git rev-parse origin/main 2>/dev/null || echo '')"
  if [ -z "$base" ]; then
    warn "AC6/NFR1: 无法确定 origin/main 基线，跳过回滚实测"
  else
    wt="$TMP/wt-rollback"
    if git worktree add --detach "$wt" HEAD >/dev/null 2>&1; then
      revs="$(git rev-list "$base..HEAD")"   # 逆时序（新→旧）——撤销一批提交须从新到旧
      ok=1
      for c in $revs; do ( cd "$wt" && git revert --no-edit "$c" ) >/dev/null 2>&1 || ok=0; done
      if [ "$ok" = "1" ]; then
        pass "AC6/NFR1: worktree 内逆时序 revert 本批 $(printf '%s' "$revs" | wc -w | tr -d ' ') 个提交成功"
        assert_dir "$wt/.agents/prompts" "AC6/NFR1: revert 后旧机制目录 .agents/prompts 恢复"
        assert_absent "$wt/.agents/.nao-migrated" "AC6/NFR1: revert 后 .nao-migrated 消失（回到 0.11 形态）"
        ( cd "$wt" && env -u NAO_SKILLS -u NAO_SHIM_ENTERED bash .agents/scripts/nao-fleet.sh check ) >"$TMP/rb.log" 2>&1; rc=$?
        assert_eq "0" "$rc" "AC6/NFR1: revert 后 .agents/scripts/nao-fleet.sh check exit=0"
        tail -1 "$TMP/rb.log" | sed 's/^/[rollback check] /'
      else
        fail "AC6/NFR1: worktree 内 revert 失败（冲突或非本批提交）"
      fi
      git worktree remove --force "$wt" >/dev/null 2>&1 || true
    else
      fail "AC6/NFR1: git worktree add 失败"
    fi
  fi
else
  info "AC6/NFR1 回滚：未启用（T507_ROLLBACK=1 时执行）"
fi

# ---------------------------------------------------------------------------
hdr "AC7/AC8 治理与非范围守护（git 口径）"
base="$(git merge-base origin/main HEAD 2>/dev/null || git rev-parse origin/main 2>/dev/null || echo '')"
if [ -n "$base" ]; then
  n="$(git rev-list --count "$base..HEAD" 2>/dev/null || echo '?')"
  info "AC7: 分支提交数 = $n（AC7 的「恰好 1 条」在 squash 合并后的 main 上成立；分支可含 PRD 收录 + 迁移等多条）"
  [ "$n" = "1" ] && pass "AC7: 分支已收敛为 1 条提交" || warn "AC7: 分支为 $n 条提交（squash 后 main 应为 1 条）"
  wipn="$(git log --format=%s "$base..HEAD" 2>/dev/null | grep -ciE '^wip[(:]' || true)"
  info "AC7: 分支 wip() 提交数 = $wipn（PM 2026-10-08 裁定：squash 前预期，仅记录不作失败；squash 由 RD 在验收后执行）"
  mwip="$(git log --format=%s origin/main 2>/dev/null | grep -ciE '^wip[(:]' || true)"
  assert_eq "0" "$mwip" "AC7: origin/main 上无 wip() 提交"
  clean="$(git status --porcelain 2>/dev/null)"
  assert_eq "" "$clean" "AC7: 工作区干净"
  st="$(git diff --name-status "$base..HEAD" 2>/dev/null)"
  hit="$(printf '%s\n' "$st" | awk '{print $2}' | grep -E "$NON_SCOPE_RE" | head -5 | tr '\n' ' ')"
  assert_eq "" "$hit" "AC8: 本批未触碰 Go 源码/构建/go.mod/go.sum/ci.yml"
  ds="$(printf '%s\n' "$st" | awk '$1!="A"{print $2}' | grep -E '^docs/(devlogs|fix-reports|plans)/|^docs/prds/' | grep -vE 'docs/prds/2026-10-08-nao-fleet-0\.12\.0-migration\.md' | head -5 | tr '\n' ' ')"
  assert_eq "" "$ds" "AC8: 本批未修改/删除历史 docs 归档（仅允许新增本批 PRD）"
  # QA 脚本须由 QA 提交引入，不得混入 RD 迁移提交
  adds="$(git log --all --format=%s --diff-filter=A -- docs/reports/ 2>/dev/null | tr '\n' '|')"
  if [ -z "$adds" ]; then
    warn "AC8: docs/reports/ 尚未入库（提交前属预期）"
  elif printf '%s' "$adds" | grep -qi 'qa'; then
    pass "AC8: docs/reports/ 由 QA 提交引入（非 RD 迁移提交）：$adds"
  else
    fail "AC8: docs/reports/ 由非 QA 提交引入：$adds"
  fi
  assert_eq "" "$(git ls-files .agents/.nao-obsolete | head -1)" "AC8/BR4: .nao-obsolete 备份不入库"
else
  warn "AC7/AC8: 无 origin/main 基线，跳过"
fi

# ---------------------------------------------------------------------------
hdr "汇总"
printf 'PASS=%d  FAIL=%d  WARN=%d\n' "$PASS" "$FAIL" "$WARN"
if [ "$FAIL" -eq 0 ]; then printf '%sALL GREEN%s（warn=%d）\n' "$G" "$N" "$WARN"; exit 0; else printf '%sFAIL=%d%s\n' "$R" "$FAIL" "$N"; exit 1; fi
