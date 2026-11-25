#!/usr/bin/env bash
set -euo pipefail

e2e_dir=$(cd "$(dirname "$0")/.." && pwd)
output="$e2e_dir/../../build/.cache/kvtool"
templates="$output/config/templates"
mkdir -p "$templates"
cp -R "$e2e_dir/kvtool/config/templates/." "$templates/"

# The pinned kvtool has the funded v0.27 fixture. Start its v0.28 counterpart
# with the packet-forward genesis schema and escrow totals from the released upgrade.
template="$templates/kava/v0.28"
mkdir -p "$template"
cp -R "$templates/kava/v0.27/." "$template/"
genesis=initstate/.kava/config/genesis.json
jq --slurpfile amounts "$e2e_dir/fixtures/v0.28-escrow-updates.json" '
  del(.app_state.packetfowardmiddleware.params)
  | .app_state.transfer.total_escrowed |= map(.amount = ($amounts[0][.denom] // .amount))
' "$templates/kava/v0.27/$genesis" > "$template/$genesis"

# Keep the validator database writable by the developer across bootstrap runs.
# kvtool uses /root/.kava for both the node home and upgrade proposals.
docker compose -f "$templates/kava/v0.27/docker-compose.yaml" config \
  --no-interpolate --no-path-resolution --format json |
  jq --arg user "$(id -u):$(id -g)" '{
    services: {kavanode: (.services.kavanode |
      .image = "kava/kava:${KAVA_TAG:-v0.28.2-goleveldb}" |
      .user = $user | .environment.HOME = "/root" | .tmpfs = ["/root:mode=0755"]
    )}
  }' > "$template/docker-compose.yaml"

cd "$e2e_dir/kvtool"
go build -ldflags "-X 'github.com/kava-labs/kvtool/config/generate.ConfigTemplatesDir=$templates'" -o "$output/kvtool" .
