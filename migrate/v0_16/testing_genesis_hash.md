### Testing kava-9 migration
To verify that the migration procedure for kava-9 is deterministic, run the following commands to compute the genesis hash of a known block.

```sh
# install latest v0.16 release candidate
cd $HOME/kava  # replace if location of kava directory is different
git fetch
git checkout v0.16.0-rc3
make install
cd $HOME

# download block 1627000 from kava-8
wget https://kava-genesis-files.s3.amazonaws.com/kava-8/block-export-1627000.json

# run the migration (make sure there are no other kava processes running)
kava migrate block-export-1627000.json > block-export-1627000-migrated.json

# calculate hash of migrated file
jq -S -c -M '' block-export-1627000-migrated.json | shasum -a 256
```

