//+build wireinject

package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"go-drive/common"
	"go-drive/common/i18n"
	"go-drive/common/registry"
	"go-drive/common/task"
	"go-drive/common/utils"
	"go-drive/drive"
	"go-drive/server"
	"go-drive/server/token"
	"go-drive/storage"
)

func Initialize(ctx context.Context, ch *registry.ComponentsHolder) (*gin.Engine, error) {
	wire.Build(
		common.InitConfig,
		storage.NewDB,
		storage.NewUserDAO,
		storage.NewPathPermissionDAO,
		storage.NewDriveCacheDAO,
		storage.NewGroupDAO,
		storage.NewPathMountDAO,
		storage.NewDriveDAO,
		storage.NewDriveDataDAO,
		wire.Bind(new(task.Runner), new(*task.TunnyRunner)),
		task.NewTunnyRunner,
		utils.NewSigner,
		token.NewTokenStore,
		server.NewChunkUploader,
		server.NewThumbnail,
		drive.NewRootDrive,
		wire.Bind(new(i18n.MessageSource), new(*i18n.FileMessageSource)),
		i18n.NewFileMessageSource,
		server.InitServer,
	)
	return &gin.Engine{}, nil
}
