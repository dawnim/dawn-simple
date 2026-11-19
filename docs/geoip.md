# IP地域查询与更新

GeoIP是扩展提供的公共能力，使用容器共享实例，不需要业务代码重新打开数据文件。框架审计和客服访问资料使用相同解析器。

## 数据来源

安装与每天03:30的现有计划任务更新五类离线文件。每个文件先下载到临时位置并校验，成功后原子替换，失败时继续使用已有文件。长期运行的进程每分钟检查本地文件是否变化，查询本身不执行文件状态检查。

| 文件 | 用途 |
|---|---|
| `ip2region_v4.xdb` | IPv4国内地域与运营商 |
| `ip2region_v6.xdb` | IPv6国内地域与运营商 |
| `GeoLite2-City.mmdb` | 海外国家、省州、城市及坐标信息 |
| `GeoLite2-Country.mmdb` | 国家信息补充 |
| `GeoLite2-ASN.mmdb` | ASN与网络组织 |

国内优先采用ip2region的地域结果；海外优先GeoLite2 City，国家数据补充未知国家。ASN组织与运营商分别保存，不把ASN组织当作已经验证的接入运营商。私网、回环、链路本地、文档地址与保留地址不查询公网定位库。

结果包含国家代码、地域名称、城市、ISP、ASN、组织、时区、坐标、精度半径及各字段来源。离线IP定位表示网络出口的估计区域，不能证明设备位置，也不用于恢复访客身份。

## 配置与命令

`dawn.geoip.directory`指定保存目录，`dawn.geoip.sources`按文件名配置下载地址。环境变量见`.env.example`中的`GEOIP_DIRECTORY`、`GEOIP_CITY_URL`、`GEOIP_COUNTRY_URL`、`GEOIP_ASN_URL`、`GEOIP_IPV4_URL`和`GEOIP_IPV6_URL`。默认使用GeoLite镜像和ip2region官方数据地址，也可以替换为获得授权的数据源或自己的镜像。

```sh
# dawn-simple仓库，使用业务应用配置更新全部已配置文件
go run . dawn:geoip
```

cron在同一部署中只运行一份，组合运行已包含cron，不需要增加另一个更新进程。

## 开发者调用

```go
import "github.com/dawnim/dawn-server/support/geoip"

manager, err := geoip.ManagerFrom(application.Container())
if err != nil {
    return err
}
result, err := manager.Lookup(ctx, "8.8.8.8")
if err != nil {
    return err
}
_ = result.CountryCode
_ = result.ISP
_ = result.ASN
_ = manager.Status()
```

也可以使用`geoip.Lookup(ctx, ip)`与`geoip.Update(ctx)`调用当前应用实例。`Manager.Status()`提供文件是否就绪、大小与更新时间；容器退出时关闭数据库和重载协程。

```go
manager.Extend(func(ctx context.Context, address netip.Addr, result *geoip.Result) error {
    // 内置解析后补充自己的网络标签，不能保留result指针
    return nil
})
```

可使用`geoip.New(geoip.Config{...})`创建独立实例用于测试或独立程序，调用方负责`Close()`；正常业务应用优先使用容器实例。
