# NFS Driver

When a volume is created, this driver will automatically create the corresponding folder on the NFS Server
and provide a mountpoint locally.

**NOTE**: This driver requires `flock` feature, so it only supports NFSv4.

## Driver Options

|Name|Type|Description|Optional|
|:-|:-|:-|:-|
|address|String|NFS server address. Note that if the value is "nfs-server.mock", NFS mounting will be skipped|false|
|remotePath|String|Remote path of NFS exported|false|
|mountOptions|String|Mount options when mount NFS|true|
|purgeAfterDelete|Bool|PurgeAfterDelete indicates whether to purge the volume data after deletion, default is false|true|

## Volume Options

|Name|Type|Description|Optional|
|:-|:-|:-|:-|
|purgeAfterDelete|string|Replace the purgeAfterDelete in the driver options for this volume|true|
