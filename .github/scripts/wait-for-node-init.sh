get_block_number() {
  local BLOCK_NUMBER=$(curl -sS -X POST \
    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
    -H "Content-Type: application/json" \
    http://localhost:8545 | jq .result)
  echo $BLOCK_NUMBER
}

BLOCK_NUMBER=$(get_block_number)

while [ "$BLOCK_NUMBER" == "" ]
do
  BLOCK_NUMBER=$(get_block_number)
  sleep 0.5
done
