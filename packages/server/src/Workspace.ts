import fs from 'fs';
import yaml from 'js-yaml';
import { Type, plainToClass } from "class-transformer";
import Project from './Project';

class Workspace {
  name: string;
  slug: string;

  @Type(() => Project)
  projects: Project[];

  id(): string {
    return this.name
  }

  static load(file: string): Workspace[] {
    return plainToClass(Workspace, yaml.safeLoad(fs.readFileSync(file, 'utf8')).workspaces);
  }
}

export default Workspace;
