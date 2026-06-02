package checkitem

import (
	"context"
	"naotodoserver/domain/checkitem/service"
	"naotodoserver/interfaces/types"
)

// 检查事项应用接口
type CheckItemApp interface {
	// 获取检查事项详情
	GetCheckItemById(ctx context.Context, itemId string) (*types.GetCheckItemRes, error)

	// 创建检查事项
	CreateCheckItem(ctx context.Context, req *types.CreateCheckItemReq) (*types.CreateCheckItemRes, error)

	// 更新检查事项
	UpdateCheckItem(
		ctx context.Context,
		itemId string,
		req *types.UpdateCheckItemReq,
	) error

	// 删除检查事项
	DeleteCheckItem(ctx context.Context, itemId string) error

	// 获取检查事项列表
	ListCheckItem(ctx context.Context, taskId string) (types.ListCheckItemRes, error)

	// 排序检查事项
	ResortCheckItems(ctx context.Context, req *types.ResortCheckItemsReq) (*types.ResortCheckItemsRes, error)

	// 批量更新检查事项
	BatchUpdateCheckItems(ctx context.Context, req *types.BatchUpdateCheckItemReq) (*types.BatchUpdateCheckItemRes, error)
}

// 检查事项应用实现
type CheckItemAppImpl struct {
	checkitemDomain service.CheckItemDomain
}

// 检查事项应用实例
