# CIFS Driver

The CIFS (Common Internet File System) driver enables Docker volumes to be
backed by SMB/CIFS shares.

When a volume is created, this driver will automatically create the
corresponding folder on the SMB Server and provide a mountpoint locally.

## Example Driver Options

```json
{
    "address": "cifs-server.example.com",
    "remotePath": "/share",
    "username": "user",
    "password": "pass",
    "mountOptions": [
        "uid=99",
        "gid=100"
    ]
}
```

## Driver Options

|Name|Type|Description|Default|Optional|
|:-|:-|:-|:-|:-|
|address|string|CIFS server address||false|
|remotePath|string|Remote path of CIFS exported||false|
|username|string|CIFS server username||false|
|password|string|CIFS server password||true|
|mountOptions|list|Mount options when mount CIFS|[]|true|
|purgeAfterDelete|bool|Indicates whether to purge volumes data from CIFS after delete docker volume|false|true|
|mock|bool|Indicates whether to run in mock mode (no actual CIFS mount)|false|true|

## Volume Options

|Name|Type|Description|Optional|
|:-|:-|:-|:-|
|purgeAfterDelete|string|Replace the purgeAfterDelete in the driver options for this volume|true|
