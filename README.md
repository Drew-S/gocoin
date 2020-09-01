# Go Coin

A simple little golang program that connects to nomics api to obtain cryptocurrency information from their public API. Usage is for simple terminal interaction, such as within Polybar or Conky.

## Usage

First go to [nomics.com](https://p.nomics.com/cryptocurrency-bitcoin-api) and obtain an API key. Place this key into the key file in `$HOME/.gocoin/key`, if the environment variable `$XDG_CONFIG_HOME` is set, the key should be in `$XDG_CONFIG_HOME/gocoin/key`.

After run the program by calling `gocoin [options]`. If no options are set the defaults of printing out some basic information about Bitcoin is presented in CAD.

Use `-c` to set your currency locale (default: CAD).

Use `-x` to set the target cryptocurrency (default: BTC).

Use `-f "<format>"` to format the output of the call (default: "%C: %P %1D:P %1D:C", which outputs "BTC: 15627.42669435 279.96446090 ")
