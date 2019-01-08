import express from 'express';
import cors from 'cors';
import graphqlHTTP from 'express-graphql';
import { buildSchema } from 'graphql';
import fs from 'fs';
import yaml from 'js-yaml';
import {plainToClass} from "class-transformer";

const schema = buildSchema(fs.readFileSync('server/schema.graphql', 'utf8'));

class Workspace {
  name: string;
  slug: string;
  projects: Project[];

  constructor(name: string, slug: string, projects: Project[]) {
    this.name = name;
    this.slug = slug;
    this.projects = projects;
  }

  id(): string {
    return this.name
  }
}

class Project {
  name: string;
  repo: string;
  branch: string;

  constructor(name: string, repo: string, branch: string) {
    this.name = name;
    this.repo = repo;
    this.branch = branch;
  }
  
  id(): string {
    return this.repo
  }
}

function LoadWorkspaces(file: string): Workspace[] {
  return plainToClass(Workspace, yaml.safeLoad(fs.readFileSync(file, 'utf8')).workspaces);
}

const workspaces = LoadWorkspaces('groundcontrol.yml');

class Root {
  workspaces():any[] {
    return workspaces;
  }
};

const app = express();
app.use(cors());
app.use('/graphql', graphqlHTTP({
  schema: schema,
  rootValue: new Root(),
  graphiql: true,
}));
app.listen(4000);
console.log('Running a GraphQL API server at localhost:4000/graphql');
