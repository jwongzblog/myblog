大拿们都说云服务是三分开发七分运维，在我看来运维占比更大，填坑之路依然有很长的路要走，本篇尝试讲述一个优化运维的场景，即ceph存储的集群坏盘后如何快速定位到硬盘的槽位，并正确的替换。其中重要的一个步骤是如何快速找出坏盘的Linux磁盘符（/dev/sd*）对应在服务器上硬盘的插槽位置。

**首先，利用MegaRAID CLI找出virtual disk（逻辑盘）与物理磁盘的关联**
```
#/opt/MegaRAID/MegaCli/MegaCli64 -LdPdInfo -aALL
```
结果如下：
```
Virtual Drive: 22 (Target Id: 22)
Name                :
RAID Level          : Primary-0, Secondary-0, RAID Level Qualifier-0
Size                : 3.637 TB
Sector Size         : 512
Is VD emulated      : Yes
Parity Size         : 0
State               : Optimal
Strip Size          : 64 KB
Number Of Drives    : 1
Span Depth          : 1
Default Cache Policy: WriteThrough, ReadAheadNone, Cached, No Write Cache if Bad BBU
Current Cache Policy: WriteThrough, ReadAheadNone, Cached, No Write Cache if Bad BBU
Default Access Policy: Read/Write
Current Access Policy: Read/Write
Disk Cache Policy   : Enabled
Encryption Type     : None
Bad Blocks Exist: No
PI type: No PI

Is VD Cached: No
Number of Spans: 1
Span: 0 - Number of PDs: 1

PD: 0 Information
Enclosure Device ID: 9
Slot Number: 11
Drive's position: DiskGroup: 21, Span: 0, Arm: 0
Enclosure position: 1
Device Id: 31
WWN: 5000c500a3fccfec
Sequence Number: 2
Media Error Count: 0
Other Error Count: 0
Predictive Failure Count: 0
Last Predictive Failure Event Seq Number: 0
PD Type: SATA

Raw Size: 3.638 TB [0x1d1c0beb0 Sectors]
Non Coerced Size: 3.637 TB [0x1d1b0beb0 Sectors]
Coerced Size: 3.637 TB [0x1d1b00000 Sectors]
Sector Size:  512
Logical Sector Size:  512
Physical Sector Size:  4096
Firmware state: Online, Spun Up
SAS Address(0): 0x56c92bf000b0eb17
Connected Port Number: 1(path0) 
Inquiry Data:             ZC1282GHST4000NM0115-1YZ107                     SN02    
Port's Linkspeed: 6.0Gb/s 
Drive has flagged a S.M.A.R.T alert : No

Exit Code: 0x00
```
- Virtual Drive: 22，代表逻辑卷序列号
- Enclosure Device ID: 9 ， Slot Number: 11，即[9:11]，硬盘的插槽序号
- WWN: 5000c500a3fccfec，硬盘的全球唯一name，一般印刷在硬盘外壳
- Inquiry Data:             ZC1282GHST4000NM0115-1YZ107                     SN02 ，这个也是硬盘外壳印刷的几条序列号组合信息
**命令找出Linux磁盘符与逻辑卷（virtual disk）的关联**
```
$ /opt/MegaRAID/storcli/storcli64 /c0/v0 show all | grep NAA
SCSI NAA Id = 6001676001750006201086de0bd7f605
$ ls -al /dev/disk/by-id/ | grep wwn-0x6001676001750006201086de0bd7f605
lrwxrwxrwx 1 root root   9 Jan 23 10:55 wwn-0x6001676001750006201086de0bd7f605 -> ../../sdk
```
