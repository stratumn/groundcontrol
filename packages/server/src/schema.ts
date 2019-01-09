import fs from 'fs';
import { join } from 'path';
import { buildSchema } from 'graphql';

export default buildSchema(fs.readFileSync(join(__dirname, '../schema.graphql'), 'utf8'));