package project

import (
	"context"
	"naotodoserver/domain/project/service"
	"naotodoserver/interfaces/types"
	"sync"
)

type ProjectApp interface {
	Create(ctx context.Context, req *types.CreateProjectReq) (*types.CreateProjectRes, error)
	Get(ctx context.Context, req *types.GetProjectReq) (*types.GetProjectRes, error)
	Update(ctx context.Context, req *types.UpdateProjectReq) (*types.UpdateProjectRes, error)
	Delete(ctx context.Context, req *types.DeleteProjectReq) (*types.DeleteProjectRes, error)
	Restore(ctx context.Context, req *types.RestoreProjectReq) (*types.RestoreProjectRes, error)
	HardDelete(
		ctx context.Context,
		req *types.HardDeleteProjectReq,
	) (*types.HardDeleteProjectRes, error)
	Archive(ctx context.Context, req *types.ArchiveProjectReq) (*types.ArchiveProjectRes, error)
	Unarchive(
		ctx context.Context,
		req *types.UnarchiveProjectReq,
	) (*types.UnarchiveProjectRes, error)
	List(ctx context.Context) (types.ListProjectRes, error)
	GetPreference(ctx context.Context, req *types.GetProjectPreferenceReq) (*types.ProjectPreferenceRes, error)
	SavePreference(
		ctx context.Context,
		req *types.UpdateProjectPreferenceReq,
	) (*types.UpdateProjectPreferenceRes, error)
}

type projectAppImpl struct {
	projectDomain service.ProjectDomain
}

var (
	App  *projectAppImpl
	once sync.Once
)
