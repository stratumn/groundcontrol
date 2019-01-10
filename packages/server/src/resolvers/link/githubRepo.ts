import { IResolverOptions } from "graphql-tools";

import { Project } from "../../__generated__/groundcontrol";

import github from "../../schemas/github";

const resolver: IResolverOptions = {
  fragment: `... on Project { repo }`,

  async resolve(project: Project, args, context, info) {
    const githubSchema = await github();
    const segments = project.repo.split("/");

    return info.mergeInfo.delegateToSchema({
      args: {
        name: segments[2],
        owner: segments[1],
      },
      context,
      fieldName: "repository",
      info,
      operation: "query",
      schema: githubSchema,
    });
  },
};

export default resolver;
