# NFS Driver Options

|Name|Type|Description|Optional|
|:-|:-|:-|:-|
|address|String|NFS server address. Note that if the value is "nfs-server.mock", NFS mounting will be skipped|false|
|remotePath|String|Remote path of NFS exported|false|
|mountOptions|String|Mount options when mount NFS|true|
|purgeAfterDelete|Bool|PurgeAfterDelete indicates whether to purge the volume data after deletion, default is false|true|
