# NFS Driver Options

|Name|Type|Description|Optional|
|:-|:-|:-|:-|
|address|String|NFS server address. Note that if the value is "nfs-server.mock", NFS mounting will be skipped|false|
|remotePath|String|RemotePath of NFS exported|false|
|mountOptions|String|MountOptions when mount NFS|true|
|purgeAfterDelete|Bool|PurgeAfterDelete indicates whether to purge the volume data after deletion, default is false|true|
