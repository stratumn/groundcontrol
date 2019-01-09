import fs from "fs";
import { buildSchema } from "graphql";
import { join } from "path";

export default buildSchema(fs.readFileSync(join(__dirname, "../schema.graphql"), "utf8"));
