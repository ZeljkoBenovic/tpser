# TPSER (tp-es-er)
Is a small program to get the block information from any `Ethereum` compatible blockchain network.

The output of this tool, will show some basic information about the blocks, like:   
* *timestamp*
* *block number*
* *the number of transactions in a block*
* *block gas size*
* *block gas utilization*

Finally, the program will show the average **TPS** (transactions per second), **TOTAL TRANSACTIONS** and **TOTAL DURATION**

**TPS** is calculated in a very rudimentary fashion, by *dividing* the `total duration` in seconds by the `total number of transactions`


## Usage
`tpser -json-rpc <JSON-RPC URL> -block-start <START_BLOCK_NUMBER> -block-end <END_BLOCK_NUMBER>`     
*if `block-start` is omitted, it will be set to block `1`*
