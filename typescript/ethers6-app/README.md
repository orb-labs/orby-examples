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
   cd typescript/ethers6-app
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
   
   Note: [contact us](https://x.com/0xOrbLabs) for your orby engine admin url

   ```
   # Orby Engine Admin url 
   ORBY_ENGINE_ADMIN_URL=orby_engine_admin_url

   # Instance name
   ORBY_INSTANCE_NAME=some_name

   # Input token address (with 0x prefix)
   INPUT_TOKEN_ADDRESS=your_input_token_address (with 0x prefix)

   # Output token address (with 0x prefix)
   OUTPUT_TOKEN_ADDRESS=your_output_token_address

   # Chain ID your input tokens are on
   INPUT_TOKEN_CHAIN_ID=1000000000001

   # Chain ID your output tokens are on
   OUTPUT_TOKEN_CHAIN_ID=1000000000002

   # Your private key (without 0x prefix)
   PRIVATE_KEY=your_private_key_here

   # Amount of input token to use (e.g. 1)
   AMOUNT=input_token_amount

   # Example type (one of: getOperationsToSwap, getOperationsToExecuteTransaction getOperationsToSignTypedData, getFungibleTokenPortfolio)
   EXAMPLE=example_type
   ```

## Usage

Run the application:

```
yarn start
```

The application will:

1. Create an account cluster based on your private key
2. Create a virtual node based on the account cluster
3. Formulate the correct input params for the desired example
4. Call corresponding example_type function
5. (For those with operations) Call sendOperationSet to sign and send the operations

## Security Considerations

- **Never share your private key**: Keep your private key secure at all times.
- **Test on testnets first**: Always test on Ethereum testnets before using on mainnet.
- **Use environment variables**: Avoid hardcoding private keys in your code.

## License

MIT
