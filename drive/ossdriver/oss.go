package ossdriver

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go-drive/common/drive_util"
	"time"
)

//func init() {
//	drive_util.RegisterDrive(drive_util.DriveFactoryConfig{
//		Type:        "s3",
//		DisplayName: i18n.T("drive.s3.name"),
//		README:      i18n.T("drive.s3.readme"),
//		ConfigForm: []types.FormItem{
//			{Field: "id", Label: i18n.T("drive.s3.form.ak.label"), Type: "text", Required: true},
//			{Field: "secret", Label: i18n.T("drive.s3.form.sk.label"), Type: "password", Required: true},
//			{Field: "bucket", Label: i18n.T("drive.s3.form.bucket.label"), Type: "text", Required: true},
//			{Field: "path_style", Label: i18n.T("drive.s3.form.path_style.label"), Type: "checkbox", Description: i18n.T("drive.s3.form.path_style.description")},
//			{Field: "region", Label: i18n.T("drive.s3.form.region.label"), Type: "text"},
//			{Field: "endpoint", Label: i18n.T("drive.s3.form.endpoint.label"), Type: "text", Description: i18n.T("drive.s3.form.endpoint.description")},
//			{Field: "proxy_upload", Label: i18n.T("drive.s3.form.proxy_in.label"), Type: "checkbox", Description: i18n.T("drive.s3.form.proxy_in.description")},
//			{Field: "proxy_download", Label: i18n.T("drive.s3.form.proxy_out.label"), Type: "checkbox", Description: i18n.T("drive.s3.form.proxy_out.description")},
//			{Field: "cache_ttl", Label: i18n.T("drive.s3.form.cache_ttl.label"), Type: "text", Description: i18n.T("drive.s3.form.cache_ttl.description")},
//		},
//		Factory: drive_util.DriveFactory{Create: NewOSSDrive},
//	})
//}

type OSSDrive struct {
	c             *oss.Client
	bucket        *string
	uploadProxy   bool
	downloadProxy bool
	cache         drive_util.DriveCache
	cacheTTL      time.Duration

	tempDir string
}

//func NewOSSDrive(ctx context.Context, config drive_util.DriveConfig, utils drive_util.DriveUtils) (types.IDrive, error) {
//   return &OSSDrive{}, nil
//}
