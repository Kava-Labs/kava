# Devnet

Devnet contains the configuration files to launch an example kava chain.

It is a single validator chain that runs locally.

Add the home flag `--home <path to kava dir>/contrib/devnet/home` to any kava command to use this configuration.

## Example usage

1. set bash alias for convenience `alias dkava='kava --home <path to kava dir>/contrib/devnet/home'`
2. launch chain `dkava start`
3. interact with the chain in another terminal, e.g.
   - query some data `dkava query staking validators`
   - list available private keys `dkava keys list`
   - send some coins `dkava tx bank send whale kava1wuzhkn2f8nqe2aprnwt3jkjvvr9m7dlkpumtz2 1000000ukava`
   - show available commands `dkava --help`

Note, any change to the genesis file requires the data dir to be reset before the chain will start again: `kava --home <path to kava dir>/contrib/devnet/home unsafe-reset-all`

## Developer Usage

As new features are developed, the genesis file can be updated on the same branch.
Then reviewers can easily manually test out new features.
