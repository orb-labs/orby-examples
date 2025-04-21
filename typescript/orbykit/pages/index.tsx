import type { NextPage } from "next";
import Head from "next/head";
import styles from "../styles/Home.module.css";
import Link from "next/link";

const linkClassName = "hover:underline hover:text-blue-500";

const Home: NextPage = () => {
  return (
    <div className={styles.container}>
      <Head>
        <title>OrbyKit Examples</title>
        <meta content="OrbyKit examples" name="description" />
        <link href="/favicon.ico" rel="icon" />
      </Head>

      <main className="flex flex-col items-center justify-end pt-8">
        <h1 className="text-2xl font-bold mb-8">OrbyKit Examples</h1>
        <Link href="/vanilla" className={linkClassName}>
          Vanilla OrbyKit
        </Link>
        <Link href="/privy" className={linkClassName}>
          Privy + OrbyKit
        </Link>
        <Link href="/rainbowkit" className={linkClassName}>
          RainbowKit + OrbyKit
        </Link>
        <Link href="/dynamic" className={linkClassName}>
          Dynamic + OrbyKit
        </Link>
        <Link href="/appkit" className={linkClassName}>
          Appkit + OrbyKit (WIP)
        </Link>
        <Link href="/standalone-modals" className={linkClassName}>
          Using OrbyKit Modals Standalone
        </Link>
      </main>
    </div>
  );
};

export default Home;
