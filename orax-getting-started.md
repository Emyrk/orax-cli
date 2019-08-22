# Orax: first PegNet mining pool

Orax is the first PegNet mining pool that requires only your mining machines. No costly data subscription, no EC management. Start mining in less than a minute.

## Preambule

Joining Orax is only possible on invitation. You will need your email address to be whitelisted to sign-up.

## Download orax-cli

Get the binary for your platform:

- Linux: https://oraxpool.s3.ca-central-1.amazonaws.com/orax-cli/orax-cli-v0.2.5
- MacOs: https://oraxpool.s3.ca-central-1.amazonaws.com/orax-cli/orax-cli-v0.2.5.app
- Windows: https://oraxpool.s3.ca-central-1.amazonaws.com/orax-cli/orax-cli-v0.2.5.exe

## Sign up

Start by creating a new account. You only need an email address and a Factoid address to receive payouts.

```bash
./orax-cli register
```

## Start mining

```bash
./orax-cli mine
```

```bash
./orax-cli mine
# You probably want to start the mining process in the background
# and save logs into a file
nohup ./orax-cli mine >> orax-cli.log &
# or you may use your favorite process manager or supervisor
```

That's it, your machine is now connected to Orax and will start mining as soon as work is available. Please note that at the first launch the LXR hash needs to initialized which can take 10 minutes or more depending on your machine. This is a one time operation.

By default the miner uses all the cores of the machine, if you wish to limit that you can use the flag `-n` and set the number of concurrent subminers.

## Adding more miners

You can link as many miners (machines) to your Orax account as you wish. Just run `./orax-cli register` on the said machine and authenticate using your existing account.

## Get info and stats about your account

Get info about your rewards and mining performance:

```bash
./orax-cli info
```

## Benchmarking a machine

You can use the following command to test the mining performance of a machine:

```bash
./orax-cli bench
```
