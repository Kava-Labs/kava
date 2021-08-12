# Kava-8 Rollback Instructions

> This document is provided to assist Chinese speaking validators upgrade, but might not be updated as fast, and might not be the most current version. When in doubt, consult the English version
> 本文档用于帮助说中文的验证节点升级，但可能不是最新的版本。如有疑问，请查阅英文版

In the event that the kava-8 relaunch is unsuccessful, we will restart the kava-7 chain using the last known state.
如果 kava-8 的升级失败，我们将使用最后一个区块链状态重新启动 kava-7 链。

In order to restore the previous chain, the following data must be recovered by validators:
为了恢复之前的链，验证者必须恢复以下数据：

- The database that contains the state of the previous chain (in ~/.kvd/data by default)
  - 包含之前链状态的数据库（默认在 ~/.kvd/data 中）
- The priv_validator_state.json file of the validator (in ~/.kvd/data by default)
- 验证节点的 priv_validator_state.json 文件（默认在 ~/.kvd/data 中）

If you don't have the database data, the Kava developer team or another validator will share a copy of the database via Amazon s3 or a similar service. You will be able to download a copy of the data and verify it before starting your node.
如果您没有数据库数据，Kava 开发团队或其他验证者将通过 Amazon s3 或类似服务共享数据库副本。您将能够下载数据副本并在启动节点之前对其进行验证。

If you don't have the backup priv_validator_state.json file, you will not have double sign protection on the first block. If this is the case, it's best to consult in the validator discord before starting your node.
如果您没有备份 priv_validator_state.json 文件，您将在新链的第一个区块不会有双签保护。如果是这种情况，最好在启动节点之前在 discord 里咨询其他验证者。

## Restoring state procedure 恢复状态过程

1. Copy the contents of your backup data directory back to the $KVD_HOME/data directory. By default this is ~/.kvd/data.
   将备份数据目录的内容复制回 $KVD_HOME/data 目录。默认情况下，这是 ~/.kvd/data。

```sh
# Assumes backup is stored in "backup" directory 假设备份存储在“备份”目录中：
rm -rf ~/.kvd/data
mv backup/.kvd/data ~/.kvd/data
```

2. Install the previous version of kava 安装之前版本的 kava

```sh
# from kava directory 在kava目录里
git checkout v0.14.2
make install

## verify version 验证版本
kvd version --long
```

3. Restart kvd process

```sh
### be sure to remove --halt-time flag if it is set  如果设置了 --halt-time 标志，请删除它
sudo systemctl daemon-reload
sudo systemctl restart kvd
```
