import { ApolloServer } from "@apollo/server";
import { startStandaloneServer } from "@apollo/server/standalone";
import { typeDefs } from "./schema.js";
import { resolvers } from "./resolvers.js";
import { defaultIndexer } from "./indexer.js";
import { MemoryStore } from "./store/memory.js";
import { RpcIngester } from "./ingest/rpc.js";

export async function startServer(port = 4000) {
  const store = new MemoryStore();
  // The resolvers read from this same singleton (see resolvers.ts) — a
  // fresh PoolIndexer here would make ingestion a no-op for every query.
  const indexer = defaultIndexer;

  const rpcUrl = process.env.SOROBAN_RPC_URL;
  const contractIdsEnv = process.env.CONTRACT_IDS;
  const pollIntervalMs = process.env.POLL_INTERVAL_MS ? parseInt(process.env.POLL_INTERVAL_MS, 10) : undefined;
  const startLedger = process.env.START_LEDGER ? parseInt(process.env.START_LEDGER, 10) : undefined;

  let ingester: RpcIngester | undefined;

  if (rpcUrl && contractIdsEnv) {
    const contractIds = contractIdsEnv.split(",").map((id) => id.trim()).filter(Boolean);
    if (contractIds.length > 0) {
      ingester = new RpcIngester(
        {
          rpcUrl,
          contractIds,
          pollIntervalMs,
          startLedger,
        },
        async (event) => {
          await store.appendEvent(event);
          indexer.indexEvent(event);
        },
        async (err) => {
          console.error("[RpcIngester error]:", err);
        },
        store,
      );
      ingester.start();
      console.log(`[RpcIngester] Started ingesting events for contracts: ${contractIds.join(", ")}`);
    }
  }

  const server = new ApolloServer({ typeDefs, resolvers });
  const { url } = await startStandaloneServer(server, {
    listen: { port },
  });

  return { url, ingester, store, indexer };
}

if (process.env.NODE_ENV !== "test") {
  startServer().then(
    ({ url }) => console.log(`GraphQL API ready at ${url}`),
    (error: unknown) => {
      console.error("GraphQL API failed to start:", error);
      process.exitCode = 1;
    },
  );
}
