/*
Migrate handles translating the state of the blockchain between software versions.

For example, as modules are changed over time the structure of the data they store changes. The data structure must be
migrated to the new structure of the newer versions.

There are two types of migration:
- **genesis migration** a script converts an exported genesis file from the old software to a new genesis file for the new software
- **live upgrade** a handler in the upgrade module converts data in the database itself from the old version to the new

Genesis migration starts a whole new blockchain (with new chain-id) for the new software version.
Live upgrade keeps the blockchain (and chain-id) the same for the new software version.

We only support migrations between mainnet kava releases.
We only support migrations from the previous mainnet kava version to the current. We don't support migrating between two old versions, use the old software version for this.
We only support migrations from old to new versions, not the other way around.

Genesis Migration
The process is:
- unmarshal the current genesis file into the old `GenesisState` type that has been copied into a `legacy` folder (ideally using the old codec version)
- convert that `GenesisState` to the current `GenesisState` type
- marshal it to json (using current codec)

On each release we can delete the previous releases migration and old GenesisState type.
eg kava-3 migrates `auth.GenesisState` from kava-2 to `auth.GenesisState` from kava-3,
but for kava-4 we don't need to keep around kava-2's `auth.GenesisState` type.

This folder contains old types from several sdk modules because they needed custom migrations for kava-3.
The sdk version for kava-2 was a master commit 18de63.

Live Upgrade
The process is:
- submit upgrade proposal on old chain
- old chain halts
- people download new version and restart their validators
- on start the new upgrade handler runs
- use copypasted old keeper and types to read from db, convert to current types and write with current keeper

*/
package migrate
