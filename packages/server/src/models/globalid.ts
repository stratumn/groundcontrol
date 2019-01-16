import Type from "./type";

export function toGlobalId(type: Type, ...parts: string[]): string {
  return Buffer.from([type, ...parts].join(":"), "utf8").toString("base64");
}

export function fromGlobalId(gid: string): [Type, string[]] {
  const parts = Buffer.from(gid, "base64").toString("utf8").split(":");

  return [parts[0] as Type, parts.slice(1)];
}