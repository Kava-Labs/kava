# Kava-11 Rollback Instructions

In the event that the Kava-11 upgrade is unsuccessful, we will restart the Kava-11 chain using the last known state.

In order to restore the previous chain, the following data must be recovered by validators:

- The database that contains the state of the previous chain (in ~/.kava/data by default)
- The priv_validator_state.json file of the validator (in ~/.kava/data by default)

If you don't have the database data, the Kava developer team or another validator will share a copy of the database via Amazon s3 or a similar service. You will be able to download a copy of the data and verify it before starting your node.
If you don't have the backup priv_validator_state.json file, you will not have double sign protection on the first block. If this is the case, it's best to consult in the validator discord before starting your node.

## Restoring state procedure

1. Stop your node

```sh
kava stop
```

2. Copy the contents of your backup data directory back to the $KAVA_HOME/data directory. By default this is ~/.kava/data.

```sh
# Assumes backup is stored in "backup" directory
rm -rf ~/.kava/data
mv backup/.kava/data ~/.kava/data
```

3. Install the previous version of kava

```sh
# from kava directory
git checkout v0.18.0
make install
## verify version
kava version --long
```

4. Start kava process

```sh
### be sure to remove --halt-time flag if it is set
sudo systemctl daemon-reload
sudo systemctl start kava
```