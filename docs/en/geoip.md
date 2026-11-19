# IP geolocation

GeoIP is a public extension capability backed by a container-shared instance. Business code does not need to reopen data files. Audit logs and support visitor profiles use the same resolver.

## Data sources

Installation and the existing daily 03:30 schedule update five offline files. Each download is staged and validated before atomic replacement. Failed downloads leave existing data available. Long-running processes check for changes every minute; lookups do not stat files themselves.

| File | Purpose |
|---|---|
| `ip2region_v4.xdb` | IPv4 regions and ISPs in China |
| `ip2region_v6.xdb` | IPv6 regions and ISPs in China |
| `GeoLite2-City.mmdb` | Overseas country, subdivision, city, and coordinates |
| `GeoLite2-Country.mmdb` | Additional country data |
| `GeoLite2-ASN.mmdb` | ASN and network organization |

Chinese regions prefer ip2region. Overseas regions prefer GeoLite2 City; Country fills unknown countries. ASN organization and ISP remain separate; an ASN organization does not establish the access ISP. Private, loopback, link-local, documentation, and reserved addresses skip public geolocation databases.

Results include country code, region, city, ISP, ASN, organization, timezone, coordinates, accuracy radius, and per-field sources. Offline IP data estimates the network exit's location; it cannot establish device location or recover visitor identity.

## Configuration and commands

`dawn.geoip.directory` sets the directory. `dawn.geoip.sources` maps filenames to download URLs. See `.env.example` for `GEOIP_DIRECTORY`, `GEOIP_CITY_URL`, `GEOIP_COUNTRY_URL`, `GEOIP_ASN_URL`, `GEOIP_IPV4_URL`, and `GEOIP_IPV6_URL`. Defaults use a GeoLite mirror and official ip2region data; replace them with authorized sources or your own mirror as needed.

```sh
# dawn-simple repository: update every configured file using application settings
go run . dawn:geoip
```

Run one cron instance per deployment. Combined service mode already includes cron and needs no additional updater process.

## Developer APIs

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

Use `geoip.Lookup(ctx, ip)` and `geoip.Update(ctx)` to call the current application's instance directly. `Manager.Status()` exposes readiness, size, and update time. Container shutdown closes databases and the reload goroutine.

```go
manager.Extend(func(ctx context.Context, address netip.Addr, result *geoip.Result) error {
    // Add custom network labels after built-in resolution; do not retain result.
    return nil
})
```

`geoip.New(geoip.Config{...})` creates a standalone instance for tests or independent programs; the caller must `Close()` it. Normal applications should use the container instance.
