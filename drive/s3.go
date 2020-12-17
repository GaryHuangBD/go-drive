package drive

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"go-drive/common/drive_util"
	"go-drive/common/errors"
	"go-drive/common/i18n"
	"go-drive/common/task"
	"go-drive/common/types"
	"go-drive/common/utils"
	"io"
	"math"
	"net/url"
	"os"
	"strings"
	"time"
)

func init() {
	drive_util.RegisterDrive(drive_util.DriveFactoryConfig{
		Type:        "s3",
		DisplayName: i18n.T("drive.s3.name"),
		README:      i18n.T("drive.s3.readme"),
		ConfigForm: []types.FormItem{
			{Field: "id", Label: i18n.T("drive.s3.form.ak.label"), Type: "text", Required: true},
			{Field: "secret", Label: i18n.T("drive.s3.form.sk.label"), Type: "password", Required: true},
			{Field: "bucket", Label: i18n.T("drive.s3.form.bucket.label"), Type: "text", Required: true},
			{Field: "path_style", Label: i18n.T("drive.s3.form.path_style.label"), Type: "checkbox", Description: i18n.T("drive.s3.form.path_style.description")},
			{Field: "region", Label: i18n.T("drive.s3.form.region.label"), Type: "text"},
			{Field: "endpoint", Label: i18n.T("drive.s3.form.endpoint.label"), Type: "text", Description: i18n.T("drive.s3.form.endpoint.description")},
			{Field: "proxy_upload", Label: i18n.T("drive.s3.form.proxy_in.label"), Type: "checkbox", Description: i18n.T("drive.s3.form.proxy_in.description")},
			{Field: "proxy_download", Label: i18n.T("drive.s3.form.proxy_out.label"), Type: "checkbox", Description: i18n.T("drive.s3.form.proxy_out.description")},
			{Field: "cache_ttl", Label: i18n.T("drive.s3.form.cache_ttl.label"), Type: "text", Description: i18n.T("drive.s3.form.cache_ttl.description")},
		},
		Factory: drive_util.DriveFactory{Create: NewS3Drive},
	})
}

type S3Drive struct {
	c             *s3.S3
	bucket        *string
	uploadProxy   bool
	downloadProxy bool
	cache         drive_util.DriveCache
	cacheTTL      time.Duration

	tempDir string
}

// NewS3Drive creates a S3 compatible storage
func NewS3Drive(ctx context.Context, config drive_util.DriveConfig,
	utils drive_util.DriveUtils) (types.IDrive, error) {
	id := config["id"]
	checkBucket := config["check_bucket"]
	secret := config["secret"]
	bucket := config["bucket"]
	pathStyle := config["path_style"]
	region := config["region"]
	endpoint := config["endpoint"]
	proxyUpload := config["proxy_upload"]
	proxyDownload := config["proxy_download"]
	cacheTtl, e := time.ParseDuration(config["cache_ttl"])
	if e != nil {
		cacheTtl = -1
	}

	sess, e := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(id, secret, ""),
		S3ForcePathStyle: aws.Bool(pathStyle != ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
	})
	if e != nil {
		return nil, e
	}
	client := s3.New(sess)
	d := &S3Drive{
		c:             client,
		bucket:        aws.String(bucket),
		uploadProxy:   proxyUpload != "",
		downloadProxy: proxyDownload != "",
		cacheTTL:      cacheTtl,
		tempDir:       utils.Config.TempDir,
	}
	if cacheTtl <= 0 {
		d.cache = drive_util.DummyCache()
	} else {
		d.cache = utils.CreateCache(d.deserializeEntry, nil)
	}
	if checkBucket != "" {
		return d, d.check(ctx)
	}
	return d, nil
}

func (s *S3Drive) check(ctx context.Context) error {
	_, e := s.c.HeadBucketWithContext(ctx, &s3.HeadBucketInput{
		Bucket: s.bucket,
	})
	if e != nil {
		if ae, ok := e.(awserr.Error); ok {
			switch ae.Code() {
			case s3.ErrCodeNoSuchBucket:
				return err.NewNotFoundMessageError(i18n.T("drive.s3.bucket_not_exists", *s.bucket))
			}
		}
	}
	return e
}

func (s *S3Drive) deserializeEntry(dat string) (types.IEntry, error) {
	ec, e := drive_util.DeserializeEntry(dat)
	if e != nil {
		return nil, e
	}
	return &s3Entry{key: ec.Path, c: s, size: ec.Size, modTime: ec.ModTime, isDir: ec.Type.IsDir()}, nil
}

func (s *S3Drive) Meta(context.Context) types.DriveMeta {
	return types.DriveMeta{CanWrite: true}
}

func (s *S3Drive) get(path string, ctx context.Context) (*s3Entry, error) {
	obj, e := s.c.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: s.bucket,
		Key:    aws.String(path),
	})
	if e != nil {
		if errCodeMatches(e, "NotFound") {
			if strings.HasSuffix(path, "/") {
				return nil, err.NewNotFoundError()
			}
			return s.get(path+"/", ctx)
		}
		return nil, e
	}
	if strings.HasSuffix(path, "/") {
		return s.newS3DirEntry(path, obj.LastModified), nil
	}
	return s.newS3ObjectEntry(path, obj.ContentLength, obj.LastModified), nil
}

func (s *S3Drive) Get(ctx context.Context, path string) (types.IEntry, error) {
	if utils.IsRootPath(path) {
		return s.newS3DirEntry(path, nil), nil
	}
	if cached, _ := s.cache.GetEntry(path); cached != nil {
		return cached, nil
	}
	entry, e := s.get(path, ctx)
	if e != nil {
		return nil, e
	}
	_ = s.cache.PutEntry(entry, s.cacheTTL)
	return entry, nil
}

func (s *S3Drive) Save(ctx types.TaskCtx, path string, _ int64,
	override bool, reader io.Reader) (types.IEntry, error) {
	if !override {
		if _, e := drive_util.RequireFileNotExists(ctx, s, path); e != nil {
			return nil, e
		}
	}
	var readSeeker io.ReadSeeker
	if rs, ok := reader.(io.ReadSeeker); ok {
		readSeeker = rs
	} else {
		file, e := drive_util.CopyReaderToTempFile(ctx, reader, s.tempDir)
		if e != nil {
			return nil, e
		}
		defer func() {
			_ = file.Close()
			_ = os.Remove(file.Name())
		}()
		readSeeker = file
	}
	_, e := s.c.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: s.bucket,
		Key:    aws.String(path),
		Body:   readSeeker,
	})
	if e != nil {
		return nil, e
	}
	_ = s.cache.Evict(path, false)
	_ = s.cache.Evict(utils.PathParent(path), false)
	get, e := s.Get(ctx, path)
	if e != nil {
		return nil, e
	}
	ctx.Progress(get.Size(), false)
	return get, nil
}

func (s *S3Drive) MakeDir(ctx context.Context, path string) (types.IEntry, error) {
	path = path + "/"
	if dir, e := s.Get(ctx, path); e == nil {
		if !dir.Type().IsDir() {
			return nil, err.NewNotAllowedMessageError(i18n.T("drive.file_exists"))
		}
		return dir, nil
	}
	_, e := s.c.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: s.bucket,
		Key:    aws.String(path),
	})
	if e != nil {
		return nil, e
	}
	_ = s.cache.Evict(utils.PathParent(path), false)
	return s.newS3DirEntry(path, nil), nil
}

func (s *S3Drive) isSelf(e types.IEntry) bool {
	if fe, ok := e.(*s3Entry); ok {
		return fe.c == s
	}
	return false
}

func (s *S3Drive) Copy(ctx types.TaskCtx, from types.IEntry, to string, override bool) (types.IEntry, error) {
	from = drive_util.GetIEntry(from, s.isSelf)
	if from == nil || from.Type().IsDir() {
		return nil, err.NewUnsupportedError()
	}
	entry, _, e := s.copy(from.(*s3Entry), to, override, ctx)
	return entry, e
}

func (s *S3Drive) copy(from *s3Entry, to string, override bool, ctx types.TaskCtx) (*s3Entry, bool, error) {
	if !override {
		_, e := s.Get(ctx, to)
		if e == nil {
			modTime := utils.Time(from.modTime)
			// skip
			return s.newS3ObjectEntry(to, &from.size, &modTime), true, nil
		}
		if !err.IsNotFoundError(e) {
			return nil, false, e
		}
	}
	ctx.Total(from.size, false)
	obj, e := s.c.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     s.bucket,
		Key:        aws.String(to),
		CopySource: aws.String(url.QueryEscape(*s.bucket + "/" + from.key)),
	})
	if e != nil {
		return nil, false, e
	}
	_ = s.cache.Evict(to, true)
	_ = s.cache.Evict(utils.PathParent(to), false)
	ctx.Progress(from.Size(), false)
	return s.newS3ObjectEntry(to, &from.size, obj.CopyObjectResult.LastModified), false, nil
}

func (s *S3Drive) Move(ctx types.TaskCtx, from types.IEntry, to string, override bool) (types.IEntry, error) {
	from = drive_util.GetIEntry(from, s.isSelf)
	if from == nil || from.Type().IsDir() {
		return nil, err.NewUnsupportedError()
	}
	fromEntry := from.(*s3Entry)
	entry, skip, e := s.copy(fromEntry, to, override, ctx)
	if e != nil {
		return nil, e
	}
	if !skip {
		e = s.delete(fromEntry.key, task.DummyContext())
		_ = s.cache.Evict(fromEntry.key, true)
		_ = s.cache.Evict(utils.PathParent(fromEntry.key), false)
	}
	return entry, e
}

func (s *S3Drive) List(ctx context.Context, path string) ([]types.IEntry, error) {
	if cached, _ := s.cache.GetChildren(path); cached != nil {
		return cached, nil
	}
	s3Path := path
	if !utils.IsRootPath(s3Path) {
		s3Path = s3Path + "/"
	}
	objs, e := s.c.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket:    s.bucket,
		Prefix:    aws.String(s3Path),
		Delimiter: aws.String("/"),
	})
	if e != nil {
		return nil, e
	}
	entries := make([]types.IEntry, 0)
	pathSet := make(map[string]bool, 0)
	for _, o := range objs.Contents {
		if *o.Key == s3Path {
			// fake dir
			continue
		}
		entries = append(entries, s.newS3ObjectEntry(*o.Key, o.Size, o.LastModified))
		pathSet[*o.Key] = true
	}
	for _, p := range objs.CommonPrefixes {
		if _, ok := pathSet[(*p.Prefix)[:len(*p.Prefix)-1]]; ok {
			// skip dir with same name
			continue
		}
		entries = append(entries, s.newS3DirEntry(*p.Prefix, nil))
	}
	_ = s.cache.PutChildren(path, entries, s.cacheTTL)
	return entries, nil
}

func (s *S3Drive) delete(path string, ctx types.TaskCtx) error {
	entry, e := s.Get(ctx, path)
	if e != nil {
		return e
	}
	tree, e := drive_util.BuildEntriesTree(ctx, entry, false)
	if e != nil {
		return e
	}
	entries := drive_util.FlattenEntriesTree(tree)
	n := int(math.Ceil(float64(len(entries)) / 1000))
	for i := 0; i < n; i += 1 {
		batches := entries[i*1000 : int(math.Min(float64((i+1)*1000), float64(len(entries))))]
		deletes := make([]*s3.ObjectIdentifier, len(batches))
		for i, o := range batches {
			key := o.Path()
			if o.Type().IsDir() {
				key += "/"
			}
			deletes[i] = &s3.ObjectIdentifier{
				Key: aws.String(key),
			}
		}
		r, e := s.c.DeleteObjectsWithContext(ctx, &s3.DeleteObjectsInput{
			Bucket: s.bucket,
			Delete: &s3.Delete{
				Objects: deletes,
				Quiet:   aws.Bool(true),
			},
		})
		if e != nil {
			return e
		}
		if r.Errors != nil && len(r.Errors) > 0 {
			return errors.New(fmt.Sprintf("%s: %s", *r.Errors[0].Key, *r.Errors[0].Code))
		}
		ctx.Progress(int64(len(batches)), false)
	}
	return nil
}

func (s *S3Drive) Delete(ctx types.TaskCtx, path string) error {
	e := s.delete(path, ctx)
	_ = s.cache.Evict(utils.PathParent(path), false)
	_ = s.cache.Evict(path, true)
	return e
}

func (s *S3Drive) Upload(ctx context.Context, path string, size int64,
	override bool, config types.SM) (*types.DriveUploadConfig, error) {
	action := config["action"]
	uploadId := config["uploadId"]
	partsEtag := config["parts"]
	seq := utils.ToInt64(config["seq"], -1)

	if !override {
		if _, e := drive_util.RequireFileNotExists(ctx, s, path); e != nil {
			return nil, e
		}
	}

	r := types.DriveUploadConfig{
		Provider: types.S3Provider,
		Config:   types.SM{},
	}
	preSigned := ""

	var e error
	switch action {
	case "UploadPart":
		req, _ := s.c.UploadPartRequest(&s3.UploadPartInput{
			Bucket:     s.bucket,
			Key:        aws.String(path),
			PartNumber: aws.Int64(seq + 1),
			UploadId:   aws.String(uploadId),
		})
		preSigned, e = req.Presign(2 * time.Hour)
		break
	case "CompleteMultipartUpload":
		_, e := s.c.CompleteMultipartUploadWithContext(ctx, &s3.CompleteMultipartUploadInput{
			Bucket:   s.bucket,
			Key:      aws.String(path),
			UploadId: aws.String(uploadId),
			MultipartUpload: &s3.CompletedMultipartUpload{
				Parts: buildCompleteUploadBody(partsEtag),
			},
		})
		_ = s.cache.Evict(path, false)
		_ = s.cache.Evict(utils.PathParent(path), false)
		return nil, e
	case "AbortMultipartUpload":
		_, e := s.c.AbortMultipartUploadWithContext(ctx, &s3.AbortMultipartUploadInput{
			Bucket:   s.bucket,
			Key:      aws.String(path),
			UploadId: aws.String(uploadId),
		})
		return nil, e
	case "CompletePutObject":
		_ = s.cache.Evict(path, false)
		_ = s.cache.Evict(utils.PathParent(path), false)
		return nil, nil
	default:
		if s.uploadProxy {
			return types.UseLocalProvider(size), nil
		}
		if size <= 5*1024*1024 {
			req, _ := s.c.PutObjectRequest(&s3.PutObjectInput{
				Bucket: s.bucket,
				Key:    aws.String(path),
			})
			preSigned, e = req.Presign(2 * time.Hour)
		} else {
			req, _ := s.c.CreateMultipartUploadRequest(&s3.CreateMultipartUploadInput{
				Bucket: s.bucket,
				Key:    aws.String(path),
			})
			preSigned, e = req.Presign(2 * time.Hour)
			r.Config["multipart"] = "1"
		}
	}
	if e != nil {
		return nil, e
	}
	if preSigned != "" {
		r.Config["url"] = preSigned
	}
	return &r, e
}

func buildCompleteUploadBody(etag string) []*s3.CompletedPart {
	temp := strings.Split(etag, ";")
	r := make([]*s3.CompletedPart, len(temp))
	for i, e := range temp {
		r[i] = &s3.CompletedPart{
			PartNumber: aws.Int64(int64(i + 1)),
			ETag:       aws.String(e),
		}
	}
	return r
}

func (s *S3Drive) Dispose() error {
	return nil
}

func (s *S3Drive) newS3DirEntry(path string, lastModified *time.Time) *s3Entry {
	var mtime int64 = -1
	if lastModified != nil {
		mtime = utils.Millisecond(*lastModified)
	}
	path = utils.CleanPath(path)
	return &s3Entry{
		isDir:   true,
		key:     path,
		modTime: mtime,
		c:       s,
	}
}

func (s *S3Drive) newS3ObjectEntry(path string, size *int64, lastModified *time.Time) *s3Entry {
	path = utils.CleanPath(path)
	return &s3Entry{
		isDir:   false,
		key:     path,
		size:    *size,
		modTime: utils.Millisecond(*lastModified),
		c:       s,
	}
}

type s3Entry struct {
	key     string
	c       *S3Drive
	size    int64
	modTime int64
	isDir   bool
}

func (s *s3Entry) Path() string {
	return s.key
}

func (s *s3Entry) Type() types.EntryType {
	if s.isDir {
		return types.TypeDir
	}
	return types.TypeFile
}

func (s *s3Entry) Size() int64 {
	if s.isDir {
		return -1
	}
	return s.size
}

func (s *s3Entry) Meta() types.EntryMeta {
	return types.EntryMeta{
		CanRead:  true,
		CanWrite: true,
	}
}

func (s *s3Entry) ModTime() int64 {
	if s.isDir {
		return -1
	}
	return s.modTime
}

func (s *s3Entry) Drive() types.IDrive {
	return s.c
}

func (s *s3Entry) Name() string {
	return utils.PathBase(s.key)
}

func (s *s3Entry) GetReader(ctx context.Context) (io.ReadCloser, error) {
	obj, e := s.c.c.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: s.c.bucket,
		Key:    aws.String(s.key),
	})
	if e != nil {
		return nil, e
	}
	return obj.Body, nil
}

func (s *s3Entry) GetURL(context.Context) (*types.ContentURL, error) {
	req, _ := s.c.c.GetObjectRequest(&s3.GetObjectInput{
		Bucket: s.c.bucket,
		Key:    aws.String(s.key),
	})
	downloadUrl, e := req.Presign(8 * time.Hour)
	if e != nil {
		return nil, e
	}
	return &types.ContentURL{URL: downloadUrl, Proxy: s.c.downloadProxy}, nil
}

func errCodeMatches(e error, code string) bool {
	if ae, ok := e.(awserr.Error); ok {
		return ae.Code() == code
	}
	return false
}
