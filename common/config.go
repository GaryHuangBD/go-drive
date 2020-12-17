package common

import (
	"errors"
	"flag"
	"fmt"
	"github.com/jinzhu/configor"
	"go-drive/common/registry"
	"os"
	"path/filepath"
	"time"
)

var (
	version = "unknown"
	hash    = "unknown"
	build   = "unknown"
)

const (
	LocalFsDir = "local"
	Listen     = ":8089"
)

func InitConfig(ch *registry.ComponentsHolder) (Config, error) {
	config := Config{}

	var v bool
	flag.BoolVar(&v, "v", false, "print version")

	var cfg string
	flag.StringVar(&cfg, "c", "config.yml", "config file path")
	if cfgFile, err := os.Stat(cfg); err == nil && !cfgFile.IsDir() {
		if err = configor.Load(&config, cfg); err != nil {
			return config, errors.Unwrap(err)
		}
	}

	flag.StringVar(&config.Listen, "l", config.Listen, "address listen on")
	flag.StringVar(&config.DBType, "db", config.DBType, "db type")
	flag.StringVar(&config.DBConnectStr, "db-connect", config.DBConnectStr, "db connect string")
	flag.StringVar(&config.DataDir, "d", config.DataDir, "path to the data dir")
	flag.StringVar(&config.ResDir, "s", config.ResDir, "path to the static files")
	flag.BoolVar(&config.FreeFs, "f", config.FreeFs, "enable unlimited local fs drive(absolute path)")

	flag.StringVar(&config.LangDir, "lang-dir", config.LangDir, "languages configuration folder")
	flag.StringVar(&config.DefaultLang, "lang", config.DefaultLang, "default language code")

	flag.StringVar(&config.OAuthRedirectURI, "oauth-redirect-uri", config.OAuthRedirectURI, "OAuth2 redirect_uri")

	flag.Int64Var(&config.ProxyMaxSize, "proxy-max-size", config.ProxyMaxSize, "maximum file size that can be proxied")

	flag.Int64Var(&config.ThumbnailMaxSize, "thumbnail-max-size", config.ThumbnailMaxSize, "maximum file size to create thumbnail")
	flag.IntVar(&config.ThumbnailMaxPixels, "thumbnail-max-pixels", config.ThumbnailMaxPixels, "maximum pixels(W*H) of original image to thumbnails")
	flag.IntVar(&config.ThumbnailConcurrent, "thumbnail-concurrent", config.ThumbnailConcurrent, "maximum number of concurrent creation of thumbnails")
	flag.DurationVar(&config.ThumbnailCacheTTL, "thumbnail-cache-ttl", config.ThumbnailCacheTTL, "thumbnail cache validity")

	flag.IntVar(&config.MaxConcurrentTask, "max-concurrent-task", config.MaxConcurrentTask, "maximum concurrent task(copy, move, upload, delete files)")

	flag.DurationVar(&config.TokenValidity, "token-validity", config.TokenValidity, "token validity")
	flag.BoolVar(&config.TokenRefresh, "token-refresh", config.TokenRefresh, "enable auto refresh token")

	flag.Parse()

	if v {
		fmt.Printf("%s %s build-%s\n", version, hash, build)
		os.Exit(0)
	}

	if _, e := os.Stat(config.DataDir); os.IsNotExist(e) {
		return config, errors.New(fmt.Sprintf("DataDir '%s' does not exist", config.DataDir))
	}
	tempDir, e := config.GetDir("temp", true)
	if e != nil {
		return config, e
	}
	config.TempDir = tempDir

	ch.Add("config", config)
	return config, nil
}

type Config struct {
	Listen       string `required:"true" `
	DBType       string `required:"true" yaml:"db.type"`
	DBConnectStr string `required:"true" yaml:"db.connect"`
	DataDir      string `default:"./"  yaml:"data.dir"`
	// static files(web) dir
	ResDir string       `default:"./web" yaml:"resource.dir"`
	// unlimited fs drive path,
	// fs drive path will be limited in DataDir/local if FreeFs is false
	FreeFs bool         `default:"false" yaml:"free.fs"`

	LangDir string      `default:"./lang" yaml:"lang.dir"`
	// DefaultLang is the default language
	DefaultLang string  `default:"en-US" yaml:"default.lang"`

	TempDir string      `yaml:"-"`

	OAuthRedirectURI string `yaml:"oauth.uri"`

	// ProxyMaxSize is the maximum file size can be proxied when
	// the API call explicitly specifies
	// that it needs to be proxied.
	// The size is unlimited when maxProxySize is <= 0
	ProxyMaxSize int64                 `default:"1048576" yaml:"proxy.maxSize"`

	// ThumbnailMaxSize is the maximum file size(MB) to create thumbnail
	ThumbnailMaxSize    int64          `default:"16777216" yaml:"thumbnail.maxSize"`
	ThumbnailCacheTTL   time.Duration  `default:"48h" yaml:"thumbnail.cacheTTL"`
	ThumbnailConcurrent int            `default:"16" yaml:"thumbnail.concurrent"`
	ThumbnailMaxPixels  int            `default:"22369621" yaml:"thumbnail.maxPixels"`

	MaxConcurrentTask int              `default:"100" yaml:"maxConcurrentTask"`

	TokenType     string               `default:"file" yaml:"tokenType"`
	TokenValidity time.Duration        `default:"2h" yaml:"tokenValidity"`
	TokenRefresh  bool                 `default:"true" yaml:"tokenRefresh"`
}

func (c Config) GetDir(name string, create bool) (string, error) {
	name = filepath.Join(c.DataDir, name)
	if create {
		if _, e := os.Stat(name); os.IsNotExist(e) {
			if e := os.Mkdir(name, 0755); e != nil {
				return "", e
			}
		}
	}
	return name, nil
}

func (c Config) GetResDir() string {
	return c.ResDir
}

func (c Config) GetLangDir() string {
	return c.LangDir
}

func (c Config) GetLocalFsDir() (string, error) {
	if c.FreeFs {
		return "", nil
	}
	return c.GetDir(LocalFsDir, true)
}
