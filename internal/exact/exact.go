package exact

// From https://github.com/coinbase/x402/blob/094dcd2b95b5e13e8673264cc026d080417ee142/python/x402/src/x402/chains.py#L4
// KNOWN_TOKENS = {
//     "84532": [
//         {
//             "human_name": "usdc",
//             "address": "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
//             "name": "USDC",
//             "decimals": 6,
//             "version": "2",
//         }
//     ],
//     "8453": [
//         {
//             "human_name": "usdc",
//             "address": "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
//             "name": "USD Coin",  # needs to be exactly what is returned by name() on contract
//             "decimals": 6,
//             "version": "2",
//         }
//     ],
//     "43113": [
//         {
//             "human_name": "usdc",
//             "address": "0x5425890298aed601595a70AB815c96711a31Bc65",
//             "name": "USD Coin",
//             "decimals": 6,
//             "version": "2",
//         }
//     ],
//     "43114": [
//         {
//             "human_name": "usdc",
//             "address": "0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E",
//             "name": "USDC",
//             "decimals": 6,
//             "version": "2",
//         }
//     ],
// }

// From https://github.com/coinbase/x402/blob/094dcd2b95b5e13e8673264cc026d080417ee142/go/pkg/gin/middleware.go#L120
// if options.Testnet {
// 	network = "base-sepolia"
// 	usdcAddress = "0x036CbD53842c5426634e7929541eC2318f3dCF7e"
// }
