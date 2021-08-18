# kava-8 Upgrade Instructions 升级指南

> This document is provided to assist Chinese speaking validators upgrade, but might not be updated as fast, and might not be the most current version. When in doubt, consult the English version
> 本文档用于帮助说中文的验证节点升级，但可能不是最新的版本。如有疑问，请查阅英文版

## Software Version and Key Dates 软件版本和关键日期

- We will be upgrading from chain-id "kava-7" to chain-id "kava-8".
  - 将升级 chain-id 从"kava-7"到"kava-8"。
- The version of Kava for kava-8 is [TODO: version]
- The kava-7 chain will be shutdown with a `SoftwareUpgradeProposal` that activates at approximately 13:00 UTC on August, 30 2021.
  - kava-7 链将通过 `SoftwareUpgradeProposal` 关闭，该 `SoftwareUpgradeProposal` 在 2021 年 8 月 30 日大约 13:00 UTC 激活。
- kava-8 genesis time is set to August 30th, 2021 at 15:00 UTC.
  - kava-8 创世块设置为 2021 年 8 月 30 日 15:00 UTC。
- The version of cosmos-sdk for kava-8 is v0.39.3.
  - kava-8 的 cosmos-sdk 版本为 v0.39.3。
- The version of tendermint for kava-8 v0.33.9.
  - kava-8 的 tendermint 版本为 v0.33.9。
- The minimum version of golang for kava-8 is 1.13+, 1.15+ has been tested and is recommended.
  - kava-8 的 golang 最低版本为 1.13+。 版本 1.15+ 已经过测试，推荐使用。

### Risks 风险

As a validator, performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.
作为验证者，在您的共识节点上执行升级程序会增加两种风险：双签名被削减。此过程中最重要的部分是在启动验证节点和签名之前确保您的软件版本和创世文件哈希无错。

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it. If the network is halted and you have started with a different genesis file than the expected one, seek advice from a Kava developer before resetting your validator.
验证者能做的最危险的事情是发现自己犯了错误并在网络启动期间再次重复升级程序。如果您在此过程中发现错误，最好的做法是等待网络启动后再进行更正。如果网络停止运行并且您开始使用的创世文件与预期的不同，请在重置验证节点之前向 Kava 开发人员寻求建议。

### Recovery 恢复

Prior to exporting kava-7 state, validators are encouraged to take a full data snapshot at the export height before proceeding. Snap-shotting depends heavily on infrastructure, but generally this can be done by backing up the .kvd and .kvcli directories.
在导出 kava-7 状态之前，建议验证者在导出区块的高度获取完整的数据快照。数据快照基于节点基础设施，但通常可以通过备份 .kvd 和 .kvcli 目录来完成。

It is critically important to back-up the .kvd/data/priv_validator_state.json file after stopping your kvd process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.
很重要的是，在停止 kvd 程序之后，必须备份 .kvd/data/priv_validator_state.json 文件。当您的验证者参与共识轮次时，此文件会在每个区块更新。 万一升级失败并需要重新启动之前的链，它是防止双签的关键文件。

In the event that the upgrade does not succeed, validators and operators must downgrade back to v0.14.2 of the Kava software and restore to their latest snapshot before restarting their nodes.
如果升级不成功，验证者和节点操作者必须降级回 Kava 软件的 v0.14.2 并在重新启动节点之前恢复到最后的快照。

## Upgrade Procedure 升级流程

### Before the upgrade 升级前

Kava Labs will submit a `SoftwareUpgradeProposal` that specifies the exact block _height_ at which the chain should shut down. This height will correspond to approximately 13:00 UTC on August 30th. Once the proposal passes, the chain will shutdown automatically at the specified height and does not require manual intervention by validators.
Kava Labs 将提交一个“SoftwareUpgradeProposal”，该消息会指定关闭链的确切区块高度。这个高度将对应于 8 月 30 日大约 13:00 UTC。一旦提案通过，链将在指定高度自动关闭，不需要验证者的任何人工操作。

### On the day of the upgrade 升级当天

**The kava chain is expected to halt at 13:00 UTC, and restart with new software at 15:00 UTC August 30th. Do not stop your node and begin the upgrade before 13:00UTC on August 30th, or you may go offline and be unable to recover until after the upgrade!**

**kava 链预计将在 13:00 UTC 停止，并在 8 月 30 日 15:00 UTC 重新启动新版本。 8 月 30 日 13:00 UTC 之前不要提前停止您的节点并开始升级，否则您可能会下线，直到升级后才能恢复！**

Kava developers will update this PR with the final block number when it is reached. **Make sure the kvd process is stopped before proceeding and that you have backed up your validator**. Failure to backup your validator could make it impossible to restart your node if the upgrade fails.
Kava 开发人员将在达到最终区块编号时更新此文件。 **确保在继续之前 kvd 程序已停止并且您已经备份了您的验证节点**。如果升级失败，如果没有备份过验证节点可能会导致无法重新启动节点。

1. Export State (this **MUST** be done using **v0.14.3**, previous v0.14.x versions will not produce the same genesis hash!)
   导出状态（这个**必须**使用**v0.14.3**完成，以前的 v0.14.x 版本不会产生相同的创世哈希！）

```sh
kvd export --for-zero-height --height PLACEHOLDER > export-genesis.json
```

**Note:** This can take a while!
**注：** 这可能需要一段时间！

2. Update to kava-8 更新到 kava-8

```sh
  # in the `kava` folder 在kava文件夹中
    git pull
    git checkout [TODO: version]
    make install

  # verify versions 确保版本无错
  kvd version --long
  # name: kava
  # server_name: kvd
  # client_name: kvcli
  # version: [TODO: version]
  # commit: PLACEHOLDER
  # build_tags: netgo,ledger
  # go: go version go1.15.8 linux/amd64


  # Migrate genesis state 迁移创世状态
  kvd migrate export-genesis.json > genesis.json

  # Verify output of genesis migration  验证创世迁移的输出
  kvd validate-genesis genesis.json # should say it's valid  应该说它是合法的
  jq -S -c -M '' genesis.json | shasum -a 256
  # PLACEHOLDER

  # Restart node with migrated genesis state 重新启动具有迁移的创世状态的节点
  cp genesis.json ~/.kvd/config/genesis.json
  kvd unsafe-reset-all

  # Restart node - 重启节点
  # ! Be sure to remove --halt-time flag if it is set in systemd/docker 如果在 systemd/docker 中设置了 --halt-time 标志，请删除它
  kvd start
```

### Coordination 协调

If the kava-8 chain does not launch by August 30, 2021 at 19:00 UTC, the launch should be considered a failure and validators should refer to the [rollback](./rollback.md) instructions to restart the previous kava-7 chain. In the event of launch failure, coordination will occur in the [Kava discord](https://discord.com/invite/kQzh3Uv).

如果 kava-8 链在 2021 年 8 月 30 日 19:00 UTC 之前没有启动，则升级应被视为失败，验证者应参考 [rollback](./rollback.md) 指令重新启动之前的 kava- 7 链。如果升级失败，将在[Kava discord](https://discord.com/invite/kQzh3Uv)中进行协调。
