# Orax: first PegNet mining pool

Orax is the first PegNet mining pool that requires only your mining machines. No costly data subscription, no EC management. Start mining in less than a minute.

## Preambule

Joining Orax is only possible on invitation. You will need your email address to be whitelisted to sign-up.

## Download orax-cli

Get the binary for your platform:

* Linux: https://oraxpool.s3.ca-central-1.amazonaws.com/orax-cli-test/orax-cli-v0.2.1-test
* MaxOs: https://oraxpool.s3.ca-central-1.amazonaws.com/orax-cli-test/orax-cli-v0.2.1-test.app
* Windows: https://oraxpool.s3.ca-central-1.amazonaws.com/orax-cli-test/orax-cli-v0.2.1-test.exe

## Sign up

Start by creating a new account. You only need an (whitelisted) email address and a Factoid address to receive payouts.
```bash
./orax-cli register
```

## Start mining

```bash
./orax-cli mine
```

That's it, your machine is now connected to orax and mining OPRs. By default the miner uses all the cores of the machine, if you wish to limit that you can use the flag `-n` and set the number of concurrent subminers. 

## Adding more miners

You link as many miners (machines) to your Orax account as you wish. Just run `./orax-cli register` on the said machine and authenticate using your existing account.

## Benchmarking a machine

You can use the following command to test the performance of a machine:

```bash
./orax-cli bench
```