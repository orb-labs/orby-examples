// config.js
require('dotenv').config();

module.exports = {
  PRIVATE_KEY: process.env.PRIVATE_KEY ?? "",
  ORBY_ENGINE_ADMIN_URL: process.env.ORBY_ENGINE_ADMIN_URL ?? "",
  ORBY_INSTANCE_NAME: process.env.ORBY_INSTANCE_NAME ?? "",
  INPUT_TOKEN_ADDRESS: process.env.INPUT_TOKEN_ADDRESS ?? "",
  INPUT_TOKEN_CHAIN_ID: process.env.INPUT_TOKEN_CHAIN_ID ?? "",
  OUTPUT_TOKEN_ADDRESS: process.env.OUTPUT_TOKEN_ADDRESS ?? "",
  OUTPUT_TOKEN_CHAIN_ID: process.env.OUTPUT_TOKEN_CHAIN_ID ?? "",
  AMOUNT: process.env.AMOUNT ?? "0",
  EXAMPLE_TYPE: process.env.EXAMPLE_TYPE ?? "",
};
