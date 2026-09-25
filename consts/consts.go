// Package consts 提供跨包共享的全局常量
package consts

// SnowflakeEpochMS 雪花 ID 纪元（Epoch），单位毫秒。
// 跨端契约：前后端必须使用相同值（2023-03-01T00:00:00Z = 1677628800000 ms）。
// 禁止修改——存量雪花 ID 的时间位已基于该值生成，改动会造成 ID 时间语义分裂。
const SnowflakeEpochMS int64 = 1677628800000
