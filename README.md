# TPSER (ti-pi-es-er)

Is a set of tools dedicated to test the performance of Ethereum compliant blockchain networks.    
The main goal, is to provide the blockchain operators some insight into its performance.   
This data can than be used to "fine tune" the blockchain client and possibly improve performance.

## MODES

The `tpser` program consists of only two modes for now, more to be added later.    
For now, only the EOA transfers are supported. ERC20 and ERC721 will follow soon.

### BlocksFetcher
The `blocks fetcher` mode fetches the information about the specified range of blocks and 
shows a nicely formatted table result, with the basic information about the blocks, like:   
* *timestamp*
* *block number*
* *the number of transactions in a block*
* *block gas size*
* *block gas utilization*

Finally, the mode will show the average **TPS** (transactions per second), **TOTAL TRANSACTIONS** and **TOTAL DURATION**

**TPS** is calculated in a very rudimentary fashion, by *dividing* the `total duration` in seconds by the `total number of transactions`

### LongSender

The `long-sender` modes is used for sending a specific number of transactions per second for a set period of time.   
This is useful when testing the network for stability after the new version has been released, memory leaks, etc.   
As the transactions can be sent for a long time, operators can get useful data from the blockchain logs and metrics.
The `long-sender` won't wait for receipts, and as such there is no information about the state of the transactions.

Optionally, the TPS report can be generated, but it's not advisable to do so if the `long-sender` 
was running for a long time, as there can be a huge number of blocks with transactions.

## Usage

### Common Flags
* `-json-rpc` - the ethernet `json-rpc` http(s)/ws(s) endpoint
* `-log-level` - the log level output - default: info
  * `info`
  * `debug`
* `-mode` - the mode of operation:
  * `blocks-fetcher` - runs in the BlockFetcher mode
  * `long-sender` - runs in the LongSender mode
* `-duration` - time in minutes of how long the `long-sender` will run
* `-to` - the account to which the funds will be sent
* `-report <bool>` - should the final TPS report be generated
* `-tps` - how much transactions per second will be sent


### BlocksFetcher

#### BlockStart/BlockEnd
* `-block-start` - the starting block
* `-block-end` - the end block

```bash
tpser \
    -mode blocks-fetcher \   
    -json-rpc <JSON-RPC URL> \     
    -block-start <START_BLOCK_NUMBER> \
    -block-end <END_BLOCK_NUMBER>
```   
*if `block-start` is omitted, it will be set to block `1`*

#### BlockRange
* `-block-range` - fetch a defined range of blocks, starting from the latest available

```bash
go run . -json-rpc <JSON_RPC_URL> -block-range <RANGE_OF_BLOCKS>
```
*blocks-fetcher is default mode, so the flag can be omitted*

### LongSender

#### Using private key
* `-pk` - the private key of the account that has funds

```bash
tpser \
    -mode long-sender \   
    -json-rpc <JSON-RPC URL>      
    -pk <PRIVATE_KEY> \
    -to <ADDRESS> \
    -tps <NUMBER_OF_TX_PER_SEC> \
    -duration <DURATION_OF_THE_TEST_IN_MIN>
```

#### Using mnemonic
* `-mnemonic` - the mnemonic string used to derive accounts
* `-mnemonic-addr` - the number of derived accounts to use, starting from 0
```bash
tpser \
    -mode long-sender \   
    -json-rpc <JSON-RPC URL>      
    -mnemonic <MNEMONIC_STRING> \
    -mnemonic-addr <NUMBER_OF_ACCOUNTS> \
    -to <ADDRESS> \
    -tps <NUMBER_OF_TX_PER_SEC> \
    -duration <DURATION_OF_THE_TEST_IN_MIN>
```


LongSender mode can be effectively used to find your blockchain most stable TPS. Its job is to send a defined number
of transactions every second for a specified duration. If your blockchain client can handle this load, without any 
transaction errors, you can feel confident that the specified TPS can be processed in production.     

How would you do this:
* Run `long-sender` with, for example `-tps 300` and `-timeout 1440`.   
  This will send 300 transactions every second non-stop for 24h.
* Periodically check for `long-sender` error output, for any transaction errors
* Periodically check block utilisation and transactions mined, using`blocks-fetcher` module with `-block-range 100` flag
* Presuming that block time is set to `2s`, 100 blocks should be processed in 200s, and they should contain
  `~60 000` transactions (`300tx * 200s (100blocks, each mined in 2s)`) 