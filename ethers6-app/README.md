# Ethereum Transaction Signer

An ethers6 application for creating, signing, and sending ethereum transactions using Orby.

## Prerequisites

- ethers6
- An Ethereum private key
- Internet connection (for fetching network data and optional broadcasting)

## Installation

1. Clone this repository:

   ```
   git clone https://github.com/orb-labs/orby-examples
   cd ethers6-app
   ```

2. Install dependencies:

   ```
   yarn install
   ```

3. Create a `.env` file from the example:

   ```
   cp .env.example .env
   ```

4. Edit the `.env` file with your information:

   ```
   # Orby Engine Admin url ([contact us](https://x.com/0xOrbLabs) for your url)
   ORBY_ENGINE_ADMIN_URL=orby_engine_admin_url

   # Orby url ([contact us](https://x.com/0xOrbLabs) for your url)
   ORBY_URL=your_orby_url

   # Instance name
   ORBY_INSTANCE_NAME=some_name

   # Source chain ID
   SOURCE_CHAIN_ID=1000000000001

   # Input token address (with 0x prefix)
   INPUT_TOKEN_ADDRESS=your_input_token_address (with 0x prefix)

   # Output token address (with 0x prefix)
   OUTPUT_TOKEN_ADDRESS=your_output_token_address

   # Chain ID your tokens are on
   TOKEN_CHAIN_ID=1000000000001

   # Your private key (without 0x prefix)
   PRIVATE_KEY=your_private_key_here

   # Amount of input token to use (e.g. 1)
   AMOUNT=input_token_amount

   # Verbose printing (true/false)
   VERBOSE=true

   ```

## Usage

Run the application:

```
yarn start
```

The application will:

1. Create an account cluster based on your private key
2. Create a virtual node based on the account cluster
3. Fetch standardized token information based on input/output token addresses
4. Call getOperationsToSwap to figure out what transactions to make
5. Sign any transactions that have been returned with your private key
6. Call sendSignedOperations on those signed transactions

## Security Considerations

- **Never share your private key**: Keep your private key secure at all times.
- **Test on testnets first**: Always test on Ethereum testnets (Sepolia) before using on mainnet.
- **Use environment variables**: Avoid hardcoding private keys in your code.

## License

MIT
