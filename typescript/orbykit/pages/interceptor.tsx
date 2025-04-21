import { useEffect, useState } from "react";

const alchemyUrl = "YOUR_ALCHEMY_URL";

export default function Interceptor() {
  const [loading, setLoading] = useState(false);
  const [ethBalance, setEthBalance] = useState<number>();

  useEffect(() => {
    const fetchBalance = async () => {
      try {
        setLoading(true);
        const response = await fetch(alchemyUrl, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            jsonrpc: "2.0",
            method: "eth_getBalance",
            params: ["0x3CA010f1512018236a26A518AE97038A226A98D5", "latest"],
            id: 1,
          }),
        });

        const data = await response.json();
        const balanceInWei = parseInt(data.result, 16);
        const balanceInEth = balanceInWei / 1e18;

        setEthBalance(balanceInEth);
        setLoading(false);
      } catch (error) {
        setLoading(false);
        console.log("error", error);
      }
    };

    fetchBalance();
  }, []);

  return (
    <div>
      <h1>Interceptor</h1>
      <p>
        ETH Balance:{" "}
        {loading ? "loading..." : `${JSON.stringify(ethBalance)} ETH`}
      </p>
    </div>
  );
}
