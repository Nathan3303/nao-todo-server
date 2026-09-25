-- ============================================================================
-- backfill_counts.sql —— 领域统计属性联动回填（ADR 2026-09-12 §7）
--
-- 执行时机：首次「带计数发布」之前执行一次（ADR §11.4：发布读列 ⇒ 回填即权威建立时机）；
-- 幂等可重跑。随部署手动执行：mysql < scripts/backfill_counts.sql
--
-- 口径（§6 拍板，用户全项通过）：
--   a. 项目 task_count 含子任务、含已放弃/已归档；不含已删除
--   b. task.subtask_count = 直接子（1 层）
--   c. check_item_count 含已完成项（总数，不分列）
--   d. comment_count 不含已删除评论
--   e. 删除即出局（父/项目计数只反映「现存」子/任务）
--
-- 全部 bump updated_at（B2 红线）：否则已拉过（有游标）的存量客户端永不重拉新计数。
-- ============================================================================

-- 1. tasks.check_item_count：该任务名下未删除检查项数
UPDATE tasks t
SET t.check_item_count = (
        SELECT COUNT(*) FROM task_check_items ci
        WHERE ci.task_id = t.id AND ci.deleted_at IS NULL
    ),
    t.updated_at = NOW()
WHERE t.deleted_at IS NULL;

-- 2. tasks.comment_count：该任务名下未删除评论数（口径 d）
UPDATE tasks t
SET t.comment_count = (
        SELECT COUNT(*) FROM task_comments c
        WHERE c.task_id = t.id AND c.deleted_at IS NULL
    ),
    t.updated_at = NOW()
WHERE t.deleted_at IS NULL;

-- 3. tasks.subtask_count：直接子任务数（口径 b；父已删除的子任务不计，口径 e）
-- 注意：MySQL 不允许 UPDATE 的目标表出现在子查询 FROM 中（ERROR 1093），
-- 故先按 parent_task_id 聚合出派生表，再 LEFT JOIN 回填（无子任务 → COALESCE 0）。
UPDATE tasks t
LEFT JOIN (
    SELECT parent_task_id, COUNT(*) AS cnt
    FROM tasks
    WHERE parent_task_id > 0 AND deleted_at IS NULL
    GROUP BY parent_task_id
) c ON c.parent_task_id = t.id
SET t.subtask_count = COALESCE(c.cnt, 0),
    t.updated_at = NOW()
WHERE t.deleted_at IS NULL;

-- 4. projects.task_count：含子任务/含归档/含放弃；不含已删除（口径 a）
UPDATE projects p
SET p.task_count = (
        SELECT COUNT(*) FROM tasks t
        WHERE t.project_id = p.id AND t.deleted_at IS NULL
    ),
    p.updated_at = NOW()
WHERE p.deleted_at IS NULL;
