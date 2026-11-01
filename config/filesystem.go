package config

func init() {
	Add("filesystem", func() map[string]interface{} {
		return map[string]interface{}{
			/*
				|--------------------------------------------------------------------------
				| Default Filesystem Disk
				|--------------------------------------------------------------------------
				|
				| Here you may specify the filesystem disk that should be used by
				| default for storing and retrieving files.
				|
			*/
			"default": Env("FILESYSTEM_DISK", "local"),
			"cloud":   Env("FILESYSTEM_CLOUD", "oss"),

			/*
				|--------------------------------------------------------------------------
				| Temporary URL Signing
				|--------------------------------------------------------------------------
				|
				| Local private disks may use signed temporary URLs. When the signing key
				| is empty, the application key may be used by the framework.
				|
			*/
			"temporary_url": map[string]interface{}{
				"signing_key": Env("FILESYSTEM_SIGNING_KEY", ""),
			},

			/*
				|--------------------------------------------------------------------------
				| Filesystem Disks
				|--------------------------------------------------------------------------
				|
				| You may configure as many filesystem disks as necessary. Each disk
				| describes a storage driver and its location or public URL.
				|
			*/
			"disks": map[string]interface{}{
				"local": map[string]interface{}{
					"driver":     "local",
					"root":       Env("FILESYSTEM_LOCAL_ROOT", "storage/app/private"),
					"url":        Env("FILESYSTEM_LOCAL_URL", ""),
					"visibility": Env("FILESYSTEM_LOCAL_VISIBILITY", "private"),
					"serve":      Env("FILESYSTEM_LOCAL_SERVE", true),
				},
				"public": map[string]interface{}{
					"driver":     "local",
					"root":       Env("FILESYSTEM_PUBLIC_ROOT", "storage/app/public"),
					"url":        Env("FILESYSTEM_PUBLIC_URL", "http://localhost:5123/storage"),
					"visibility": Env("FILESYSTEM_PUBLIC_VISIBILITY", "public"),
					"serve":      Env("FILESYSTEM_PUBLIC_SERVE", true),
				},
				"oss": map[string]interface{}{
					"driver":     "oss",
					"bucket":     Env("FILESYSTEM_OSS_BUCKET", ""),
					"endpoint":   Env("FILESYSTEM_OSS_ENDPOINT", ""),
					"access_key": Env("FILESYSTEM_OSS_ACCESS_KEY_ID", ""),
					"secret_key": Env("FILESYSTEM_OSS_ACCESS_KEY_SECRET", ""),
					"prefix":     Env("FILESYSTEM_OSS_PREFIX", ""),
					"url":        Env("FILESYSTEM_OSS_URL", ""),
					"visibility": Env("FILESYSTEM_OSS_VISIBILITY", "private"),
					"timeout":    Env("FILESYSTEM_OSS_TIMEOUT", 30),
				},
				"qiniu": map[string]any{
					"driver":     "qiniu",
					"bucket":     Env("FILESYSTEM_QINIU_BUCKET", ""),
					"endpoint":   Env("FILESYSTEM_QINIU_ENDPOINT", ""),
					"region":     Env("FILESYSTEM_QINIU_REGION", "cn-east-1"),
					"access_key": Env("FILESYSTEM_QINIU_ACCESS_KEY", ""),
					"secret_key": Env("FILESYSTEM_QINIU_SECRET_KEY", ""),
					"prefix":     Env("FILESYSTEM_QINIU_PREFIX", ""),
					"url":        Env("FILESYSTEM_QINIU_URL", ""),
					"path_style": Env("FILESYSTEM_QINIU_PATH_STYLE", false),
				},
				"cos": map[string]any{
					"driver":     "cos",
					"bucket":     Env("FILESYSTEM_COS_BUCKET", ""),
					"endpoint":   Env("FILESYSTEM_COS_ENDPOINT", ""),
					"region":     Env("FILESYSTEM_COS_REGION", "ap-guangzhou"),
					"access_key": Env("FILESYSTEM_COS_ACCESS_KEY", ""),
					"secret_key": Env("FILESYSTEM_COS_SECRET_KEY", ""),
					"prefix":     Env("FILESYSTEM_COS_PREFIX", ""),
					"url":        Env("FILESYSTEM_COS_URL", ""),
					"path_style": Env("FILESYSTEM_COS_PATH_STYLE", false),
				},
				"s3": map[string]any{
					"driver":     "s3",
					"bucket":     Env("FILESYSTEM_S3_BUCKET", ""),
					"endpoint":   Env("FILESYSTEM_S3_ENDPOINT", ""),
					"region":     Env("FILESYSTEM_S3_REGION", "us-east-1"),
					"access_key": Env("FILESYSTEM_S3_ACCESS_KEY", ""),
					"secret_key": Env("FILESYSTEM_S3_SECRET_KEY", ""),
					"prefix":     Env("FILESYSTEM_S3_PREFIX", ""),
					"url":        Env("FILESYSTEM_S3_URL", ""),
					"path_style": Env("FILESYSTEM_S3_PATH_STYLE", false),
				},
			},

			/*
				|--------------------------------------------------------------------------
				| Symbolic Links
				|--------------------------------------------------------------------------
				|
				| Here you may configure the symbolic links that will be created when the
				| `storage:link` command is executed. The map keys should be the locations
				| of the links and the values should be their targets.
				|
			*/
			"links": map[string]interface{}{
				"public/storage": "storage/app/public",
				// "public/images": "storage/app/images",
			},
		}
	})
}
