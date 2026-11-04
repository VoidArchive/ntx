import { defineRailway, project, service, volume } from "railway/iac";
// This repository manages only its own resources in the environment. Other
// repositories export their own partial name.
// See https://docs.railway.com/infrastructure-as-code#multi-repo-projects
export const partial = "fearless-adaptation";
export default defineRailway(() => {
  const fearlessAdaptation = service("fearless-adaptation", {
    build: { buildEnvironment: "V3", builder: "RAILPACK", dockerfilePath: "apps/api/Dockerfile" },
    start: "/bin/ntx serve",
    healthcheck: "/health",
    healthcheckTimeout: 100,
    replicas: { "asia-southeast1-eqsg3a": 1 },
    deploy: { restartPolicyType: "ON_FAILURE" },
    // SQLite lives on this volume (NTX_DB_PATH=/data/market.db in the Dockerfile)
    volumeMounts: { "/data": volume("fearless-adaptation-volume", { sizeMB: 5000, region: "asia-southeast1-eqsg3a" }) },
  });

  return project("ntx-production", {
    variables: { managed: false },
    resources: [fearlessAdaptation],
  });
});
