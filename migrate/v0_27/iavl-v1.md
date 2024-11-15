# IAVL V1

IAVL V1 is an updated data format for the low-level storage of application data in the Kava blockchain.
The change in format results in much more performant syncing of the chain, and greatly reduces the
storage footprint required for Kava nodes.

As of `v0.27.0`, IAVL V1 is the data storage format for the Kava blockchain.

# Configuration

> [!IMPORTANT]
> Nodes running IAVL V1 should set `iavl-disable-fastnode = true` in their `app.toml` file.

> [!IMPORTANT]
> Before starting kava on existing IAVL V0 data, please read the note about pruning changes.

### `goleveldb` nodes
1. Replace or recreate your node data with IAVL V1: see [Data](#data).
2. Install an IAVL V1 binary:
```sh
git checkout v0.27.0
make install
```
3. Start kava as usual: `kava start`

### `rocksdb` nodes

> [!TIP]
> Since v0.26.2, databases are opened with [opendb](https://github.com/Kava-Labs/opendb/) which
> makes rocksdb more configurable. For best memory performance, node operators are encouraged to add
> these settings to their `app.toml` when running with `db_backend = rocksdb`:
> ```toml
> [rocksdb]
> max_open_files = 16384 # increase default max # of open files from 4096
> block_size = 16384 # decreases block index memory by 4x!
> ```

1. Replace or recreate your node data with IAVL V1: see [Data](#data).
2. Update your default memory allocator: see [default memory allocator](#default-memory-allocator)
3. Install an IAVL V1 binary:
```sh
git checkout v0.26.2-iavl-v1
make install COSMOS_BUILD_OPTIONS=rocksdb
```
4. Start kava with the desired memory allocator: `LD_PRELOAD=/path/to/tcmalloc kava start`

## Data

The IAVL V1 binaries are compatible with IAVL V0 data, but it is recommended to start from state sync or a IAVL v1 snapshot.
However, for maximum performance & storage benefit, node operators should use data that is completely
built with IAVL V1.

### Validators & Pruning nodes

For nodes with minimal historical state, we recommend bootstrapping your node with statesync.
For an example of how to setup a node with statesync, see [here](https://www.polkachu.com/state_sync/kava).

A public RPC server is available at `https://rpc.kava.io:443`.

### Full-archive node

For nodes that need complete historical data, a full archive IAVL v1 snapshot is avaiable at https://quicksync.io/kava.

In addition to the most recent version on this file, the Release Notes for
[`v0.26.2-iavl-v1`](https://github.com/Kava-Labs/kava/releases/tag/v0.26.2-iavl-v1) will contain
links to data when they are available.

Node operators can expect a more performant kava process that uses less than half the data storage
required of IAVL V0.

## Default memory allocator

For nodes using rocksdb, greatly improved memory performance was seen when using [`tcmalloc`](https://github.com/google/tcmalloc)
as the memory allocator.

To update your node to use `tcmalloc`:
1. install `tcmalloc`
    * via gperftools collection package (recommended)
      * package manager: `apt-get install google-perftools`
      * [gperftools Releases](https://github.com/gperftools/gperftools/releases)
    * from source: see [tcmalloc Quickstart](https://google.github.io/tcmalloc/quickstart.html)
2. whenever running kava, do so with the new memory allocator: `LD_PRELOAD=/path/to/tcmalloc kava start`
    * replace `/path/to/tcmalloc` above with correct path

## IAVL V1 Pruning Changes

There is an important difference in how changes to pruning settings are handled in IAVL V1.

In IAVL V0, a change to pruning settings takes effect only for new blocks. In IAVL V1, all blocks are
pruned to the new settings the moment the change is made.

When switching to IAVL V1 in this upgrade, you will experience a long startup time if your node has
had a change to pruning settings. If your node has changed its pruning settings during its runtime,
you are encouraged to upgrade with `pruning = "nothing"`.

If you experience a long start time or an out of memory error on startup, try again with `pruning = "nothing"`.
For optimal performance, node operators are encouraged to reinitialize their node with fully IAVL V1
data.

**Example:**
A node is started with `pruning = "nothing"` on IAVL V0 (any release prior to `kava@v0.26.2-iavl-v1`).

The node runs for awhile, is stopped, and then switched to `pruning = "everything"`.

On startup, the node does not prune the existing blocks. Only new blocks are processed with the current
pruning setting.

In IAVL V1 (when you start v0.27 on top of the above data), the `pruning = "everything"` setting is
applied to the entire state. Depending on how many blocks there are that should be pruned, the intial
start up can take a long time and may fail due to a lack of memory.
