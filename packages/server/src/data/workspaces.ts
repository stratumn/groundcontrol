import { readFile } from "fs";
import yaml from "js-yaml";
import { join } from "path";
import { promisify } from "util";

import { Workspace } from "../__generated__/groundcontrol";

const filename = join(__dirname, "../../../../groundcontrol.yml");

export async function all(): Promise<Workspace[]> {
    const data = await promisify(readFile)(filename, { encoding: "utf8" });

    return yaml.safeLoad(data).workspaces;
}

export async function get(slug: string) {
    const items = await all();

    return items.find((item) => item.slug === slug) || null;
}
