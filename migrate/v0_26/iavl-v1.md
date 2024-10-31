# IAVL V1

IAVL V1 is an updated data format for the low-level storage of application data in the Kava blockchain.
The change in format results in much more performant syncing of the chain, and greatly reduces the
storage footprint required for Kava nodes.

The first version of Kava using IAVL V1 is `v0.26.2-iavl-v1`. Future major versions will use IAVL V1
unless otherwise specified.

Steps for using IAVL V1:

**Note:** Nodes running IAVL V1 should set `iavl-disable-fastnode = true` in their `app.toml` file.

**For `goleveldb` nodes:**
1. Replace or recreate your node data with IAVL V1: see [Data](#data).
2. Install an IAVL V1 binary:
```sh
git checkout v0.26.2-iavl-v1
make install
```
1. Start kava as usual: `kava start`

**For `rocksdb` nodes:**
1. Replace or recreate your node data with IAVL V1: see [Data](#data).
2. Update your default memory allocator: see [default memory allocator](#default-memory-allocator)
3. Install an IAVL V1 binary:
```sh
git checkout v0.26.2-iavl-v1
make install COSMOS_BUILD_OPTIONS=rocksdb
```
4. Start kava with the desired memory allocator: `LD_PRELOAD=/path/to/tcmalloc kava start`

## Data

The IAVL V1 binaries are compatible with IAVL V0 data (all chain data snapshots available up until this time).
However, for maximum performance & storage benefit, node operators should use data that is completely
built with IAVL V1.

### Validators & Pruning nodes

For nodes with minimal historical state, we recommend bootstrapping your node with statesync.
For an example of how to setup a node with statesync, see [here](https://www.polkachu.com/state_sync/kava).

A public RPC server is available at `https://rpc.kava.io:443`.

### Full-archive node

For nodes with the complete historical data, we are in the process of creating and sharing a snapshot
of full-archive node data. An update will be made when they are publicly available.

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
