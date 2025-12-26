# NFS Driver

When a volume is created, this driver will automatically create the corresponding folder on the NFS Server
and provide a mountpoint locally.

**NOTE**: This driver requires `flock` feature, so it only supports NFSv4.

## Driver Options

|Name|Type|Description|Default|Optional|
|:-|:-|:-|:-|:-|
|address|string|NFS server address||false|
|remotePath|string|Remote path of NFS exported||false|
|mountOptions|list|Mount options when mount NFS|["nfsvers=4","rw","noatime","rsize=8192","wsize=8192","tcp","timeo=14","sync"]|true|
|purgeAfterDelete|bool|Indicates whether to purge volumes data from NFS after delete docker volume|false|true|
|allowMultipleMount|bool|Indicates whether to allow multiple containers to mount the same volume|true|true|
|mock|bool|Indicates whether to run in mock mode (no actual NFS mount)|false|true|

## Volume Options

|Name|Type|Description|Optional|
|:-|:-|:-|:-|
|purgeAfterDelete|string|Replace the purgeAfterDelete in the driver options for this volume|true|
|allowMultipleMount|bool|Replace the allowMultipleMount in the driver options for this volume|true|
