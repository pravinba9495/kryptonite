<p align="center">
  <img width="100" src="https://www.maxpixel.net/static/photo/1x/Growth-Green-Cartoon-Crystal-Green-Cartoon-307264.png" alt="Kryptonite Logo">
  <br/>
  <h1>kryptonite</h1>
</p>

[![Go Reference](https://pkg.go.dev/badge/github.com/pravinba9495/kryptonite.svg)](https://pkg.go.dev/github.com/pravinba9495/kryptonite) ![Go Report Card](https://goreportcard.com/badge/github.com/pravinba9495/kryptonite) ![Issues](https://img.shields.io/github/issues-raw/pravinba9495/kryptonite) ![License](https://img.shields.io/github/license/pravinba9495/kryptonite) ![Release](https://img.shields.io/github/v/release/pravinba9495/kryptonite?include_prereleases)

Automated crypto swapping bot, written in Go. Supports swapping on Ethereum, BSC, Polygon, Optimisim and Arbitrum networks using 1inch AggregatorV4 router. Under active development.

## Table of Contents
- [Introduction](#introduction)
- [Setup](#setup)
  - [Requirements](#requirements)
  - [Parameters](#parameters)
  - [Usage](#usage)
- [Instructions](#instructions)
- [Documentation](#documentation)
- [Development](#development)
- [Maintainers](#maintainers)
- [License](#license)

## Introduction
This project started as a hobby to figure out a way to keep an eye on the crypto market while juggling my day job. The Crypto market is highly volatile. Cryptocurrencies can fluctuate in price drastically within seconds you have your eyes off the screen. Unless you are a trader by profession, you cannot actively manage your portfolio, make any meaningful and profitable moves or prevent a loss. You can swap your tokens into stable coins to prevent losses. However, this requires you to pay constant attention to the market, which is not an easy task for everyone. What if there is a way to protect your crypto investment from major pullbacks like the one everyone witnessed in November 2021?

With Kryptonite, you can now automatically set limit buy/sell and stop-loss orders, like a watchdog protecting your crypto assets from losses, even while you are sleeping. Kryptonite does technical analysis for you on the fly, every day, every minute. It uses historical as well as real-time data to calculate reasonable support and resistance levels and places its trades accordingly. It can react to a market crash more swiftly than any human could. Importantly, Kryptonite strives to reduce your anxiety levels in an uncertain and rigged market.

## Setup

### Requirements

### Parameters

The following command line parameters are supported.

<div align="center">

<table>
<thead>
<tr>
<th>Parameter</th>
<th>Description</th>
<th>Type</th>
<th>Default</th>
</tr>
</thead>
<tbody>

<tr>
<td>--privateKey</td>
<td>Your wallet private key</td>
<td>string</td>
<td></td>
</tr>

<tr>
<td>--publicKey</td>
<td>Your wallet public address</td>
<td>string</td>
<td></td>
</tr>

<tr>
<td>--chainId</td>
<td>Chain to use. Allowed options: 1 (Ethereum), 10 (Optimism), 56 (Binance Smart Chain), 137 (Polygon/Matic), 42161 (Arbitrum)</td>
<td>integer</td>
<td>1</td>
</tr>

<tr>
<td>--stableToken</td>
<td>Stable token (ERC20) to use. Example: USDC, USDT, DAI</td>
<td>string</td>
<td>USDC</td>
</tr>

<tr>
<td>--targetToken</td>
<td>Target ERC20 token to hold. Example: WETH, WMATIC, LINK.</td>
<td>string</td>
<td>WETH</td>
</tr>

<tr>
<td>--redisAddress</td>
<td>Redis server host. Example: 192.168.1.100:6379</td>
<td>string</td>
<td></td>
</tr>

<tr>
<td>--botToken</td>
<td>Telegram bot token used to send and receive transaction confirmations</td>
<td>string</td>
<td></td>
</tr>

<tr>
<td>--chatId</td>
<td>Your telegram chat id. You will receive this when you authorize yourself with the bot for the first time.</td>
<td>string</td>
<td></td>
</tr>

<tr>
<td>--password</td>
<td>Password to share with the bot to authorize yourself as the admin</td>
<td>string</td>
<td>kryptonite</td>
</tr>

<tr>
<td>--days</td>
<td>No. of days to use to calculate moving average</td>
<td>integer</td>
<td>30</td>
</tr>

<tr>
<td>--profitPercent</td>
<td>Profit percent at which the bot will execute a sell order</td>
<td>integer</td>
<td>50</td>
</tr>

<tr>
<td>--stopLossPercent</td>
<td>Loss percent at which the bot will execute a stop loss order</td>
<td>integer</td>
<td>25</td>
</tr>

</tbody>
</table>

</div>

### Usage
```shell
docker run pravinba9495/kryptonite:latest kryptonite \
                --privateKey=<PRIVATE_KEY> \
                --publicKey=<PUBLIC_ADDRESS> \
                --chainId=<CHAIN_ID> \
                --stableToken=<STABLE_TOKEN> \
                --targetToken=<TARGET_TOKEN> \
                --redisAddress=<REDIS_ADDRESS> \
                --botToken=<BOT_TOKEN> \
                --chatId=<CHAT_ID> \
                --password=<PASSWORD> \
                --days=<DAYS> \
                --profitPercent=<PROFIT_PERCENT> \
                --stopLossPercent=<STOP_LOSS_PERCENT>
```
## Instructions


## Documentation
Kryptonite documentation is hosted at [Read the docs](https://pkg.go.dev/github.com/pravinba9495/kryptonite).

## Development
Kryptonite is still under development. Contributions are always welcome!

## Maintainers
* [@pravinba9495](https://github.com/pravinba9495)
## License
MIT
