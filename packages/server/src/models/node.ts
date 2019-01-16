import { Node } from "../__generated__/groundcontrol";

const allNodes = new Map<string, Node>();

export function get(gid: string) {
  return allNodes.get(gid) || null;
}

export function set(gid: string, node: Node) {
  allNodes.set(gid, node);
}

export default {
  get,
  set,
};
