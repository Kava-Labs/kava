# Kava-5-1 Rollback Instructions

In the event that the kava-5.1 relaunch is unsuccessful, we will restart the kava-6 chain using the last known state.

In order to restore the previous chain, the following data must be recovered by validators:

* The database that contains the state of the previous chain (in ~/.kvd/data by default)
* The priv_validator_state.json file of the validator (in ~/.kvd/data by default)

If you don't have the database data, the Kava developer team or another validator will share a copy of the database via Amazon s3 or a similar service. You will be able to download a copy of the data and verify it before starting your node.
If you don't have the backup priv_validator_state.json file, you will not have double sign protection on the first block. If this is the case, it's best to consult in the validator discord before starting your node.

## Restoring state procedure

1. Copy the contents of your backup data directory back to the $KVD_HOME/data directory. By default this is ~/.kvd/data.

```
# Assumes backup is stored in "backup" directory
rm -rf ~/.kvd/data
mv backup/.kvd/data ~/.kvd/data
```

2. Install the previous version of kava

```
# from kava directory
git checkout v0.12.4
make install

## verify version
kvd version --long
```

3. Restart kvd process

```
### be sure to remove --halt-time flag if it is set
sudo systemctl daemon-reload
sudo systemctl restart kvd
```
