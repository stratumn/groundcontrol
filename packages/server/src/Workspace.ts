import { plainToClass, Type } from "class-transformer";
import fs from "fs";
import yaml from "js-yaml";
import Project from "./Project";

class Workspace {
  public static load(file: string): Workspace[] {
    return plainToClass(Workspace, yaml.safeLoad(fs.readFileSync(file, "utf8")).workspaces);
  }

  @Type(() => Project)
  public projects: Project[];

  constructor(
    public name: string,
    public slug: string,
    projects: Project[],
  ) {
    this.projects = projects;
  }

  public id(): string {
    return this.name;
  }
}

export default Workspace;
