import { Abi, AbiFallback, AbiReceive } from "abitype";

export function getAbiFallbackFunction(abi: Abi): AbiFallback | undefined {
  for (const f of abi) {
    if (f.type === "fallback") {
      return f;
    }
  }

  return undefined;
}

export function getAbiReceiveFunction(abi: Abi): AbiReceive | undefined {
  for (const f of abi) {
    if (f.type === "receive") {
      return f;
    }
  }

  return undefined;
}
